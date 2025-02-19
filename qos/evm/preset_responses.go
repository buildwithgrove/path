package evm

import (
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// errResponseEmptyEndpointResponse is a pre-defined JSON-RPC response structure used when an EVM endpoint 
// returns an empty response body. This indicates:
// - Server misbehavior 
// - Triggers endpoint removal from selection pool
// - Allows client retry with different endpoint
var errResponseEmptyEndpointResponse = jsonrpc.GetErrorResponse(
	jsonrpc.ID{}, // Use request's original ID if present
	-32000,       // JSON-RPC standard server error code; https://www.jsonrpc.org/historical/json-rpc-2-0.html
	"Endpoint (data/service node error): Received an empty response. The endpoint will be dropped from the selection pool. Please try again.", // Error Response Message
	map[string]string{
		// Marks the error as retryable to allow clients to safely retry their request.
		"retryable": "true",
	},
)

// NewEmptyResponse creates a JSON-RPC error response for empty endpoint responses:
// - Preserves original request ID
// - Marks error as retryable for safe client retry
func NewEmptyResponseError(requestID jsonrpc.ID) jsonrpc.Response {
	response := errResponseEmptyEndpointResponse
	response.ID = requestID
	return response
}
