package solana

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// maximum length of the error message stored in request validation failure observations and logs.
// This is used to prevent overly verbose error messages from being stored in logs and metrics leading to excessive memory usage and cost.
const maxErrMessageLen = 1000

// TODO_MVP(@adshmh): Drop this struct once JUDGE framework is merged.
//
// requestValidator handles request validation, generating:
// - Error contexts if validation fails: e.g. error parsing into a JSONRPC request.
// - Request context if valdiation succeeds.
type requestValidator struct {
	logger       polylog.Logger
	chainID      string
	serviceID    protocol.ServiceID
	serviceState *ServiceState
}

// validateHTTPRequest validates an HTTP request, extracting and validating its EVM JSONRPC payload.
// If validation fails, an errorContext is returned along with false.
// If validation succeeds, a fully initialized requestContext is returned along with true.
func (rv *requestValidator) validateHTTPRequest(req *http.Request) (gateway.RequestQoSContext, bool) {
	logger := rv.logger.With(
		"qos", "Solana",
		"method", "validateHTTPRequest",
	)

	// Read the HTTP request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Warn().Err(err).Msg("HTTP request body read failed - returning generic error response")
		return rv.createHTTPBodyReadFailureContext(err), false
	}

	// Parse and validate the JSONRPC request
	jsonrpcReq, err := parseJSONRPCFromRequestBody(logger, body)
	if err != nil {
		return rv.createRequestUnmarshalingFailureContext(jsonrpcReq.ID, err), false
	}

	// Request is valid, return a fully initialized requestContext
	return &requestContext{
		logger:               rv.logger,
		chainID:              rv.chainID,
		serviceID:            rv.serviceID,
		requestPayloadLength: uint(len(body)),
		JSONRPCReq:           jsonrpcReq,
		// Set the origin of the request as USER:
		// i.e. the request is from a user.
		requestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_USER,
	}, true
}

// createHTTPBodyReadFailureContext creates an error context for HTTP body read failures.
func (rv *requestValidator) createHTTPBodyReadFailureContext(err error) gateway.RequestQoSContext {
	// TODO_IMPROVE(@adshmh): Propagate a request ID parameter on internal errors
	// that occur after successful request parsing.
	// There are no such cases as of PR #186.
	//
	// Create the JSON-RPC error response
	response := jsonrpc.NewErrResponseInternalErr(jsonrpc.ID{}, err)

	// Create the observations object with the HTTP body read failure observation
	observations := createHTTPBodyReadFailureObservation(rv.serviceID, rv.chainID, err, response)

	// Build and return the error context
	return &qos.RequestErrorContext{
		Logger:   rv.logger,
		Response: response,
		Observations: qosobservations.Observations{
			ServiceObservations: observations,
		},
	}
}

// createRequestUnmarshalingFailureContext creates an error context for request unmarshaling failures.
func (rv *requestValidator) createRequestUnmarshalingFailureContext(id jsonrpc.ID, err error) gateway.RequestQoSContext {
	// Create the JSON-RPC error response
	response := jsonrpc.NewErrResponseInvalidRequest(id, err)

	// Create the observations object with the request unmarshaling failure observation
	observations := createRequestUnmarshalingFailureObservation(id, rv.serviceID, rv.chainID, err, response)

	// Build and return the error context
	return &qos.RequestErrorContext{
		Logger:   rv.logger,
		Response: response,
		Observations: qosobservations.Observations{
			ServiceObservations: observations,
		},
	}
}

// createRequestUnmarshalingFailureObservation creates an observation for an Solana request
// that failed to unmarshal from JSON.
//
// This observation:
// - Captures details about the validation failure (request ID, error message, chain ID)
// - Is used for both reporting metrics and providing context for debugging
//
// Parameters:
// - id: The JSON-RPC request ID associated with the failed request
// - err: The error that occurred during unmarshaling
// - chainID: The Solana chain identifier for which the request was intended
//
// Returns:
// - qosobservations.Observations: A structured observation containing details about the validation failure
func createRequestUnmarshalingFailureObservation(
	_ jsonrpc.ID,
	serviceID protocol.ServiceID,
	chainID string,
	err error,
	jsonrpcResponse jsonrpc.Response,
) *qosobservations.Observations_Solana {
	return &qosobservations.Observations_Solana{
		Solana: &qosobservations.SolanaRequestObservations{
			ServiceId: string(serviceID),
			ChainId:   chainID,
			RequestError: &qosobservations.RequestError{
				ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_JSONRPC_PARSE_ERROR,
				ErrorDetails:   err.Error(),
				HttpStatusCode: int32(jsonrpcResponse.GetRecommendedHTTPStatusCode()),
			},
		},
	}
}

// createHTTPBodyReadFailureObservation creates an observation for cases where
// reading the HTTP request body for an Solana service request has failed.
//
// This observation:
// - Includes the chainID and detailed error information
// - Is useful for diagnosing connectivity or HTTP parsing issues
//
// Parameters:
// - chainID: The Solana chain identifier for which the request was intended
// - err: The error that occurred during HTTP body reading
//
// Returns:
// - qosobservations.Observations: A structured observation containing details about the HTTP read failure
func createHTTPBodyReadFailureObservation(
	serviceID protocol.ServiceID,
	chainID string,
	err error,
	jsonrpcResponse jsonrpc.Response,
) *qosobservations.Observations_Solana {
	return &qosobservations.Observations_Solana{
		Solana: &qosobservations.SolanaRequestObservations{
			ServiceId: string(serviceID),
			ChainId:   chainID,
			RequestError: &qosobservations.RequestError{
				ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_JSONRPC_PARSE_ERROR,
				ErrorDetails:   err.Error(),
				HttpStatusCode: int32(jsonrpcResponse.GetRecommendedHTTPStatusCode()),
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
		logger.With(
			"request_preview", requestPreview,
		).Info().Err(err).Msg("Request failed JSON-RPC validation - returning generic error response")
	}

	return jsonrpcRequest, err
}
