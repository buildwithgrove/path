package qos

import (
	"errors"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

func GetRequestErrorForProtocolError() *qosobservations.RequestError {
	err := errors.New("internal error: protocol error: no endpoint responses received.")
	// initialize a JSONRPC error response to derive the HTTP status code.
	jsonrpcErrorResponse := jsonrpc.NewErrResponseInternalErr(jsonrpc.ID{}, err)

	return &qosobservations.RequestError{
		// Protocol-level error: e.g. selected endpoint timed out.
		ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_INTERNAL_PROTOCOL_ERROR,
		ErrorDetails:   err.Error(),
		HttpStatusCode: int32(jsonrpcErrorResponse.GetRecommendedHTTPStatusCode()),
	}
}
