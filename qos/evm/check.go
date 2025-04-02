package evm

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	// Each endpoint check should use its own ID to avoid potential conflicts.
	// ID of JSON-RPC requests for any new checks should be added to the list below.
	_              = iota
	idChainIDCheck = 1000 + iota
	idBlockNumberCheck
)

// EndpointStore provides the endpoint check generator required by
// the gateway package to augment endpoints' quality data,
// using synthetic service requests.
var _ gateway.QoSEndpointCheckGenerator = &EndpointStore{}

// TODO_IMPROVE(@commoddity): implement QoS check expiry functionality and use protocol.EndpointAddr
// to filter out checks for any endpoint which has acurrently valid QoS data point.
func (es *EndpointStore) GetRequiredQualityChecks(_ protocol.EndpointAddr) []gateway.RequestQoSContext {
	return []gateway.RequestQoSContext{
		getEndpointCheck(es.logger, es, withChainIDCheck),
		getEndpointCheck(es.logger, es, withBlockHeightCheck),
		// TODO_FUTURE: add an archival endpoint check.
	}
}

// getEndpointCheck prepares a request context for a specific endpoint check.
func getEndpointCheck(
	logger polylog.Logger,
	endpointStore *EndpointStore,
	options ...func(*requestContext),
) *requestContext {
	requestCtx := requestContext{
		logger:        logger,
		endpointStore: endpointStore,
	}

	for _, option := range options {
		option(&requestCtx)
	}

	return &requestCtx
}

// withChainIDCheck updates the request context to make an EVM JSON-RPC eth_chainId request.
func withChainIDCheck(requestCtx *requestContext) {
	requestCtx.jsonrpcReq = buildJSONRPCReq(idChainIDCheck, methodChainID)
}

// withBlockHeightCheck updates the request context to make an EVM JSON-RPC eth_blockNumber request.
func withBlockHeightCheck(requestCtx *requestContext) {
	requestCtx.jsonrpcReq = buildJSONRPCReq(idBlockNumberCheck, methodBlockNumber)
}

func buildJSONRPCReq(id int, method jsonrpc.Method) jsonrpc.Request {
	return jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(id),
		Method:  method,
	}
}
