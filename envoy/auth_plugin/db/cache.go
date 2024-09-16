//go:build auth_plugin

package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/authorizer-plugin/types"
)

// userDataCache is an in-memory cache that stores gateway endpoints and their associated data.
type userDataCache struct {
	db DBDriver

	gatewayEndpoints     map[types.EndpointID]types.GatewayEndpoint
	gatewayEndpointsMu   sync.RWMutex
	cacheRefreshInterval time.Duration

	logger polylog.Logger
}

func NewUserDataCache(driver DBDriver, cacheRefreshInterval time.Duration, logger polylog.Logger) (*userDataCache, error) {
	cache := &userDataCache{
		db: driver,

		gatewayEndpoints:     make(map[types.EndpointID]types.GatewayEndpoint),
		cacheRefreshInterval: cacheRefreshInterval,
		gatewayEndpointsMu:   sync.RWMutex{},

		logger: logger.With("component", "user_data_cache"),
	}

	if err := cache.updateCache(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to set cache: %w", err)
	}

	go cache.cacheRefreshHandler(context.Background())

	return cache, nil
}

func (c *userDataCache) GetGatewayEndpoint(ctx context.Context, endpointID types.EndpointID) (types.GatewayEndpoint, bool) {
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
