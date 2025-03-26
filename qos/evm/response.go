package evm

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseUnmarshaller is the entrypoint function for any
// new supported response types.
// E.g. to handle "eth_getBalance" requests, the following need to be defined:
//  1. A new custom responseUnmarshaller
//  2. A new custom struct to handle the details of the particular response.
type responseUnmarshaller func(
	logger polylog.Logger,
	jsonrpcReq jsonrpc.Request,
	jsonrpcResp jsonrpc.Response,
) (response, error)

var (
	// All response types need to implement the response interface.
	// Any new response struct needs to be added to the following list.
	_ response = &responseToChainID{}
	_ response = &responseToBlockNumber{}
	_ response = &responseToArchival{}
	_ response = &responseGeneric{}

	methodResponseMappings = map[jsonrpc.Method]responseUnmarshaller{
		methodChainID:          responseUnmarshallerChainID,
		methodBlockNumber:      responseUnmarshallerBlockNumber,
		methodGetBlockByNumber: responseUnmarshallerArchival,
	}
)

// unmarshalResponse converts raw endpoint bytes into a JSONRPC response struct.
// As of PR #164, generates endpoint observations for responses to:
//   - eth_chainId
//   - eth_blockNumber
//   - any empty response, regardless of method
func unmarshalResponse(
	logger polylog.Logger,
	jsonrpcReq jsonrpc.Request,
	data []byte,
) (
	response, error,
) {
	// Create a specialized response for empty endpoint response.
	if len(data) == 0 {
		return responseEmpty{
			logger:     logger,
			jsonrpcReq: jsonrpcReq,
		}, nil
	}

	// Unmarshal the raw response payload into a JSONRPC response.
	var jsonrpcResponse jsonrpc.Response
	err := json.Unmarshal(data, &jsonrpcResponse)
	if err != nil {
		// The response raw payload (e.g. as received from an endpoint) could not be unmarshalled as a JSONRC response.
		// Return a generic response to the user.
		return getGenericJSONRPCErrResponse(logger, jsonrpcReq.ID, data, err), err
	}

	// Validate the JSONRPC response.
	if err := jsonrpcResponse.Validate(jsonrpcReq.ID); err != nil {
		return getGenericJSONRPCErrResponse(logger, jsonrpcReq.ID, data, err), err
	}

	// We intentionally skip checking whether the JSONRPC response indicates an error.
	// This allows the method-specific handler to determine how to respond to the user.

	// Unmarshal the JSONRPC response into a method-specific response.
	unmarshaller, found := methodResponseMappings[jsonrpcReq.Method]
	if found {
		return unmarshaller(logger, jsonrpcReq, jsonrpcResponse)
	}

	// Default to a generic response if no method-specific response is found.
	return responseUnmarshallerGeneric(logger, jsonrpcReq, data)
}
