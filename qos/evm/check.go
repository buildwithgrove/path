package evm

import (
	"time"

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

type endpointCheckName string

type evmEndpointCheck interface {
	CheckName() string
	IsValid(serviceState *ServiceState) error
	ExpiresAt() time.Time
}

var (
	// EndpointStore provides the endpoint check generator required by
	// the gateway package to augment endpoints' quality data,
	// using synthetic service requests.
	_ gateway.QoSEndpointCheckGenerator = &EndpointStore{}
	// evmQualityCheck implements the QualityCheck interface for EVM-based endpoints.
	_ gateway.QualityCheck = &evmQualityCheck{}
)

type evmQualityCheck struct {
	evmEndpointCheck
	requestContext *requestContext
}

func (q *evmQualityCheck) GetRequestContext() gateway.RequestQoSContext {
	return q.requestContext
}

func (q *evmQualityCheck) EndpointAddr() protocol.EndpointAddr {
	return q.requestContext.preSelectedEndpointAddr
}

func (es *EndpointStore) GetRequiredQualityChecks(endpointAddr protocol.EndpointAddr) []gateway.QualityCheck {
	endpoint, ok := es.endpoints[endpointAddr]
	if !ok {
		endpoint = newEndpoint()
	}

	return []gateway.QualityCheck{
		&evmQualityCheck{
			requestContext:   getEndpointCheck(es.logger, es, endpointAddr, withChainIDCheck),
			evmEndpointCheck: endpoint.checks[endpointCheckNameChainID],
		},
		&evmQualityCheck{
			requestContext:   getEndpointCheck(es.logger, es, endpointAddr, withBlockHeightCheck),
			evmEndpointCheck: endpoint.checks[endpointCheckNameBlockHeight],
		},
	}
}

// getEndpointCheck prepares a request context for a specific endpoint check.
func getEndpointCheck(
	logger polylog.Logger,
	endpointStore *EndpointStore,
	endpointAddr protocol.EndpointAddr,
	options ...func(*requestContext),
) *requestContext {
	requestCtx := requestContext{
		logger:                  logger,
		endpointStore:           endpointStore,
		isValid:                 true,
		preSelectedEndpointAddr: endpointAddr,
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
