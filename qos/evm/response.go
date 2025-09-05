package evm

import (
	"encoding/json"
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/log"
	"github.com/buildwithgrove/path/protocol"
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
	_ response = &responseToGetBalance{}
	_ response = &responseGeneric{}

	methodResponseMappings = map[jsonrpc.Method]responseUnmarshaller{
		methodChainID:     responseUnmarshallerChainID,
		methodBlockNumber: responseUnmarshallerBlockNumber,
		methodGetBalance:  responseUnmarshallerGetBalance,
	}
)

// unmarshalResponse converts raw endpoint bytes into a JSONRPC response struct.
// As of PR #194, generates endpoint observations for responses to:
//   - eth_chainId
//   - eth_blockNumber
//   - eth_getBalance
//   - any empty response, regardless of method
func unmarshalResponse(
	logger polylog.Logger,
	jsonrpcReqs map[string]jsonrpc.Request,
	data []byte,
	endpointAddr protocol.EndpointAddr,
) (response, error) {
	// Create a specialized response for empty endpoint response.
	if len(data) == 0 {
		return responseEmpty{
			logger:      logger,
			jsonrpcReqs: jsonrpcReqs,
		}, nil
	}

	// Unmarshal the raw response payload into a JSONRPC response.
	var jsonrpcResponse jsonrpc.Response
	err := json.Unmarshal(data, &jsonrpcResponse)
	if err != nil {
		// The response raw payload (e.g. as received from an endpoint) could not be unmarshalled as a JSONRC response.
		// Return a generic response to the user.
		payloadStr := string(data)
		logger.With(
			"requests", jsonrpcReqs,
			"unmarshal_err", err,
			"raw_payload", log.Preview(payloadStr),
			"endpoint_addr", endpointAddr,
		).Error().Msg("❌ Failed to unmarshal response payload as JSON-RPC")

		return getGenericJSONRPCErrResponse(logger, getJsonRpcIDForErrorResponse(jsonrpcReqs), data, err), err
	}

	// Get the corresponding JSONRPC request for the response.
	jsonrpcReq, ok := jsonrpcReqs[jsonrpcResponse.ID.String()]
	if !ok {
		logger.Error().Msg("SHOULD NEVER HAPPEN: JSON-RPC ID not found in the response")
		err := fmt.Errorf("JSON-RPC ID not found in the response")
		return getGenericJSONRPCErrResponse(logger, getJsonRpcIDForErrorResponse(jsonrpcReqs), data, err), err
	}

	// Validate the JSONRPC response.
	if err := jsonrpcResponse.Validate(jsonrpcReq.ID); err != nil {
		payloadStr := string(data)
		logger.With(
			"request", jsonrpcReq,
			"validation_err", err,
			"raw_payload", log.Preview(payloadStr),
			"endpoint_addr", endpointAddr,
		).Debug().Msg("JSON-RPC response validation failed")
		return getGenericJSONRPCErrResponse(logger, jsonrpcReq.ID, data, err), err
	}

	// Unmarshal the JSONRPC response into a method-specific response.
	unmarshaller, found := methodResponseMappings[jsonrpcReq.Method]
	if found {
		return unmarshaller(logger, jsonrpcReq, jsonrpcResponse)
	}

	// Default to a generic response if no method-specific response is found.
	// Pass the already unmarshaled jsonrpcResponse to avoid double unmarshaling.
	return responseUnmarshallerGenericFromResponse(logger, jsonrpcReq, jsonrpcResponse)
}

// getJsonRpcIDForErrorResponse determines the appropriate ID to use in error responses when no endpoint response was received.
// Follows JSON-RPC 2.0 specification guidelines for ID handling in error scenarios:
//
// Single request (len == 1):
//   - Returns the original request's ID to maintain proper request-response correlation
//   - Allows client to match the error response back to the specific request that failed
//
// Batch request or no requests (len != 1):
//   - Returns null ID (empty jsonrpc.ID{}) per JSON-RPC spec requirement
//   - Per spec: "If there was an error in detecting the id in the Request object, it MUST be Null"
//   - For batch requests, no single ID represents the entire failed batch
//   - For zero requests, no valid ID exists to return
//
// This approach ensures specification compliance and clear error semantics for clients.
// Reference: https://www.jsonrpc.org/specification#response_object
func getJsonRpcIDForErrorResponse(jsonrpcReqs map[string]jsonrpc.Request) jsonrpc.ID {
	if len(jsonrpcReqs) == 1 {
		for id := range jsonrpcReqs {
			return jsonrpc.IDFromStr(id)
		}
	}
	return jsonrpc.ID{}
}
