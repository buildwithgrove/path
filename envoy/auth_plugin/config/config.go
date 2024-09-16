//go:build auth_plugin

package config

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultCacheRefreshInterval = 5 * time.Minute

/* ---------------------------------  Authorizer Plugin Config Struct -------------------------------- */

// AuthorizerPluginConfig contains the configuration for the authorize plugin.
type AuthorizerPluginConfig struct {
	PostgresConnectionString string        `yaml:"postgres_connection_string"`
	CacheRefreshInterval     time.Duration `yaml:"cache_refresh_interval"`
}

// LoadAuthorizerPluginConfig reads a YAML configuration file from the specified path
// and unmarshals its content into a AuthorizerPluginConfig instance.
func LoadAuthorizerPluginConfigFromYAML(path string) (AuthorizerPluginConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AuthorizerPluginConfig{}, err
	}

	var config AuthorizerPluginConfig
	if err = yaml.Unmarshal(data, &config); err != nil {
		return AuthorizerPluginConfig{}, err
	}

	config.hydrateConfig()

	return config, config.validate()
}

/* --------------------------------- Authorizer Plugin Config Helpers -------------------------------- */

func (c *AuthorizerPluginConfig) hydrateConfig() {
	if c.CacheRefreshInterval == 0 {
		c.CacheRefreshInterval = defaultCacheRefreshInterval
	}
}

func (c AuthorizerPluginConfig) validate() error {
	if !isValidPostgresConnectionString(c.PostgresConnectionString) {
		return fmt.Errorf("invalid postgres connection string: %s", c.PostgresConnectionString)
	}
	return nil
}

// isValidPostgresConnectionString checks if a string is a valid PostgreSQL connection string.
func isValidPostgresConnectionString(s string) bool {
	// Regular expression to match a valid PostgreSQL connection string
	var dbConnStringRegex = regexp.MustCompile(`^postgres://[^:]+:[^@]+@[^:]+:\d+/.+$`)
	return dbConnStringRegex.MatchString(s)
}
