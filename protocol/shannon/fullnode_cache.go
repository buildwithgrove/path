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
	grpcoptions "google.golang.org/grpc"

	"github.com/buildwithgrove/path/protocol"
)

// TODO_IMPROVE(@commoddity): Implement a FullNode interface that adapts caching strategy based on GatewayMode:
// - 1. Centralized Mode:
//   - List of owned apps is predefined.
//   - Onchain data can be proactively cached before any user requests.
//
// - 2. Delegated Mode:
//   - Apps are specified dynamically by incoming user requests.
//   - Cache must be built incrementally (lazy-loading) as new apps are requested.
//
// - 3. Add more documentation around lazy mode

// ---------------- Cache Configuration ----------------
const (
	// Retry base delay for exponential backoff on failed refreshes
	retryBaseDelay = 100 * time.Millisecond

	// cacheCapacity: Maximum number of entries the cache can hold across all shards.
	// This is the total capacity, not per-shard. When capacity is exceeded, the cache
	// will evict a percentage of the least recently used entries from each shard.
	// 100k entries should handle a large number of apps and sessions for most deployments.
	//
	// TODO_TECHDEBT(@commoddity): Revisit cache capacity based on real-world usage patterns.
	// Consider making this configurable and potentially different for apps vs sessions.
	cacheCapacity = 100_000

	// numShards: Number of independent cache shards for concurrent access.
	// SturdyC divides the cache into multiple shards to reduce lock contention and
	// improve performance under concurrent read/write operations. Each shard operates
	// independently with its own mutex, allowing parallel operations across shards.
	// 10 shards provides good balance between concurrency and memory overhead.
	numShards = 10

	// evictionPercentage: Percentage of entries to evict from each shard when capacity is reached.
	// When a shard reaches its capacity limit, this percentage of the least recently used (LRU)
	// entries will be removed to make space for new entries. 10% provides incremental cleanup
	// without causing large memory spikes during eviction cycles.
	// SturdyC also runs background eviction jobs to remove expired entries automatically.
	evictionPercentage = 10

	// accountCacheTTL: TTL for the account cache.
	// This is a stupidly long TTL because account data never changes.
	// It could theoretically be cached indefinitely but SturdyC requires a TTL.
	// TODO_TECHDEBT: Re-evaluate if/how we cache this for the entirety of PATH's lifecycle.
	accountCacheTTL = 120 * time.Minute
)

// getCacheDelays gets the delays for the SturdyC Early Refresh Strategy.
//
// "Refreshing" in SturdyC means proactively fetching fresh data in the background
// BEFORE the cached entry expires. This prevents cache misses and eliminates latency
// spikes by ensuring hot data is always available immediately.
//
// Cache refresh timing is 30-90% of TTL (e.g. 1.2-3.6 minutes for 4-minute TTL).
// This spread is to avoid overloading the full node with too many simultaneous requests.
//
// Reference: https://github.com/viccon/sturdyc?tab=readme-ov-file#early-refreshes
func getCacheDelays(minRefreshPercentage, maxRefreshPercentage float64, ttl time.Duration) (min, max time.Duration) {
	minFloat := float64(ttl) * minRefreshPercentage
	maxFloat := float64(ttl) * maxRefreshPercentage

	// Round to the nearest second
	min = time.Duration(minFloat/float64(time.Second)+0.5) * time.Second
	max = time.Duration(maxFloat/float64(time.Second)+0.5) * time.Second
	return
}

// Use cache prefixes to avoid collisions with other cache keys.
// This is a simple way to namespace the cache keys.
const (
	appCacheKeyPrefix     = "app"
	sessionCacheKeyPrefix = "session"
	accountCacheKeyPrefix = "account"
)

var _ FullNode = &cachingFullNode{}

// cachingFullNode implements the FullNode interface by wrapping a LazyFullNode
// and caching results to improve performance with automatic refresh-ahead.
//
// Early Refresh Strategy:
// Uses SturdyC's early refresh to prevent thundering herd and eliminate latency spikes.
// Background refreshes happen before entries expire, so GetApp/GetSession never block.
//
// Example times (values may change):
//   - 4min TTL, refresh at 1.2-3.6min (30-90% of TTL)
//
// Benefits: Zero-latency reads for active traffic, thundering herd protection,
// automatic load balancing, and graceful degradation.
//
// Docs reference: https://github.com/viccon/sturdyc
type cachingFullNode struct {
	logger polylog.Logger

	// Use a LazyFullNode as the underlying node
	// for fetching data from the protocol.
	lazyFullNode *lazyFullNode

	// Applications can be cached indefinitely. They're invalidated only when they unstake.
	//
	// TODO_MAINNET_MIGRATION(@Olshansk): Ensure applications are invalidated during unstaking.
	//   Revisit these values after mainnet migration to ensure no race conditions.
	appCache *sturdyc.Client[*apptypes.Application]

	// As of #275, on Beta TestNet, sessions are 5 minutes.
	//
	// TODO_MAINNET_MIGRATION(@Olshansk): Revisit these values after mainnet migration to ensure no race conditions.
	sessionCache *sturdyc.Client[sessiontypes.Session]
}

