package evm

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// emptyResponse provides the functionality required from a response by a requestContext instance.
var _ response = responseEmpty{}

// TODO_MVP(@adshmh): Implement request retry support:
//  1. Add ShouldRetry() method to gateway.RequestQoSContext
//  2. Integrate ShouldRetry() into gateway request handler
//  3. Extend evm.response interface with ShouldRetry()
//  4. Add ShouldRetry() to evm.requestContext to evaluate retry eligibility based on responses
//
// responseEmpty processes empty endpoint responses by:
//  1. Creating an observation to penalize the endpoint and track metrics
//  2. Generating a JSONRPC error to return to the client
type responseEmpty struct {
	logger      polylog.Logger
	jsonrpcReqs map[string]jsonrpc.Request
}

// GetObservation returns an observation indicating the endpoint returned an empty response.
// Implements the response interface.
func (r responseEmpty) GetObservation() qosobservations.EVMEndpointObservation {

	return qosobservations.EVMEndpointObservation{
		ResponseObservation: &qosobservations.EVMEndpointObservation_EmptyResponse{
			EmptyResponse: &qosobservations.EVMEmptyResponse{
				HttpStatusCode: int32(r.getHTTPStatusCode()), // EmptyResponse always returns a 500 Internal error HTTP status code.
				// An empty response is always invalid.
				ResponseValidationError: qosobservations.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_EMPTY,
			},
		},
	}
}

// GetHTTPResponse builds and returns the httpResponse matching the responseEmpty instance.
// Implements the response interface.
func (r responseEmpty) GetHTTPResponse() httpResponse {
	return httpResponse{
		responsePayload: r.getResponsePayload(),
		// HTTP Status 500 Internal Server Error for an empty response
		httpStatusCode: r.getHTTPStatusCode(),
	}
}

// getResponsePayload constructs a JSONRPC error response indicating endpoint failure.
// Uses request ID in response per JSONRPC spec: https://www.jsonrpc.org/specification#response_object
func (r responseEmpty) getResponsePayload() []byte {
	userResponse := newErrResponseEmptyEndpointResponse(getJsonRpcIDForErrorResponse(r.jsonrpcReqs))
	bz, err := json.Marshal(userResponse)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.logger.Warn().Err(err).Msg("responseEmpty: Marshaling JSONRPC response failed.")
	}
	return bz
}

// getHTTPStatusCode returns the HTTP status code to be returned to the client.
// Always returns returns 500 Internal Server Error on responseEmpty struct.
func (r responseEmpty) getHTTPStatusCode() int {
	return httpStatusResponseValidationFailureEmptyResponse
}

// GetJSONRPCID returns the JSONRPC ID of the response.
// Implements the response interface.
func (r responseEmpty) GetJSONRPCID() jsonrpc.ID {
	return getJsonRpcIDForErrorResponse(r.jsonrpcReqs)
}
