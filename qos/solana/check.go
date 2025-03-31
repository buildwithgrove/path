package solana

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	// Each endpoint check should use its own ID to avoid potential conflicts.
	// ID of JSON-RPC requests for any new checks should be added to the list below.
	_           = iota
	idGetHealth = 1000 + iota
	idGetEpochInfo
	idGetBlock
)

// EndpointStore provides the endpoint check generator required by
// the gateway package to augment endpoints' quality data,
// using synthetic service requests.
var _ gateway.QoSEndpointCheckGenerator = &EndpointStore{}

func (es *EndpointStore) GetRequiredQualityChecks() []gateway.RequestQoSContext {
	// TODO_IMPROVE(@adshmh): skip any checks for which the endpoint already has
	// a valid (i.e. not expired) QoS data point.

	return []gateway.RequestQoSContext{
		getEndpointCheck(es.logger, es, withGetHealth),
		getEndpointCheck(es.logger, es, withGetEpochInfo),
		// TODO_MVP(@adshmh): Add a check for a `getBlock` request
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
		isValid:       true,
	}

	for _, option := range options {
		option(&requestCtx)
	}

	return &requestCtx
}

// withGetHealth updates the request context to make a Solana JSON-RPC getHealth request.
func withGetHealth(requestCtx *requestContext) {
	requestCtx.JSONRPCReq = buildJSONRPCReq(idGetHealth, methodGetHealth)
}

// withGetEpochInfo updates the request context to make a Solana JSON-RPC getEpochInfo request.
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
