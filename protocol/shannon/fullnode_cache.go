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
//
// TODO_IMPROVE: Implement interface to adapt caching strategy based on GatewayMode
// TODO_IMPROVE: Make cache TTLs configurable in config YAML file
const (
	// Cache TTLs and cleanup intervals
	defaultAppCacheTTL     = 5 * time.Minute
	defaultSessionCacheTTL = 30 * time.Second

	defaultAppCacheCleanupInterval     = defaultAppCacheTTL * 2
	defaultSessionCacheCleanupInterval = defaultSessionCacheTTL * 2

	// Preemptive refresh thresholds (20% of TTL)
	appRefreshThreshold     = 1 * time.Minute
	sessionRefreshThreshold = 6 * time.Second

	// Cache key prefixes
	appCacheKeyPrefix     = "app"
	sessionCacheKeyPrefix = "session"
)

var _ FullNode = &cachingFullNode{}

type cachingFullNode struct {
	lazyFullNode *lazyFullNode

	appCache     *cache.Cache
	sessionCache *cache.Cache

	// Track ongoing refresh operations to prevent duplicates
	appRefreshInProgress     map[string]bool
	appRefreshMutex          sync.Mutex
	sessionRefreshInProgress map[string]bool
	sessionRefreshMutex      sync.Mutex
}

func NewCachingFullNode(lazyFullNode *lazyFullNode) *cachingFullNode {
	return &cachingFullNode{
		lazyFullNode:             lazyFullNode,
		appCache:                 cache.New(defaultAppCacheTTL, defaultAppCacheCleanupInterval),
		sessionCache:             cache.New(defaultSessionCacheTTL, defaultSessionCacheCleanupInterval),
		appRefreshInProgress:     make(map[string]bool),
		sessionRefreshInProgress: make(map[string]bool),
	}
}

func (cfn *cachingFullNode) GetApp(ctx context.Context, appAddr string) (*apptypes.Application, error) {
	key := createCacheKey(appCacheKeyPrefix, appAddr)

	if cached, expiration, found := cfn.appCache.GetWithExpiration(key); found {
		app := cached.(*apptypes.Application)

		if cfn.shouldRefresh(expiration, appRefreshThreshold) {
			cfn.refreshAppAsync(ctx, appAddr, key)
		}

		return app, nil
	}

	// Cache miss - fetch and store
	app, err := cfn.lazyFullNode.GetApp(ctx, appAddr)
	if err != nil {
		return nil, err
	}

	cfn.appCache.Set(key, app, cache.DefaultExpiration)
	return app, nil
}

func (cfn *cachingFullNode) GetSession(ctx context.Context, serviceID protocol.ServiceID, appAddr string) (sessiontypes.Session, error) {
	key := createCacheKey(sessionCacheKeyPrefix, fmt.Sprintf("%s:%s", serviceID, appAddr))

	if cached, expiration, found := cfn.sessionCache.GetWithExpiration(key); found {
		session := cached.(sessiontypes.Session)

		if cfn.shouldRefresh(expiration, sessionRefreshThreshold) {
			cfn.refreshSessionAsync(ctx, serviceID, appAddr, key)
		}

		return session, nil
	}

	// Cache miss - fetch and store
	session, err := cfn.lazyFullNode.GetSession(ctx, serviceID, appAddr)
	if err != nil {
		return sessiontypes.Session{}, err
	}

	cfn.sessionCache.Set(key, session, cache.DefaultExpiration)
	return session, nil
}

// shouldRefresh determines if a cache entry should be refreshed preemptively
func (cfn *cachingFullNode) shouldRefresh(expiration time.Time, threshold time.Duration) bool {
	if expiration.IsZero() {
		return false
	}

	timeUntilExpiry := time.Until(expiration)
	return timeUntilExpiry <= threshold && timeUntilExpiry > 0
}

// refreshAppAsync triggers background refresh for app cache entry
func (cfn *cachingFullNode) refreshAppAsync(_ context.Context, appAddr, key string) {
	cfn.appRefreshMutex.Lock()
	if cfn.appRefreshInProgress[key] {
		cfn.appRefreshMutex.Unlock()
		return
	}
	cfn.appRefreshInProgress[key] = true
	cfn.appRefreshMutex.Unlock()

	go func() {
		defer func() {
			cfn.appRefreshMutex.Lock()
			delete(cfn.appRefreshInProgress, key)
			cfn.appRefreshMutex.Unlock()
		}()

		refreshCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if app, err := cfn.lazyFullNode.GetApp(refreshCtx, appAddr); err == nil {
			cfn.appCache.Set(key, app, cache.DefaultExpiration)
		}
	}()
}

// refreshSessionAsync triggers background refresh for session cache entry
func (cfn *cachingFullNode) refreshSessionAsync(_ context.Context, serviceID protocol.ServiceID, appAddr, key string) {
	cfn.sessionRefreshMutex.Lock()
	if cfn.sessionRefreshInProgress[key] {
		cfn.sessionRefreshMutex.Unlock()
		return
	}
	cfn.sessionRefreshInProgress[key] = true
	cfn.sessionRefreshMutex.Unlock()

	go func() {
		defer func() {
			cfn.sessionRefreshMutex.Lock()
			delete(cfn.sessionRefreshInProgress, key)
			cfn.sessionRefreshMutex.Unlock()
		}()

		refreshCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if session, err := cfn.lazyFullNode.GetSession(refreshCtx, serviceID, appAddr); err == nil {
			cfn.sessionCache.Set(key, session, cache.DefaultExpiration)
		}
	}()
}

// Delegate methods to underlying lazy full node
func (cfn *cachingFullNode) ValidateRelayResponse(supplierAddr sdk.SupplierAddress, responseBz []byte) (*servicetypes.RelayResponse, error) {
	return cfn.lazyFullNode.ValidateRelayResponse(supplierAddr, responseBz)
}

func (cfn *cachingFullNode) IsHealthy() bool {
	return cfn.lazyFullNode.IsHealthy()
}

func (cfn *cachingFullNode) GetAccountClient() *sdk.AccountClient {
	return cfn.lazyFullNode.GetAccountClient()
}

func createCacheKey(prefix, key string) string {
	return fmt.Sprintf("%s:%s", prefix, key)
}
