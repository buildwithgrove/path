package evm

import (
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

// endpointCheckName is a type for the names of the checks applied to an endpoint.
type endpointCheckName string

// check is an interface for the checks applied to an endpoint.
// It is embedded in the struct that satisfies the gateway.QualityCheck interface.
type check interface {
	name() endpointCheckName
	isValid(serviceState *ServiceState) error
	shouldRun() bool
}

// EndpointStore provides the endpoint check generator required by
// the gateway package to augment endpoints' quality data,
// using synthetic service requests.
var _ gateway.QoSEndpointCheckGenerator = &EndpointStore{}

// evmQualityCheck provides:
//  1. The validity and expiry of the check.
//  2. The request context used to perform a quality check.
//
// An evmQualityCheck may have an empty request context if the check
// derives its validity from applying other observations.
// For example: if the check is for an empty response to any request.
type evmQualityCheck struct {
	check
	requestContext *requestContext
}

func (q *evmQualityCheck) shouldRun() bool {
	return q.requestContext != nil && q.check.shouldRun()
}

func (q *evmQualityCheck) getRequestContext() *requestContext {
	return q.requestContext
}

// GetRequiredQualityChecks returns the list of quality checks required for an endpoint.
// It is called in the `gateway/hydrator.go` file on each run of the hydrator.
func (es *EndpointStore) GetRequiredQualityChecks(endpointAddr protocol.EndpointAddr) []gateway.RequestQoSContext {
	es.endpointsMu.RLock()
	endpoint, ok := es.endpoints[endpointAddr]
	es.endpointsMu.RUnlock()

	if !ok {
		endpoint = newEndpoint(es)
	}

	return endpoint.getChecks(endpointAddr)
}

// getEndpointCheck prepares a request context for a specific endpoint check.
// The pre-selected endpoint address is assigned to the request context in the `endpoint.getChecks` method.
// It is called in the individual `check_*.go` files to build the request context.
func getEndpointCheck(endpointStore *EndpointStore, options ...func(*requestContext)) *requestContext {
	requestCtx := requestContext{
		logger:        endpointStore.logger,
		endpointStore: endpointStore,
		isValid:       true,
	}

	for _, option := range options {
		option(&requestCtx)
	}

	return &requestCtx
}

// buildJSONRPCReq builds a JSON-RPC request with the given ID and method.
// It is called in the individual `check_*.go` files to build the request context.
func buildJSONRPCReq(id int, method jsonrpc.Method) jsonrpc.Request {
	return jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(id),
		Method:  method,
	}
}
