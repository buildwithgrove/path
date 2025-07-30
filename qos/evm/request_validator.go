package evm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/log"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_TECHDEBT(@adshmh): Simplify the qos package by refactoring gateway.QoSContextBuilder.
// Proposed change: Create a new ServiceRequest type containing raw payload data ([]byte)
// Benefits: Decouples the qos package from HTTP-specific error handling.

// maximum length of the error message stored in request validation failure observations and logs.
// This is used to prevent overly verbose error messages from being stored in logs and metrics leading to excessive memory usage and cost.
const maxErrMessageLen = 1000

// TODO_TECHDEBT(@adshmh): Refactor the evmRequestValidator struct to be more generic and reusable.
//
// evmRequestValidator handles request validation, generating:
// - Error contexts when validation fails
// - Request contexts when validation succeeds
// TODO_IMPROVE(@adshmh): Consider creating an interface with method-specific JSONRPC request validation
type evmRequestValidator struct {
	logger       polylog.Logger
	chainID      string
	serviceID    protocol.ServiceID
	serviceState *serviceState
}

// validateHTTPRequest validates an HTTP request, extracting and validating its EVM JSONRPC payload.
// If validation fails, an errorContext is returned along with false.
// If validation succeeds, a fully initialized requestContext is returned along with true.
func (erv *evmRequestValidator) validateHTTPRequest(req *http.Request) (gateway.RequestQoSContext, bool) {
	logger := erv.logger.With(
		"qos", "EVM",
		"method", "validateHTTPRequest",
	)

	// Read the HTTP request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Warn().Err(err).Msg("HTTP request body read failed - returning generic error response")
		return erv.createHTTPBodyReadFailureContext(err), false
	}

	// Parse and validate the JSONRPC request(s) - handles both single and batch requests
	jsonrpcReqs, err := parseJSONRPCFromRequestBody(logger, body)
	if err != nil {
		// For error context, use the first request ID if available, otherwise use empty ID
		var requestID jsonrpc.ID
		if len(jsonrpcReqs) > 0 {
			requestID = jsonrpcReqs[0].ID
		}
		// If no requests parsed or empty ID, requestID will be zero value (empty)
		return erv.createRequestUnmarshalingFailureContext(requestID, err), false
	}

	// TODO_MVP(@adshmh): Add JSON-RPC request validation to block invalid requests
	// TODO_IMPROVE(@adshmh): Add method-specific JSONRPC request validation

	// Request is valid, return a fully initialized requestContext
	return &requestContext{
		logger:               erv.logger,
		chainID:              erv.chainID,
		serviceID:            erv.serviceID,
		requestPayloadLength: uint(len(body)),
		jsonrpcReqs:          jsonrpcReqs,
		serviceState:         erv.serviceState,
		// Set the origin of the request as ORGANIC (i.e. from a user).
		requestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
	}, true
}

// createHTTPBodyReadFailureContext creates an error context for HTTP body read failures.
func (erv *evmRequestValidator) createHTTPBodyReadFailureContext(err error) gateway.RequestQoSContext {
	// Create the observations object with the HTTP body read failure observation
	observations := createHTTPBodyReadFailureObservation(erv.serviceID, erv.chainID, err)

	// TODO_IMPROVE(@adshmh): Propagate a request ID parameter on internal errors
	// that occur after successful request parsing.
	// There are no such cases as of PR #186.
	//
	// Create the JSON-RPC error response
	response := newErrResponseInternalErr(jsonrpc.ID{}, err)

	// Build and return the error context
	return &errorContext{
		logger:                 erv.logger,
		response:               response,
		responseHTTPStatusCode: httpStatusRequestValidationFailureReadHTTPBodyFailure,
		evmObservations:        observations,
	}
}

// createRequestUnmarshalingFailureContext creates an error context for request unmarshaling failures.
func (erv *evmRequestValidator) createRequestUnmarshalingFailureContext(id jsonrpc.ID, err error) gateway.RequestQoSContext {

	// Create the observations object with the request unmarshaling failure observation
	observations := createRequestUnmarshalingFailureObservation(id, erv.serviceID, erv.chainID, err)
	// Create the JSON-RPC error response
	response := newErrResponseInvalidRequest(err, id)

	// Build and return the error context
	return &errorContext{
		logger:                 erv.logger,
		response:               response,
		responseHTTPStatusCode: httpStatusRequestValidationFailureUnmarshalFailure,
		evmObservations:        observations,
	}
}

