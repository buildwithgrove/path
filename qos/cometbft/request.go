package cometbft

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
