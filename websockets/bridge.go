package websockets

import (
	"context"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/observation"
)

// Message represents a websocket message that can be:
//   - Client request
//   - Endpoint response
//   - Subscription push event (e.g. eth_subscribe)
type Message struct {
	// Data is the message payload
	Data []byte

	// MessageType is an int returned by the gorilla/websocket package
	MessageType int
}

// MessageHandler processes websocket messages.
// It can transform the message data before forwarding it.
type MessageHandler interface {
	// HandleMessage processes a message and returns the data to forward.
	// If an error is returned, the connection will be closed.
	HandleMessage(msg Message) ([]byte, error)
}

// ObservationPublisher handles publishing observations after processing endpoint messages.
type ObservationPublisher interface {
	// SetObservationContext sets the gateway observations and data reporter.
	// Set once per Bridge initialization.
	SetObservationContext(*observation.GatewayObservations, gateway.RequestResponseReporter)
	// InitializeMessageObservations initializes the observations for the current message.
	// Called once per message.
	InitializeMessageObservations() *observation.RequestResponseObservations
	// UpdateMessageObservations updates the observations for the current message.
	// Called once per message if the message handler does not return an error.
	UpdateMessageObservationsFromSuccess(*observation.RequestResponseObservations)
	// UpdateMessageObservationsFromError updates the observations for the current message.
	// Called once per message if the message handler returns an error.
	UpdateMessageObservationsFromError(*observation.RequestResponseObservations, error)
	// PublishObservations publishes protocol-specific observations.
	// Called once per message.
	PublishMessageObservations(*observation.RequestResponseObservations)
}

// Bridge routes data between an Endpoint and a Client.
// One Bridge represents a single WebSocket connection
// between a Client and a WebSocket Endpoint.
//
// This is a generic websocket bridge that handles the websocket protocol
// and message routing, while delegating protocol-specific logic to the
// provided message handlers.
//
// Full data flow: Client <---clientConn---> PATH Bridge <---endpointConn---> Relay Miner Bridge <------> Endpoint
type Bridge struct {
	// ctx is used to stop the bridge when the context is canceled from either connection
	ctx context.Context

	logger polylog.Logger

	// endpointConn is the connection to the WebSocket Endpoint
	endpointConn *websocketConnection
	// clientConn is the connection to the Client
	clientConn *websocketConnection

	// msgChan receives messages from the Client and Endpoint and passes them to the other side of the bridge.
	msgChan chan message

	// clientMessageHandler processes messages from the client before forwarding to the endpoint
	clientMessageHandler MessageHandler
	// endpointMessageHandler processes messages from the endpoint before forwarding to the client
	endpointMessageHandler MessageHandler

	// observationPublisher handles publishing observations after endpoint message processing
	observationPublisher ObservationPublisher
}

