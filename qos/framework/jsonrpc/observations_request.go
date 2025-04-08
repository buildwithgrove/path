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

	// Only set if validation was successful
	JsonrpcRequest *JsonRpcRequest
	// Only set if validation failed
	ValidationFailure *ValidationFailure
		ErrorType              RequestValidationError
		ValidationErrorDetails
		ErrorResponse *jsonrpc.Response
}
