package websockets

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/poktroll/pkg/relayer/proxy"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/request"
)

// FullNode represents a Shannon FullNode as only Shannon supports websocket connections.
// It is used only to validate the relay responses returned by the Endpoint.
type FullNode interface {
	// ValidateRelayResponse validates the raw bytes returned from an endpoint (in response to a relay request) and returns the parsed response.
	ValidateRelayResponse(supplierAddr sdk.SupplierAddress, responseBz []byte) (*servicetypes.RelayResponse, error)
}

// RelayRequestSigner is used by the request context to sign the relay request.
// It takes an unsigned relay request and an application, and returns a relay request signed either by the gateway that has delegation from the app.
// If/when the Permissionless Gateway Mode is supported by the Shannon integration, the app's own private key may also be used for signing the relay request.
type RelayRequestSigner interface {
	SignRelayRequest(req *servicetypes.RelayRequest, app apptypes.Application) (*servicetypes.RelayRequest, error)
}

// SelectedEndpoint represents a Shannon Endpoint that has been selected to service a persistent websocket connection.
type SelectedEndpoint interface {
	Addr() protocol.EndpointAddr
	PublicURL() string
	WebsocketURL() (string, error)
	Supplier() string
	Session() *sessiontypes.Session
}

// bridge routes data between an Endpoint and a Client.
// One bridge represents a single WebSocket connection between
// a Client and a WebSocket Endpoint.
//
// Full data flow: Client <---clientConn---> PATH Bridge <---endpointConn---> Relay Miner Bridge <------> Endpoint
type bridge struct {
	// ctx is used to stop the bridge when the context is canceled from either connection
	ctx context.Context

	logger polylog.Logger

	// endpointConn is the connection to the WebSocket Endpoint
	endpointConn *websocketConnection
	// clientConn is the connection to the Client
	clientConn *websocketConnection

	// msgChan receives messages from the Client and Endpoint and passes them to the other side of the bridge.
	msgChan chan message

	// selectedEndpoint is the Endpoint that the bridge is connected to
	selectedEndpoint SelectedEndpoint
	// relayRequestSigner is the RelayRequestSigner that the bridge uses to sign relay requests
	relayRequestSigner RelayRequestSigner
	// fullNode is the FullNode that the bridge uses to validate relay responses
	fullNode FullNode
}

// NewBridge creates a new Bridge instance and a new connection to the Endpoint from the Endpoint URL
func NewBridge(
	logger polylog.Logger,
	clientWSSConn *websocket.Conn,
	selectedEndpoint SelectedEndpoint,
	relayRequestSigner RelayRequestSigner,
	fullNode FullNode,
) (*bridge, error) {
	logger = logger.With(
		"component", "bridge",
		"endpoint_url", selectedEndpoint.PublicURL(),
	)

	// Connect to the Endpoint
	endpointWSSConn, err := connectWebsocketEndpoint(logger, selectedEndpoint)
	if err != nil {
		return nil, fmt.Errorf("NewBridge: %s", err.Error())
	}

	// Create a context that can be canceled from either connection
	ctx, cancelCtx := context.WithCancel(context.Background())

	// Create a channel to pass messages between the Client and Endpoint
	msgChan := make(chan message)

	// Create bridge instance without connections first
	b := &bridge{
		logger:             logger,
		msgChan:            msgChan,
		selectedEndpoint:   selectedEndpoint,
		relayRequestSigner: relayRequestSigner,
		fullNode:           fullNode,
		ctx:                ctx,
	}

	// Initialize connections with context and cancel function
	b.endpointConn = newConnection(
		b.ctx,
		cancelCtx,
		logger.With("conn", "endpoint"),
		endpointWSSConn,
		messageSourceEndpoint,
		msgChan,
	)
	b.clientConn = newConnection(
		b.ctx,
		cancelCtx,
		logger.With("conn", "client"),
		clientWSSConn,
		messageSourceClient,
		msgChan,
	)

	return b, nil
}

// connectWebsocketEndpoint makes a websocket connection to the websocket Endpoint.
func connectWebsocketEndpoint(logger polylog.Logger, selectedEndpoint SelectedEndpoint) (*websocket.Conn, error) {
	logger.Info().Msgf("🔗 Connecting to endpoint: %s", selectedEndpoint.PublicURL())

	websocketURL, err := selectedEndpoint.WebsocketURL()
	if err != nil {
		logger.Error().Err(err).Msgf("❌ Selected endpoint does not support websocket RPC type: %s", selectedEndpoint.Addr())
		return nil, err
	}

	u, err := url.Parse(websocketURL)
	if err != nil {
		logger.Error().Err(err).Msgf("❌ Error parsing endpoint URL: %s", selectedEndpoint.PublicURL())
		return nil, err
	}

	headers := getBridgeRequestHeaders(selectedEndpoint.Session())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		logger.Error().Err(err).Msgf("❌ Error connecting to endpoint: %s", u.String())
		return nil, err
	}

	return conn, nil
}

// TODO_DOCUMENT(@commoddity): Document these headers and how bridge connections work in more detail.
//
// getBridgeRequestHeaders returns the headers that should be sent to the RelayMiner
// when establishing a new websocket connection to the Endpoint.
//
// The headers are:
//   - `Target-Service-Id`: The service ID of the target service.
//   - `App-Address:` The address of the session's application.
//   - `Rpc-Type`: The type of RPC request. Always "websocket" for websocket connection requests.
func getBridgeRequestHeaders(session *sessiontypes.Session) http.Header {
	headers := http.Header{}
	headers.Add(request.HTTPHeaderTargetServiceID, session.Header.ServiceId)
	headers.Add(request.HTTPHeaderAppAddress, session.Header.ApplicationAddress)

	// Get the "WEBSOCKET" RPC type enum value and add it to the headers.
	rpcTypeWebsocket := strconv.Itoa(int(sharedtypes.RPCType_WEBSOCKET))
	headers.Add(proxy.RPCTypeHeader, rpcTypeWebsocket)
	return headers
}

