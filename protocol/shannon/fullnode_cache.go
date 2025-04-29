package shannon

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pokt-network/poktroll/pkg/polylog"
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

// NewCachingFullNode creates a new CachingFullNode that wraps the given LazyFullNode.
func NewCachingFullNode(logger polylog.Logger, lazyFullNode *lazyFullNode) *cachingFullNode {
	return &cachingFullNode{
		logger: logger.With("component", "CachingFullNode"),

		lazyFullNode: lazyFullNode,

		appCache:     cache.New(defaultCacheTTL, defaultCacheCleanupInterval),
		sessionCache: cache.New(defaultCacheTTL, defaultCacheCleanupInterval),
	}
}

var _ FullNode = &cachingFullNode{}

// cachingFullNode implements the FullNode interface by wrapping a LazyFullNode
// and caching results to improve performance.
type cachingFullNode struct {
	logger polylog.Logger

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

// GetApp returns the application with the given address, using a cached version if available.
func (cfn *cachingFullNode) GetApp(ctx context.Context, appAddr string) (*apptypes.Application, error) {
	appCacheKey := createCacheKey(appCacheKeyPrefix, appAddr)

	// Check cache first
	if cachedApp, found := cfn.appCache.Get(appCacheKey); found {
		cfn.logger.Debug().Str("app_addr", appAddr).Msg("Returning cached application")

		// Type assertion is safe because we know the cache value can only be *apptypes.Application.
		return cachedApp.(*apptypes.Application), nil
	}

	// Use mutex to prevent multiple concurrent cache updates for the same app
	cfn.appMutex.Lock()
	defer cfn.appMutex.Unlock()

	// Double-check cache after acquiring lock (follows standard double-checked locking pattern)
	if cachedApp, found := cfn.appCache.Get(appCacheKey); found {
		cfn.logger.Debug().Str("app_addr", appAddr).Msg("Returning cached application after lock")
		return cachedApp.(*apptypes.Application), nil
	}

	// Cache miss - get from underlying node
	app, err := cfn.lazyFullNode.GetApp(ctx, appAddr)
	if err != nil {
		return nil, err
	}

	// Store in cache if the app is found.
	cfn.appCache.Set(appCacheKey, app, cache.DefaultExpiration)

	cfn.logger.Debug().Str("app_addr", appAddr).Msg("Cached application")

	return app, nil
}

// GetSession returns the session for the given service and app, using a cached version if available.
func (cfn *cachingFullNode) GetSession(
	ctx context.Context,
	serviceID protocol.ServiceID,
	appAddr string,
) (sessiontypes.Session, error) {
	// Create a unique cache key for this service+app combination
	sessionCacheKey := createCacheKey(sessionCacheKeyPrefix, fmt.Sprintf("%s:%s", serviceID, appAddr))

	logger := cfn.logger.With("service_id", string(serviceID), "app_addr", appAddr)

	// Check cache first
	if cachedSession, found := cfn.sessionCache.Get(sessionCacheKey); found {
		logger.Debug().Msg("Returning cached session")

		// Type assertion is safe because we know the cache value can only be sessiontypes.Session.
		return cachedSession.(sessiontypes.Session), nil
	}

	// Use mutex to prevent multiple concurrent cache updates for the same session
	cfn.sessionMutex.Lock()
	defer cfn.sessionMutex.Unlock()

	// Double-check cache after acquiring lock (follows standard double-checked locking pattern)
	if cachedSession, found := cfn.sessionCache.Get(sessionCacheKey); found {
		logger.Debug().Msg("Returning cached session after lock")
		return cachedSession.(sessiontypes.Session), nil
	}

	// Cache miss - get from underlying node
	session, err := cfn.lazyFullNode.GetSession(ctx, serviceID, appAddr)
	if err != nil {
		return sessiontypes.Session{}, err
	}

	// Store in cache if the session is found.
	cfn.sessionCache.Set(sessionCacheKey, session, cache.DefaultExpiration)

	logger.Debug().Str("session_id", session.SessionId).Msg("Cached session")

	return session, nil
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
