package evm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

func (qos *QoS) ParseHTTPRequest(_ context.Context, req *http.Request) (gateway.RequestQoSContext, bool) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return requestContextFromInternalError(err), false
	}

	var jsonrpcReq jsonrpc.Request
	if err := json.Unmarshal(body, &jsonrpcReq); err != nil {
		return requestContextFromUserError(err), false
	}

	// TODO_IMPROVE: method-specific validation of the JSONRPC request.
	return &requestContext{
		endpointStore: qos.endpointStore,
		jsonrpcReq:    jsonrpcReq,
		isValid:       true,
	}, true
}

// TODO_UPNEXT(@adshmh): return a request context to handle internal errors.
// requestContextFromInternalError returns a request context
// for an internal error, e.g. error on reading the HTTP request body.
func requestContextFromInternalError(err error) *requestContext {
	return nil
}

// TODO_UPNEXT(@adshmh): return a request context to handle user errors.
// requestContextFromUserError returns a request context
// for a user error, e.g. an unmarshalling error is a
// user error because the request body, provided by the user,
// cannot be parsed as a valid JSONRPC request.
func requestContextFromUserError(err error) *requestContext {
	return nil
}
