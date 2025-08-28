package grpc

import (
	"crypto/tls"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO_TECHDEBT: Make all of these configurable
const (
	defaultBackoffBaseDelay  = 1 * time.Second
	defaultBackoffMaxDelay   = 60 * time.Second
	defaultMinConnectTimeout = 10 * time.Second
	defaultKeepAliveTime     = 30 * time.Second
	defaultKeepAliveTimeout  = 30 * time.Second
)

type GRPCConfig struct {
	HostPort          string        `yaml:"host_port"`
	Insecure          bool          `yaml:"insecure"`
	BackoffBaseDelay  time.Duration `yaml:"backoff_base_delay"`
	BackoffMaxDelay   time.Duration `yaml:"backoff_max_delay"`
	MinConnectTimeout time.Duration `yaml:"min_connect_timeout"`
	KeepAliveTime     time.Duration `yaml:"keep_alive_time"`
	KeepAliveTimeout  time.Duration `yaml:"keep_alive_timeout"`
}

// ConnectGRPC creates a new gRPC client connection.
//
// Configuration:
// - TLS is enabled by default; set `grpc_config.insecure` to disable.
// - Backoff parameters can be customized under `grpc_config` in YAML.
//
// Notes:
//   - gRPC settings are intentionally minimal to keep E2E tests focused on
//     gateway functionality rather than gRPC configuration.
//
// TODO_TECHDEBT: migrate to an enhanced gRPC connection with reconnect logic.
func ConnectGRPC(config GRPCConfig) (*grpc.ClientConn, error) {
	if config.Insecure {
		transport := grpc.WithTransportCredentials(insecure.NewCredentials())
		dialOptions := []grpc.DialOption{transport}
		return grpc.NewClient(
			config.HostPort,
			dialOptions...,
		)
	}

	// TODO_TECHDEBT: make the necessary changes to allow using grpc.NewClient here.
	// Currently using the grpc.NewClient method fails the E2E tests.
	return grpc.Dial( //nolint:all
		config.HostPort,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
	)
}

func (c *GRPCConfig) HydrateDefaults() GRPCConfig {
	if c.BackoffBaseDelay == 0 {
		c.BackoffBaseDelay = defaultBackoffBaseDelay
	}
	if c.BackoffMaxDelay == 0 {
		c.BackoffMaxDelay = defaultBackoffMaxDelay
	}
	if c.MinConnectTimeout == 0 {
		c.MinConnectTimeout = defaultMinConnectTimeout
	}
	if c.KeepAliveTime == 0 {
		c.KeepAliveTime = defaultKeepAliveTime
	}
	if c.KeepAliveTimeout == 0 {
		c.KeepAliveTimeout = defaultKeepAliveTimeout
	}
	return *c
}
