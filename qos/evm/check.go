package evm

import (
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/buildwithgrove/path/relayer"
)

var (
	idChainIDCheck     = jsonrpc.IDFromInt(1001)
	idBlockNumberCheck = jsonrpc.IDFromInt(1002)
)

// EndpointStore provides the endpoint check generator required by
// the gateway package to augment endpoints' quality data,
// using synthetic service requests.
var _ gateway.QoSEndpointCheckGenerator = &EndpointStore{}

func (es *EndpointStore) GetRequiredQualityChecks(endpointAddr relayer.EndpointAddr) []gateway.RequestQoSContext {
	// TODO_IMPROVE: skip any checks for which the endpoint already has
	// a valid (e.g. not expired) quality data point.
	requestCtx := requestContext{
		isValid:                 true,
		preSelectedEndpointAddr: endpointAddr,
	}

	return []gateway.RequestQoSContext{
		withChainIDCheck(requestCtx),
		withBlockHeightCheck(requestCtx),
		// TODO_FUTURE: add an archival endpoint check.
	}
}

func withChainIDCheck(requestCtx requestContext) *requestContext {
	requestCtx.jsonrpcReq = jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      idChainIDCheck,
		Method:  methodChainID,
	}

	return &requestCtx
}

func withBlockHeightCheck(requestCtx requestContext) *requestContext {
	requestCtx.jsonrpcReq = jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      idBlockNumberCheck,
		Method:  methodBlockNumber,
	}

	return &requestCtx
}
