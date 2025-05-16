package jsonrpc

import (
	"fmt"
)

const (
	ResponseCodeDefaultInternalErr = -32000 // JSON-RPC standard server error code; https://www.jsonrpc.org/historical/json-rpc-2-0.html
)

// NewErrResponseInternalErr creates a JSON-RPC error response when an internal error has occurred (e.g. reading HTTP request's body)
// Marks the error as retryable to allow clients to safely retry their request.
func NewErrResponseInternalErr(requestID ID, err error) Response {
	return GetErrorResponse(
		requestID,
		ResponseCodeDefaultInternalErr, // JSON-RPC standard server error code; https://www.jsonrpc.org/historical/json-rpc-2-0.html
		fmt.Sprintf("internal error: %s", err.Error()), // Error Message
		map[string]string{
			"error": err.Error(),
			// Custom extension - not part of the official JSON-RPC spec
			// Marks the error as retryable to allow clients to safely retry their request
			"retryable": "true",
		},
	)
}

// NewErrResponseInvalidRequest returns a JSON-RPC error response for malformed or invalid requests.
// The error indicates the request cannot be processed due to issues like:
//   - Failed JSON-RPC deserialization
//   - Missing required JSON-RPC fields (e.g. `method`)
//   - Unsupported JSON-RPC method
//
// If the request contains a valid JSON-RPC ID, it is included in the error response.
// The error is marked as permanent since retrying without correcting the request will fail.
func NewErrResponseInvalidRequest(requestID ID, err error) Response {
	return GetErrorResponse(
		requestID,                      // Use request's original ID if present
		ResponseCodeDefaultInternalErr, // JSON-RPC standard server error code; https://www.jsonrpc.org/historical/json-rpc-2-0.html
		fmt.Sprintf("invalid request: %s", err.Error()), // Error Message
		map[string]string{
			"error": err.Error(),
			// Custom extension - not part of the official JSON-RPC spec
			// Indicates this error is permanent - the request must be corrected as retrying will not succeed
			"retryable": "false",
		},
	)
}
