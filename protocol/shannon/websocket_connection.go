package shannon

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"
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
	logger = logger.With(
		"component", "shannon_websocket_bridge",
		"endpoint_url", rc.selectedEndpoint.PublicURL(),
	)

	protocolObservations := buildWebsocketBridgeEndpointObservation(rc.logger, rc.serviceID, rc.selectedEndpoint)

	// Upgrade HTTP request from client to websocket connection.
	// - Connection is passed to websocket bridge for Client <-> Gateway communication.
	clientConn, err := websockets.UpgradeClientWebsocketConnection(logger, req, w)
	if err != nil {
		return nil, fmt.Errorf("createShannonWebsocketBridge: %s", err.Error())
	}

	// Connect to the endpoint
	endpointConn, err := connectWebsocketEndpoint(logger, rc.selectedEndpoint)
	if err != nil {
		return nil, fmt.Errorf("createShannonWebsocketBridge: %s", err.Error())
	}

	// Create Shannon-specific message handlers
	clientHandler := &shannonClientMessageHandler{
		logger:             logger,
		selectedEndpoint:   rc.selectedEndpoint,
		relayRequestSigner: rc.relayRequestSigner,
		serviceID:          rc.serviceID,
	}
	endpointHandler := &shannonEndpointMessageHandler{
		logger:           logger,
		selectedEndpoint: rc.selectedEndpoint,
		fullNode:         rc.fullNode,
		serviceID:        rc.serviceID,
	}

	// Create observation publisher
	observationPublisher := &shannonObservationPublisher{
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

// connectWebsocketEndpoint makes a websocket connection to the websocket Endpoint.
func connectWebsocketEndpoint(logger polylog.Logger, selectedEndpoint endpoint) (*websocket.Conn, error) {
	// Get the websocket-specific URL from the selected endpoint.
	websocketURL, err := selectedEndpoint.WebsocketURL()
	if err != nil {
		logger.Error().Err(err).Msgf("❌ Selected endpoint does not support websocket RPC type: %s", selectedEndpoint.Addr())
		return nil, err
	}

	logger.Info().Msgf("🔗 Connecting to websocket endpoint: %s", websocketURL)

	// Ensure the websocket URL is valid.
	u, err := url.Parse(websocketURL)
	if err != nil {
		logger.Error().Err(err).Msgf("❌ Error parsing endpoint URL: %s", websocketURL)
		return nil, err
	}

	// Prepare the headers for the websocket connection.
	headers := http.Header{}

	// If the selected endpoint is a protocol endpoint, add the headers
	// that the RelayMiner requires to forward the request to the Endpoint.
	//
	// Requests to fallback endpoints bypass the protocol so RelayMiner headers are not needed.
	if !selectedEndpoint.IsFallback() {
		headers = getRelayMinerConnectionHeaders(logger, selectedEndpoint.Session().GetHeader())
	}

	// Connect to the websocket endpoint using the default websocket dialer.
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		logger.Error().Err(err).Msgf("❌ Error connecting to endpoint: %s", u.String())
		return nil, err
	}

	logger.Debug().Msgf("🔗 Connected to websocket endpoint: %s", websocketURL)

	return conn, nil
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
