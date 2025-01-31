package solana

import (
	"github.com/buildwithgrove/path/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	// Each endpoint check should use its own ID to avoid potential conflicts.
	// ID of JSONRPC requests for any new checks should be added to the list below.
	_           = iota
	idGetHealth = 1000 + iota
	idGetEpochInfo
	idGetBlock
)

func getEndpointCheck(endpointStore *qos.EndpointStore, options ...func(*requestContext),
) *requestContext {
	requestCtx := requestContext{
		endpointStore: endpointStore,
		isValid:       true,
	}

	for _, option := range options {
		option(&requestCtx)
	}

	return &requestCtx
}

func withGetHealth(requestCtx *requestContext) {
	requestCtx.jsonrpcReq = buildJSONRPCReq(idGetHealth, methodGetHealth)
}

func withGetEpochInfo(requestCtx *requestContext) {
	requestCtx.jsonrpcReq = buildJSONRPCReq(idGetEpochInfo, methodGetEpochInfo)
}

func buildJSONRPCReq(id int, method jsonrpc.Method) jsonrpc.Request {
	return jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(id),
		Method:  method,
	}
}
