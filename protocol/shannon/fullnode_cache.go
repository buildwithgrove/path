// Package shannon provides blockchain data fetching and caching for Shannon full nodes.
//
// This package contains:
//   - cachingFullNode: Intelligent caching layer with block-based session refresh
//   - LazyFullNode: Direct connection to Shannon full nodes
//   - Configuration types for flexible client setup
//
// The caching system uses SturdyC to provide:
//   - Block-aware session refresh (triggers at SessionEndBlockHeight+1)
//   - Zero-downtime cache swaps during session transitions
//   - Stampede protection for concurrent requests
//   - Infinite TTL for account public keys (immutable data)
package shannon

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"
	"github.com/viccon/sturdyc"

	shannonmetrics "github.com/buildwithgrove/path/metrics/protocol/shannon"
	"github.com/buildwithgrove/path/protocol"
)

// cachingFullNode implements FullNode interface.
var _ FullNode = &cachingFullNode{}

// Cache configuration constants
const (
	// SturdyC cache configuration
	// Docs: https://github.com/viccon/sturdyc
	cacheCapacity      = 100_000 // Max entries across all shards
	numShards          = 10      // Number of cache shards for concurrency
	evictionPercentage = 10      // Percentage of LRU entries evicted when full

	// Cache key prefixes to avoid collisions
	sessionCacheKeyPrefix       = "session"
	accountPubKeyCacheKeyPrefix = "pubkey"
)

// noTTL represents infinite cache duration (~292 years)
// Used for immutable data like account public keys
const noTTL = time.Duration(math.MaxInt64)

// cachingFullNode provides intelligent caching for Shannon blockchain data.
//
// Key features:
//   - Block-based session refresh: Monitors SessionEndBlockHeight instead of time-based TTL
//   - Zero-downtime transitions: Creates new cache instances and atomically swaps them
//   - Intelligent polling: Switches to 1-second polling when approaching session end
//   - Stampede protection: SturdyC prevents duplicate requests for the same data
//   - Infinite caching: Account public keys cached forever (immutable data)
//
// Documentation: https://github.com/viccon/sturdyc
type cachingFullNode struct {
	logger             polylog.Logger
	onchainDataFetcher FullNode

	// Session cache with block-based refresh monitoring
	sessionCache        *sturdyc.Client[sessiontypes.Session]
	sessionCacheMu      sync.RWMutex
	sessionRefreshState *sessionRefreshState

	// Account public key cache with infinite TTL
	accountPubKeyCache *sturdyc.Client[cryptotypes.PubKey]
}

// NewCachingFullNode creates a new caching layer around a FullNode.
//
// The cache automatically starts background session monitoring and will refresh
// sessions based on blockchain height rather than time-based TTL.
func NewCachingFullNode(
	logger polylog.Logger,
	dataFetcher FullNode,
) (*cachingFullNode, error) {
	logger = logger.With("client", "caching_full_node")

	cfn := &cachingFullNode{
		logger:             logger,
		onchainDataFetcher: dataFetcher,

		sessionCache: getCache[sessiontypes.Session](),
		sessionRefreshState: &sessionRefreshState{
			activeSessionKeys: make(map[string]sessionKeyInfo),
		},

		accountPubKeyCache: getCache[cryptotypes.PubKey](),
	}

	// Start background session monitoring
	cfn.startSessionMonitoring()

	return cfn, nil
}

// getCache creates a SturdyC cache instance with infinite TTL
func getCache[T any]() *sturdyc.Client[T] {
	return sturdyc.New[T](
		cacheCapacity,
		numShards,
		noTTL,
		evictionPercentage,
	)
}

// GetApp fetches application data directly from the full node without caching.
//
// Applications are not cached because:
//   - Only needed during gateway startup for service configuration
//   - Runtime access to applications happens via sessions (which contain the app)
//   - Reduces cache complexity for rarely-accessed data
func (cfn *cachingFullNode) GetApp(ctx context.Context, appAddr string) (*apptypes.Application, error) {
	return cfn.onchainDataFetcher.GetApp(ctx, appAddr)
}

