package cosmos

import (
	"errors"

	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_UPNEXT(@commoddity): Add the ability to perform QoS checks to
// determine if an endpoint supports WebSocket connections.

// validateWebsocketRequest validates a WebSocket request by:
// 1. Checking if it's a valid WebSocket upgrade request
// 2. Checking if the WebSocket RPC type is supported
// 3. Creating the request context with all necessary information
func (rv *requestValidator) validateWebsocketRequest() (gateway.RequestQoSContext, bool) {
	logger := rv.logger.With(
		"validator", "WebSocket",
		"method", "validateWebsocketRequest",
	)

	// Set the RPC type to WebSocket as specified in the user requirements
	rpcType := sharedtypes.RPCType_WEBSOCKET

	// Hydrate the logger with detected RPC type
	logger = logger.With("detected_rpc_type", rpcType.String())

	// Check if WebSocket RPC type is supported by the service
	if _, supported := rv.supportedAPIs[sharedtypes.RPCType_WEBSOCKET]; !supported {
		logger.Warn().Msg("Request uses unsupported WebSocket RPC type")
		return rv.createWebsocketUnsupportedRPCTypeContext(rpcType), false
	}

	// Build and return the request context
	return rv.buildWebsocketRequestContext(
		rpcType,
		qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
	)
}

func (rv *requestValidator) buildWebsocketRequestContext(
	rpcType sharedtypes.RPCType,
	requestOrigin qosobservations.RequestOrigin,
) (gateway.RequestQoSContext, bool) {
	logger := rv.logger.With(
		"method", "buildWebsocketRequestContext",
	)

	// Generate the QoS observation for the request
	requestObservation := rv.buildWebsocketRequestObservations(
		rpcType,
		requestOrigin,
	)

	// Create specialized WebSocket context
	return &requestContext{
		logger:                          logger,
		serviceState:                    rv.serviceState,
		observations:                    requestObservation,
		protocolErrorObservationBuilder: buildProtocolErrorObservation,
	}, true
}

func (rv *requestValidator) buildWebsocketRequestObservations(
	rpcType sharedtypes.RPCType,
	requestOrigin qosobservations.RequestOrigin,
) *qosobservations.CosmosRequestObservations {

	return &qosobservations.CosmosRequestObservations{
		CosmosChainId: rv.cosmosChainID,
		ServiceId:     string(rv.serviceID),
		RequestOrigin: requestOrigin,
		RequestProfile: &qosobservations.CosmosRequestProfile{
			BackendServiceDetails: &qosobservations.BackendServiceDetails{
				BackendServiceType: convertToProtoBackendServiceType(rpcType),
				SelectionReason:    "WebSocket upgrade request detection",
			},
		},
	}
}

// createWebsocketUnsupportedRPCTypeContext creates an error context for unsupported WebSocket RPC type
func (rv *requestValidator) createWebsocketUnsupportedRPCTypeContext(rpcType sharedtypes.RPCType) gateway.RequestQoSContext {
	// Create a JSONRPC error response for unsupported RPC type
	err := errors.New("unsupported RPC type: " + rpcType.String())
	response := jsonrpc.NewErrResponseInvalidRequest(jsonrpc.ID{}, err)

	// Create the observations object with the unsupported RPC type observation
	observations := rv.createWebsocketUnsupportedRPCTypeObservation(rpcType, response)

	// Build and return the error context
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

func (rv *requestValidator) createWebsocketUnsupportedRPCTypeObservation(
	rpcType sharedtypes.RPCType,
	jsonrpcResponse jsonrpc.Response,
) *qosobservations.CosmosRequestObservations {
	return &qosobservations.CosmosRequestObservations{
		ServiceId:     string(rv.serviceID),
		CosmosChainId: rv.cosmosChainID,
		RequestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
		RequestProfile: &qosobservations.CosmosRequestProfile{
			BackendServiceDetails: &qosobservations.BackendServiceDetails{
				BackendServiceType: convertToProtoBackendServiceType(rpcType),
				SelectionReason:    "WebSocket upgrade request detection (unsupported)",
			},
		},
		RequestLevelError: &qosobservations.RequestError{
			ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_JSONRPC_UNSUPPORTED_RPC_TYPE,
			ErrorDetails:   "Unsupported RPC type: " + rpcType.String(),
			HttpStatusCode: int32(jsonrpcResponse.GetRecommendedHTTPStatusCode()),
		},
	}
}
