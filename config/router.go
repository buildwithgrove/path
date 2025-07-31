package config

import (
	"time"
)

/* --------------------------------- Router Config Defaults -------------------------------- */

// TODO_IMPROVE: Make all of these configurable for PATH users
const (
	// default PATH port
	defaultPort = 3069

	// defaultMaxRequestHeaderBytes is the default maximum size of the HTTP request header.
	defaultMaxRequestHeaderBytes = 2 * 1e6 // 2 MB

	// https://pkg.go.dev/net/http#Server
	// HTTP server's default timeout values.
	defaultHTTPServerReadTimeout  = 20 * time.Second
	defaultHTTPServerWriteTimeout = 30 * time.Second
	defaultHTTPServerIdleTimeout  = 120 * time.Second
)

/* --------------------------------- Router Config Struct -------------------------------- */

// RouterConfig contains server configuration settings.
// See default values above.
type RouterConfig struct {
	Port                  int           `yaml:"port"`
	MaxRequestHeaderBytes int           `yaml:"max_request_header_bytes"`
	ReadTimeout           time.Duration `yaml:"read_timeout"`
	WriteTimeout          time.Duration `yaml:"write_timeout"`
	IdleTimeout           time.Duration `yaml:"idle_timeout"`
}

/* --------------------------------- Router Config Private Helpers -------------------------------- */

// hydrateRouterDefaults assigns default values to RouterConfig fields if they are not set.
func (c *RouterConfig) hydrateRouterDefaults() {
	if c.Port == 0 {
		c.Port = defaultPort
	}
	if c.MaxRequestHeaderBytes == 0 {
		c.MaxRequestHeaderBytes = defaultMaxRequestHeaderBytes
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
