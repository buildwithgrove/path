package evm

import (
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// errResponseEmptyEndpointResponse is a pre-defined JSON-RPC response structure to be used (i.e. returned)
// for all EVM JSON-RPC requests when the endpoint returned an empty response body.
// This indicates server misbehavior, triggers removal of the endpoint from the selection pool,
// and tells the client they can retry with a different endpoint.
var errResponseEmptyEndpointResponse = jsonrpc.GetErrorResponse(
	jsonrpc.ID{}, // Use request's original ID if present
	-32000,       // JSON-RPC standard server error code; https://www.jsonrpc.org/historical/json-rpc-2-0.html
	"Endpoint (data/service node error): Received an empty response. The endpoint will be dropped from the selection pool. Please try again.", // Error Response Message
	map[string]string{
		"retryable": "true",
	},
)

// NewEmptyResponse creates a JSON-RPC error response when an endpoint returns no data.
// Preserves the request ID from the original request.
// Marks the error as retryable to allow clients to safely retry their request.
func NewEmptyResponseError(requestID jsonrpc.ID) jsonrpc.Response {
	response := errResponseEmptyEndpointResponse
	response.ID = requestID
	return response
}
