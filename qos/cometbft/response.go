package cometbft

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

var (
	// CometBFT response IDs for different request types:
	// - JSON-RPC success: 1
	// - REST success: -1
	// - Any error: 1
	jsonrpcSuccessID = jsonrpc.IDFromInt(1)
	restSuccessID    = jsonrpc.IDFromInt(-1)
	errorID          = jsonrpc.IDFromInt(1)
)

// getExpectedResponseID returns the expected ID for a CometBFT response depending
// on the request type (REST/JSON-RPC) and the response result (error/success).
func getExpectedResponseID(response jsonrpc.Response, isJSONRPC bool) jsonrpc.ID {
	if response.IsError() {
		return errorID
	}
	if isJSONRPC {
		return jsonrpcSuccessID
	}
	return restSuccessID
}

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
	isJSONRPC bool,
) (response, error) {
	// Unmarshal the raw response payload into a JSON-RPC response.
	var jsonrpcResponse jsonrpc.Response
	if err := json.Unmarshal(data, &jsonrpcResponse); err != nil {
		// The response raw payload (e.g. as received from an endpoint) could not be unmarshalled as a JSONRC response.
		// Return a generic response to the user.
		return getGenericJSONRPCErrResponse(logger, jsonrpcResponse, data, err), err
	}

	// Validate the JSON-RPC response.
	if err := jsonrpcResponse.Validate(getExpectedResponseID(jsonrpcResponse, isJSONRPC)); err != nil {
		return getGenericJSONRPCErrResponse(logger, jsonrpcResponse, data, err), err
	}

	// NOTE: We intentionally skip checking whether the JSON-RPC response indicates an error.
	// This allows the method-specific handler to determine how to respond to the user.

	// Unmarshal the JSON-RPC response into a method-specific response.
	unmarshaller, found := apiPathResponseMappings[apiPath]
	if found {
		return unmarshaller(logger, jsonrpcResponse)
	}

	// Default to a generic response if no method-specific response is found.
	return responseUnmarshallerGeneric(logger, jsonrpcResponse, data)
}
