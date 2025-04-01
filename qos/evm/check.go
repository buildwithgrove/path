package evm

import (
	"encoding/json"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// EndpointStore provides the endpoint check generator required by
// the gateway package to augment endpoints' quality data,
// using synthetic service requests.
var _ gateway.QoSEndpointCheckGenerator = &EndpointStore{}

const (
	// Each endpoint check should use its own ID to avoid potential conflicts.
	// ID of JSON-RPC requests for any new checks should be added to the list below.
	_              = iota
	idChainIDCheck = 1000 + iota
	idBlockNumberCheck
	idArchivalCheck
)

// check is an interface for the checks applied to an endpoint.
type evmQualityCheck interface {
	isValid(serviceState *ServiceState) error
	shouldRun() bool
	setRequestContext(*requestContext)
}

// GetRequiredQualityChecks returns the list of quality checks required for an endpoint.
// It is called in the `gateway/hydrator.go` file on each run of the hydrator.
func (es *EndpointStore) GetRequiredQualityChecks(endpointAddr protocol.EndpointAddr) []gateway.RequestQoSContext {
	es.endpointsMu.RLock()
	endpoint, ok := es.endpoints[endpointAddr]
	es.endpointsMu.RUnlock()

	// If the endpoint is not yet in the store, use an endpoint with the default empty checks.
	// e.g. if `GetRequiredQualityChecks` is called before the first observation is received for an endpoint.
	if !ok {
		endpoint = newEndpoint()
	}

	return endpoint.getChecks(es)
}

// getEndpointCheck prepares a request context for a specific endpoint check.
// The pre-selected endpoint address is assigned to the request context in the `endpoint.getChecks` method.
// It is called in the individual `check_*.go` files to build the request context.
func getEndpointCheck(es *EndpointStore, check evmQualityCheck) *requestContext {
	requestCtx := requestContext{
		logger:        es.logger,
		endpointStore: es,
	}

	check.setRequestContext(&requestCtx)

	return &requestCtx
}

func buildJSONRPCReq(id int, method jsonrpc.Method, params ...any) jsonrpc.Request {
	request := jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(id),
		Method:  method,
	}

	if len(params) > 0 {
		jsonParams, err := json.Marshal(params)
		if err == nil {
			request.Params = jsonrpc.NewParams(jsonParams)
		}
	}

	return request
}
