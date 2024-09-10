package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/user"
)

type cache struct {
	userApps             map[user.UserAppID]user.UserApp
	db                   Driver
	cacheRefreshInterval time.Duration
	mu                   sync.RWMutex
	logger               polylog.Logger
}

type Driver interface {
	GetUserApps(ctx context.Context) (map[user.UserAppID]user.UserApp, error)
}

func NewCache(driver Driver, cacheRefreshInterval time.Duration, logger polylog.Logger) (*cache, error) {
	cache := &cache{
		userApps:             make(map[user.UserAppID]user.UserApp),
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

func (c *cache) GetUserApp(ctx context.Context, userAppID user.UserAppID) (user.UserApp, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	userApp, ok := c.userApps[userAppID]
	return userApp, ok
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
	userApps, err := c.db.GetUserApps(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user apps: %w", err)
	}

	c.mu.Lock()
	c.userApps = userApps
	c.mu.Unlock()

	return nil
}
