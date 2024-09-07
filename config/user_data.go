package config

import (
	"fmt"
	"time"

	"github.com/buildwithgrove/path/config/utils"
)

const defaultCacheRefreshInterval = 5 * time.Minute

/* --------------------------------- User Data Config Struct -------------------------------- */

// UserDataConfig contains user data configuration settings.
type UserDataConfig struct {
	DBConnectionString   string        `yaml:"db_connection_string"`
	CacheRefreshInterval time.Duration `yaml:"cache_refresh_interval"`
}

/* --------------------------------- User Data Config Private Helpers -------------------------------- */

func (c *UserDataConfig) validate() error {
	if !utils.IsValidDBConnectionString(c.DBConnectionString) {
		return fmt.Errorf("invalid DB connection string: %s", c.DBConnectionString)
	}
	return nil
}

func (c *UserDataConfig) hydrateDefaults() {
	if c.CacheRefreshInterval == 0 {
		c.CacheRefreshInterval = defaultCacheRefreshInterval
	}
}
