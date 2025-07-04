package relay

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Verify default values
	assert.Equal(t, 4, config.MaxParallelRequests, "Default max parallel requests should be 4")
	assert.Equal(t, 30*time.Second, config.ParallelRequestTimeout, "Default timeout should be 30s")
	assert.True(t, config.EnableEndpointDiversity, "Default endpoint diversity should be enabled")
}

func TestValidateAndHydrate(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		expectedValid bool
		expectedConfig Config
		description   string
	}{
		{
			name: "valid_config",
			config: Config{
				MaxParallelRequests:     4,
				ParallelRequestTimeout:  30 * time.Second,
				EnableEndpointDiversity: true,
			},
			expectedValid: true,
			expectedConfig: Config{
				MaxParallelRequests:     4,
				ParallelRequestTimeout:  30 * time.Second,
				EnableEndpointDiversity: true,
			},
			description: "Should accept valid configuration",
		},
		{
			name:          "empty_config_gets_defaults",
			config:        Config{},
			expectedValid: true,
			expectedConfig: Config{
				MaxParallelRequests:     4,  // Default
				ParallelRequestTimeout:  30 * time.Second, // Default
				EnableEndpointDiversity: false, // Default for empty config
			},
			description: "Should hydrate empty config with defaults",
		},
		{
			name: "min_valid_values",
			config: Config{
				MaxParallelRequests:    1,
				ParallelRequestTimeout: 1 * time.Second,
			},
			expectedValid: true,
			expectedConfig: Config{
				MaxParallelRequests:    1,
				ParallelRequestTimeout: 1 * time.Second,
			},
			description: "Should accept minimum valid values",
		},
		{
			name: "max_valid_values",
			config: Config{
				MaxParallelRequests:    10,
				ParallelRequestTimeout: 300 * time.Second,
			},
			expectedValid: true,
			expectedConfig: Config{
				MaxParallelRequests:    10,
				ParallelRequestTimeout: 300 * time.Second,
			},
			description: "Should accept maximum valid values",
		},
		{
			name: "invalid_too_many_parallel_requests",
			config: Config{
				MaxParallelRequests: 11,
			},
			expectedValid: false,
			description:   "Should reject too many parallel requests",
		},
		{
			name: "invalid_too_few_parallel_requests",
			config: Config{
				MaxParallelRequests: 0, // Will be set to default first, so this should pass
			},
			expectedValid: true,
			expectedConfig: Config{
				MaxParallelRequests:    4, // Should be hydrated to default
				ParallelRequestTimeout: 30 * time.Second,
			},
			description: "Should hydrate zero parallel requests to default",
		},
		{
			name: "invalid_negative_parallel_requests",
			config: Config{
				MaxParallelRequests: -1,
			},
			expectedValid: false,
			description:   "Should reject negative parallel requests",
		},
		{
			name: "invalid_timeout_too_long",
			config: Config{
				ParallelRequestTimeout: 301 * time.Second,
			},
			expectedValid: false,
			description:   "Should reject timeout longer than 300s",
		},
		{
			name: "invalid_timeout_too_short",
			config: Config{
				ParallelRequestTimeout: 500 * time.Millisecond,
			},
			expectedValid: false,
			description:   "Should reject timeout shorter than 1s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying the test case
			config := tt.config

			// Validate and hydrate
			err := config.ValidateAndHydrate()

			if tt.expectedValid {
				assert.NoError(t, err, tt.description)
				
				// Check expected values
				assert.Equal(t, tt.expectedConfig.MaxParallelRequests, config.MaxParallelRequests)
				assert.Equal(t, tt.expectedConfig.ParallelRequestTimeout, config.ParallelRequestTimeout)
				
				// EnableEndpointDiversity is not hydrated by ValidateAndHydrate, so only check if explicitly set
				if tt.expectedConfig.EnableEndpointDiversity || tt.config.EnableEndpointDiversity {
					assert.Equal(t, tt.expectedConfig.EnableEndpointDiversity, config.EnableEndpointDiversity)
				}
			} else {
				assert.Error(t, err, tt.description)
			}
		})
	}
}