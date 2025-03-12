package evm

import (
	"encoding/json"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseNone satisfies the response interface and handles the case
// where no response has been received from any endpoint.
// This differs from empty responses (where an endpoint responded with empty data),
// as this represents a case where no endpoint communication occurred at all.
var _ response = responseNone{}

// responseNone represents the absence of any endpoint response.
// This can occur due to protocol-level failures or when no endpoint was selected.
type responseNone struct {
	logger     polylog.Logger
	jsonrpcReq jsonrpc.Request
}

// GetObservation returns an observation indicating no endpoint provided a response.
// This allows tracking metrics for scenarios where endpoint selection or communication failed.
// Implements the response interface.
func (r responseNone) GetObservation() qosobservations.EVMEndpointObservation {
	validationError := qosobservations.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_NO_RESPONSE

	return qosobservations.EVMEndpointObservation{
		ResponseObservation: &qosobservations.EVMEndpointObservation_NoResponse{
			NoResponse: &qosobservations.EVMNoResponse{
				HttpStatusCode:          r.getHTTPStatusCode(), // NoResponse always returns a 500 Internal error HTTP status code.
				ResponseValidationError: &validationError,      // NoResponse is always an invalid response.
			},
		},
	}
}

// GetHTTPResponse creates and returns a predefined httpResponse for cases when QoS has received no responses from the protocol.
// Implements the response interface.
func (r responseNone) GetHTTPResponse() httpResponse {
	return httpResponse{
		responsePayload: r.getResponsePayload(),
		httpStatusCode:  r.getHTTPStatusCode(),
	}
}

// getResponsePayload constructs a JSONRPC error response indicating no endpoint response was received.
// Uses request ID in response per JSONRPC spec: https://www.jsonrpc.org/specification#response_object
func (r responseNone) getResponsePayload() []byte {
	userResponse := newErrResponseNoEndpointResponse(r.jsonrpcReq.ID)
	bz, err := json.Marshal(userResponse)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.logger.Warn().Err(err).Msg("responseNone: Marshaling JSONRPC response failed.")
	}
	return bz
}

// getHTTPStatusCode returns the HTTP status code to be returned to the client.
// Always returns returns 500 Internal Server Error on responseNone struct.
func (r responseNone) getHTTPStatusCode() int {
	// TODO_IN_THIS_PR: decide whether 503 Service Unavailable - is more appropriate for no endpoint response
	return http.StatusInternalServerError
}
