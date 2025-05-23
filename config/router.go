package config

import (
	"time"
)

/* --------------------------------- Router Config Defaults -------------------------------- */

const (
	defaultPort               = 3069
	defaultMaxRequestBodySize = 1 << 20 // 1 MB
)

/* --------------------------------- Router Config Struct -------------------------------- */

// RouterConfig contains server configuration settings.
// See default values above.
type RouterConfig struct {
	Port               int           `yaml:"port"`
	MaxRequestBodySize int           `yaml:"max_request_body_size"`
	ReadTimeout        time.Duration `yaml:"read_timeout"`
	WriteTimeout       time.Duration `yaml:"write_timeout"`
	IdleTimeout        time.Duration `yaml:"idle_timeout"`
}

/* --------------------------------- Router Config Private Helpers -------------------------------- */

// hydrateRouterDefaults assigns default values to RouterConfig fields if they are not set.
func (c *RouterConfig) hydrateRouterDefaults() {
	if c.Port == 0 {
		c.Port = defaultPort
	}
	if c.MaxRequestBodySize == 0 {
		c.MaxRequestBodySize = defaultMaxRequestBodySize
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = defaultHTTPServerReadTimeout
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = defaultHTTPServerWriteTimeout
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = defaultHTTPServerIdleTimeout
	}
}
