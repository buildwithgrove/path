package qos

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// RequestContextFromProtocolError returns a request context for a protocol error, e.g. the selected endpoint did not return a response.
func RequestContextFromInternalError(logger polylog.Logger, jsonrpcRequestID jsonrpc.ID, err error) *RequestErrorContext {
	return &RequestErrorContext{
		Logger:   logger,
		Response: jsonrpc.NewErrResponseInternalErr(jsonrpcRequestID, err),
	}
}
