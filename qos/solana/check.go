package solana

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
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

// EndpointStore provides the endpoint check generator required by
// the gateway package to augment endpoints' quality data,
// using synthetic service requests.
var _ gateway.QoSEndpointCheckGenerator = &EndpointStore{}

func (es *EndpointStore) GetRequiredQualityChecks(endpointAddr protocol.EndpointAddr) []gateway.RequestQoSContext {
	// TODO_IMPROVE: skip any checks for which the endpoint already has
	// a valid (e.g. not expired) quality data point.

	return []gateway.RequestQoSContext{
		getEndpointCheck(es.Logger, endpointAddr, es, withGetHealth),
		getEndpointCheck(es.Logger, endpointAddr, es, withGetEpochInfo),
		// TODO_MVP(@adshmh): Add a check for a `getBlock` request
	}
}

func getEndpointCheck(
	logger polylog.Logger,
	endpointAddr protocol.EndpointAddr,
	endpointStore *EndpointStore,
	options ...func(*requestContext),
) *requestContext {
	requestCtx := requestContext{
		EndpointStore:           endpointStore,
		Logger:                  logger,
		isValid:                 true,
		preSelectedEndpointAddr: endpointAddr,
	}

	for _, option := range options {
		option(&requestCtx)
	}

	return &requestCtx
}

func withGetHealth(requestCtx *requestContext) {
	requestCtx.JSONRPCReq = buildJSONRPCReq(idGetHealth, methodGetHealth)
}

func withGetEpochInfo(requestCtx *requestContext) {
	requestCtx.JSONRPCReq = buildJSONRPCReq(idGetEpochInfo, methodGetEpochInfo)
}

func buildJSONRPCReq(id int, method jsonrpc.Method) jsonrpc.Request {
	return jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(id),
		Method:  method,
	}
}
