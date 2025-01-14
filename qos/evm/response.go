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
//  2. A new custom struct  to handle the details of the particular response.
type responseUnmarshaller func(jsonrpcReq jsonrpc.Request, jsonrpcResp jsonrpc.Response, logger polylog.Logger) (response, error)

var (
	// All response types needs to implement the response interface.
	// Any new response struct needs to be added to the following list.
	_ response = &responseToChainID{}
	_ response = &responseToBlockNumber{}
	_ response = &responseGeneric{}

	methodResponseMappings = map[jsonrpc.Method]responseUnmarshaller{
		methodChainID:     responseUnmarshallerChainID,
		methodBlockNumber: responseUnmarshallerBlockNumber,
	}
)

// unmarshalResponse parses the supplied raw byte slice, received from an endpoint, into a JSONRPC response.
// As of PR #72, responses to the following JSONRPC methods are processed into endpoint observations:
//   - eth_chainId
//   - eth_blockNumber
func unmarshalResponse(jsonrpcReq jsonrpc.Request, data []byte, logger polylog.Logger) (response, error) {
	var jsonrpcResponse jsonrpc.Response
	err := json.Unmarshal(data, &jsonrpcResponse)
	if err != nil {
		// The response raw payload (e.g. as received from an endpoint) could not be unmarshalled as a JSONRC response.
		// Return a generic response to the user.
		return getGenericJSONRPCErrResponse(jsonrpcReq.ID, data, err, logger), err
	}

	if err := jsonrpcResponse.Validate(jsonrpcReq.ID); err != nil {
		return getGenericJSONRPCErrResponse(jsonrpcReq.ID, data, err, logger), err
	}

	// Note: we intentionally skip checking whether the JSONRPC response indicates an error. This allows the method-specific handler
	// to determine how to respond to the user.
	unmarshaller, found := methodResponseMappings[jsonrpcReq.Method]
	if found {
		return unmarshaller(jsonrpcReq, jsonrpcResponse, logger)
	}

	return responseUnmarshallerGeneric(jsonrpcReq, data, logger)
}
