package framework

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_MVP(@adshmh): Allow custom QoS services to supply custom request validation logic.
// Example use case: specifying a list of allowed JSONRPC request methods.
// This would require:
// 1. Declaring a public RequestValidator interface.
// 2. Helper functions, e.g. BuildRequestValidatorForAllowedMethods.
//
// maximum length of the error message stored in request validation failure observations and logs.
// This is used to prevent overly verbose error messages from being stored in logs and metrics leading to excessive memory usage and cost.
const maxErrMessageLen = 1000

type  requestDetails struct {
	// The client's JSONRPC request
	// Only set if the request was successfully parsed.
	request *jsonrpc.Request

	// Request error, if any.
	requestError *requestError
}

func (rd *requestDetails) isValid() bool {
	return rd.requestError == nil
}

// buildRequestContextFromHTTPRequest builds and returns a context for processing the HTTP request:
// - Reads and processes the HTTP request
// - Parses a JSONRPC request from the HTTP request's payload.
// - Validates the resulting JSONRPC request.
// - Initializes the context for processing the request in the following scenarios:
//   - Internal errors: e.g. reading the HTTP request.
//   - Invalid request: e.g. malformed payload.
//   - Valid request: proper JSONRPC request.
func buildRequestDetailsFromHTTP(
	logger polylog.Logger,
	httpReq *http.Request,
) *requestDetails {
	// Read the HTTP request body
	body, err := io.ReadAll(httpReq.Body)
	defer httpReq.Body.Close()

	// TODO_IMPROVE(@adshmh): Propagate a request ID parameter on internal errors that occur after successful request parsing.
	// There are no such cases as of PR #186.
	if err != nil {
		// Handle read error (internal server error)
		logger.Error().Err(err).Msg("Failed to read request body")

		// return the error details to be stored in the request journal.
		return buildRequestDetailsForInternalErrHTTPRead(err)
	}

	// Parse the JSON-RPC request
	var jsonrpcReq jsonrpc.JsonRpcRequest
	if err := json.Unmarshal(body, &jsonrpcReq); err != nil {
		// TODO_IN_THIS_PR: log the first 1K bytes of the body.
		// Handle parse error (client error)
		logger.Error().Err(err).Msg("Failed to parse JSON-RPC request")

		return buildRequestDetailsForParseError(err)
	}

	// Validate the request
	requestErr := validateRequest(jsonrpcReq)
	if requestErr != nil {
		// Request is invalid according to the validator
		logger.Info().
			Str("method", jsonrpcReq.Method).
			Msg("Request validation failed")

		return requestDetails{
			request: jsonrpcReq,
			requestError: requestErr,
		}
	}

	// Request is valid
	logger.Debug().
		Str("method", jsonrpcReq.Method).
		Msg("Request validation successful")

	return requestDetails {
		request: &jsonrpcReq,
	}

}

// validateRequest provides a basic validation of JSONRPC requests.
// It checks:
// - JSONRPC version (must be "2.0")
// - Method presence
//
// Returns a non-nil requestError if validation fails.
func validateRequest(request *jsonrpc.Request) *requestError {
	// Check JSONRPC version
	if request.Jsonrpc != jsonrpc.Version2 {
		return buildRequestErrorJSONRPCErrInvalidVersion(request.ID)
	}

	// Check method presence
	if request.Method == "" {
		return buildRequestErrorJSONRPCErrMissingMethod(request.ID)
	}

	// Request is valid
	return nil
}
