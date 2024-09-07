package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/config"
	"github.com/buildwithgrove/path/db/driver"
)

type (
	cache struct {
		userApps             map[driver.UserAppID]driver.UserApp
		db                   dbDriver
		cacheRefreshInterval time.Duration
		mu                   sync.RWMutex
		logger               polylog.Logger
	}
	dbDriver interface {
		GetUserApps(ctx context.Context) (map[driver.UserAppID]driver.UserApp, error)
	}
)

func NewCache(config config.UserDataConfig, logger polylog.Logger) (*cache, func() error, error) {

	db, cleanup, err := driver.NewPostgresDriver(config.DBConnectionString)
	if err != nil {
		return nil, cleanup, fmt.Errorf("failed to initialize db: %w", err)
	}

	cache := &cache{
		userApps:             make(map[driver.UserAppID]driver.UserApp),
		db:                   db,
		cacheRefreshInterval: config.CacheRefreshInterval,
		mu:                   sync.RWMutex{},
		logger:               logger,
	}

	if err := cache.setCache(context.Background()); err != nil {
		return nil, cleanup, fmt.Errorf("failed to set cache: %w", err)
	}

	go cache.cacheRefreshHandler(context.Background())

	return cache, cleanup, nil
}

func (c *cache) GetUserApp(ctx context.Context, userAppID driver.UserAppID) (driver.UserApp, bool) {
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
