//go:build auth_server

package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/buildwithgrove/auth-server/db/postgres"
)

const (
	defaultPort                 = 10003
	defaultCacheRefreshInterval = 5 * time.Minute
)

/* ---------------------------------  Authorizer Server Config Struct -------------------------------- */

// AuthServerConfig contains the configuration for the authorizer server.
type AuthServerConfig struct {
	PostgresConnectionString string        `yaml:"postgres_connection_string"`
	CacheRefreshInterval     time.Duration `yaml:"cache_refresh_interval"`
	Port                     int           `yaml:"port"`
}

// LoadAuthServerConfig reads a YAML configuration file from the specified path
// and unmarshals its content into a AuthServerConfig instance.
func LoadAuthServerConfigFromYAML(path string) (AuthServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AuthServerConfig{}, err
	}

	var config AuthServerConfig
	if err = yaml.Unmarshal(data, &config); err != nil {
		return AuthServerConfig{}, err
	}

	config.hydrateConfig()

	return config, config.validate()
}

/* --------------------------------- Authorizer Server Config Helpers -------------------------------- */

func (c *AuthServerConfig) hydrateConfig() {
	if c.CacheRefreshInterval == 0 {
		c.CacheRefreshInterval = defaultCacheRefreshInterval
	}
	if c.Port == 0 {
		c.Port = defaultPort
	}
}

func (c AuthServerConfig) validate() error {
	if !postgres.IsValidPostgresConnectionString(c.PostgresConnectionString) {
		return fmt.Errorf("invalid postgres connection string: %s", c.PostgresConnectionString)
	}
	return nil
}