// NewCachingFullNode creates a new CachingFullNode that creates a LazyFullNode with caching account fetcher.
func NewCachingFullNode(
	logger polylog.Logger,
	lazyFullNode *lazyFullNode,
	cacheConfig CacheConfig,
) (*cachingFullNode, error) {
	// Set default TTLs if not set
	cacheConfig.hydrateDefaults()
	// Configure app cache with early refreshes
	appMinRefreshDelay, appMaxRefreshDelay := getCacheDelays(0.3, 0.9, cacheConfig.AppTTL)

	// Create the app cache with early refreshes
	appCache := sturdyc.New[*apptypes.Application](
		cacheCapacity,
		numShards,
		cacheConfig.AppTTL,
		evictionPercentage,
		// See: https://github.com/viccon/sturdyc?tab=readme-ov-file#early-refreshes
		sturdyc.WithEarlyRefreshes(
			appMinRefreshDelay,
			appMaxRefreshDelay,
			cacheConfig.AppTTL,
			retryBaseDelay,
		),
	)

	// Configure session cache with early refreshes
	sessionMinRefreshDelay, sessionMaxRefreshDelay := getCacheDelays(0.3, 0.9, cacheConfig.SessionTTL)

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

	// Create the account cache, which is used to cache account responses from the full node.
	accountCache := initAccountCache()

	// Wrap the original account fetcher with the caching account fetcher
	// and replace the lazy full node's account fetcher with the caching one.
	replaceLazyFullNodeAccountFetcher(logger, lazyFullNode, accountCache)

	return &cachingFullNode{
		logger:       logger,
		lazyFullNode: lazyFullNode,
		appCache:     appCache,
		sessionCache: sessionCache,
	}, nil
}

// GetApp returns the application with the given address, using a cached version if available.
// The cache will automatically refresh the app in the background before it expires.
func (cfn *cachingFullNode) GetApp(ctx context.Context, appAddr string) (*apptypes.Application, error) {
	// See: https://github.com/viccon/sturdyc?tab=readme-ov-file#get-or-fetch
	return cfn.appCache.GetOrFetch(
		ctx,
		getAppCacheKey(appAddr),
		func(fetchCtx context.Context) (*apptypes.Application, error) {
			cfn.logger.Debug().Str("app_key", getAppCacheKey(appAddr)).Msgf(
				"[cachingFullNode.GetApp] Making request to full node",
			)
			return cfn.lazyFullNode.GetApp(fetchCtx, appAddr)
		},
	)
}

// getAppCacheKey returns the cache key for the given app address.
// It uses the appCacheKeyPrefix and the app address to create a unique key.
//
// eg. "app:pokt1up7zlytnmvlsuxzpzvlrta95347w322adsxslw"
func getAppCacheKey(appAddr string) string {
	return fmt.Sprintf("%s:%s", appCacheKeyPrefix, appAddr)
}

// GetSession returns the session for the given service and app, using a cached version if available.
// The cache will automatically refresh the session in the background before it expires.
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
				"[cachingFullNode.GetSession] Making request to full node",
			)
			return cfn.lazyFullNode.GetSession(fetchCtx, serviceID, appAddr)
		},
	)
}

// getSessionCacheKey returns the cache key for the given service and app address.
// It uses the sessionCacheKeyPrefix, service ID, and app address to create a unique key.
//
// eg. "session:eth:pokt1up7zlytnmvlsuxzpzvlrta95347w322adsxslw"
func getSessionCacheKey(serviceID protocol.ServiceID, appAddr string) string {
	return fmt.Sprintf("%s:%s:%s", sessionCacheKeyPrefix, serviceID, appAddr)
}

// ValidateRelayResponse delegates to the underlying node.
func (cfn *cachingFullNode) ValidateRelayResponse(
	supplierAddr sdk.SupplierAddress,
	responseBz []byte,
) (*servicetypes.RelayResponse, error) {
	return cfn.lazyFullNode.ValidateRelayResponse(supplierAddr, responseBz)
}

// GetAccountClient delegates to the underlying node.
func (cfn *cachingFullNode) GetAccountClient() *sdk.AccountClient {
	return cfn.lazyFullNode.GetAccountClient()
}

