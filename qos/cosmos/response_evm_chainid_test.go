package cosmos

import (
	"encoding/json"
	"testing"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/stretchr/testify/require"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

func TestResponseValidatorEVMChainID(t *testing.T) {
	logger := polylog.NewNopLogger()

	tests := []struct {
		name               string
		jsonrpcResponse    jsonrpc.Response
		expectedChainID    string
		expectValidationOK bool
	}{
		{
			name: "valid eth_chainId response - mainnet",
			jsonrpcResponse: jsonrpc.Response{
				JSONRPC: jsonrpc.Version2,
				ID:      jsonrpc.IDFromInt(1),
				Result:  json.RawMessage(`"0x1"`),
			},
			expectedChainID:    "0x1",
			expectValidationOK: true,
		},
		{
			name: "valid eth_chainId response - custom chain",
			jsonrpcResponse: jsonrpc.Response{
				JSONRPC: jsonrpc.Version2,
				ID:      jsonrpc.IDFromInt(1),
				Result:  json.RawMessage(`"0x15f900"`),
			},
			expectedChainID:    "0x15f900",
			expectValidationOK: true,
		},
		{
			name: "error response from endpoint",
			jsonrpcResponse: jsonrpc.Response{
				JSONRPC: jsonrpc.Version2,
				ID:      jsonrpc.IDFromInt(1),
				Error: &jsonrpc.Error{
					Code:    -32601,
					Message: "Method not found",
				},
			},
			expectedChainID:    "",
			expectValidationOK: false,
		},
		{
			name: "invalid result format - not a string",
			jsonrpcResponse: jsonrpc.Response{
				JSONRPC: jsonrpc.Version2,
				ID:      jsonrpc.IDFromInt(1),
				Result:  json.RawMessage(`123`),
			},
			expectedChainID:    "",
			expectValidationOK: false,
		},
		{
			name: "empty result",
			jsonrpcResponse: jsonrpc.Response{
				JSONRPC: jsonrpc.Version2,
				ID:      jsonrpc.IDFromInt(1),
				Result:  json.RawMessage(`""`),
			},
			expectedChainID:    "",
			expectValidationOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the validator
			resp := responseValidatorEVMChainID(logger, tt.jsonrpcResponse)

			// Type assert to get the concrete type
			evmResp, ok := resp.(*responseEVMChainID)
			require.True(t, ok, "expected responseEVMChainID type")

			// Check the chain ID
			require.Equal(t, tt.expectedChainID, evmResp.evmChainID)

			// Check the observation
			obs := evmResp.GetObservation()
			require.NotNil(t, obs.EndpointResponseValidationResult)
			require.Equal(t, qosobservations.CosmosResponseValidationType_COSMOS_RESPONSE_VALIDATION_TYPE_JSONRPC, 
				obs.EndpointResponseValidationResult.ResponseValidationType)

			// Check if parsed response is present
			if tt.expectValidationOK {
				parsedResp, ok := obs.EndpointResponseValidationResult.ParsedResponse.(*qosobservations.CosmosEndpointResponseValidationResult_ResponseEvmJsonrpcChainId)
				require.True(t, ok, "expected ResponseEvmJsonrpcChainId type")
				require.NotNil(t, parsedResp.ResponseEvmJsonrpcChainId)
				require.Equal(t, tt.expectedChainID, parsedResp.ResponseEvmJsonrpcChainId.EvmChainId)
			}
		})
	}
}

func TestResponseEVMChainID_GetHTTPResponse(t *testing.T) {
	logger := polylog.NewNopLogger()

	tests := []struct {
		name               string
		jsonrpcResponse    jsonrpc.Response
		expectedStatusCode int
	}{
		{
			name: "successful response",
			jsonrpcResponse: jsonrpc.Response{
				JSONRPC: jsonrpc.Version2,
				ID:      jsonrpc.IDFromInt(1),
				Result:  json.RawMessage(`"0x1"`),
			},
			expectedStatusCode: 200,
		},
		{
			name: "error response - method not found",
			jsonrpcResponse: jsonrpc.Response{
				JSONRPC: jsonrpc.Version2,
				ID:      jsonrpc.IDFromInt(1),
				Error: &jsonrpc.Error{
					Code:    -32601,
					Message: "Method not found",
				},
			},
			expectedStatusCode: 404,
		},
		{
			name: "error response - invalid params",
			jsonrpcResponse: jsonrpc.Response{
				JSONRPC: jsonrpc.Version2,
				ID:      jsonrpc.IDFromInt(1),
				Error: &jsonrpc.Error{
					Code:    -32602,
					Message: "Invalid params",
				},
			},
			expectedStatusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := responseEVMChainID{
				logger:          logger,
				jsonRPCResponse: tt.jsonrpcResponse,
			}

			httpResp := resp.GetHTTPResponse()
			require.Equal(t, tt.expectedStatusCode, httpResp.StatusCode)
			require.Equal(t, tt.expectedStatusCode, resp.getHTTPStatusCode())
		})
	}
}