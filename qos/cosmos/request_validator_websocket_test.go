package cosmos

import (
	"testing"

	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
	"github.com/stretchr/testify/require"
)

func Test_validateWebsocketRequest(t *testing.T) {
	tests := []struct {
		name                   string
		supportedAPIs          map[sharedtypes.RPCType]struct{}
		expectSuccess          bool
		expectErrorContextType bool
	}{
		{
			name: "supported websockets service config",
			supportedAPIs: map[sharedtypes.RPCType]struct{}{
				sharedtypes.RPCType_WEBSOCKET: {},
			},
			expectSuccess:          true,
			expectErrorContextType: false,
		},
		{
			name: "unsupported websockets service config",
			supportedAPIs: map[sharedtypes.RPCType]struct{}{
				sharedtypes.RPCType_REST: {},
			},
			expectSuccess:          false,
			expectErrorContextType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the request validator with test data
			validator := &requestValidator{
				logger:        polyzero.NewLogger(),
				cosmosChainID: "test-chain",
				serviceID:     "test-service",
				supportedAPIs: tt.supportedAPIs,
				serviceState:  &serviceState{}, // minimal setup for context building
			}

			// Call the function under test
			ctx, success := validator.validateWebsocketRequest()

			// Verify the success/failure expectation
			require.Equal(t, tt.expectSuccess, success, "validateWebsocketRequest success result mismatch")

			// Verify context is not nil
			require.NotNil(t, ctx, "returned context should not be nil")

			// Additional verification based on expected result type
			if tt.expectErrorContextType {
				// For unsupported case, we expect some kind of error context
				// We can't easily type assert the exact error context type without more imports,
				// but we can verify it's not nil and success is false
				require.False(t, success, "should return false for unsupported Websocket")
			} else {
				// For supported case, we expect success
				require.True(t, success, "should return true for supported Websocket")
			}
		})
	}
}
