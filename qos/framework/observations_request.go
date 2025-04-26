package jsonrpc

import (
	observations "github.com/buildwithgrove/path/observation/qos/framework"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_IN_THIS_PR: ensure the observations will contain:
// - HTTP Status code: e.g. httpStatusRequestValidationFailureUnmarshalFailure,
// - Validation error: e.g. qosobservations.EVMRequestValidationError_EVM_REQUEST_VALIDATION_ERROR_REQUEST_UNMARSHALING_FAILURE,
// - Error details.
//
// buildObservations builds and returns request observations of of the requestDetails struct.
func (rd *requestDetails) buildObservations() *qosobservations.RequestObservation {
	// build a JSONRPC request observation, if one was parsed.
	var jsonrpcRequestObs *qosobservations.JsonrpcRequest
	if rd.request != nil {
		jsonrpcRequestObs = rd.request.GetObservation(),
	}

	// build request failure observation, if the request parsing failed.
	var errorObs *qosobservations.RequestError
	if rd.requestError != nil {
		errorObs = rd.requestError.buildObservations()
	}

	return &qosobservations.RequestObservation {
		// Only set if validation was successful
		JsonrpcRequest: jsonrpcRequestObs,
		// Only set if the request failed for any reason.
		// A valid request can still fail due to a gateway internal error, e.g.:
		// - error reading HTTP request's body.
		// - protocol-level error: e.g. selected endpoint timed out.
		RequestError: errorObs,
	}
}
