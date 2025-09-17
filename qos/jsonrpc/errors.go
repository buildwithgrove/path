package jsonrpc

import (
	"fmt"
)

const (
	// DEV_NOTE: Intentionallly using Non-reserved JSONRPC error code for internal errors.
	// Allows distinguishing errors coming from PATH from backend server errors.
	// e.g.
	// -32000: indicates a backend server (ETH, Solana, etc.) returned the error.
	// -31001: indicates PATH consturcted the JSONRPC error response (e.g. if the endpoint payload failed to parse as a JSONRPC response)
	ResponseCodeDefaultInternalErr = -31001 // JSON-RPC server error code; https://www.jsonrpc.org/specification#error_object
	ResponseCodeBackendServerErr   = -31002 // Indicates a backend server error: e.g. payload could not be parsed into a valid JSON-RPC response.

	ResponseCodeDefaultBadRequest = -32600 // JSON-RPC error code indicating bad user request
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
		requestID,                    // Use request's original ID if present
		ResponseCodeBackendServerErr, // Indicates a backend server error caught by PATH.
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
		requestID,                      // Use request's original ID if present
		ResponseCodeDefaultInternalErr, // Used to indicate an internal PATH/protocol error (e.g. endpoint timed out).
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
		ID{},                           // Use null ID for batch-level failures per JSON-RPC spec
		ResponseCodeDefaultInternalErr, // Used to indicate an internal PATH/protocol error: here it indicates failure to marshal a batch JSON-RPC response.
		fmt.Sprintf("Failed to marshal batch response: %s", err.Error()), // Error Message
		map[string]string{
			"error": err.Error(),
			// Custom extension - not part of the official JSON-RPC spec
			// Marks the error as retryable since this is an internal processing issue
			"retryable": "true",
		},
	)
}
