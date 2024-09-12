package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/user"
)

// cache is an in-memory cache that stores gateway endpoints and their associated data.
type cache struct {
	db DBDriver

	gatewayEndpoints     map[user.EndpointID]user.GatewayEndpoint
	cacheRefreshInterval time.Duration
	mu                   sync.RWMutex

	logger polylog.Logger
}

func NewCache(driver DBDriver, cacheRefreshInterval time.Duration, logger polylog.Logger) (*cache, error) {
	cache := &cache{
		gatewayEndpoints:     make(map[user.EndpointID]user.GatewayEndpoint),
		db:                   driver,
		cacheRefreshInterval: cacheRefreshInterval,
		mu:                   sync.RWMutex{},
		logger:               logger,
	}

	if err := cache.setCache(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to set cache: %w", err)
	}

	go cache.cacheRefreshHandler(context.Background())

	return cache, nil
}

func (c *cache) GetGatewayEndpoint(ctx context.Context, endpointID user.EndpointID) (user.GatewayEndpoint, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	gatewayEndpoint, ok := c.gatewayEndpoints[endpointID]
	return gatewayEndpoint, ok
}

func (c *cache) cacheRefreshHandler(ctx context.Context) {
	for {
		<-time.After(c.cacheRefreshInterval)

		err := c.setCache(ctx)
		if err != nil {
			c.logger.Error().Err(err).Msg("failed to refresh cache")
		}
	}
}

func (c *cache) setCache(ctx context.Context) error {
	gatewayEndpoints, err := c.db.GetGatewayEndpoints(ctx)
	if err != nil {
		return fmt.Errorf("failed to get gateway endpoints: %w", err)
	}

	c.mu.Lock()
	c.gatewayEndpoints = gatewayEndpoints
	c.mu.Unlock()

	return nil
}