// createRequestUnmarshalingFailureObservation creates an observation for an EVM request
// that failed to unmarshal from JSON.
//
// This observation:
// - Captures details about the validation failure (request ID, error message, chain ID)
// - Is used for both reporting metrics and providing context for debugging
//
// Parameters:
// - id: The JSON-RPC request ID associated with the failed request
// - err: The error that occurred during unmarshaling
// - chainID: The EVM chain identifier for which the request was intended
//
// Returns:
// - qosobservations.Observations: A structured observation containing details about the validation failure
func createRequestUnmarshalingFailureObservation(
	_ jsonrpc.ID,
	serviceID protocol.ServiceID,
	chainID string,
	err error,
) *qosobservations.Observations_Evm {
	errorDetails := err.Error()
	return &qosobservations.Observations_Evm{
		Evm: &qosobservations.EVMRequestObservations{
			ServiceId: string(serviceID),
			ChainId:   chainID,
			RequestValidationFailure: &qosobservations.EVMRequestObservations_EvmRequestUnmarshalingFailure{
				EvmRequestUnmarshalingFailure: &qosobservations.EVMRequestUnmarshalingFailure{
					HttpStatusCode:  httpStatusRequestValidationFailureUnmarshalFailure,
					ValidationError: qosobservations.EVMRequestValidationError_EVM_REQUEST_VALIDATION_ERROR_REQUEST_UNMARSHALING_FAILURE,
					ErrorDetails:    &errorDetails,
				},
			},
		},
	}
}

// createHTTPBodyReadFailureObservation creates an observation for cases where
// reading the HTTP request body for an EVM service request has failed.
//
// This observation:
// - Includes the chainID and detailed error information
// - Is useful for diagnosing connectivity or HTTP parsing issues
//
// Parameters:
// - chainID: The EVM chain identifier for which the request was intended
// - err: The error that occurred during HTTP body reading
//
// Returns:
// - qosobservations.Observations: A structured observation containing details about the HTTP read failure
func createHTTPBodyReadFailureObservation(
	serviceID protocol.ServiceID,
	chainID string,
	err error,
) *qosobservations.Observations_Evm {
	errorDetails := err.Error()
	return &qosobservations.Observations_Evm{
		Evm: &qosobservations.EVMRequestObservations{
			ServiceId: string(serviceID),
			ChainId:   chainID,
			RequestValidationFailure: &qosobservations.EVMRequestObservations_EvmHttpBodyReadFailure{
				EvmHttpBodyReadFailure: &qosobservations.EVMHTTPBodyReadFailure{
					HttpStatusCode:  httpStatusRequestValidationFailureReadHTTPBodyFailure,
					ValidationError: qosobservations.EVMRequestValidationError_EVM_REQUEST_VALIDATION_ERROR_HTTP_BODY_READ_FAILURE,
					ErrorDetails:    &errorDetails,
				},
			},
		},
	}
}

// TODO_MVP(@adshmh): Add a JSON-RPC request validator to reject invalid/unsupported
// method calls early in request flow.
//
// parseJSONRPCFromRequestBody is the main entry point for parsing HTTP request bodies into
// JSON-RPC request structures. It orchestrates the parsing process by:
//  1. Validating the request body format
//  2. Detecting whether it's a single request or batch request
//  3. Delegating to appropriate parsing methods
//  4. Returning a normalized slice of JSON-RPC requests
//
// Supports both single requests and batch requests according to the JSON-RPC 2.0 specification.
// Reference: https://www.jsonrpc.org/specification#batch
func parseJSONRPCFromRequestBody(
	logger polylog.Logger,
	requestBody []byte,
) ([]jsonrpc.Request, error) {
	// Step 1: Validate the request body format
	trimmedBody, err := validateRequestBody(logger, requestBody)
	if err != nil {
		return nil, err
	}

	// Step 2: Detect request format (batch vs single)
	requestFormat := detectRequestFormat(trimmedBody)

	// Step 3: Parse based on detected format
	switch requestFormat {
	case jsonrpcBatchRequest:
		return parseBatchRequest(logger, requestBody)
	case jsonrpcSingleRequest:
		return parseSingleRequest(logger, requestBody)
	default:
		return handleInvalidFormat(logger, requestBody)
	}
}

