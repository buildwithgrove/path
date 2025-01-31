package evm

import (
	"github.com/buildwithgrove/path/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	// Each endpoint check should use its own ID to avoid potential conflicts.
	// ID of JSONRPC requests for any new checks should be added to the list below.
	_              = iota
	idChainIDCheck = 1000 + iota
	idBlockNumberCheck
)

func getEndpointCheck(endpointStore *qos.EndpointStore, options ...func(*requestContext)) *requestContext {
	requestCtx := requestContext{
		endpointStore: endpointStore,
		isValid:       true,
	}

	for _, option := range options {
		option(&requestCtx)
	}

	return &requestCtx
}

func withChainIDCheck(requestCtx *requestContext) {
	requestCtx.jsonrpcReq = jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(idChainIDCheck),
		Method:  methodChainID,
	}
}

func withBlockHeightCheck(requestCtx *requestContext) {
	requestCtx.jsonrpcReq = jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(idBlockNumberCheck),
		Method:  methodBlockNumber,
	}
}
