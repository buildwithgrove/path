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
const maxErrMessageLen = 1000

// requestValidator handles EVM request validation, generating appropriate error contexts
// when validation fails or request contexts when validation succeeds.
type requestValidator struct {
	logger        polylog.Logger
	chainID       string
	endpointStore *EndpointStore
}

// validateHTTPRequest validates an HTTP request, extracting and validating its JSONRPC payload.
// If validation fails, an errorContext is returned along with false.
// If validation succeeds, a fully initialized requestContext is returned along with true.
func (rv *requestValidator) validateHTTPRequest(req *http.Request) (gateway.RequestQoSContext, bool) {
	logger := rv.logger.With(
		"qos", "EVM",
		"method", "validateHTTPRequest",
	)

	// TODO_TECHDEBT(@adshmh): Simplify the qos package by refactoring gateway.QoSContextBuilder.
	// Proposed change: Create a new ServiceRequest type containing raw payload data ([]byte)
	// Benefits: Decouples the qos package from HTTP-specific error handling.
	//
	// Read the request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Warn().Err(err).Msg("HTTP request body read failed - returning generic error response")
		return rv.createHTTPBodyReadFailureContext(err), false
	}

	// TODO_TECHDEBT(@adshmh): support Batch JSONRPC requests, as per the JSONRPC spec:
	// https://www.jsonrpc.org/specification#batch
	//
	// TODO_MVP(@adshmh): Add a JSON-RPC request validator to reject invalid/unsupported method calls early in request flow.
	//
	// Parse the JSONRPC request
	var jsonrpcReq jsonrpc.Request
	if err := json.Unmarshal(body, &jsonrpcReq); err != nil {
		logger.With(
			"request_preview", string(body[:min(maxErrMessageLen, len(body))]), // truncate body to first 1000 bytes for logging
		).Info().Err(err).Msg("Request failed validation - returning generic error response")
		return rv.createRequestUnmarshalingFailureContext(jsonrpcReq.ID, err), false
	}

	// TODO_MVP(@adshmh): Add JSON-RPC request validation to block invalid requests
	// TODO_IMPROVE(@adshmh): Add method-specific JSONRPC request validation

	// Request is valid, return a fully initialized requestContext
	return &requestContext{
		logger:        rv.logger,
		chainID:       rv.chainID,
		jsonrpcReq:    jsonrpcReq,
		endpointStore: rv.endpointStore,
	}, true
}

// createHTTPBodyReadFailureContext creates an error context for HTTP body read failures.
func (rv *requestValidator) createHTTPBodyReadFailureContext(err error) gateway.RequestQoSContext {

	// Create the observations object with the HTTP body read failure observation
	observations := createHTTPBodyReadFailureObservation(rv.chainID, err)

	// TODO_IMPROVE(@adshmh): Propagate a request ID parameter on internal errors that occur after successful request parsing.
	// There are no such cases as of PR #165.
	//
	// Create the JSON-RPC error response
	response := newErrResponseInternalErr(jsonrpc.ID{}, err)

	// Build and return the error context
	return &errorContext{
		logger:                 rv.logger,
		response:               response,
		responseHTTPStatusCode: httpStatusRequestValidationFailureReadHTTPBodyFailure,
		observations:           observations,
	}
}

// createRequestUnmarshalingFailureContext creates an error context for request unmarshaling failures.
func (rv *requestValidator) createRequestUnmarshalingFailureContext(id jsonrpc.ID, err error) gateway.RequestQoSContext {

	// Create the observations object with the request unmarshaling failure observation
	observations := createRequestUnmarshalingFailureObservation(id, rv.chainID, err)
	// Create the JSON-RPC error response
	response := newErrResponseInvalidRequest(err, id)

	// Build and return the error context
	return &errorContext{
		logger:                 rv.logger,
		response:               response,
		responseHTTPStatusCode: httpStatusRequestValidationFailureUnmarshalFailure,
		observations:           observations,
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
) qosobservations.Observations {
	errorDetails := err.Error()
	return qosobservations.Observations{
		ServiceObservations: &qosobservations.Observations_Evm{
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
) qosobservations.Observations {
	errorDetails := err.Error()
	return qosobservations.Observations{
		ServiceObservations: &qosobservations.Observations_Evm{
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
		},
	}
}
