package cosmos

import (
	"errors"
	"fmt"

	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_TECHDEBT(@adshmh): Build the request context with necessary functions for websocket requests:
// - Do this for ALL QoS services: EVM, Cosmos, Solana, etc.
// - Validate the request's payload
// - Track the request to enable the validation of endpoint responses.
// - Revisit the request context's logic with regard to websockets:
// - Examples:
//   - Consider recreating the QoS context per endpoint message
//   - How to apply the observations for early detection of endpoint errors.
//
// TODO_IMPROVE(@commoddity): Add endpoint-level QoS checks to determine WebSocket support.
// Currently validates WebSocket upgrade requests at the service level only.

// validateWebsocketRequest validates WebSocket upgrade requests for Cosmos SDK services.
// Returns (requestContext, true) if WebSocket is supported.
// Returns (errorContext, false) if WebSocket is not configured for this service.
func (rv *requestValidator) validateWebsocketRequest() (gateway.RequestQoSContext, bool) {
	logger := rv.logger.With(
		"validator", "WebSocket",
		"method", "validateWebsocketRequest",
	)
	rpcType := sharedtypes.RPCType_WEBSOCKET
	logger = logger.With("rpc_type", rpcType.String())

	// Verify WebSocket support in service configuration
	if _, supported := rv.supportedAPIs[sharedtypes.RPCType_WEBSOCKET]; !supported {
		logger.Warn().Msg("Request uses unsupported WebSocket RPC type")
		return rv.createWebsocketUnsupportedRPCTypeContext(rpcType), false
	}

	// Build and return the request context
	return rv.buildWebsocketRequestContext(
		rpcType,
		qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
	), true
}

// buildWebsocketRequestContext builds a request context for WebSocket upgrade requests.
func (rv *requestValidator) buildWebsocketRequestContext(
	rpcType sharedtypes.RPCType,
	requestOrigin qosobservations.RequestOrigin,
) gateway.RequestQoSContext {
	logger := rv.logger.With(
		"method", "buildWebsocketRequestContext",
	)
	requestObservation := rv.buildWebsocketRequestObservations(
		rpcType,
		requestOrigin,
	)
	return &requestContext{
		logger:                          logger,
		serviceState:                    rv.serviceState,
		observations:                    requestObservation,
		protocolErrorObservationBuilder: buildProtocolErrorObservation,
	}
}

// buildWebsocketRequestObservations builds a request observation for WebSocket upgrade requests.
func (rv *requestValidator) buildWebsocketRequestObservations(
	rpcType sharedtypes.RPCType,
	requestOrigin qosobservations.RequestOrigin,
) *qosobservations.CosmosRequestObservations {

	return &qosobservations.CosmosRequestObservations{
		CosmosChainId: rv.cosmosChainID,
		ServiceId:     string(rv.serviceID),
		RequestOrigin: requestOrigin,
		RequestProfiles: []*qosobservations.CosmosRequestProfile{
			{
				BackendServiceDetails: &qosobservations.BackendServiceDetails{
					BackendServiceType: convertToProtoBackendServiceType(rpcType),
					SelectionReason:    "WebSocket upgrade request detection",
				},
			},
		},
	}
}

// createWebsocketUnsupportedRPCTypeContext creates error context when WebSocket is not configured
func (rv *requestValidator) createWebsocketUnsupportedRPCTypeContext(
	rpcType sharedtypes.RPCType,
) gateway.RequestQoSContext {
	err := errors.New("WebSocket not supported for this service")
	response := jsonrpc.NewErrResponseInvalidRequest(jsonrpc.ID{}, err)

	observations := rv.createWebsocketUnsupportedRPCTypeObservation(rpcType, response)

	return &qos.RequestErrorContext{
		Logger:   rv.logger,
		Response: response,
		Observations: &qosobservations.Observations{
			ServiceObservations: &qosobservations.Observations_Cosmos{
				Cosmos: observations,
			},
		},
	}
}

// createWebsocketUnsupportedRPCTypeObservation creates an observation for unsupported WebSocket requests
func (rv *requestValidator) createWebsocketUnsupportedRPCTypeObservation(
	rpcType sharedtypes.RPCType,
	jsonrpcResponse jsonrpc.Response,
) *qosobservations.CosmosRequestObservations {
	return &qosobservations.CosmosRequestObservations{
		ServiceId:     string(rv.serviceID),
		CosmosChainId: rv.cosmosChainID,
		RequestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
		RequestProfiles: []*qosobservations.CosmosRequestProfile{
			{
				BackendServiceDetails: &qosobservations.BackendServiceDetails{
					BackendServiceType: convertToProtoBackendServiceType(rpcType),
					SelectionReason:    "WebSocket upgrade request detection (unsupported)",
				},
			},
		},
		RequestLevelError: &qosobservations.RequestError{
			ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_JSONRPC_UNSUPPORTED_RPC_TYPE,
			ErrorDetails:   fmt.Sprintf("Unsupported RPC type %s for service %s", rpcType.String(), string(rv.serviceID)),
			HttpStatusCode: int32(jsonrpcResponse.GetRecommendedHTTPStatusCode()),
		},
	}
}
