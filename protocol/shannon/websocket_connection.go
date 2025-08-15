package shannon

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/pokt-network/poktroll/pkg/relayer/proxy"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/request"
	"github.com/buildwithgrove/path/websockets"
)

// TODO_TECHDEBT(@adshmh): Move this functionality to the gateway package.
// - Build the required protocol observations in the protocol/shannon package.
// - Coordinate the building and publication of observation in the gateway package.
// - Move the construction of the bridge to gateway package.
//
// TODO_NEXT(@commoddity): Integrate session rollover detection for faster WebSocket disconnection.
// See: https://github.com/buildwithgrove/path/issues/408
//
// Currently, WebSocket connections during session rollovers experience dead connection periods
// of up to 30 seconds due to ping/pong timeout detection. This should be improved by:
//
// 1. Integrating existing sessionRolloverState monitoring from fullnode_session_rollover.go
//    into the WebSocket bridge to proactively detect when endpoints become unresponsive
// 2. Immediately dropping client connections when session rollovers are detected rather than
//    waiting for ping/pong timeouts
// 3. Making ping/pong timeouts configurable for faster dead connection detection
//
// This will eliminate multi-second dead connection periods and improve client experience
// during session transitions while maintaining client-side reconnection responsibility.

// createWebsocketBridge creates a websocket bridge.
// This function encapsulates all Shannon-specific logic for websocket handling.
func (rc *requestContext) createWebsocketBridge(
	req *http.Request,
	w http.ResponseWriter,
) (gateway.WebsocketsBridge, error) {
	// Hydrate the logger with the websocket bridge specific information.
	logger := rc.logger.With(
		"connection_type", "websocket",
		"component", "shannon_websocket_bridge",
		"endpoint_address", rc.selectedEndpoint.Addr(),
		"is_fallback", rc.selectedEndpoint.IsFallback(),
	)

	// Get the websocket-specific URL from the selected endpoint.
	websocketURL, err := rc.selectedEndpoint.WebsocketURL()
	if err != nil {
		logger.Error().Err(err).Msg("❌ Selected endpoint does not support websocket RPC type")
		return nil, err
	}
	logger = logger.With("websocket_url", websocketURL)

	// Build the protocol observations for the websocket bridge.
	protocolObservations := buildWebsocketBridgeEndpointObservation(
		logger,
		rc.serviceID,
		rc.selectedEndpoint,
	)

	// Upgrade HTTP request from client to websocket connection.
	// - Connection is passed to websocket bridge for Client <-> Gateway communication.
	clientConn, err := websockets.UpgradeClientWebsocketConnection(rc.logger, req, w)
	if err != nil {
		return nil, fmt.Errorf("createWebsocketBridge: %s", err.Error())
	}

	// Get the headers for the websocket connection.
	headers := rc.getWebsocketConnectionHeaders()

	// Connect to the endpoint
	endpointConn, err := websockets.ConnectWebsocketEndpoint(rc.logger, websocketURL, headers)
	if err != nil {
		return nil, fmt.Errorf("createWebsocketBridge: %s", err.Error())
	}

	// TODO_TECHDEBT(@adshmh): Refactor to keep gateway package as the coordinator of components' interactions:
	// - Update the gateway.ProtocolRequestContext interface to enable handling client/endpoint messages.
	// - Replace the gateway.ProtocolRequestContext's `HandleWebsocketRequest` with proper methods to be called from gateway package.
	// - Drop the creation of bridge below.
	// - Minimize the responsibilities of bridge struct (see TODO comments in websockets/bridge.go)
	//
	// Create Shannon-specific message handlers
	clientHandler := &websocketClientMessageHandler{
		logger:             logger.With("component", "shannon_client_message_handler"),
		selectedEndpoint:   rc.selectedEndpoint,
		relayRequestSigner: rc.relayRequestSigner,
		serviceID:          rc.serviceID,
	}
	endpointHandler := &endpointMessageHandler{
		logger:           logger.With("component", "shannon_endpoint_message_handler"),
		selectedEndpoint: rc.selectedEndpoint,
		fullNode:         rc.fullNode,
		serviceID:        rc.serviceID,
	}

	// Create observation publisher
	observationPublisher := &observationPublisher{
		logger:               logger.With("component", "shannon_observation_publisher"),
		serviceID:            rc.serviceID,
		protocolObservations: protocolObservations,
	}

	// Create the generic websocket bridge with Shannon-specific handlers
	bridge, err := websockets.NewBridge(
		rc.logger,
		clientConn,
		endpointConn,
		clientHandler,
		endpointHandler,
		observationPublisher,
	)
	if err != nil {
		return nil, err
	}

	return bridge, nil
}

// getWebsocketConnectionHeaders returns headers for the websocket connection:
//   - Target-Service-Id: The service ID of the target service
//   - App-Address: The address of the session's application
//   - Rpc-Type: Always "websocket" for websocket connection requests
func (rc *requestContext) getWebsocketConnectionHeaders() http.Header {
	headers := http.Header{}

	// If the selected endpoint is a protocol endpoint, add the headers
	// that the RelayMiner requires to forward the request to the Endpoint.
	//
	// Requests to fallback endpoints bypass the protocol so RelayMiner headers are not needed.
	if !rc.getSelectedEndpoint().IsFallback() {
		headers = rc.getRelayMinerConnectionHeaders(rc.getSelectedEndpoint().Session().GetHeader())
	}

	return headers
}

// getRelayMinerConnectionHeaders returns headers for RelayMiner websocket connections:
//   - Target-Service-Id: The service ID of the target service
//   - App-Address: The address of the session's application
//   - Rpc-Type: Always "websocket" for websocket connection requests
func (rc *requestContext) getRelayMinerConnectionHeaders(sessionHeader *sessiontypes.SessionHeader) http.Header {
	if sessionHeader == nil {
		rc.logger.Error().Msg("❌ SHOULD NEVER HAPPEN: Error getting relay miner connection headers: session header is nil")
		return http.Header{}
	}

	return http.Header{
		request.HTTPHeaderTargetServiceID: {sessionHeader.ServiceId},
		request.HTTPHeaderAppAddress:      {sessionHeader.ApplicationAddress},
		proxy.RPCTypeHeader:               {strconv.Itoa(int(sharedtypes.RPCType_WEBSOCKET))},
	}
}
