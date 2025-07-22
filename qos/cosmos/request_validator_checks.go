package cosmos

import (
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

	// TODO_NEXT(@commoddity): Add endpoint checks for the following:
	//
	//  1. CosmosSDK URL paths (sharedtypes.RPCType_REST):
	//     - Node Info (/cosmos/base/tendermint/v1beta1/node_info)
	//     https://docs.cosmos.network/api#tag/Service/operation/GetNodeInfo
	//     - Syncing Status (/cosmos/base/tendermint/v1beta1/syncing)
	//     https://docs.cosmos.network/api#tag/Service/operation/GetSyncing
	//
	//  2. EVM JSON-RPC methods (sharedtypes.RPCType_JSON_RPC):
	//     - `{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}`
	//     - `{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`

	return checks
}

// getCometBFTEndpointChecks generates the endpoint checks for the CometBFT RPC type.
// API reference: https://docs.cometbft.com/v1.0/rpc/
func (rv *requestValidator) getCometBFTEndpointChecks(endpoint endpoint) []gateway.RequestQoSContext {
	checks := []gateway.RequestQoSContext{}

	if rv.shouldHealthCheckRun(endpoint.checkHealth) {
		checks = append(checks, rv.getJSONRPCRequestContextFromRequest(
			sharedtypes.RPCType_COMET_BFT,
			endpoint.checkHealth.getRequest(),
		))
	}

	if rv.shouldStatusCheckRun(endpoint.checkStatus) {
		checks = append(checks, rv.getJSONRPCRequestContextFromRequest(
			sharedtypes.RPCType_COMET_BFT,
			endpoint.checkStatus.getRequest(),
		))
	}

	return checks
}

// shouldHealthCheckRun returns true if the health check is not yet initialized or has expired.
func (rv *requestValidator) shouldHealthCheckRun(check endpointCheckHealth) bool {
	return check.expiresAt.IsZero() || check.IsExpired()
}

// shouldStatusCheckRun returns true if the status check is not yet initialized or has expired.
func (rv *requestValidator) shouldStatusCheckRun(check endpointCheckStatus) bool {
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
