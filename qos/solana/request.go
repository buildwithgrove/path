package solana

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// requestContextFromInternalError returns a request context
// for an internal error, e.g. error on reading the HTTP request body.
func requestContextFromInternalError(logger polylog.Logger, err error) *qos.RequestErrorContext {
	return &qos.RequestErrorContext{
		Logger:   logger,
		Response: jsonrpc.NewErrResponseInternalErr(jsonrpc.ID{}, err),
	}
}

// requestContextFromUserError returns a request context
// for a user error, e.g. an unmarshaling error is a
// user error because the request body, provided by the user,
// cannot be parsed as a valid JSONRPC request.
func requestContextFromUserError(logger polylog.Logger, err error) *qos.RequestErrorContext {
	return &qos.RequestErrorContext{
		Logger:   logger,
		Response: jsonrpc.NewErrResponseInternalErr(jsonrpc.ID{}, err),
	}
}
