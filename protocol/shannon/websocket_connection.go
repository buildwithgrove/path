package shannon

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/poktroll/pkg/relayer/proxy"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/request"
	"github.com/buildwithgrove/path/websockets"
)

// createShannonWebsocketBridge creates a Shannon-specific websocket bridge.
// This function encapsulates all Shannon-specific logic for websocket handling.
func (rc *requestContext) createShannonWebsocketBridge(
	logger polylog.Logger,
	req *http.Request,
	w http.ResponseWriter,
) (gateway.WebsocketsBridge, error) {
	// Hydrate the logger with the websocket bridge specific information.
	logger = logger.With(
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
	clientConn, err := websockets.UpgradeClientWebsocketConnection(logger, req, w)
	if err != nil {
		return nil, fmt.Errorf("createShannonWebsocketBridge: %s", err.Error())
	}

	// Get the headers for the websocket connection.
	headers := getShannonWebsocketConnectionHeaders(logger, rc.selectedEndpoint)

	// Connect to the endpoint
	endpointConn, err := websockets.ConnectWebsocketEndpoint(logger, websocketURL, headers)
	if err != nil {
		return nil, fmt.Errorf("createShannonWebsocketBridge: %s", err.Error())
	}

	// Create Shannon-specific message handlers
	clientHandler := &shannonClientMessageHandler{
		logger:             logger.With("component", "shannon_client_message_handler"),
		selectedEndpoint:   rc.selectedEndpoint,
		relayRequestSigner: rc.relayRequestSigner,
		serviceID:          rc.serviceID,
	}
	endpointHandler := &shannonEndpointMessageHandler{
		logger:           logger.With("component", "shannon_endpoint_message_handler"),
		selectedEndpoint: rc.selectedEndpoint,
		fullNode:         rc.fullNode,
		serviceID:        rc.serviceID,
	}

	// Create observation publisher
	observationPublisher := &shannonObservationPublisher{
		logger:               logger.With("component", "shannon_observation_publisher"),
		serviceID:            rc.serviceID,
		protocolObservations: protocolObservations,
	}

	// Create the generic websocket bridge with Shannon-specific handlers
	bridge, err := websockets.NewBridge(
		logger,
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

// getShannonWebsocketConnectionHeaders returns the headers that should be sent to the websocket connection.
//
// The headers are:
//   - `Target-Service-Id`: The service ID of the target service.
//   - `App-Address:` The address of the session's application.
//   - `Rpc-Type`: The type of RPC request. Always "websocket" for websocket connection requests.
func getShannonWebsocketConnectionHeaders(logger polylog.Logger, selectedEndpoint endpoint) http.Header {
	headers := http.Header{}

	// If the selected endpoint is a protocol endpoint, add the headers
	// that the RelayMiner requires to forward the request to the Endpoint.
	//
	// Requests to fallback endpoints bypass the protocol so RelayMiner headers are not needed.
	if !selectedEndpoint.IsFallback() {
		headers = getRelayMinerConnectionHeaders(logger, selectedEndpoint.Session().GetHeader())
	}

	return headers
}

// getRelayMinerConnectionHeaders returns the headers that should be sent to the RelayMiner
// when establishing a new websocket connection to the Endpoint.
//
// The headers are:
//   - `Target-Service-Id`: The service ID of the target service.
//   - `App-Address:` The address of the session's application.
//   - `Rpc-Type`: The type of RPC request. Always "websocket" for websocket connection requests.
func getRelayMinerConnectionHeaders(logger polylog.Logger, sessionHeader *sessiontypes.SessionHeader) http.Header {
	if sessionHeader == nil {
		logger.Error().Msg("❌ SHOULD NEVER HAPPEN: Error getting relay miner connection headers: session header is nil")
		return http.Header{}
	}

	return http.Header{
		request.HTTPHeaderTargetServiceID: {sessionHeader.ServiceId},
		request.HTTPHeaderAppAddress:      {sessionHeader.ApplicationAddress},
		proxy.RPCTypeHeader:               {strconv.Itoa(int(sharedtypes.RPCType_WEBSOCKET))},
	}
}
