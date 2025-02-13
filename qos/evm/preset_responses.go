package evm

import (
	"fmt"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// errResponseEmptyEndpointResponse represents a JSON-RPC error response for when
// an endpoint returns an empty response body.
// This indicates server misbehavior, triggers removal of the endpoint from the selection pool,
// and tells the client they can retry with a different endpoint.
var errResponseEmptyEndpointResponse = jsonrpc.GetErrorResponse(
	nil,    // Use request's original ID if present
	-32000, // JSON-RPC server error code
	"Server error: Received an empty response. Server will be dropped from the selection pool. Please try again.", // Error Response Message
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
