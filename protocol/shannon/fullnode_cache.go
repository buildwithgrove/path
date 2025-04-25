package shannon

import (
	"context"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/protocol"
)

// TODO_IN_THIS_PR(@commoddity): implement a FullNode interface which caches the results.
// This needs to consider the GatewayMode:
//		A. Centralized: the list of owned apps is specified in advance, and onchain data can be cached before any requests are received.
//		B. Delegated: cache needs to be done in Lazy/incremental way, as user requests specifying different apps are received.

const (
	// TODO_IN_THIS_PR(@commoddity): make the cache TTL configurable in config YAML file.
	defaultCacheTTL             = 30 * time.Second
	defaultCacheCleanupInterval = 1 * time.Minute

	// Cache key prefixes to avoid collisions
	appCacheKeyPrefix     = "app:"
	sessionCacheKeyPrefix = "session:"
)

// CachingFullNode implements the FullNode interface by wrapping a LazyFullNode
// and caching results to improve performance.
type CachingFullNode struct {
	logger polylog.Logger

	// Use a LazyFullNode as the underlying node
	// for fetching data from the protocol.
	underlyingNode *LazyFullNode

	// Separate caches for different entity types
	appCache     *cache.Cache
	sessionCache *cache.Cache
}

// NewCachingFullNode creates a new CachingFullNode that wraps the given LazyFullNode.
func NewCachingFullNode(logger polylog.Logger, underlyingNode *LazyFullNode) *CachingFullNode {
	return &CachingFullNode{
		logger: logger.With("component", "CachingFullNode"),

		underlyingNode: underlyingNode,

		appCache:     cache.New(defaultCacheTTL, defaultCacheCleanupInterval),
		sessionCache: cache.New(defaultCacheTTL, defaultCacheCleanupInterval),
	}
}

// GetApp returns the application with the given address, using a cached version if available.
func (cfn *CachingFullNode) GetApp(ctx context.Context, appAddr string) (*apptypes.Application, error) {
	cacheKey := createCacheKey(appCacheKeyPrefix, appAddr)

	// Check cache first
	if cachedValue, found := cfn.appCache.Get(cacheKey); found {
		cfn.logger.Debug().Str("app_addr", appAddr).Msg("Returning cached application")
		return cachedValue.(*apptypes.Application), nil
	}

	// Cache miss - get from underlying node
	app, err := cfn.underlyingNode.GetApp(ctx, appAddr)
	if err != nil {
		return nil, err
	}

	// Store in cache
	cfn.appCache.Set(cacheKey, app, cache.DefaultExpiration)

	cfn.logger.Debug().Str("app_addr", appAddr).Msg("Cached application")

	return app, nil
}

// GetSession returns the session for the given service and app, using a cached version if available.
func (cfn *CachingFullNode) GetSession(
	ctx context.Context,
	serviceID protocol.ServiceID,
	appAddr string,
) (sessiontypes.Session, error) {
	// Create a unique cache key for this service+app combination
	cacheKey := createCacheKey(sessionCacheKeyPrefix, fmt.Sprintf("%s:%s", serviceID, appAddr))

	// Check cache first
	if cachedValue, found := cfn.sessionCache.Get(cacheKey); found {
		cfn.logger.Debug().
			Str("service_id", string(serviceID)).
			Str("app_addr", appAddr).
			Msg("Returning cached session")
		return cachedValue.(sessiontypes.Session), nil
	}

	// Cache miss - get from underlying node
	session, err := cfn.underlyingNode.GetSession(ctx, serviceID, appAddr)
	if err != nil {
		return sessiontypes.Session{}, err
	}

	// Store in cache
	cfn.sessionCache.Set(cacheKey, session, cache.DefaultExpiration)

	cfn.logger.Debug().
		Str("service_id", string(serviceID)).
		Str("app_addr", appAddr).
		Str("session_id", session.SessionId).
		Msg("Cached session")

	return session, nil
}

// ValidateRelayResponse delegates to the underlying node's implementation.
// This operation doesn't benefit from caching as it's validating external data.
func (cfn *CachingFullNode) ValidateRelayResponse(
	supplierAddr sdk.SupplierAddress,
	responseBz []byte,
) (*servicetypes.RelayResponse, error) {
	return cfn.underlyingNode.ValidateRelayResponse(supplierAddr, responseBz)
}

// IsHealthy returns true if the cache is ready and the underlying node is healthy.
func (cfn *CachingFullNode) IsHealthy() bool {
	// The cache is always ready, so we only need to check if the underlying node is healthy
	return cfn.underlyingNode.IsHealthy()
}

// GetAccountClient delegates to the underlying node.
func (cfn *CachingFullNode) GetAccountClient() *sdk.AccountClient {
	return cfn.underlyingNode.GetAccountClient()
}

// createCacheKey creates a cache key for the given prefix and key.
// eg. createCacheKey("app:", "0x123") -> "app:0x123"
// eg. createCacheKey("session:", "0x123:0x456") -> "session:0x123:0x456"
func createCacheKey(prefix string, key string) string {
	return fmt.Sprintf("%s%s", prefix, key)
}
