package config

import (
	"fmt"
	"time"
)

// GatewayRelayConfig contains configuration for the PATH gateway relay functionality
type GatewayRelayConfig struct {
	// Relay contains configuration for relay request handling
	Relay RelayConfig `yaml:"relay" mapstructure:"relay"`
}

// RelayConfig contains configuration for relay request handling
type RelayConfig struct {
	// MaxParallelRequests controls the maximum number of parallel requests sent to different endpoints
	// for a single user request. This improves resilience by racing multiple requests.
	// Default: 4, Min: 1, Max: 10
	MaxParallelRequests int `yaml:"max_parallel_requests" mapstructure:"max_parallel_requests"`

	// ParallelRequestTimeout controls the maximum time to wait for parallel requests to complete
	// before cancelling remaining requests. This prevents zombie goroutines.
	// Default: 30s, Min: 1s, Max: 300s
	ParallelRequestTimeout time.Duration `yaml:"parallel_request_timeout" mapstructure:"parallel_request_timeout"`

	// EnableEndpointDiversity controls whether to prefer endpoints with different TLDs
	// when selecting multiple endpoints for parallel requests.
	// Default: true
	EnableEndpointDiversity bool `yaml:"enable_endpoint_diversity" mapstructure:"enable_endpoint_diversity"`
}

// DefaultGatewayRelayConfig returns a GatewayRelayConfig with sensible defaults
func DefaultGatewayRelayConfig() GatewayRelayConfig {
	return GatewayRelayConfig{
		Relay: RelayConfig{
			MaxParallelRequests:     4,
			ParallelRequestTimeout:  30 * time.Second,
			EnableEndpointDiversity: true,
		},
	}
}

// hydrateDefaults applies default values to missing configuration
func (grc *GatewayRelayConfig) hydrateDefaults() {
	if err := grc.Relay.validateAndHydrate(); err != nil {
		// If validation fails, use defaults
		*grc = DefaultGatewayRelayConfig()
	}
}

// Validate validates the gateway relay configuration
func (grc *GatewayRelayConfig) Validate() error {
	// Validate relay config
	if err := grc.Relay.validateAndHydrate(); err != nil {
		return fmt.Errorf("invalid relay config: %w", err)
	}

	return nil
}

// validateAndHydrate validates and applies defaults to relay configuration
func (rc *RelayConfig) validateAndHydrate() error {
	// Validate MaxParallelRequests
	if rc.MaxParallelRequests == 0 {
		rc.MaxParallelRequests = 4 // Default
	}
	if rc.MaxParallelRequests < 1 {
		return fmt.Errorf("max_parallel_requests must be at least 1, got %d", rc.MaxParallelRequests)
	}
	if rc.MaxParallelRequests > 10 {
		return fmt.Errorf("max_parallel_requests must be at most 10, got %d", rc.MaxParallelRequests)
	}

	// Validate ParallelRequestTimeout
	if rc.ParallelRequestTimeout == 0 {
		rc.ParallelRequestTimeout = 30 * time.Second // Default
	}
	if rc.ParallelRequestTimeout < time.Second {
		return fmt.Errorf("parallel_request_timeout must be at least 1s, got %v", rc.ParallelRequestTimeout)
	}
	if rc.ParallelRequestTimeout > 300*time.Second {
		return fmt.Errorf("parallel_request_timeout must be at most 300s, got %v", rc.ParallelRequestTimeout)
	}

	return nil
}