package relay

import (
	"fmt"
	"time"
)

// Config contains configuration for relay request handling
type Config struct {
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

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		MaxParallelRequests:     4,
		ParallelRequestTimeout:  30 * time.Second,
		EnableEndpointDiversity: true,
	}
}

// ValidateAndHydrate validates and applies defaults to relay configuration
func (c *Config) ValidateAndHydrate() error {
	// Validate MaxParallelRequests
	if c.MaxParallelRequests == 0 {
		c.MaxParallelRequests = 4 // Default
	}
	if c.MaxParallelRequests < 1 {
		return fmt.Errorf("max_parallel_requests must be at least 1, got %d", c.MaxParallelRequests)
	}
	if c.MaxParallelRequests > 10 {
		return fmt.Errorf("max_parallel_requests must be at most 10, got %d", c.MaxParallelRequests)
	}

	// Validate ParallelRequestTimeout
	if c.ParallelRequestTimeout == 0 {
		c.ParallelRequestTimeout = 30 * time.Second // Default
	}
	if c.ParallelRequestTimeout < time.Second {
		return fmt.Errorf("parallel_request_timeout must be at least 1s, got %v", c.ParallelRequestTimeout)
	}
	if c.ParallelRequestTimeout > 300*time.Second {
		return fmt.Errorf("parallel_request_timeout must be at most 300s, got %v", c.ParallelRequestTimeout)
	}

	return nil
}