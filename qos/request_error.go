package qos

import (
	"errors"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// GetRequestErrorForProtocolError returns a request error for a protocol error
// E.g. the selected endpoint did not return a response.
func GetRequestErrorForProtocolError() *qosobservations.RequestError {
	err := errors.New("internal error: protocol error: no endpoint responses received.")
	// initialize a JSONRPC error response to derive the HTTP status code.
	jsonrpcErrorResponse := jsonrpc.NewErrResponseInternalErr(jsonrpc.ID{}, err)

	// Protocol-level error
	// Examples:
	// - No endpoint responses received.
	// - Selected endpoint timed out.
	return &qosobservations.RequestError{
		ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_INTERNAL_PROTOCOL_ERROR,
		ErrorDetails:   err.Error(),
		HttpStatusCode: int32(jsonrpcErrorResponse.GetRecommendedHTTPStatusCode()),
	}
}
