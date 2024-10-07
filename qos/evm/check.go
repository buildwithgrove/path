package evm

import (
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/buildwithgrove/path/relayer"
)

const (
	// Each endpoint check should use its own ID to avoid potential conflicts.
	// ID of JSONRPC requests for any new checks should be added to the list below.
	_              = iota
	idChainIDCheck = 1000 + iota
	idBlockNumberCheck
)

// EndpointStore provides the endpoint check generator required by
// the gateway package to augment endpoints' quality data,
// using synthetic service requests.
var _ gateway.QoSEndpointCheckGenerator = &EndpointStore{}

func (es *EndpointStore) GetRequiredQualityChecks(endpointAddr relayer.EndpointAddr) []gateway.RequestQoSContext {
	// TODO_IMPROVE: skip any checks for which the endpoint already has
	// a valid (e.g. not expired) quality data point.

	return []gateway.RequestQoSContext{
		getEndpointCheck(endpointAddr, withChainIDCheck),
		getEndpointCheck(endpointAddr, withBlockHeightCheck),
		// TODO_FUTURE: add an archival endpoint check.
	}
}

func getEndpointCheck(endpointAddr relayer.EndpointAddr, options ...func(*requestContext)) *requestContext {
	requestCtx := requestContext{
		isValid:                 true,
		preSelectedEndpointAddr: endpointAddr,
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
