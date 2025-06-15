package shannon

import (
	"context"
	"fmt"
	"time"

	accounttypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"
	"github.com/viccon/sturdyc"

	"github.com/buildwithgrove/path/protocol"
)

// ---------------- Cache Configuration ----------------
const (
	// Retry base delay for exponential backoff on failed refreshes
	retryBaseDelay = 100 * time.Millisecond

	// cacheCapacity:
	//   - Max entries across all shards (not per-shard)
	//   - Exceeding capacity triggers LRU eviction per shard
	//   - 100k supports most large deployments
	//   - TODO_TECHDEBT(@commoddity): Revisit based on real-world usage; consider making configurable
	cacheCapacity = 100_000

	// numShards:
	//   - Number of independent cache shards for concurrency
	//   - Reduces lock contention, improves parallelism
	//   - 10 is a good balance for most workloads
	numShards = 10

	// evictionPercentage:
	//   - % of LRU entries evicted per shard when full
	//   - 10% = incremental cleanup, avoids memory spikes
	//   - SturdyC also evicts expired entries in background
	evictionPercentage = 10

	// TODO_TECHDEBT(@commoddity): See Issue #291 for improvements to refresh logic
	// minEarlyRefreshPercentage:
	//   - Earliest point (as % of TTL) to start background refresh
	//   - 0.75 = 75% of TTL (e.g. 22.5s for 30s TTL)
	minEarlyRefreshPercentage = 0.75

	// maxEarlyRefreshPercentage:
	//   - Latest point (as % of TTL) to start background refresh
	//   - 0.9 = 90% of TTL (e.g. 27s for 30s TTL)
	//   - Ensures refresh always completes before expiry
	maxEarlyRefreshPercentage = 0.9
)

// getCacheDelays returns the min/max delays for SturdyC's Early Refresh strategy.
// - Proactively refreshes cache before expiry (prevents misses/latency spikes)
// - Refresh window: 75-90% of TTL (e.g. 22.5-27s for 30s TTL)
// - Spreads requests to avoid thundering herd
// See: https://github.com/viccon/sturdyc?tab=readme-ov-file#early-refreshes
func getCacheDelays(ttl time.Duration) (min, max time.Duration) {
	minFloat := float64(ttl) * minEarlyRefreshPercentage
	maxFloat := float64(ttl) * maxEarlyRefreshPercentage

	// Round to the nearest second
	min = time.Duration(minFloat/float64(time.Second)+0.5) * time.Second
	max = time.Duration(maxFloat/float64(time.Second)+0.5) * time.Second
	return
}

// Prefix for session cache keys to avoid collisions with other keys.
const sessionCacheKeyPrefix = "session"

var _ FullNode = &cachingFullNode{}

// cachingFullNode wraps a LazyFullNode with SturdyC-based caching.
// - Early refresh: background updates before expiry (prevents thundering herd/latency spikes)
// - Example: 30s TTL, refresh at 22.5–27s (75–90%)
// - Benefits: zero-latency reads, graceful degradation, auto load balancing
// Docs: https://github.com/viccon/sturdyc
type cachingFullNode struct {
	logger polylog.Logger

	// Underlying node for protocol data fetches
	lazyFullNode *LazyFullNode

	// Session cache (5 min on Beta TestNet, see #275)
	// TODO_MAINNET_MIGRATION(@Olshansk): Revisit after mainnet
	sessionCache *sturdyc.Client[sessiontypes.Session]

	// Account client wrapped with SturdyC cache
	cachingAccountClient *sdk.AccountClient
}

