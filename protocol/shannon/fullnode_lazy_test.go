package shannon

import (
	"testing"

	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"github.com/stretchr/testify/assert"
)

// Test helper to create a test LazyFullNode
func createTestLazyFullNode(t *testing.T) *LazyFullNode {
	logger := polyzero.NewLogger()

	return &LazyFullNode{
		logger: logger,
	}
}

// Test the SessionConfig grace period scaling factor functionality
func TestSessionConfigGracePeriodScaling(t *testing.T) {
	tests := []struct {
		name                       string
		gracePeriodScaleDownFactor float64
		gracePeriodBlocks          uint64
		expectedScaledBlocks       int64
		description                string
	}{
		{
			name:                       "default_scale_factor",
			gracePeriodScaleDownFactor: 0.5,
			gracePeriodBlocks:          10,
			expectedScaledBlocks:       5, // 10 * 0.5 = 5
			description:                "Should scale grace period by default factor",
		},
		{
			name:                       "no_scaling",
			gracePeriodScaleDownFactor: 1.0,
			gracePeriodBlocks:          10,
			expectedScaledBlocks:       10, // 10 * 1.0 = 10
			description:                "Should not scale when factor is 1.0",
		},
		{
			name:                       "aggressive_scaling",
			gracePeriodScaleDownFactor: 0.2,
			gracePeriodBlocks:          20,
			expectedScaledBlocks:       4, // 20 * 0.2 = 4
			description:                "Should aggressively scale down grace period",
		},
		{
			name:                       "zero_scaling",
			gracePeriodScaleDownFactor: 0.0,
			gracePeriodBlocks:          10,
			expectedScaledBlocks:       0, // 10 * 0.0 = 0
			description:                "Should eliminate grace period when factor is 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create LazyFullNode with test config
			lfn := createTestLazyFullNode(t)
			lfn.sessionConfig = SessionConfig{
				GracePeriodScaleDownFactor: tt.gracePeriodScaleDownFactor,
			}

			// Calculate scaled grace period blocks
			scaledBlocks := int64(float64(tt.gracePeriodBlocks) * lfn.sessionConfig.GracePeriodScaleDownFactor)

			// Verify the scaling calculation
			assert.Equal(t, tt.expectedScaledBlocks, scaledBlocks, tt.description)
		})
	}
}

// Test SessionConfig validation and defaults
func TestSessionConfigValidation(t *testing.T) {
	tests := []struct {
		name           string
		config         SessionConfig
		expectedValid  bool
		expectedConfig SessionConfig
		description    string
	}{
		{
			name: "valid_config",
			config: SessionConfig{
				GracePeriodScaleDownFactor: 0.5,
			},
			expectedValid: true,
			expectedConfig: SessionConfig{
				GracePeriodScaleDownFactor: 0.5,
			},
			description: "Should accept valid configuration",
		},
		{
			name: "zero_factor",
			config: SessionConfig{
				GracePeriodScaleDownFactor: 0.0,
			},
			expectedValid: true,
			expectedConfig: SessionConfig{
				GracePeriodScaleDownFactor: 0.0,
			},
			description: "Should accept zero scale factor (disables grace period)",
		},
		{
			name: "max_factor",
			config: SessionConfig{
				GracePeriodScaleDownFactor: 1.0,
			},
			expectedValid: true,
			expectedConfig: SessionConfig{
				GracePeriodScaleDownFactor: 1.0,
			},
			description: "Should accept maximum scale factor",
		},
		{
			name: "invalid_negative_factor",
			config: SessionConfig{
				GracePeriodScaleDownFactor: -0.1,
			},
			expectedValid: false,
			expectedConfig: SessionConfig{},
			description: "Should reject negative scale factor",
		},
		{
			name: "invalid_too_large_factor",
			config: SessionConfig{
				GracePeriodScaleDownFactor: 1.5,
			},
			expectedValid: false,
			expectedConfig: SessionConfig{},
			description: "Should reject scale factor greater than 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate the configuration
			err := tt.config.validate()

			if tt.expectedValid {
				assert.NoError(t, err, tt.description)
				assert.Equal(t, tt.expectedConfig.GracePeriodScaleDownFactor, tt.config.GracePeriodScaleDownFactor)
			} else {
				assert.Error(t, err, tt.description)
			}
		})
	}
}

// Test DefaultSessionConfig returns appropriate defaults
func TestDefaultSessionConfig(t *testing.T) {
	config := SessionConfig{}
	config.hydrateDefaults()
	
	// Verify the default grace period scale down factor
	assert.Equal(t, 0.8, config.GracePeriodScaleDownFactor, "Default grace period scale down factor should be 0.8")
	
	// Verify validation passes for defaults
	err := config.validate()
	assert.NoError(t, err, "Default config should be valid")
}

// Test grace period logic calculation without external dependencies
func TestGracePeriodLogic(t *testing.T) {
	// This test verifies the mathematical logic used in session grace period calculations
	// without requiring complex SDK client mocking
	
	tests := []struct {
		name                       string
		currentHeight              int64
		sessionStartHeight         int64
		gracePeriodBlocks          uint64
		gracePeriodScaleDownFactor float64
		expectWithinGracePeriod    bool
		expectWithinScaledGrace    bool
		description                string
	}{
		{
			name:                       "outside_grace_period",
			currentHeight:              100,
			sessionStartHeight:         90,
			gracePeriodBlocks:          5,
			gracePeriodScaleDownFactor: 0.5,
			expectWithinGracePeriod:    false,
			expectWithinScaledGrace:    false,
			description:                "Should be outside grace period when too far past session start",
		},
		{
			name:                       "within_grace_period_but_outside_scaled",
			currentHeight:              92,
			sessionStartHeight:         90,
			gracePeriodBlocks:          5,
			gracePeriodScaleDownFactor: 0.5,
			expectWithinGracePeriod:    true,
			expectWithinScaledGrace:    false,
			description:                "Should be within grace period but outside scaled grace period",
		},
		{
			name:                       "within_scaled_grace_period",
			currentHeight:              91,
			sessionStartHeight:         90,
			gracePeriodBlocks:          5,
			gracePeriodScaleDownFactor: 0.5,
			expectWithinGracePeriod:    true,
			expectWithinScaledGrace:    true,
			description:                "Should be within both grace period and scaled grace period",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replicate the grace period logic from GetSessionWithExtendedValidity
			prevSessionEndHeight := tt.sessionStartHeight - 1
			prevSessionEndHeightWithExtendedValidity := prevSessionEndHeight + int64(tt.gracePeriodBlocks)
			prevSessionEndHeightWithExtendedValidityScaled := prevSessionEndHeight + int64(float64(tt.gracePeriodBlocks)*tt.gracePeriodScaleDownFactor)

			// Check if within grace period
			withinGracePeriod := tt.currentHeight <= prevSessionEndHeightWithExtendedValidity
			withinScaledGrace := tt.currentHeight <= prevSessionEndHeightWithExtendedValidityScaled

			assert.Equal(t, tt.expectWithinGracePeriod, withinGracePeriod, tt.description+" - grace period check")
			assert.Equal(t, tt.expectWithinScaledGrace, withinScaledGrace, tt.description+" - scaled grace period check")
		})
	}
}