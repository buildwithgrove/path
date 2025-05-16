package evm

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
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

	// Parse and validate the JSONRPC request
	jsonrpcReq, err := parseJSONRPCFromRequestBody(logger, body)
	if err != nil {
		return erv.createRequestUnmarshalingFailureContext(jsonrpcReq.ID, err), false
	}

	// TODO_MVP(@adshmh): Add JSON-RPC request validation to block invalid requests
	// TODO_IMPROVE(@adshmh): Add method-specific JSONRPC request validation

	// Request is valid, return a fully initialized requestContext
	return &requestContext{
		logger:               erv.logger,
		chainID:              erv.chainID,
		serviceID:            erv.serviceID,
		requestPayloadLength: uint(len(body)),
		jsonrpcReq:           jsonrpcReq,
		serviceState:         erv.serviceState,
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

// TODO_TECHDEBT(@adshmh): support Batch JSONRPC requests, as per the JSONRPC spec:
// https://www.jsonrpc.org/specification#batch
//
// TODO_MVP(@adshmh): Add a JSON-RPC request validator to reject invalid/unsupported
// method calls early in request flow.
//
// parseJSONRPCFromRequestBody attempts to unmarshal the HTTP request body into a JSONRPC
// request structure.
//
// If parsing fails, it:
// - Logs the first portion of the request body (truncated for security/performance)
// - Includes the specific error information
func parseJSONRPCFromRequestBody(
	logger polylog.Logger,
	requestBody []byte,
) (jsonrpc.Request, error) {
	var jsonrpcRequest jsonrpc.Request
	err := json.Unmarshal(requestBody, &jsonrpcRequest)
	if err != nil {
		// Only log a preview of the request body (first 1000 bytes or less) to avoid excessive logging
		requestPreview := string(requestBody[:min(maxErrMessageLen, len(requestBody))])
		logger.Info().Err(err).Msgf("Request failed JSON-RPC validation - returning generic error response. Request preview: %s", requestPreview)
	}

	return jsonrpcRequest, err
}
