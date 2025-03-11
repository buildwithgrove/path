package evm

import (
	"encoding/json"
	"net/http"

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
	logger     polylog.Logger
	jsonrpcReq jsonrpc.Request
}

// GetObservation returns an observation indicating the endpoint returned an empty response.
// Implements the response interface.
func (r responseEmpty) GetObservation() qosobservations.EVMEndpointObservation {
	validationError := qosobservations.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_EMPTY

	return qosobservations.EVMEndpointObservation{
		ResponseObservation: &qosobservations.EVMEndpointObservation_EmptyResponse{
			EmptyResponse: &qosobservations.EVMEmptyResponse{
				Valid:                   false, // Empty responses are inherently invalid - explicitly set for clarity
				ResponseValidationError: &validationError,
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
		httpStatusCode: http.StatusInternalServerError,
	}
}

// getResponsePayload constructs a JSONRPC error response indicating endpoint failure.
// Uses request ID in response per JSONRPC spec: https://www.jsonrpc.org/specification#response_object
func (r responseEmpty) getResponsePayload() []byte {
	userResponse := newErrResponseEmptyEndpointResponse(r.jsonrpcReq.ID)
	bz, err := json.Marshal(userResponse)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.logger.Warn().Err(err).Msg("responseEmpty: Marshaling JSONRPC response failed.")
	}
	return bz
}