// Run starts the bridge and establishes a bidirectional communication
// through PATH between the Client and the selected websocket endpoint.
//
// Full data flow: Client <---clientConn---> PATH Bridge <---endpointConn---> Relay Miner Bridge <------> Endpoint
func (b *bridge) Run() {
	b.logger.Info().Msg("bridge operation started successfully")

	// Listen for the context to be canceled and shut down the bridge
	go func() {
		<-b.ctx.Done()
		b.Shutdown(fmt.Errorf("context canceled"))
	}()

	for msg := range b.msgChan {
		switch msg.source {
		case messageSourceClient:
			b.handleClientMessage(msg)

		case messageSourceEndpoint:
			b.handleEndpointMessage(msg)
		}
	}
}

// Shutdown stops the bridge and closes both connections.
// This method is passed to each connection and is called when an error is encountered.
//
// It ensures that both Client and Endpoint connections are closed and the message channel is closed.
//
// This is important as it is expected that the RelayMiner connection will be closed on every session rollover
// and it is critical that the closing of the connection propagates to the Client so they can reconnect.
func (b *bridge) Shutdown(err error) {
	b.logger.Error().Err(err).Msg("🔌 ❌ Websocket bridge shutting down due to error!")

	// Send close message to both connections and close the connections
	errMsg := fmt.Sprintf("bridge shutting down: %s", err.Error())
	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, errMsg)

	if b.clientConn != nil {
		if err := b.clientConn.WriteMessage(websocket.CloseMessage, closeMsg); err != nil {
			b.logger.Error().Err(err).Msg("error writing close message to client connection")
		}
		b.clientConn.Close()
	}
	if b.endpointConn != nil {
		if err := b.endpointConn.WriteMessage(websocket.CloseMessage, closeMsg); err != nil {
			b.logger.Error().Err(err).Msg("error writing close message to endpoint connection")
		}
		b.endpointConn.Close()
	}

	// Close the message channel to stop the message loop
	close(b.msgChan)
}

// handleClientMessage processes a message from the Client and sends it to the Endpoint
// It signs the request using the RelayRequestSigner and sends the signed request to the Endpoint
func (b *bridge) handleClientMessage(msg message) {
	b.logger.Debug().Msgf("received message from client: %s", string(msg.data))

	// Sign the client message before sending it to the Endpoint
	signedClientMessageBz, err := b.signClientMessage(msg)
	if err != nil {
		b.clientConn.handleDisconnect(fmt.Errorf("handleClientMessage: error signing client message: %w", err))
		return
	}

	// Send the signed request to the RelayMiner, which will forward it to the Endpoint
	if err := b.endpointConn.WriteMessage(msg.messageType, signedClientMessageBz); err != nil {
		b.endpointConn.handleDisconnect(fmt.Errorf("handleClientMessage: error writing client message to endpoint: %w", err))
		return
	}
}

// signClientMessage signs the client message in order to send it to the Endpoint
// It uses the RelayRequestSigner to sign the request and returns the signed request
func (b *bridge) signClientMessage(msg message) ([]byte, error) {
	unsignedRelayRequest := &servicetypes.RelayRequest{
		Meta: servicetypes.RelayRequestMetadata{
			SessionHeader:           b.selectedEndpoint.Session().GetHeader(),
			SupplierOperatorAddress: b.selectedEndpoint.Supplier(),
		},
		Payload: msg.data,
	}

	app := b.selectedEndpoint.Session().GetApplication()
	signedRelayRequest, err := b.relayRequestSigner.SignRelayRequest(unsignedRelayRequest, *app)
	if err != nil {
		return nil, fmt.Errorf("error signing client message: %s", err.Error())
	}

	relayRequestBz, err := signedRelayRequest.Marshal()
	if err != nil {
		return nil, fmt.Errorf("error marshalling signed client message: %s", err.Error())
	}

	return relayRequestBz, nil
}

// handleEndpointMessage processes a message from the Endpoint and sends it to the Client
// It validates the relay response using the Shannon FullNode and sends the relay response to the Client
// Subscription events pushed from the Endpoint to the Client will be handled here as well.
func (b *bridge) handleEndpointMessage(msg message) {
	b.logger.Debug().Msgf("received message from endpoint: %s", string(msg.data))

	// Validate the relay response using the Shannon FullNode
	relayResponse, err := b.fullNode.ValidateRelayResponse(sdk.SupplierAddress(b.selectedEndpoint.Supplier()), msg.data)
	if err != nil {
		b.endpointConn.handleDisconnect(fmt.Errorf("handleEndpointMessage: error validating relay response: %w", err))
		return
	}

	// Send the relay response or subscription push event to the Client
	if err := b.clientConn.WriteMessage(msg.messageType, relayResponse.Payload); err != nil {
		// NOTE: On session rollover, the RelayMiner will disconnect the Endpoint connection, which will trigger this
		// error. This is expected and the Client is expected to handle the reconnection in their connection logic.
		b.clientConn.handleDisconnect(fmt.Errorf("handleEndpointMessage: error writing endpoint message to client: %w", err))
		return
	}
}
