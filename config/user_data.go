package config

import (
	"fmt"
	"time"

	"github.com/buildwithgrove/path/config/utils"
)

const defaultCacheRefreshInterval = 5 * time.Minute

/* --------------------------------- User Data Config Struct -------------------------------- */

// UserDataConfig contains user data configuration settings, which are only relevant if user data handling
// is enabled for the gateway by setting the 'user_data_config' field in the config YAML file.
//
// The DB connection string must be for a valid Postgres database, which will
// contain user data for the Gateway. A cache refresh interval may also be set.
type UserDataConfig struct {
	PostgresConnectionString string `yaml:"postgres_connection_string"`
	// The interval at which the local user data cache should be refreshed from the
	// connected Postgres DB. Must be set in valid YAML time syntax, eg 30s, 5m, etc.
	CacheRefreshInterval time.Duration `yaml:"cache_refresh_interval"`
}

/* --------------------------------- User Data Config Private Helpers -------------------------------- */

func (c *UserDataConfig) validate() error {
	if !utils.IsValidPostgresConnectionString(c.PostgresConnectionString) {
		return fmt.Errorf("invalid DB connection string: %s", c.PostgresConnectionString)
	}
	return nil
}

func (c *UserDataConfig) hydrateDefaults() {
	if c.CacheRefreshInterval == 0 {
		c.CacheRefreshInterval = defaultCacheRefreshInterval
	}
}
