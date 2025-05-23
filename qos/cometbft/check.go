package cometbft

import (
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
)

// EndpointStore provides the endpoint check generator required by
// the gateway package to augment endpoints' quality data,
// using synthetic service requests.
var _ gateway.QoSEndpointCheckGenerator = &EndpointStore{}

// TODO_IMPROVE(@commoddity): implement QoS check expiry functionality and use protocol.EndpointAddr
// to filter out checks for any endpoint which has acurrently valid QoS data point.
func (es *EndpointStore) GetRequiredQualityChecks(_ protocol.EndpointAddr) []gateway.RequestQoSContext {
	// TODO_IMPROVE(@adshmh): skip any checks for which the endpoint already has
	// a valid (i.e. not expired) QoS data point.

	return []gateway.RequestQoSContext{
		getEndpointCheck(es.logger, es, withHealthCheck),
		getEndpointCheck(es.logger, es, withStatusCheck),
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

// withHealthCheck updates the request context to make a CometBFT GET /health-check request.
func withHealthCheck(requestCtx *requestContext) {
	request, _ := http.NewRequest(http.MethodGet, apiPathHealthCheck, nil)
	requestCtx.httpReq = request
}

// withStatusCheck updates the request context to make a CometBFT GET /status request.
func withStatusCheck(requestCtx *requestContext) {
	request, _ := http.NewRequest(http.MethodGet, apiPathStatus, nil)
	requestCtx.httpReq = request
}