// NewBridge creates a new Bridge instance with connections to both client and endpoint.
func NewBridge(
	logger polylog.Logger,
	clientWSSConn *websocket.Conn,
	endpointWSSConn *websocket.Conn,
	clientMessageHandler MessageHandler,
	endpointMessageHandler MessageHandler,
	observationPublisher ObservationPublisher,
) (*Bridge, error) {
	logger = logger.With("component", "websocket_bridge")

	// Create a context that can be canceled from either connection
	ctx, cancelCtx := context.WithCancel(context.Background())

	// Create a channel to pass messages between the Client and Endpoint
	msgChan := make(chan message)

	// Create bridge instance
	b := &Bridge{
		logger: logger,
		ctx:    ctx,

		msgChan:                msgChan,
		clientMessageHandler:   clientMessageHandler,
		endpointMessageHandler: endpointMessageHandler,
		observationPublisher:   observationPublisher,
	}
	if err := b.validateComponents(); err != nil {
		return nil, fmt.Errorf("invalid bridge components: %w", err)
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

// validateComponents ensures the Bridge is not created with nil components.
// This is done to avoid panics and to make the Bridge's behavior more predictable.
func (b *Bridge) validateComponents() error {
	switch {
	case b.observationPublisher == nil:
		return fmt.Errorf("observationPublisher is nil")
	case b.clientMessageHandler == nil:
		return fmt.Errorf("clientMessageHandler is nil")
	case b.endpointMessageHandler == nil:
		return fmt.Errorf("endpointMessageHandler is nil")
	}
	return nil
}

// StartAsync starts the bridge and establishes a bidirectional communication
// through PATH between the Client and the selected websocket endpoint.
//
// This method implements the gateway.WebsocketsBridge interface.
//
// Full data flow: Client <---clientConn---> PATH Bridge <---endpointConn---> Relay Miner Bridge <------> Endpoint
func (b *Bridge) StartAsync(
	gatewayObservations *observation.GatewayObservations,
	dataReporter gateway.RequestResponseReporter,
) {
	b.logger.Info().Msg("🏗️ Websocket bridge operation started successfully")

	// Set the observation context for observation publishing.
	//
	// These values, both gateway and protocol, are static for the duration of
	// the bridge's operation. New observations will be set when a new Bridge is created.
	if b.observationPublisher != nil {
		b.observationPublisher.SetObservationContext(gatewayObservations, dataReporter)
	}

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
func (b *Bridge) Shutdown(err error) {
	b.logger.Error().Err(err).Msg("🔌 ❌ Websocket bridge shutting down due to error!")

	// Send close message to both connections and close the connections
	errMsg := fmt.Sprintf("bridge shutting down: %s", err.Error())
	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, errMsg)

	if b.clientConn != nil {
		if err := b.clientConn.WriteMessage(websocket.CloseMessage, closeMsg); err != nil {
			b.logger.Error().Err(err).Msg("❌ error writing close message to client connection")
		}
		b.clientConn.Close()
	}
	if b.endpointConn != nil {
		if err := b.endpointConn.WriteMessage(websocket.CloseMessage, closeMsg); err != nil {
			b.logger.Error().Err(err).Msg("❌ error writing close message to endpoint connection")
		}
		b.endpointConn.Close()
	}

	// Close the message channel to stop the message loop
	close(b.msgChan)
}

// handleClientMessage processes a message from the Client and sends it to the endpoint.
func (b *Bridge) handleClientMessage(msg message) {
	// Create a Message struct for the handler
	handlerMsg := Message{
		Data:        msg.data,
		MessageType: msg.messageType,
	}

	// Process the message through the client message handler
	processedData, err := b.clientMessageHandler.HandleMessage(handlerMsg)
	if err != nil {
		b.clientConn.handleDisconnect(fmt.Errorf("handleClientMessage: %w", err))
		return
	}

	// Send the processed message to the endpoint
	if err := b.endpointConn.WriteMessage(msg.messageType, processedData); err != nil {
		b.endpointConn.handleDisconnect(fmt.Errorf("handleClientMessage: error writing to endpoint: %w", err))
		return
	}
}

// handleEndpointMessage processes a message from the Endpoint and sends it to the Client.
func (b *Bridge) handleEndpointMessage(msg message) {
	// Create a Message struct for the handler
	handlerMsg := Message{
		Data:        msg.data,
		MessageType: msg.messageType,
	}

	// Initialize the message observations for the current message.
	messageObservations := b.observationPublisher.InitializeMessageObservations()

	// Ensure observations are published regardless of success or failure
	defer b.observationPublisher.PublishMessageObservations(messageObservations)

	// Process the message through the endpoint message handler
	processedData, err := b.endpointMessageHandler.HandleMessage(handlerMsg)
	if err != nil {
		// Update observations with error details before disconnecting
		b.observationPublisher.UpdateMessageObservationsFromError(messageObservations, err)
		b.endpointConn.handleDisconnect(fmt.Errorf("handleEndpointMessage: %w", err))
		return
	}

	// Send the processed message to the client
	if err := b.clientConn.WriteMessage(msg.messageType, processedData); err != nil {
		// NOTE: On session rollover, the Endpoint will disconnect the Endpoint connection, which will trigger this
		// error. This is expected and the Client is expected to handle the reconnection in their connection logic.
		b.clientConn.handleDisconnect(fmt.Errorf("handleEndpointMessage: error writing to client: %w", err))
		return
	}

	// Update observations with success details
	b.observationPublisher.UpdateMessageObservationsFromSuccess(messageObservations)
}
