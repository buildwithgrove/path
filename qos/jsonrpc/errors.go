package jsonrpc

import (
	"fmt"
)

const (
	ResponseCodeDefaultInternalErr = -32000 // JSON-RPC standard server error code; https://www.org/historical/json-rpc-2-0.html
	ResponseCodeDefaultBadRequest  = -32600 // JSON-RPC error code indicating bad user request
)

// NewErrResponseInternalErr creates a JSON-RPC error response when an internal error has occurred (e.g. reading HTTP request's body)
// Marks the error as retryable to allow clients to safely retry their request.
func NewErrResponseInternalErr(requestID ID, err error) Response {
	return GetErrorResponse(
		requestID,
		ResponseCodeDefaultInternalErr, // JSON-RPC standard server error code; https://www.org/historical/json-rpc-2-0.html
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
		requestID,                     // Use request's original ID if present
		ResponseCodeDefaultBadRequest, // -32600 code indicates an invalid user request.
		fmt.Sprintf("invalid request: %s", err.Error()), // Error Message
		map[string]string{
			"error": err.Error(),
			// Custom extension - not part of the official JSON-RPC spec
			// Indicates this error is permanent - the request must be corrected as retrying will not succeed
			"retryable": "false",
		},
	)
}

// NewErrResponseEmptyEndpointResponse creates a JSON-RPC error response for empty endpoint responses:
//   - Preserves original request ID
//   - Marks error as retryable for safe client retry
func NewErrResponseEmptyEndpointResponse(requestID ID) Response {
	return GetErrorResponse(
		requestID, // Use request's original ID if present
		-32000,    // JSON-RPC standard server error code; https://www.jsonrpc.org/historical/json-rpc-2-0.html
		"Endpoint (data/service node error): Received an empty response. The endpoint will be dropped from the selection pool. Please try again.", // Error Response Message
		map[string]string{
			// Custom extension - not part of the official JSON-RPC spec
			// Marks the error as retryable to allow clients to safely retry their request.
			"retryable": "true",
		},
	)
}

// NewErrResponseNoEndpointResponse creates a JSON-RPC error response for the case
// where no endpoint response was received at all.
// This response:
//   - Preserves the original request ID
//   - Marks error as retryable for safe client retry
//   - Provides actionable message for clients
func NewErrResponseNoEndpointResponse(requestID ID) Response {
	return GetErrorResponse(
		requestID, // Use request's original ID if present
		-32000,    // JSON-RPC standard server error code; https://www.jsonrpc.org/historical/json-rpc-2-0.html
		"Failed to receive any response from endpoints. This could be due to network issues or high load. Please try again.", // Error Response Message
		map[string]string{
			// Custom extension - not part of the official JSON-RPC spec
			// Marks the error as retryable to allow clients to safely retry their request.
			"retryable": "true",
		},
	)
}

// NewErrResponseBatchMarshalFailure creates a JSON-RPC error response for batch response marshaling failures.
// This occurs when individual responses are valid but combining them into a JSON array fails.
// Uses null ID per JSON-RPC spec for batch-level errors that cannot be correlated to specific requests.
func NewErrResponseBatchMarshalFailure(err error) Response {
	return GetErrorResponse(
		ID{},   // Use null ID for batch-level failures per JSON-RPC spec
		-32000, // JSON-RPC standard server error code; https://www.jsonrpc.org/historical/json-rpc-2-0.html
		fmt.Sprintf("Failed to marshal batch response: %s", err.Error()), // Error Message
		map[string]string{
			"error": err.Error(),
			// Custom extension - not part of the official JSON-RPC spec
			// Marks the error as retryable since this is an internal processing issue
			"retryable": "true",
		},
	)
}