// NewCachingFullNode wraps a LazyFullNode with:
//   - Session cache: caches sessions, refreshes early
//   - Account cache: indefinite cache for account data
//
// Both use early refresh to avoid thundering herd/latency spikes.
func NewCachingFullNode(
	logger polylog.Logger,
	lazyFullNode *LazyFullNode,
	cacheConfig CacheConfig,
) (*cachingFullNode, error) {
	// Set default session TTL if not set
	cacheConfig.hydrateDefaults()

	// Log cache configuration
	logger.Debug().
		Str("cache_config_session_ttl", cacheConfig.SessionTTL.String()).
		Msgf("cachingFullNode - Cache Configuration")

	// Configure session cache with early refreshes
	sessionMinRefreshDelay, sessionMaxRefreshDelay := getCacheDelays(cacheConfig.SessionTTL)

	// Create the session cache with early refreshes
	sessionCache := sturdyc.New[sessiontypes.Session](
		cacheCapacity,
		numShards,
		cacheConfig.SessionTTL,
		evictionPercentage,
		// See: https://github.com/viccon/sturdyc?tab=readme-ov-file#early-refreshes
		sturdyc.WithEarlyRefreshes(
			sessionMinRefreshDelay,
			sessionMaxRefreshDelay,
			cacheConfig.SessionTTL,
			retryBaseDelay,
		),
	)

	// Account cache: infinite for app lifetime; no early refresh needed.
	accountCache := sturdyc.New[*accounttypes.QueryAccountResponse](
		accountCacheCapacity,
		numShards,
		accountCacheTTL,
		evictionPercentage,
	)

	// Initialize the caching full node with the modified lazy full node
	return &cachingFullNode{
		logger:       logger,
		lazyFullNode: lazyFullNode,
		sessionCache: sessionCache,
		// Wrap the underlying account fetcher with a SturdyC caching layer.
		cachingAccountClient: getCachingAccountClient(
			logger,
			accountCache,
			lazyFullNode.accountClient,
		),
	}, nil
}

// GetApp is a NoOp (apps fetched only at startup; relaying fetches sessions for app/session sync).
func (cfn *cachingFullNode) GetApp(ctx context.Context, appAddr string) (*apptypes.Application, error) {
	return cfn.lazyFullNode.GetApp(ctx, appAddr)
}

// GetSession returns (and auto-refreshes) the session for a service/app from cache.
func (cfn *cachingFullNode) GetSession(
	ctx context.Context,
	serviceID protocol.ServiceID,
	appAddr string,
) (sessiontypes.Session, error) {
	// See: https://github.com/viccon/sturdyc?tab=readme-ov-file#get-or-fetch
	return cfn.sessionCache.GetOrFetch(
		ctx,
		getSessionCacheKey(serviceID, appAddr),
		func(fetchCtx context.Context) (sessiontypes.Session, error) {
			cfn.logger.Debug().Str("session_key", getSessionCacheKey(serviceID, appAddr)).Msgf(
				"[cachingFullNode.GetSession] Fetching from full node",
			)
			return cfn.lazyFullNode.GetSession(fetchCtx, serviceID, appAddr)
		},
	)
}

// getSessionCacheKey builds a unique cache key for session: <prefix>:<serviceID>:<appAddr>
func getSessionCacheKey(serviceID protocol.ServiceID, appAddr string) string {
	return fmt.Sprintf("%s:%s:%s", sessionCacheKeyPrefix, serviceID, appAddr)
}

// ValidateRelayResponse: passthrough to underlying node.
func (cfn *cachingFullNode) ValidateRelayResponse(
	supplierAddr sdk.SupplierAddress,
	responseBz []byte,
) (*servicetypes.RelayResponse, error) {
	return cfn.lazyFullNode.ValidateRelayResponse(supplierAddr, responseBz)
}

// GetAccountClient: passthrough to underlying node (returns caching client).
func (cfn *cachingFullNode) GetAccountClient() *sdk.AccountClient {
	return cfn.cachingAccountClient
}

// IsHealthy: passthrough to underlying node.
// TODO_IMPROVE(@commoddity):
//   - Add smarter health checks (e.g. verify cached apps/sessions)
//   - Currently always true (cache fills as needed)
func (cfn *cachingFullNode) IsHealthy() bool {
	return cfn.lazyFullNode.IsHealthy()
}
