package cosmos

import (
	"encoding/json"
	"fmt"
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

// TODO_TECHDEBT(@adshmh): Refactor the cosmosSDKRequestValidator struct to be more generic and reusable.
//
// cosmosSDKRequestValidator handles request validation for CosmosSDK chains, generating:
// - Error contexts when validation fails
// - Request contexts when validation succeeds
// Supports both REST endpoints (/health, /status) and JSON-RPC requests
type cosmosSDKRequestValidator struct {
	logger       polylog.Logger
	chainID      string
	serviceID    protocol.ServiceID
	serviceState *serviceState
}

// validateHTTPRequest validates an HTTP request for CosmosSDK chains.
// Handles both REST endpoints (GET /health, /status) and JSON-RPC POST requests.
// If validation fails, an errorContext is returned along with false.
// If validation succeeds, a fully initialized requestContext is returned along with true.
func (crv *cosmosSDKRequestValidator) validateHTTPRequest(req *http.Request) (gateway.RequestQoSContext, bool) {
	logger := crv.logger.With(
		"qos", "CosmosSDK",
		"method", "validateHTTPRequest",
	)

	// Check if this is a REST endpoint request (GET /health or /status)
	if req.Method == http.MethodGet && crv.isRESTEndpoint(req.URL.Path) {
		return crv.createRESTRequestContext(req), true
	}

	// Handle JSON-RPC POST requests
	if req.Method == http.MethodPost {
		return crv.validateJSONRPCRequest(req, logger)
	}

	// Unsupported method or path
	logger.Warn().Msgf("Unsupported HTTP method %s or path %s", req.Method, req.URL.Path)
	return crv.createUnsupportedMethodContext(req), false
}

// isRESTEndpoint checks if the path corresponds to a supported REST endpoint
func (crv *cosmosSDKRequestValidator) isRESTEndpoint(path string) bool {
	return path == apiPathHealthCheck || path == apiPathStatus
}

// createRESTRequestContext creates a request context for REST endpoint requests
func (crv *cosmosSDKRequestValidator) createRESTRequestContext(req *http.Request) *requestContext {
	return &requestContext{
		logger:               crv.logger,
		httpReq:              req,
		chainID:              crv.chainID,
		serviceID:            crv.serviceID,
		requestPayloadLength: 0, // REST requests have no body
		serviceState:         crv.serviceState,
		requestOrigin:        qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
	}
}

// validateJSONRPCRequest validates a JSON-RPC POST request
func (crv *cosmosSDKRequestValidator) validateJSONRPCRequest(req *http.Request, logger polylog.Logger) (gateway.RequestQoSContext, bool) {
	// Read the HTTP request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Warn().Err(err).Msg("HTTP request body read failed - returning generic error response")
		return crv.createHTTPBodyReadFailureContext(err), false
	}

	// Parse and validate the JSONRPC request
	jsonrpcReq, err := parseJSONRPCFromRequestBody(logger, body)
	if err != nil {
		return crv.createRequestUnmarshalingFailureContext(jsonrpcReq.ID, err), false
	}

	// TODO_MVP(@adshmh): Add JSON-RPC request validation to block invalid requests
	// TODO_IMPROVE(@adshmh): Add method-specific JSONRPC request validation

	// Request is valid, return a fully initialized requestContext
	return &requestContext{
		logger:               crv.logger,
		httpReq:              req,
		chainID:              crv.chainID,
		serviceID:            crv.serviceID,
		requestPayloadLength: uint(len(body)),
		jsonrpcReq:           jsonrpcReq,
		serviceState:         crv.serviceState,
		// Set the origin of the request as ORGANIC (i.e. from a user).
		requestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
	}, true
}

// createUnsupportedMethodContext creates an error context for unsupported HTTP methods or paths
func (crv *cosmosSDKRequestValidator) createUnsupportedMethodContext(req *http.Request) gateway.RequestQoSContext {
	err := jsonrpc.NewErrResponseInvalidRequest(jsonrpc.ID{},
		fmt.Errorf("unsupported method %s or path %s", req.Method, req.URL.Path))

	observations := createUnsupportedMethodObservation(crv.serviceID, crv.chainID, req.Method, req.URL.Path)

	return &errorContext{
		logger:                 crv.logger,
		response:               err,
		responseHTTPStatusCode: http.StatusMethodNotAllowed,
		cosmosSDKObservations:  observations,
	}
}

// createHTTPBodyReadFailureContext creates an error context for HTTP body read failures.
func (crv *cosmosSDKRequestValidator) createHTTPBodyReadFailureContext(err error) gateway.RequestQoSContext {
	// Create the observations object with the HTTP body read failure observation
	observations := createHTTPBodyReadFailureObservation(crv.serviceID, crv.chainID, err)

	// TODO_IMPROVE(@adshmh): Propagate a request ID parameter on internal errors
	// that occur after successful request parsing.
	// There are no such cases as of PR #186.
	//
	// Create the JSON-RPC error response
	response := newErrResponseInternalErr(jsonrpc.ID{}, err)

	// Build and return the error context
	return &errorContext{
		logger:                 crv.logger,
		response:               response,
		responseHTTPStatusCode: httpStatusRequestValidationFailureReadHTTPBodyFailure,
		cosmosSDKObservations:  observations,
	}
}

