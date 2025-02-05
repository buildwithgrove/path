package cometbft

import (
	"net/http"

	"github.com/buildwithgrove/path/qos"
)

func getEndpointCheck(endpointStore *qos.EndpointStore, options ...func(*requestContext)) *requestContext {
	requestCtx := requestContext{
		endpointStore: endpointStore,
		isValid:       true,
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
