package cosmos

import (
	"io"
	"net/http"

	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

/* -------------------- QoS Endpoint Check Generator -------------------- */

// requestValidator provides the required synthetic QoS checks to the hydrator.
//
// This responsibility lies with the request validator because it is the component
// that generates the request contexts for both JSONRPC and REST requests.
var _ gateway.QoSEndpointCheckGenerator = &requestValidator{}

// GetRequiredQualityChecks returns the list of quality checks required for an endpoint.
// It is called in the `gateway/hydrator.go` file on each run of the hydrator.
func (rv *requestValidator) GetRequiredQualityChecks(endpointAddr protocol.EndpointAddr) []gateway.RequestQoSContext {
	// Get the endpoint from the endpoint store.
	endpoint := rv.serviceState.endpointStore.getEndpoint(endpointAddr)

	// Get the RPC types supported by the CosmosSDK service.
	supportedAPIs := rv.serviceState.serviceQoSConfig.getSupportedAPIs()

	// List of all synthetic QoS checks required for the endpoint.
	var checks []gateway.RequestQoSContext

	// If the service supports CometBFT, add the CometBFT endpoint checks.
	if _, ok := supportedAPIs[sharedtypes.RPCType_COMET_BFT]; ok {
		checks = append(checks, rv.getCometBFTEndpointChecks(endpoint)...)
	}

	// If the service supports CosmosSDK, add the CosmosSDK endpoint checks.
	if _, ok := supportedAPIs[sharedtypes.RPCType_REST]; ok {
		checks = append(checks, rv.getCosmosSDKEndpointChecks(endpoint)...)
	}

	// TODO_NEXT(@commoddity): Add endpoint checks for EVM JSON-RPC methods (sharedtypes.RPCType_JSON_RPC):
	//     - `{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}`
	//     - `{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`

	return checks
}

// getCometBFTEndpointChecks generates the endpoint checks for the CometBFT RPC type.
// API reference: https://docs.cometbft.com/v1.0/rpc/
func (rv *requestValidator) getCometBFTEndpointChecks(endpoint endpoint) []gateway.RequestQoSContext {
	checks := []gateway.RequestQoSContext{}

	if rv.shouldCometBFTHealthCheckRun(endpoint.checkCometBFTHealth) {
		checks = append(checks, rv.getJSONRPCRequestContextFromRequest(
			sharedtypes.RPCType_COMET_BFT,
			endpoint.checkCometBFTHealth.getRequest(),
		))
	}

	if rv.shouldCometBFTStatusCheckRun(endpoint.checkCometBFTStatus) {
		checks = append(checks, rv.getJSONRPCRequestContextFromRequest(
			sharedtypes.RPCType_COMET_BFT,
			endpoint.checkCometBFTStatus.getRequest(),
		))
	}

	return checks
}

// getCosmosSDKEndpointChecks generates the endpoint checks for the CosmosSDK RPC type.
// API reference: https://docs.cosmos.network/api
func (rv *requestValidator) getCosmosSDKEndpointChecks(endpoint endpoint) []gateway.RequestQoSContext {
	// Cosmos SDK status check should always be run.
	checks := []gateway.RequestQoSContext{
		rv.getRESTRequestContextFromRequest(
			sharedtypes.RPCType_REST,
			endpoint.checkCosmosStatus.getRequest(),
		),
	}

	return checks
}

// shouldCometBFTHealthCheckRun returns true if the health check is not yet initialized or has expired.
func (rv *requestValidator) shouldCometBFTHealthCheckRun(check endpointCheckCometBFTHealth) bool {
	return check.expiresAt.IsZero() || check.IsExpired()
}

// shouldCometBFTStatusCheckRun returns true if the status check is not yet initialized or has expired.
func (rv *requestValidator) shouldCometBFTStatusCheckRun(check endpointCheckCometBFTStatus) bool {
	return check.expiresAt.IsZero() || check.IsExpired()
}

// getJSONRPCRequestContextFromRequest prepares a gateway request context for a JSONRPC QoS endpoint check.
func (rv *requestValidator) getJSONRPCRequestContextFromRequest(
	rpcType sharedtypes.RPCType,
	jsonrpcReq jsonrpc.Request,
) gateway.RequestQoSContext {
	context, ok := rv.buildJSONRPCRequestContext(
		rpcType,
		jsonrpcReq,
		qosobservations.RequestOrigin_REQUEST_ORIGIN_SYNTHETIC,
	)
	if !ok {
		rv.logger.Error().Msg("SHOULD NEVER HAPPEN: failed to build JSONRPC request context")
	}
	return context
}

// TODO_NEXT(@commoddity): Add getRESTRequestContextFromRequest method for generating Cosmos SDK quality checks.
func (rv *requestValidator) getRESTRequestContextFromRequest(
	rpcType sharedtypes.RPCType,
	restReq *http.Request,
) gateway.RequestQoSContext {
	var httpRequestBody []byte
	if restReq.Body != nil {
		bodyBytes, err := io.ReadAll(restReq.Body)
		if err != nil {
			rv.logger.Error().Err(err).Msg("failed to read request body")
			return nil
		}
		httpRequestBody = bodyBytes
	}

	context, ok := rv.buildRESTRequestContext(
		rpcType,
		restReq.URL,
		restReq.Method,
		httpRequestBody,
		qosobservations.RequestOrigin_REQUEST_ORIGIN_SYNTHETIC,
	)
	if !ok {
		rv.logger.Error().Msg("SHOULD NEVER HAPPEN: failed to build REST request context")
	}

	return context
}