// jsonrpcRequestFormat represents the detected format of a JSON-RPC request
type jsonrpcRequestFormat int

const (
	jsonrpcInvalidFormat jsonrpcRequestFormat = iota
	jsonrpcSingleRequest
	jsonrpcBatchRequest
)

// validateRequestBody performs initial validation on the HTTP request body.
// It trims whitespace and ensures the body is not empty.
//
// Returns:
//   - trimmedBody: the request body with leading/trailing whitespace removed
//   - error: validation error if the body is empty or invalid
func validateRequestBody(logger polylog.Logger, requestBody []byte) ([]byte, error) {
	trimmedBody := bytes.TrimSpace(requestBody)
	if len(trimmedBody) == 0 {
		logger.Error().Msg("❌ Request failed JSON-RPC validation - empty request body")
		return nil, fmt.Errorf("empty request body")
	}
	return trimmedBody, nil
}

// detectRequestFormat analyzes the trimmed request body to determine if it represents
// a single JSON-RPC request or a batch request.
//
// According to JSON-RPC 2.0 specification:
//   - Single request: JSON object starting with '{'
//   - Batch request: JSON array starting with '['
//
// Returns the detected format type for routing to appropriate parsing logic.
func detectRequestFormat(trimmedBody []byte) jsonrpcRequestFormat {
	if len(trimmedBody) == 0 {
		return jsonrpcInvalidFormat
	}

	switch trimmedBody[0] {
	case '[':
		return jsonrpcBatchRequest
	case '{':
		return jsonrpcSingleRequest
	default:
		return jsonrpcInvalidFormat
	}
}

// parseBatchRequest handles parsing of JSON-RPC batch requests (JSON arrays).
// Performs validation to ensure:
//   - The array can be unmarshaled into JSON-RPC request structures
//   - The batch is not empty (per JSON-RPC 2.0 specification requirement)
//
// Returns a slice of JSON-RPC requests on success, or an error with diagnostic information.
func parseBatchRequest(logger polylog.Logger, requestBody []byte) ([]jsonrpc.Request, error) {
	var jsonrpcRequests []jsonrpc.Request
	err := json.Unmarshal(requestBody, &jsonrpcRequests)
	if err != nil {
		requestPreview := log.Preview(string(requestBody))
		logger.Error().Err(err).Msgf("❌ Batch request failed JSON-RPC validation - returning generic error response. Request preview: %s", requestPreview)
		return nil, err
	}

	// Validate that batch is not empty (per JSON-RPC spec)
	if len(jsonrpcRequests) == 0 {
		logger.Error().Msg("❌ Empty batch request not allowed per JSON-RPC specification")
		return nil, fmt.Errorf("empty batch request not allowed")
	}

	logger.Debug().Int("batch_size", len(jsonrpcRequests)).Msg("Parsed JSON-RPC batch request")
	return jsonrpcRequests, nil
}

// parseSingleRequest handles parsing of single JSON-RPC requests (JSON objects).
// Unmarshals the request body into a JSON-RPC request structure and wraps it
// in a slice for consistent return type with batch requests.
//
// Returns a single-element slice containing the parsed JSON-RPC request.
func parseSingleRequest(logger polylog.Logger, requestBody []byte) ([]jsonrpc.Request, error) {
	var jsonrpcRequest jsonrpc.Request
	err := json.Unmarshal(requestBody, &jsonrpcRequest)
	if err != nil {
		requestPreview := log.Preview(string(requestBody))
		logger.Error().Err(err).Msgf("❌ Request failed JSON-RPC validation - returning generic error response. Request preview: %s", requestPreview)
		return nil, err
	}

	logger.Debug().Msg("Parsed single JSON-RPC request")
	return []jsonrpc.Request{jsonrpcRequest}, nil
}

// handleInvalidFormat processes request bodies that don't conform to valid JSON-RPC format.
// This handles cases where the request body doesn't start with '{' (object) or '[' (array).
//
// Returns an error with diagnostic information including a preview of the invalid content.
func handleInvalidFormat(logger polylog.Logger, requestBody []byte) ([]jsonrpc.Request, error) {
	requestPreview := log.Preview(string(requestBody))
	logger.Error().Msgf("❌ Invalid JSON-RPC format - must start with '{' or '['. Request preview: %s", requestPreview)
	return nil, fmt.Errorf("invalid JSON-RPC format - must be JSON object or array")
}