// GetSession returns a session from cache or fetches it from the blockchain.
//
// On cache miss, this method:
//   - Fetches the session from the full node
//   - Updates the global session end height for monitoring
//   - Tracks the session key for background refresh
//   - Caches the session with infinite TTL (refreshed by block monitoring)
//
// SturdyC provides automatic stampede protection for concurrent requests.
func (cfn *cachingFullNode) GetSession(
	ctx context.Context,
	serviceID protocol.ServiceID,
	appAddr string,
) (sessiontypes.Session, error) {
	startTime := time.Now()
	sessionKey := getSessionCacheKey(serviceID, appAddr)

	// Get current cache instance with read lock
	// This is to ensure that the cache is not modified while we are fetching the session
	// ie. when the session cache is reset during a session rollover.
	cfn.sessionCacheMu.RLock()
	sessionCache := cfn.sessionCache
	cfn.sessionCacheMu.RUnlock()

	// Track if this will be a cache hit by checking if key exists
	var cacheHit bool
	var cacheResult string
	if _, exists := sessionCache.Get(sessionKey); exists {
		cacheHit = true
		cacheResult = "hit"
		shannonmetrics.RecordSessionCacheOperation(string(serviceID), "get", "session", "hit")
	} else {
		cacheResult = "miss"
		shannonmetrics.RecordSessionCacheOperation(string(serviceID), "get", "session", "miss")
	}

	// SturdyC GetOrFetch provides stampede protection
	session, err := sessionCache.GetOrFetch(
		ctx,
		sessionKey,
		func(fetchCtx context.Context) (sessiontypes.Session, error) {
			fetchStartTime := time.Now()
			cfn.logger.Debug().
				Str("session_key", sessionKey).
				Msgf("Cache miss - fetching session from full node for service %s", serviceID)

			shannonmetrics.RecordSessionCacheOperation(string(serviceID), "fetch", "session", "attempted")
			session, fetchErr := cfn.onchainDataFetcher.GetSession(fetchCtx, serviceID, appAddr)
			fetchDuration := time.Since(fetchStartTime).Seconds()

			if fetchErr != nil {
				shannonmetrics.RecordSessionCacheOperation(string(serviceID), "fetch", "session", "error")
				shannonmetrics.RecordSessionOperationDuration(string(serviceID), "cache_fetch", "error", false, fetchDuration)
				cacheResult = "fetch_error"
				return session, fetchErr
			}

			shannonmetrics.RecordSessionCacheOperation(string(serviceID), "fetch", "session", "success")
			shannonmetrics.RecordSessionOperationDuration(string(serviceID), "cache_fetch", "success", false, fetchDuration)
			if !cacheHit {
				cacheResult = "fetch_success"
			}

			// Register session for block-based monitoring
			cfn.updateSessionEndHeight(session)

			// Track for background refresh during session transitions
			cfn.trackActiveSession(sessionKey, serviceID, appAddr)

			return session, fetchErr
		},
	)

	duration := time.Since(startTime).Seconds()

	if err == nil {
		// Record session transition metrics
		shannonmetrics.RecordSessionTransition(string(serviceID), appAddr, "session_fetch", cacheHit)
		// Record overall operation duration
		shannonmetrics.RecordSessionOperationDuration(string(serviceID), "get_session", cacheResult, false, duration)
	}

	return session, err
}

// getSessionCacheKey creates a unique cache key: "session:<serviceID>:<appAddr>"
func getSessionCacheKey(serviceID protocol.ServiceID, appAddr string) string {
	return fmt.Sprintf("%s:%s:%s", sessionCacheKeyPrefix, serviceID, appAddr)
}

