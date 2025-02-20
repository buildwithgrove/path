package websockets

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"
)

// FullNode is represents a Shannon FullNode as only Shannon supports websocket connections.
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
	PublicURL() string
	Supplier() string
	Session() *sessiontypes.Session
}

// bridge routes data between an Endpoint and a Client.
// One bridge represents a single WebSocket connection between
// a Client and a WebSocket Endpoint.
//
// Full data flow: Client <------> PATH Bridge <------> Relay Miner Bridge <------> Endpoint
type bridge struct {
	logger polylog.Logger

	// endpointConn is the connection to the WebSocket Endpoint
	endpointConn *connection
	// clientConn is the connection to the Client
	clientConn *connection

	// msgChan and stopChan are shared between the Client and Endpoint
	// which allows a reuse of the connection struct for both connections.

	// msgChan receives messages from the Client and Endpoint
	// and passes them to the other side of the bridge.
	msgChan <-chan message
	// stopChan is a channel that signals the bridge to stop
	stopChan chan error

	selectedEndpoint   SelectedEndpoint
	relayRequestSigner RelayRequestSigner
	fullNode           FullNode
}

// NewBridge creates a new Bridge instance and a new connection to the Endpoint from the Endpoint URL
func NewBridge(
	logger polylog.Logger,
	clientWSSConn *websocket.Conn,
	selectedEndpoint SelectedEndpoint,
	relayRequestSigner RelayRequestSigner,
	fullNode FullNode,
) (*bridge, error) {
	endpointWSSConn, err := connectEndpoint(selectedEndpoint)
	if err != nil {
		return nil, fmt.Errorf("NewBridge: %s", err.Error())
	}

	msgChan := make(chan message)
	stopChan := make(chan error)

	logger = logger.With(
		"component", "bridge",
		"endpoint_url", selectedEndpoint.PublicURL(),
	)

	endpointConnection := newConnection(
		logger.With("conn", "endpoint"),
		endpointWSSConn,
		messageSourceEndpoint,
		msgChan,
		stopChan,
	)
	clientConnection := newConnection(
		logger.With("conn", "client"),
		clientWSSConn,
		messageSourceClient,
		msgChan,
		stopChan,
	)

	return &bridge{
		logger: logger,

		endpointConn: endpointConnection,
		clientConn:   clientConnection,
		msgChan:      msgChan,
		stopChan:     stopChan,

		selectedEndpoint:   selectedEndpoint,
		relayRequestSigner: relayRequestSigner,
		fullNode:           fullNode,
	}, nil
}

// Run starts the bridge and establishes a bidirectional communication
// through PATH between the Client and the selected websocket endpoint.
//
// Full data flow: Client <------> PATH Bridge <------> Relay Miner Bridge <------> Endpoint
func (b *bridge) Run() {
	// Start goroutine to read messages from message channel
	go b.messageLoop()

	b.logger.Info().Msg("bridge operation started successfully")

	// Keep the bridge open until a stop signal is received (i.e. block until told otherwise)
	<-b.stopChan
}

// Close stops the bridge and closes both connections
func (b *bridge) Close() {
	close(b.stopChan)
}

// messageLoop reads from the message channel and handles messages from the endpoint and Client
func (b *bridge) messageLoop() {
	for {
		select {
		case <-b.stopChan:
			return

		case msg := <-b.msgChan:
			switch msg.source {

			// If the message is from the Client connection, send it to the Endpoint
			case messageSourceClient:
				b.handleClientMessage(msg)

			// If the message is from the Endpoint, send it to the Client
			case messageSourceEndpoint:
				b.handleEndpointMessage(msg)
			}
		}
	}
}

// handleClientMessage processes a message from the Client and sends it to the Endpoint
// It signs the request using the RelayRequestSigner and sends the signed request to the Endpoint
func (b *bridge) handleClientMessage(msg message) {
	b.logger.Debug().Msgf("received message from client: %s", string(msg.data))

	// Sign the client message before sending it to the Endpoint
	signedClientMessageBz, err := b.signClientMessage(msg)
	if err != nil {
		b.clientConn.handleError(err, messageSourceClient)
		return
	}

	// Send the signed request to the RelayMiner, which will forward it to the Endpoint
	if err := b.endpointConn.WriteMessage(msg.messageType, signedClientMessageBz); err != nil {
		b.endpointConn.handleError(err, messageSourceEndpoint)
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
		return nil, fmt.Errorf("signClientMessage: error signing client message: %s", err.Error())
	}

	relayRequestBz, err := signedRelayRequest.Marshal()
	if err != nil {
		return nil, fmt.Errorf("signClientMessage: error marshalling signed client message: %s", err.Error())
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
		b.endpointConn.handleError(err, messageSourceEndpoint)
		return
	}

	if err := b.clientConn.WriteMessage(msg.messageType, relayResponse.Payload); err != nil {
		b.clientConn.handleError(err, messageSourceClient)
		return
	}
}
