package shannon

import (
	"context"
	"fmt"
	"time"

	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"
	"github.com/viccon/sturdyc"

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
// - 4. Test the performance of a caching node vs lazy node.
const (
	// Applications can be cached indefinitely.
	// They're invalidated only when they unstake.
	// TODO_MAINNET_MIGRATION(@Olshansk): Ensure applications are invalidated during unstaking. Revisit these values after mainnet migration to ensure no race conditions.
	defaultAppCacheTTL = 5 * time.Minute

	// As of #275, on Beta TestNet.
	// - Blocks are 30 seconds
	// - A session is 50 blocks
	// - A grace period is 1 block (i.e. 30 seconds)
	// TODO_MAINNET_MIGRATION(@Olshansk): Revisit these values after mainnet migration to ensure no race conditions.
	// TODO_TECHDEBT(@Olshansk): Update this cache refresh time if onchain block time changes.
	defaultSessionCacheTTL = 30 * time.Second

	// "Refreshing" in SturdyC means proactively fetching fresh data in the background
	// BEFORE the cached entry expires. This prevents cache misses and eliminates latency
	// spikes by ensuring hot data is always available immediately.
	//
	// Apps refresh timing (4-4.5 minutes for 5-minute TTL = 80-90% of TTL):
	// 	- Chosen to balance data freshness with background load
	// 	- Apps change infrequently, so refreshing near expiry is sufficient
	// 	- Random jitter prevents thundering herd on the FullNode
	//
	// TODO_TECHDEBT(@Olshansk): Revisit these early refresh timings and percentages.
	// Consider making them configurable and validate against real-world traffic patterns.
	//
	// Reference: https://github.com/viccon/sturdyc?tab=readme-ov-file#early-refreshes

	// Early refresh configuration for apps
	appMinRefreshDelay = 4 * time.Minute
	appMaxRefreshDelay = 4*time.Minute + 30*time.Second

	// Early refresh configuration for sessions
	sessionMinRefreshDelay = 20 * time.Second
	sessionMaxRefreshDelay = 25 * time.Second

	// Retry base delay for exponential backoff on failed refreshes
	retryBaseDelay = 100 * time.Millisecond

	// Cache configuration
	//
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
)

const (
	appCacheKeyPrefix     = "app"
	sessionCacheKeyPrefix = "session"
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
//   - Apps: 5min TTL, refresh at 4-4.5min (80-90% of TTL)
//   - Sessions: 30sec TTL, refresh at 20-25sec (67-83% of TTL)
//
// Benefits: Zero-latency reads for active traffic, thundering herd protection,
// automatic load balancing, and graceful degradation.
//
// Docs reference: https://github.com/viccon/sturdyc
type cachingFullNode struct {
	// Use a LazyFullNode as the underlying node
	// for fetching data from the protocol.
	lazyFullNode *lazyFullNode

	// Separate SturdyC caches for applications and sessions
	appCache     *sturdyc.Client[*apptypes.Application]
	sessionCache *sturdyc.Client[sessiontypes.Session]
}

// NewCachingFullNode creates a new CachingFullNode that wraps the given LazyFullNode.
func NewCachingFullNode(lazyFullNode *lazyFullNode) *cachingFullNode {
	// Configure app cache with early refreshes
	appCache := sturdyc.New[*apptypes.Application](
		cacheCapacity,
		numShards,
		defaultAppCacheTTL,
		evictionPercentage,
		// See: https://github.com/viccon/sturdyc?tab=readme-ov-file#early-refreshes
		sturdyc.WithEarlyRefreshes(
			appMinRefreshDelay,
			appMaxRefreshDelay,
			defaultAppCacheTTL,
			retryBaseDelay,
		),
	)

	// Configure session cache with early refreshes
	sessionCache := sturdyc.New[sessiontypes.Session](
		cacheCapacity,
		numShards,
		defaultSessionCacheTTL,
		evictionPercentage,
		// See: https://github.com/viccon/sturdyc?tab=readme-ov-file#early-refreshes
		sturdyc.WithEarlyRefreshes(
			sessionMinRefreshDelay,
			sessionMaxRefreshDelay,
			defaultSessionCacheTTL,
			retryBaseDelay,
		),
	)

	return &cachingFullNode{
		lazyFullNode: lazyFullNode,
		appCache:     appCache,
		sessionCache: sessionCache,
	}
}

// GetApp returns the application with the given address, using a cached version if available.
// The cache will automatically refresh the app in the background before it expires.
func (cfn *cachingFullNode) GetApp(ctx context.Context, appAddr string) (*apptypes.Application, error) {
	// See: https://github.com/viccon/sturdyc?tab=readme-ov-file#get-or-fetch
	return cfn.appCache.GetOrFetch(
		ctx,
		getAppCacheKey(appAddr),
		func(fetchCtx context.Context) (*apptypes.Application, error) {
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
