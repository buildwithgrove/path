package cosmos

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseUnmarshaller is the entrypoint for processing new supported response types.
//
// To add support for a new endpoint (e.g. "/block"):
// 1. Define a new custom responseUnmarshaller
// 2. Create a corresponding struct to handle the response details
type responseUnmarshaller func(
	logger polylog.Logger,
	jsonrpcResp jsonrpc.Response,
) (response, error)

var (
	// All response types must implement the response interface.
	_ response = &responseToHealth{}
	_ response = &responseToStatus{}
	_ response = &responseGeneric{}

	// Maps API paths to their corresponding response unmarshallers
	apiPathResponseMappings = map[string]responseUnmarshaller{
		apiPathHealthCheck: responseUnmarshallerHealth,
		apiPathStatus:      responseUnmarshallerStatus,
	}
)

// unmarshalResponse parses the supplied raw byte slice from an endpoint into a JSON-RPC response.
func unmarshalResponse(
	logger polylog.Logger,
	apiPath string,
	data []byte,
) (response, error) {
	// Unmarshal the raw response payload into a JSON-RPC response.
	var jsonrpcResponse jsonrpc.Response
	if err := json.Unmarshal(data, &jsonrpcResponse); err != nil {
		// The response raw payload (e.g. as received from an endpoint) could not be unmarshalled as a JSON-RPC response.
		// Return a generic response to the user.
		return getGenericJSONRPCErrResponse(logger, jsonrpcResponse, data, err), err
	}

	// NOTE: We intentionally skip validating the JSON-RPC response ID here because
	// CometBFT endpoints may use different ID conventions.
	// This allows the method-specific handler to determine how to respond to the user.

	// Unmarshal the JSON-RPC response into a method-specific response.
	unmarshaller, found := apiPathResponseMappings[apiPath]
	if found {
		return unmarshaller(logger, jsonrpcResponse)
	}

	// Default to a generic response if no method-specific response is found.
	return responseUnmarshallerGeneric(logger, jsonrpcResponse, data)
}
