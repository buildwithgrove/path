package shannon

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"

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
	// TODO_IMPROVE(@commoddity): make the cache TTL configurable in config YAML file.
	defaultCacheTTL             = 30 * time.Second
	defaultCacheCleanupInterval = 1 * time.Minute

	// Cache key prefixes to avoid collisions
	appCacheKeyPrefix     = "app"
	sessionCacheKeyPrefix = "session"
)

var _ FullNode = &cachingFullNode{}

// cachingFullNode implements the FullNode interface by wrapping a LazyFullNode
// and caching results to improve performance.
type cachingFullNode struct {
	// Use a LazyFullNode as the underlying node
	// for fetching data from the protocol.
	lazyFullNode *lazyFullNode

	// Caches for applications and mutexes to protect cache access.
	appCache *cache.Cache
	appMutex sync.Mutex

	// Caches for sessions and mutexes to protect cache access.
	sessionCache *cache.Cache
	sessionMutex sync.Mutex
}

// NewCachingFullNode creates a new CachingFullNode that wraps the given LazyFullNode.
func NewCachingFullNode(lazyFullNode *lazyFullNode) *cachingFullNode {
	return &cachingFullNode{
		lazyFullNode: lazyFullNode,

		appCache:     cache.New(defaultCacheTTL, defaultCacheCleanupInterval),
		sessionCache: cache.New(defaultCacheTTL, defaultCacheCleanupInterval),
	}
}

// GetApp returns the application with the given address, using a cached version if available.
func (cfn *cachingFullNode) GetApp(ctx context.Context, appAddr string) (*apptypes.Application, error) {
	appCacheKey := createCacheKey(appCacheKeyPrefix, appAddr)

	// Try to get app from cache with double-check locking pattern
	if cachedApp, found := cfn.getFromAppCache(appCacheKey); found {
		return cachedApp, nil
	}

	// Cache miss - get from underlying node
	app, err := cfn.lazyFullNode.GetApp(ctx, appAddr)
	if err != nil {
		return nil, err
	}

	// Store in cache if the app is found.
	cfn.appCache.Set(appCacheKey, app, cache.DefaultExpiration)

	return app, nil
}

// getFromAppCache retrieves an application from the cache using double-check locking pattern.
// It returns the cached application and a boolean indicating if it was found.
func (cfn *cachingFullNode) getFromAppCache(cacheKey string) (*apptypes.Application, bool) {
	// Check cache first
	if cachedApp, found := cfn.appCache.Get(cacheKey); found {
		return cachedApp.(*apptypes.Application), true
	}

	// Use mutex to prevent multiple concurrent cache updates for the same app
	cfn.appMutex.Lock()
	defer cfn.appMutex.Unlock()

	// Double-check cache after acquiring lock
	if cachedApp, found := cfn.appCache.Get(cacheKey); found {
		return cachedApp.(*apptypes.Application), true
	}

	return nil, false
}

// GetSession returns the session for the given service and app, using a cached version if available.
func (cfn *cachingFullNode) GetSession(
	ctx context.Context,
	serviceID protocol.ServiceID,
	appAddr string,
) (sessiontypes.Session, error) {
	// Create a unique cache key for this service+app combination
	sessionCacheKey := createCacheKey(sessionCacheKeyPrefix, fmt.Sprintf("%s:%s", serviceID, appAddr))

	// Try to get session from cache with double-check locking pattern
	if cachedSession, found := cfn.getFromSessionCache(sessionCacheKey); found {
		return cachedSession, nil
	}

	// Cache miss - get from underlying node
	session, err := cfn.lazyFullNode.GetSession(ctx, serviceID, appAddr)
	if err != nil {
		return sessiontypes.Session{}, err
	}

	// Store in cache if the session is found.
	cfn.sessionCache.Set(sessionCacheKey, session, cache.DefaultExpiration)

	return session, nil
}

// getFromSessionCache retrieves a session from the cache using double-check locking pattern.
// It returns the cached session and a boolean indicating if it was found.
func (cfn *cachingFullNode) getFromSessionCache(cacheKey string) (sessiontypes.Session, bool) {
	// Check cache first
	if cachedSession, found := cfn.sessionCache.Get(cacheKey); found {
		return cachedSession.(sessiontypes.Session), true
	}

	// Use mutex to prevent multiple concurrent cache updates for the same session
	cfn.sessionMutex.Lock()
	defer cfn.sessionMutex.Unlock()

	// Double-check cache after acquiring lock
	if cachedSession, found := cfn.sessionCache.Get(cacheKey); found {
		return cachedSession.(sessiontypes.Session), true
	}

	return sessiontypes.Session{}, false
}

// ValidateRelayResponse delegates to the underlying node.
func (cfn *cachingFullNode) ValidateRelayResponse(
	supplierAddr sdk.SupplierAddress,
	responseBz []byte,
) (*servicetypes.RelayResponse, error) {
	return cfn.lazyFullNode.ValidateRelayResponse(supplierAddr, responseBz)
}

// IsHealthy delegates to the underlying node.
func (cfn *cachingFullNode) IsHealthy() bool {
	return cfn.lazyFullNode.IsHealthy()
}

// GetAccountClient delegates to the underlying node.
func (cfn *cachingFullNode) GetAccountClient() *sdk.AccountClient {
	return cfn.lazyFullNode.GetAccountClient()
}

// createCacheKey creates a cache key for the given prefix and key.
//
//	eg. createCacheKey("app", "0x123") -> "app/0x123"
//	eg. createCacheKey("session", "anvil:0x456") -> "session/anvil:0x456"
func createCacheKey(prefix string, key string) string {
	return fmt.Sprintf("%s/%s", prefix, key)
}
