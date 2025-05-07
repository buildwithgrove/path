package judge

import (
	"encoding/json"
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_IN_THIS_PR: verify returned JSONRPC error codes.

const (
	// Standard JSONRPC 2.0 error codes
	ErrorCodeParseError     int64 = -32700
	ErrorCodeInvalidRequest int64 = -32600
	ErrorCodeMethodNotFound int64 = -32601
	ErrorCodeInvalidParams  int64 = -32602
	ErrorCodeInternalError  int64 = -32603

	// Server error codes (reserved from -32000 to -32099)
	ErrorCodeServerError int64 = -32000
)

// newErrResponseEmptyEndpointResponse creates a JSON-RPC error response for empty endpoint responses:
// - Preserves original request ID
// - Marks error as retryable for safe client retry
func newErrResponseEmptyEndpointResponse(requestID jsonrpc.ID) *jsonrpc.Response {
	jsonrpcResp := jsonrpc.GetErrorResponse(
		requestID, // Use request's original ID if present
		-32000,    // JSON-RPC standard server error code; https://www.jsonrpc.org/historical/json-rpc-2-0.html
		"Endpoint (data/service node error): Received an empty response. The endpoint will be dropped from the selection pool. Please try again.", // Error Response Message
		map[string]string{
			// Custom extension - not part of the official JSON-RPC spec
			// Marks the error as retryable to allow clients to safely retry their request.
			"retryable": "true",
		},
	)

	return &jsonrpcResp
}

// newErrResponseParseError creates a JSON-RPC error response for parse errors.
// This response:
// - Preserves the original request ID
// - Marks error as retryable
// - Indicates the endpoint response couldn't be parsed
func newErrResponseParseError(requestID jsonrpc.ID, parseErr error) *jsonrpc.Response {
	jsonrpcResp := jsonrpc.GetErrorResponse(
		requestID,
		ErrorCodeParseError,
		"Failed to parse endpoint response",
		map[string]string{
			"error":     parseErr.Error(),
			"retryable": "true",
		},
	)
	return &jsonrpcResp
}

// newErrResponseNoEndpointResponse creates a JSON-RPC error response for the case
// where no endpoint response was received at all.
// This response:
// - Preserves the original request ID
// - Marks error as retryable for safe client retry
// - Provides actionable message for clients
func newErrResponseNoEndpointResponse(requestID jsonrpc.ID) *jsonrpc.Response {
	jsonrpcResp := jsonrpc.GetErrorResponse(
		requestID, // Use request's original ID if present
		-32000,    // JSON-RPC standard server error code; https://www.jsonrpc.org/historical/json-rpc-2-0.html
		"Failed to receive any response from endpoints. This could be due to network issues or high load. Please try again.", // Error Response Message
		map[string]string{
			// Custom extension - not part of the official JSON-RPC spec
			// Marks the error as retryable to allow clients to safely retry their request.
			"retryable": "true",
		},
	)

	return &jsonrpcResp
}

func newJSONRPCErrResponseInternalProtocolError(requestID jsonrpc.ID) *jsonrpc.Response {
	jsonrpcResp := jsonrpc.GetErrorResponse(
		requestID,
		ErrorCodeInternalError,
		"internal error: protocol-level error has occurred", // Error Message
		map[string]string{
			"error_type": "protocol",
			// Custom extension - not part of the official JSON-RPC spec
			// Marks the error as retryable to allow clients to safely retry their request
			"retryable": "true",
		},
	)

	return &jsonrpcResp
}

// newJSONRPCErrResponseInternalReadError creates a JSON-RPC error response for HTTP request read errors.
// This response:
// - Uses an empty ID since the request couldn't be read
// - Marks error as retryable since it's likely a server issue
// - Provides the specific read error message
func newJSONRPCErrResponseInternalReadError(readErr error) *jsonrpc.Response {
	jsonrpcResp := jsonrpc.GetErrorResponse(
		jsonrpc.ID{}, // No ID for read errors
		ErrorCodeInternalError,
		"Internal server error: failed to read request",
		map[string]string{
			"error":     readErr.Error(),
			"retryable": "true",
		},
	)

	return &jsonrpcResp
}

func newJSONRPCErrResponseJSONRPCRequestValidationError(requestID jsonrpc.ID, validationErr error) *jsonrpc.Response {
	jsonrpcResp := jsonrpc.GetErrorResponse(
		requestID,
		-32000, // JSON-RPC standard server error code; https://www.jsonrpc.org/historical/json-rpc-2-0.html
		fmt.Sprintf("invalid request: %s", validationErr.Error()), // Error Message
		map[string]string{
			"error": validationErr.Error(),
			// Custom extension - not part of the official JSON-RPC spec
			// Indicates this error is permanent - the request must be corrected as retrying will not succeed
			"retryable": "false",
		},
	)

	return &jsonrpcResp
}

// newErrResponseMarshalError creates a JSON-RPC error response for marshaling errors.
// This response:
// - Preserves the original request ID if available
// - Marks error as retryable
// - Indicates the response couldn't be serialized
func newJSONRPCErrResponseMarshalError(requestID jsonrpc.ID, marshalErr error) jsonrpc.Response {
	return jsonrpc.GetErrorResponse(
		requestID,
		ErrorCodeInternalError,
		fmt.Sprintf("Failed to marshal response: %s", marshalErr.Error()),
		map[string]string{
			"retryable": "true",
		},
	)
}

// The error indicates the request cannot be processed due to issues like:
//   - Failed JSON-RPC deserialization
//   - Missing required JSON-RPC fields (e.g. `method`)
//   - Unsupported JSON-RPC method
//
// If the request contains a valid JSON-RPC ID, it is included in the error response.
// The error is marked as permanent since retrying without correcting the request will fail.
func newErrResponseInvalidRequest(err error, requestID jsonrpc.ID) jsonrpc.Response {
	return jsonrpc.GetErrorResponse(
		requestID, // Use request's original ID if present
		-32000,    // JSON-RPC standard server error code; https://www.jsonrpc.org/historical/json-rpc-2-0.html
		fmt.Sprintf("invalid request: %s", err.Error()), // Error Message
		map[string]string{
			"error": err.Error(),
			// Custom extension - not part of the official JSON-RPC spec
			// Indicates this error is permanent - the request must be corrected as retrying will not succeed
			"retryable": "false",
		},
	)
}

// newErrResponseInvalidVersionError creates a JSON-RPC error response for invalid version errors.
// This response:
// - Preserves the original request ID
// - Marks error as non-retryable since it's a client issue
// - Indicates the JSONRPC version is invalid
func newErrResponseInvalidVersionError(requestID jsonrpc.ID) jsonrpc.Response {
	return jsonrpc.GetErrorResponse(
		requestID,
		ErrorCodeInvalidRequest,
		"Invalid JSON-RPC version, expected '2.0'",
		map[string]string{
			"retryable": "false",
		},
	)
}

// newErrResponseMissingMethodError creates a JSON-RPC error response for missing method errors.
// This response:
// - Preserves the original request ID
// - Marks error as non-retryable since it's a client issue
// - Indicates that the method field is required
func newErrResponseMissingMethodError(requestID jsonrpc.ID) jsonrpc.Response {
	return jsonrpc.GetErrorResponse(
		requestID,
		ErrorCodeInvalidRequest,
		"Method is required",
		map[string]string{
			"retryable": "false",
		},
	)
}

// newJSONRPCErrResponseParseRequestError creates a JSON-RPC error response for parse errors.
// This response:
// - Uses an empty ID since we couldn't parse the request to get an ID
// - Marks error as non-retryable since it's likely a client issue with the JSONRPC format
// - Indicates the request couldn't be parsed
func newJSONRPCErrResponseParseRequestError(parseErr error) *jsonrpc.Response {
	jsonrpcResp := jsonrpc.GetErrorResponse(
		jsonrpc.ID{}, // No ID for parse errors
		ErrorCodeParseError,
		"Failed to parse JSON-RPC request",
		map[string]string{
			"error":     parseErr.Error(),
			"retryable": "false",
		},
	)

	return &jsonrpcResp
}

// marshalErrorResponse marshals a JSONRPC error response to JSON.
// This handles the serialization of the error response to bytes.
func marshalErrorResponse(
	logger polylog.Logger,
	response jsonrpc.Response,
) ([]byte, error) {
	payload, err := json.Marshal(response)
	if err != nil {
		// Create a simple fallback error response as raw JSON
		fallback := fmt.Sprintf(`{"jsonrpc":"2.0","id":"%v","error":{"code":%d,"message":"%s"}}`,
			response.ID, response.Error.Code, response.Error.Message)
		return []byte(fallback), nil
	}
	return payload, nil
}
