package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/auth-plugin/user"
)

// userDataCache is an in-memory cache that stores gateway endpoints and their associated data.
type userDataCache struct {
	db DBDriver

	gatewayEndpoints     map[user.EndpointID]user.GatewayEndpoint
	gatewayEndpointsMu   sync.RWMutex
	cacheRefreshInterval time.Duration

	logger polylog.Logger
}

// NewUserDataCache creates a new user data cache, which stores GatewayEndpoints in memory for fast access.
// It refreshes the cache from the Postgres database connection at the specified interval.
func NewUserDataCache(driver DBDriver, cacheRefreshInterval time.Duration, logger polylog.Logger) (*userDataCache, error) {
	cache := &userDataCache{
		db: driver,

		gatewayEndpoints:     make(map[user.EndpointID]user.GatewayEndpoint),
		cacheRefreshInterval: cacheRefreshInterval,
		gatewayEndpointsMu:   sync.RWMutex{},

		logger: logger.With("component", "user_data_cache"),
	}

	// Initialize the cache with the GatewayEndpoints from the Postgres database.
	if err := cache.updateCache(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to set cache: %w", err)
	}

	go cache.cacheRefreshHandler(context.Background())

	return cache, nil
}

// GetGatewayEndpoint returns a GatewayEndpoint from the cache and a bool indicating if it exists in the cache.
func (c *userDataCache) GetGatewayEndpoint(endpointID user.EndpointID) (user.GatewayEndpoint, bool) {
	c.gatewayEndpointsMu.RLock()
	defer c.gatewayEndpointsMu.RUnlock()

	gatewayEndpoint, ok := c.gatewayEndpoints[endpointID]
	return gatewayEndpoint, ok
}

// cacheRefreshHandler is intended to be run in a go routine.
func (c *userDataCache) cacheRefreshHandler(ctx context.Context) {
	for {
		<-time.After(c.cacheRefreshInterval)

		if err := c.updateCache(ctx); err != nil {
			c.logger.Error().Err(err).Msg("failed to refresh cache")
		}
	}
}

// updateCache fetches the GatewayEndpoints from the Postgres database and sets them in the cache.
// TODO_IMPROVE(@commoddity) - set up a Postgres listener to update the cache when
// the GatewayEndpoints change, rather than having to poll the database on an interval.
func (c *userDataCache) updateCache(ctx context.Context) error {
	gatewayEndpoints, err := c.db.GetGatewayEndpoints(ctx)
	if err != nil {
		return fmt.Errorf("failed to get gateway endpoints: %w", err)
	}

	c.gatewayEndpointsMu.Lock()
	defer c.gatewayEndpointsMu.Unlock()
	c.gatewayEndpoints = gatewayEndpoints

	return nil
}