// GetAccountPubKey returns an account's public key from cache or blockchain.
//
// Account public keys are cached with infinite TTL because they never change.
// The fetchFn is only called once per address during the application lifetime.
func (cfn *cachingFullNode) GetAccountPubKey(
	ctx context.Context,
	address string,
) (pubKey cryptotypes.PubKey, err error) {
	startTime := time.Now()
	accountKey := getAccountPubKeyCacheKey(address)

	// Track if this will be a cache hit by checking if key exists
	var cacheHit bool
	var cacheResult string
	if _, exists := cfn.accountPubKeyCache.Get(accountKey); exists {
		cacheHit = true
		cacheResult = "hit"
		shannonmetrics.RecordSessionCacheOperation("", "get", "account_pubkey", "hit")
	} else {
		cacheResult = "miss"
		shannonmetrics.RecordSessionCacheOperation("", "get", "account_pubkey", "miss")
	}

	pubKey, err = cfn.accountPubKeyCache.GetOrFetch(
		ctx,
		accountKey,
		func(fetchCtx context.Context) (cryptotypes.PubKey, error) {
			fetchStartTime := time.Now()
			cfn.logger.Debug().
				Str("account_key", accountKey).
				Msg("Cache miss - fetching account public key from full node")

			shannonmetrics.RecordSessionCacheOperation("", "fetch", "account_pubkey", "attempted")
			// Use the account client from the underlying full node
			accountClient := cfn.onchainDataFetcher.GetAccountClient()
			pubKey, fetchErr := accountClient.GetPubKeyFromAddress(fetchCtx, address)
			fetchDuration := time.Since(fetchStartTime).Seconds()

			if fetchErr != nil {
				shannonmetrics.RecordSessionCacheOperation("", "fetch", "account_pubkey", "error")
				shannonmetrics.RecordSessionOperationDuration("", "account_pubkey_fetch", "error", false, fetchDuration)
				cacheResult = "fetch_error"
			} else {
				shannonmetrics.RecordSessionCacheOperation("", "fetch", "account_pubkey", "success")
				shannonmetrics.RecordSessionOperationDuration("", "account_pubkey_fetch", "success", false, fetchDuration)
				if !cacheHit {
					cacheResult = "fetch_success"
				}
			}

			return pubKey, fetchErr
		},
	)

	duration := time.Since(startTime).Seconds()

	if err == nil {
		// Record overall operation duration for account public key retrieval
		shannonmetrics.RecordSessionOperationDuration("", "get_account_pubkey", cacheResult, false, duration)
	}

	return pubKey, err
}

// getAccountPubKeyCacheKey creates a unique cache key: "pubkey:<address>"
func getAccountPubKeyCacheKey(address string) string {
	return fmt.Sprintf("%s:%s", accountPubKeyCacheKeyPrefix, address)
}

// LatestBlockHeight returns the current blockchain height from the full node.
// This method is not cached as block height changes frequently.
func (cfn *cachingFullNode) LatestBlockHeight(ctx context.Context) (height int64, err error) {
	return cfn.onchainDataFetcher.LatestBlockHeight(ctx)
}

// ValidateRelayResponse validates the raw response bytes received from an endpoint.
// Uses the SDK and the caching full node's account client for validation.
func (cfn *cachingFullNode) ValidateRelayResponse(
	supplierAddr sdk.SupplierAddress,
	responseBz []byte,
) (*servicetypes.RelayResponse, error) {
	return sdk.ValidateRelayResponse(
		context.Background(),
		supplierAddr,
		responseBz,
		cfn.onchainDataFetcher.GetAccountClient(),
	)
}

// GetAccountClient returns the account client from the underlying full node.
func (cfn *cachingFullNode) GetAccountClient() *sdk.AccountClient {
	return cfn.onchainDataFetcher.GetAccountClient()
}

// IsHealthy reports the health status of the cache.
// Currently always returns true as the cache is populated on-demand.
//
// TODO_IMPROVE: Add meaningful health checks:
//   - Verify cache connectivity
//   - Check session refresh monitoring status
//   - Validate recent successful fetches
func (cfn *cachingFullNode) IsHealthy() bool {
	return cfn.onchainDataFetcher.IsHealthy()
}
