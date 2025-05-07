package judge

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// - Struct name: QualityCheckContext
// - Struct Methods:
//   - GetState(): e.g. for Archival checks
//   - AddCheck(jsonrpc.Request)

type EndpointQualityChecksContext struct {
	logger polylog.Logger

	// Service State (read-only)
	// Allows the custom QoS service to base the endpoint checks on current state.
	// Includes the endpoint store in read-only mode.
	*ServiceState

	// Endpoint loaded from the endpoint store.
	endpoint *Endpoint

	// Custom service's Endpoint Checks function
	endpointChecksBuilder EndpointQualityChecksBuilder

	endpointChecksToPerform []*jsonrpc.Request
}

func (ctx *EndpointQualityChecksContext) buildEndpointQualityCheckContexts() []gateway.RequestQoSContext {
	jsonrpcRequestsToSend := ctx.endpointChecksBuilder(ctx)

	var qualityCheckContexts []gateway.RequestQoSContext
	for _, jsonrpcReq := range jsonrpcRequestsToSend {
		// new request context for the quality check
		requestCtx := &requestQoSContext{
			logger: ctx.logger,
		}

		// initialize the context using the JSONRPC request required for endpoint quality check.
		requestCtx.initFromJSONRPCRequest(jsonrpcReq)

		qualityCheckContexts = append(qualityCheckContexts, requestCtx)
	}

	return qualityCheckContexts
}

func (ctx *EndpointQualityChecksContext) GetEndpoint() *Endpoint {
	return ctx.endpoint
}

func (ctx *EndpointQualityChecksContext) AddQualityCheck(jsonrpcReq *jsonrpc.Request) {
	ctx.endpointChecksToPerform = append(ctx.endpointChecksToPerform, jsonrpcReq)
}

// TODO_IN_THIS_PR: pick a more descriptive/fluent API name.
func (ctx *EndpointQualityChecksContext) Build() []*jsonrpc.Request {
	return ctx.endpointChecksToPerform
}