// createRequestUnmarshalingFailureContext creates an error context for request unmarshaling failures.
func (crv *cosmosSDKRequestValidator) createRequestUnmarshalingFailureContext(id jsonrpc.ID, err error) gateway.RequestQoSContext {

	// Create the observations object with the request unmarshaling failure observation
	observations := createRequestUnmarshalingFailureObservation(id, crv.serviceID, crv.chainID, err)
	// Create the JSON-RPC error response
	response := newErrResponseInvalidRequest(err, id)

	// Build and return the error context
	return &errorContext{
		logger:                 crv.logger,
		response:               response,
		responseHTTPStatusCode: httpStatusRequestValidationFailureUnmarshalFailure,
		cosmosSDKObservations:  observations,
	}
}

// createUnsupportedMethodObservation creates an observation for unsupported HTTP method or path
func createUnsupportedMethodObservation(
	serviceID protocol.ServiceID,
	chainID string,
	method string,
	path string,
) *qosobservations.Observations_Cosmos {
	return &qosobservations.Observations_Cosmos{
		Cosmos: &qosobservations.CosmosSDKRequestObservations{
			RouteRequest: fmt.Sprintf("%s %s", method, path),
			EndpointObservations: []*qosobservations.CosmosSDKEndpointObservation{
				{
					ResponseObservation: &qosobservations.CosmosSDKEndpointObservation_UnrecognizedResponse{
						UnrecognizedResponse: &qosobservations.CosmosSDKUnrecognizedResponse{
							JsonrpcResponse: &qosobservations.JsonRpcResponse{
								Id: "",
							},
						},
					},
				},
			},
		},
	}
}

// createRequestUnmarshalingFailureObservation creates an observation for a CosmosSDK request
// that failed to unmarshal from JSON.
//
// This observation:
// - Captures details about the validation failure (request ID, error message, chain ID)
// - Is used for both reporting metrics and providing context for debugging
//
// Parameters:
// - id: The JSON-RPC request ID associated with the failed request
// - err: The error that occurred during unmarshaling
// - chainID: The CosmosSDK chain identifier for which the request was intended
//
// Returns:
// - qosobservations.Observations: A structured observation containing details about the validation failure
func createRequestUnmarshalingFailureObservation(
	_ jsonrpc.ID,
	serviceID protocol.ServiceID,
	chainID string,
	err error,
) *qosobservations.Observations_Cosmos {
	return &qosobservations.Observations_Cosmos{
		Cosmos: &qosobservations.CosmosSDKRequestObservations{
			RouteRequest: "JSON-RPC unmarshaling failed",
			EndpointObservations: []*qosobservations.CosmosSDKEndpointObservation{
				{
					ResponseObservation: &qosobservations.CosmosSDKEndpointObservation_UnrecognizedResponse{
						UnrecognizedResponse: &qosobservations.CosmosSDKUnrecognizedResponse{
							JsonrpcResponse: &qosobservations.JsonRpcResponse{
								Id: "",
							},
						},
					},
				},
			},
		},
	}
}

// createHTTPBodyReadFailureObservation creates an observation for cases where
// reading the HTTP request body for a CosmosSDK service request has failed.
//
// This observation:
// - Includes the chainID and detailed error information
// - Is useful for diagnosing connectivity or HTTP parsing issues
//
// Parameters:
// - chainID: The CosmosSDK chain identifier for which the request was intended
// - err: The error that occurred during HTTP body reading
//
// Returns:
// - qosobservations.Observations: A structured observation containing details about the HTTP read failure
func createHTTPBodyReadFailureObservation(
	serviceID protocol.ServiceID,
	chainID string,
	err error,
) *qosobservations.Observations_Cosmos {
	return &qosobservations.Observations_Cosmos{
		Cosmos: &qosobservations.CosmosSDKRequestObservations{
			RouteRequest: "HTTP body read failed",
			EndpointObservations: []*qosobservations.CosmosSDKEndpointObservation{
				{
					ResponseObservation: &qosobservations.CosmosSDKEndpointObservation_UnrecognizedResponse{
						UnrecognizedResponse: &qosobservations.CosmosSDKUnrecognizedResponse{
							JsonrpcResponse: &qosobservations.JsonRpcResponse{
								Id: "",
							},
						},
					},
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
		logger.Error().Err(err).Msgf("❌ Request failed JSON-RPC validation - returning generic error response. Request preview: %s", requestPreview)
	}

	return jsonrpcRequest, err
}
