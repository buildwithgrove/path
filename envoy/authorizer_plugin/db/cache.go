//go:build authorizer_plugin

package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/authorizer-plugin/user"
)

// userDataCache is an in-memory cache that stores gateway endpoints and their associated data.
type userDataCache struct {
	db DBDriver

	gatewayEndpoints     map[user.EndpointID]user.GatewayEndpoint
	cacheRefreshInterval time.Duration
	mu                   sync.RWMutex

	logger polylog.Logger
}

func NewUserDataCache(driver DBDriver, cacheRefreshInterval time.Duration, logger polylog.Logger) (*userDataCache, error) {
	cache := &userDataCache{
		db: driver,

		gatewayEndpoints:     make(map[user.EndpointID]user.GatewayEndpoint),
		cacheRefreshInterval: cacheRefreshInterval,
		mu:                   sync.RWMutex{},

		logger: logger.With("component", "user_data_cache"),
	}

	if err := cache.setCache(context.Background()); err != nil {
		cache.logger.Error().Err(err).Msg("failed to set cache")
		return nil, fmt.Errorf("failed to set cache: %w", err)
	}

	go cache.cacheRefreshHandler(context.Background())

	return cache, nil
}

func (c *userDataCache) GetGatewayEndpoint(ctx context.Context, endpointID user.EndpointID) (user.GatewayEndpoint, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	gatewayEndpoint, ok := c.gatewayEndpoints[endpointID]
	return gatewayEndpoint, ok
}

func (c *userDataCache) cacheRefreshHandler(ctx context.Context) {
	for {
		<-time.After(c.cacheRefreshInterval)

		err := c.setCache(ctx)
		if err != nil {
			c.logger.Error().Err(err).Msg("failed to refresh cache")
		}
	}
}

func (c *userDataCache) setCache(ctx context.Context) error {
	gatewayEndpoints, err := c.db.GetGatewayEndpoints(ctx)
	if err != nil {
		return fmt.Errorf("failed to get gateway endpoints: %w", err)
	}

	c.logger.Info().Msgf("successfully set cache with %d gateway endpoints", len(gatewayEndpoints))

	c.mu.Lock()
	c.gatewayEndpoints = gatewayEndpoints
	c.mu.Unlock()

	return nil
}
