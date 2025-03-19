package evm

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// maximum length of the error message stored in request validation failure observations and logs.
// This is used to prevent overly verbose error messages from being stored in logs and metrics leading to excessive memory usage and cost.
const maxErrMessageLen = 1000

// TODO_REFACTOR: consider refactoring evmRequestValidator out of qos/evm and into qos to help reuse the code if possible.
//
// evmRequestValidator handles request validation, generating appropriate error contexts
// when validation fails or request contexts when validation succeeds.
type evmRequestValidator struct {
	logger        polylog.Logger
	// DEV_NOTE: If adapting this struct for non-blockchain QoS services, replace chainID with an appropriate serviceID
	chainID       string
	endpointStore *EndpointStore
}

// validateHTTPRequest validates an HTTP request, extracting and validating its JSONRPC payload.
// If validation fails, an errorContext is returned along with false.
// If validation succeeds, a fully initialized requestContext is returned along with true.
func (erv *evmRequestValidator) validateHTTPRequest(req *http.Request) (gateway.RequestQoSContext, bool) {
	logger := erv.logger.With(
		"qos", "EVM",
		"method", "validateHTTPRequest",
	)

	// TODO_TECHDEBT(@adshmh): Simplify the qos package by refactoring gateway.QoSContextBuilder.
	// Proposed change: Create a new ServiceRequest type containing raw payload data ([]byte)
	// Benefits: Decouples the qos package from HTTP-specific error handling.

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
		logger:        erv.logger,
		chainID:       erv.chainID,
		jsonrpcReq:    jsonrpcReq,
		endpointStore: erv.endpointStore,
	}, true
}

// createHTTPBodyReadFailureContext creates an error context for HTTP body read failures.
func (erv *evmRequestValidator) createHTTPBodyReadFailureContext(err error) gateway.RequestQoSContext {
	// Create the observations object with the HTTP body read failure observation
	observations := createHTTPBodyReadFailureObservation(erv.chainID, err)

	// TODO_IMPROVE(@adshmh): Propagate a request ID parameter on internal errors that occur after successful request parsing.
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
	observations := createRequestUnmarshalingFailureObservation(id, erv.chainID, err)
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
// that failed to unmarshal from JSON. This observation captures details about the validation
// failure, including the request ID, error message, and chain ID. It is used for both
// reporting metrics and providing context for debugging.
//
// Parameters:
//   - id: The JSON-RPC request ID associated with the failed request
//   - err: The error that occurred during unmarshaling
//   - chainID: The EVM chain identifier for which the request was intended
//
// Returns:
//   - qosobservations.Observations: A structured observation containing details about the validation failure
func createRequestUnmarshalingFailureObservation(
	id jsonrpc.ID,
	chainID string,
	err error,
) *qosobservations.Observations_Evm {
	errorDetails := err.Error()
	return &qosobservations.Observations_Evm{
		Evm: &qosobservations.EVMRequestObservations{
			ChainId: chainID,
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
// reading the HTTP request body for an EVM service request has failed. This observation
// includes the chainID and detailed error information, which is useful for diagnosing
// connectivity or HTTP parsing issues.
//
// Parameters:
//   - chainID: The EVM chain identifier for which the request was intended
//   - err: The error that occurred during HTTP body reading
//
// Returns:
//   - qosobservations.Observations: A structured observation containing details about the HTTP read failure
func createHTTPBodyReadFailureObservation(
	chainID string,
	err error,
) *qosobservations.Observations_Evm {
	errorDetails := err.Error()
	return &qosobservations.Observations_Evm{
		Evm: &qosobservations.EVMRequestObservations{
			ChainId: chainID,
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
// TODO_MVP(@adshmh): Add a JSON-RPC request validator to reject invalid/unsupported method calls early in request flow.
//
// parseJSONRPCFromRequestBody attempts to unmarshal the HTTP request body into a JSONRPC request structure.
// If parsing fails, it logs the first portion of the request body (truncated for security/performance)
// along with the specific error.
func parseJSONRPCFromRequestBody(
	logger polylog.Logger,
	requestBody []byte,
) (jsonrpc.Request, error) {
	var jsonrpcRequest jsonrpc.Request
	err := json.Unmarshal(requestBody, &jsonrpcRequest)

	if err != nil {
		// Only log a preview of the request body (first 1000 bytes or less) to avoid excessive logging
		requestPreview := string(requestBody[:min(maxErrMessageLen, len(requestBody))])

		logger.With(
			"request_preview", requestPreview,
		).Info().Err(err).Msg("Request failed JSON-RPC validation - returning generic error response")
	}

	return jsonrpcRequest, err
}
