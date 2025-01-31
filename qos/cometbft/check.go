package cometbft

import (
	"net/http"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
)

// EndpointStore provides the endpoint check generator required by
// the gateway package to augment endpoints' quality data,
// using synthetic service requests.
var _ gateway.QoSEndpointCheckGenerator = &EndpointStore{}

func (es *EndpointStore) GetRequiredQualityChecks(endpointAddr protocol.EndpointAddr) []gateway.RequestQoSContext {
	// TODO_IMPROVE: skip any checks for which the endpoint already has
	// a valid (e.g. not expired) quality data point.

	return []gateway.RequestQoSContext{
		getEndpointCheck(es, endpointAddr, withHealthCheck),
		getEndpointCheck(es, endpointAddr, withBlockHeightCheck),
	}
}

func getEndpointCheck(endpointStore *EndpointStore, endpointAddr protocol.EndpointAddr, options ...func(*requestContext)) *requestContext {
	requestCtx := requestContext{
		endpointStore:           endpointStore,
		isValid:                 true,
		preSelectedEndpointAddr: endpointAddr,
	}

	for _, option := range options {
		option(&requestCtx)
	}

	return &requestCtx
}

func withHealthCheck(requestCtx *requestContext) {
	request, _ := http.NewRequest(http.MethodGet, apiPathHealthCheck, nil)
	requestCtx.httpReq = request
}

// withBlockHeightCheck
func withBlockHeightCheck(requestCtx *requestContext) {
	request, _ := http.NewRequest(http.MethodGet, apiPathBlockHeight, nil)
	requestCtx.httpReq = request
}
