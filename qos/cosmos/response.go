package cosmos

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

var (
	// CosmosSDK response IDs for different request types:
	// - REST-like responses: -1
	// TODO_NEXT(@adshmh): Use proper JSON-RPC ID response validation that works for all CosmosSDK chains.
	restLikeResponseID = jsonrpc.IDFromInt(-1)
)

// responseUnmarshaller is the entrypoint for processing new supported response types.
//
// To add support for a new endpoint (e.g. "/block"):
// 1. Define a new custom responseUnmarshaller
// 2. Create a corresponding struct to handle the response details
type responseUnmarshaller func(
	logger polylog.Logger,
	jsonrpcResp jsonrpc.Response,
	restResponse []byte,
) (response, error)

var (
	// All response types must implement the response interface.
	_ response = &responseToCometbftHealth{}
	_ response = &responseToCometbftStatus{}
	_ response = &responseToCosmosStatus{}
	_ response = &responseGeneric{}

	// Maps API paths to their corresponding response unmarshallers
	apiPathResponseMappings = map[string]responseUnmarshaller{
		apiPathHealthCheck:  responseUnmarshallerCometbftHealth,
		apiPathStatus:       responseUnmarshallerCometbftStatus,
		apiPathCosmosStatus: responseUnmarshallerCosmosStatus,
	}
)

// unmarshalResponse parses the supplied raw byte slice from an endpoint into either a JSON-RPC or REST response.
func unmarshalResponse(
	logger polylog.Logger,
	apiPath string,
	data []byte,
	isJSONRPC bool,
	endpointAddr protocol.EndpointAddr,
) (response, error) {
	// Try to unmarshal the raw response payload into a JSON-RPC response.
	var jsonrpcResponse jsonrpc.Response
	if err := json.Unmarshal(data, &jsonrpcResponse); err != nil {
		// The response raw payload could not be unmarshalled as a JSON-RPC response.
		// Treat it as a REST response and use the generic unmarshaller.
		return responseUnmarshallerGeneric(logger, jsonrpcResponse, data, isJSONRPC)
	}

	// Validate the JSON-RPC response.
	// TODO_NEXT(@adshmh): Use proper JSON-RPC ID response validation that works for all CosmosSDK chains.

	// NOTE: We intentionally skip checking whether the JSON-RPC response indicates an error.
	// This allows the method-specific handler to determine how to respond to the user.

	// Unmarshal the JSON-RPC response into a method-specific response.
	unmarshaller, found := apiPathResponseMappings[apiPath]
	if found {
		return unmarshaller(logger, jsonrpcResponse, data)
	}

	// Default to a generic response if no method-specific response is found.
	return responseUnmarshallerGeneric(logger, jsonrpcResponse, data, isJSONRPC)
}