// IsHealthy delegates to the underlying node.
//
// TODO_IMPROVE(@commoddity):
//   - Implement a more sophisticated health check
//   - Check for the presence of cached apps and sessions (when the TODO_IMPROVE at the top of this file is addressed)
//   - For now, always returns true because the cache is populated incrementally as new apps and sessions are requested.
func (cfn *cachingFullNode) IsHealthy() bool {
	return cfn.lazyFullNode.IsHealthy()
}

// ---------------- Caching Account Fetcher ----------------

// cachingPoktNodeAccountFetcher implements the PoktNodeAccountFetcher interface.
var _ sdk.PoktNodeAccountFetcher = &cachingPoktNodeAccountFetcher{}

// cachingPoktNodeAccountFetcher wraps an sdk.PoktNodeAccountFetcher with caching capabilities.
// It implements the same PoktNodeAccountFetcher interface but adds sturdyc caching
// in order to reduce repeated and unnecessary requests to the full node.
type cachingPoktNodeAccountFetcher struct {
	logger polylog.Logger

	// The underlying account client to delegate to when cache misses occur
	underlyingAccountClient sdk.PoktNodeAccountFetcher

	// Cache for account responses
	accountCache *sturdyc.Client[*accounttypes.QueryAccountResponse]
}

// Account implements the PoktNodeAccountFetcher interface with caching.
//
// It matches the function signature of the CosmosSDK's account fetcher
// in order to satisfy the PoktNodeAccountFetcher interface.
//
// See CosmosSDK's account fetcher:
// https://github.com/cosmos/cosmos-sdk/blob/main/x/auth/types/query.pb.go#L1090
func (c *cachingPoktNodeAccountFetcher) Account(
	ctx context.Context,
	req *accounttypes.QueryAccountRequest,
	opts ...grpcoptions.CallOption,
) (*accounttypes.QueryAccountResponse, error) {
	return c.accountCache.GetOrFetch(
		ctx,
		getAccountCacheKey(req.Address),
		func(fetchCtx context.Context) (*accounttypes.QueryAccountResponse, error) {
			c.logger.Debug().Str("account_key", getAccountCacheKey(req.Address)).Msgf(
				"[cachingPoktNodeAccountFetcher.Account] Making request to full node",
			)
			return c.underlyingAccountClient.Account(fetchCtx, req, opts...)
		},
	)
}

// getAccountCacheKey returns the cache key for the given account address.
// It uses the accountCacheKeyPrefix and the account address to create a unique key.
//
// eg. "account:pokt1up7zlytnmvlsuxzpzvlrta95347w322adsxslw"
func getAccountCacheKey(address string) string {
	return fmt.Sprintf("%s:%s", accountCacheKeyPrefix, address)
}

func initAccountCache() *sturdyc.Client[*accounttypes.QueryAccountResponse] {
	// Configure account cache with early refreshes to avoid thundering herd.
	accountMinRefreshDelay, accountMaxRefreshDelay := getCacheDelays(0.7, 0.9, accountCacheTTL)

	// Create the account cache, which will be used to cache account responses.
	accountCache := sturdyc.New[*accounttypes.QueryAccountResponse](
		cacheCapacity,
		numShards,
		accountCacheTTL,
		evictionPercentage,
		// See: https://github.com/viccon/sturdyc?tab=readme-ov-file#early-refreshes
		sturdyc.WithEarlyRefreshes(
			accountMinRefreshDelay,
			accountMaxRefreshDelay,
			accountCacheTTL,
			retryBaseDelay,
		),
	)

	return accountCache
}

// replaceLazyFullNodeAccountFetcher wraps the original account fetcher with the caching
// account fetcher and replaces the lazy full node's account fetcher with the caching one.
//
// This is used to replace the lazy full node's account fetcher with the caching one.
// It is used in the NewCachingFullNode function to create a new caching full node.
func replaceLazyFullNodeAccountFetcher(
	logger polylog.Logger,
	lazyFullNode *lazyFullNode,
	accountCache *sturdyc.Client[*accounttypes.QueryAccountResponse],
) {
	// Wrap the original account fetcher with the caching account fetcher
	originalAccountFetcher := lazyFullNode.accountClient.PoktNodeAccountFetcher

	// Replace the lazy full node's account fetcher with the caching one.
	lazyFullNode.accountClient = &sdk.AccountClient{
		PoktNodeAccountFetcher: &cachingPoktNodeAccountFetcher{
			logger:                  logger,
			underlyingAccountClient: originalAccountFetcher,
			accountCache:            accountCache,
		},
	}
}
