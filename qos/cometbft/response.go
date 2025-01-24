package cometbft

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

var (
	// CometBFT always returns a `-1` ID for successful responses.
	cometBFTSuccessResponseID = jsonrpc.IDFromInt(-1)
	// CometBFT always returns a `1` ID for error responses.
	cometBFTErrResponseID = jsonrpc.IDFromInt(1)
)

func getExpectedResponseID(response jsonrpc.Response) jsonrpc.ID {
	if response.IsError() {
		return cometBFTErrResponseID
	}
	return cometBFTSuccessResponseID
}

// responseUnmarshaller is the entrypoint function for any
// new supported response types.
// E.g. to handle "/block" requests, the following need to be defined:
//  1. A new custom responseUnmarshaller
//  2. A new custom struct  to handle the details of the particular response.
type responseUnmarshaller func(
	logger polylog.Logger,
	jsonrpcResp jsonrpc.Response,
) (response, error)

var (
	// All response types needs to implement the response interface.
	// Any new response struct needs to be added to the following list.
	_ response = &responseToHealth{}
	_ response = &responseGeneric{}

	routeResponseMappings = map[string]responseUnmarshaller{
		routeHealthCheck: responseUnmarshallerHealth,
		routeBlockHeight: responseUnmarshallerBlockHeight,
	}
)

// unmarshalResponse parses the supplied raw byte slice, received from an endpoint, into a JSONRPC response.
// Responses to the following JSONRPC methods are processed into endpoint observations:
//   - eth_blockNumber
func unmarshalResponse(logger polylog.Logger, route string, data []byte) (response, error) {
	// Unmarshal the raw response payload into a JSONRPC response.
	var jsonrpcResponse jsonrpc.Response
	err := json.Unmarshal(data, &jsonrpcResponse)
	if err != nil {
		// The response raw payload (e.g. as received from an endpoint) could not be unmarshalled as a JSONRC response.
		// Return a generic response to the user.
		return getGenericJSONRPCErrResponse(logger, jsonrpcResponse, data, err), err
	}

	// Validate the JSONRPC response.
	if err := jsonrpcResponse.Validate(getExpectedResponseID(jsonrpcResponse)); err != nil {
		return getGenericJSONRPCErrResponse(logger, jsonrpcResponse, data, err), err
	}

	// We intentionally skip checking whether the JSONRPC response indicates an error.
	// This allows the method-specific handler to determine how to respond to the user.

	// Unmarshal the JSONRPC response into a method-specific response.
	unmarshaller, found := routeResponseMappings[route]
	if found {
		return unmarshaller(logger, jsonrpcResponse)
	}

	// Default to a generic response if no method-specific response is found.
	return responseUnmarshallerGeneric(logger, jsonrpcResponse, data)
}
