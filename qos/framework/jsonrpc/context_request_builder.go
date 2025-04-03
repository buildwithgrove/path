package framework

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_IN_THIS_PR: ensure the observations will contain:
// - HTTP Status code: e.g. httpStatusRequestValidationFailureUnmarshalFailure,
// - Validation error: e.g. qosobservations.EVMRequestValidationError_EVM_REQUEST_VALIDATION_ERROR_REQUEST_UNMARSHALING_FAILURE,
// - Error details.
					

// TODO_TECHDEBT(@adshmh): Simplify the qos package by refactoring gateway.QoSContextBuilder.
// Proposed change: Create a new ServiceRequest type containing raw payload data ([]byte)
// Benefits: Decouples the qos package from HTTP-specific error handling.

// maximum length of the error message stored in request validation failure observations and logs.
// This is used to prevent overly verbose error messages from being stored in logs and metrics leading to excessive memory usage and cost.
const maxErrMessageLen = 1000

// TODO_IN_THIS_PR: add hydratedLoggers.

// requestBuilder handles the construction of requestQoSContext objects
type requestBuilder struct {
	Logger polylog.Logger
	ServiceID ServiceID
	EndpointCallProcessor  endpointCallProcessor
	EndpointSelector endpointSelector

	context *requestQoSContext
}

// parseHTTPRequest reads and processes the HTTP request
// validates an HTTP request, extracting and validating its EVM JSONRPC payload.
func (rb *requestBuilder) ParseHTTPRequest(httpReq *http.Request) *requestBuilder {
	requestCtx := requestQoSContext {
		Logger: rb.Logger,
		ServiceID: rb.ServiceID,
	}

	// Read the HTTP request body
	body, err := io.ReadAll(httpReq.Body)
	defer httpReq.Body.Close()

	// TODO_IMPROVE(@adshmh): Propagate a request ID parameter on internal errors
	// that occur after successful request parsing.
	// There are no such cases as of PR #186.
	if err != nil {
		// Handle read error (internal server error)
		rb.Logger.Error().Err(err).Msg("Failed to read request body")

		// Create error response for read failure
		errResp := newErrResponseInternalReadError(err)

		requestCtx.JSONRPCErrorResponse = &errResp 
		rb.context = requestCtx
		return rb
	}

	// Parse the JSON-RPC request
	var jsonrpcReq jsonrpc.JsonRpcRequest
	if err := json.Unmarshal(body, &jsonrpcReq); err != nil {
		// TODO_IN_THIS_PR: log the first 1K bytes of the body.
		// Handle parse error (client error)
		rb.Logger.Error().Err(err).Msg("Failed to parse JSON-RPC request")

		// Create error response for parse failure
		errResp := newErrResponseParseError(err)

		requestCtx.JSONRPCErrorResponse = &errResp
		rb.context = requestCtx
		return rb
	}

	// Store the parsed request
	requestCtx.Request = &jsonrpcReq
	rb.context = requestCtx
	return rb
}

// validateRequest validates the JSONRPC request using the default validator
func (rb *requestBuilder) ValidateRequest() *requestBuilder {
	// Skip validation if we already have an error
	if rb.context.errorResponse != nil || rb.context.request == nil {
		return rb
	}

	// Validate the request
	errResp, isValid := validateRequest(rb.context.request)
	if !isValid {
		// Request is invalid according to the validator
		rb.Logger.Warn().
			Str("method", rb.context.Request.Method).
			Msg("Request validation failed")

		rb.context.JSONRPCErrorResponse = errResp
		return rb
	}

	rb.context.EndpointCallsProcessor  = rb.EndpointCallProcessor
	rb.context.EndpointSelector = rb.EndpointSelector

	// Request is valid
	rb.Logger.Info().
		Str("method", rb.context.request.Method).
		Msg("Request validation successful")

	return rb
}

// build finalizes and returns the request context
func (rb *requestBuilder) Build() (*requestQoSContext, bool) {
	return rb.context, rb.context.errorResponse == nil
}

// TODO_MVP(@adshmh): Allow custom QoS services to supply custom request validation logic.
// Example use case: specifying a list of allowed JSONRPC request methods.
// This would require:
// 1. Declaring a public RequestValidator interface.
// 2. Helper functions, e.g. BuildRequestValidatorForAllowedMethods.
//
// validateRequest provides a basic validation of JSONRPC requests.
// It checks:
// - JSONRPC version (must be "2.0")
// - Method presence
//
// It returns a JSONRPC response if validation fails and a boolean indicating whether request processing should continue.
func validateRequest(request *jsonrpc.JsonRpcRequest, allowedMethods map[string]bool) (*jsonrpc.Response, bool) {
	// Check JSONRPC version
	if request.Jsonrpc != jsonrpc.Version2 {
		resp := newErrResponseInvalidVersionError(request.Id)
		return &resp, false
	}

	// Check method presence
	if request.Method == "" {
		resp := newErrResponseMissingMethodError(request.Id)
		return &resp, false
	}

	// Request is valid
	return nil, true
}
