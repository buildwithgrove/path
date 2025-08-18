package websockets

import (
	"context"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// WebSocketMessageHandler handles websocket messages.
// It can, for example, transform the message data before forwarding it.
type WebSocketMessageHandler interface {
	// HandleMessage processes a message and returns the data to forward.
	// If an error is returned, the connection will be closed.
	HandleMessage(msgData []byte) ([]byte, error)
}

// bridge routes data between an Endpoint and a Client.
// One bridge represents a single WebSocket connection
// between a Client and a WebSocket Endpoint.
//
// This is a generic websocket bridge that handles the websocket protocol
// and message routing, while delegating protocol-specific logic to the
// provided message handlers.
//
// Architecture:
// - Protocol-agnostic: Works with any protocol (Shannon, future protocols)
// - Message handlers: Protocol-specific logic for signing/validation
// - Notification channels: Gateway observability without tight coupling
//
// Full data flow: Client <---clientConn---> PATH bridge <---endpointConn---> Relay Miner bridge <------> Endpoint
type bridge struct {
	// ctx is used to stop the bridge when the context is canceled from either connection
	ctx    context.Context
	logger polylog.Logger

	// endpointConn is the connection to the WebSocket Endpoint
	endpointConn *websocketConnection
	// clientConn is the connection to the Client
	clientConn *websocketConnection

	// msgChan receives messages from the Client and Endpoint and passes them to the other side of the bridge.
	msgChan chan message

	// clientMessageHandler processes messages from the client before forwarding to the endpoint
	clientMessageHandler WebSocketMessageHandler
	// endpointMessageHandler processes messages from the endpoint before forwarding to the client
	endpointMessageHandler WebSocketMessageHandler

	// messageSuccessChan signals successful message processing to the gateway
	messageSuccessChan chan<- struct{}
	// messageErrorChan sends error information when message processing fails
	messageErrorChan chan<- error
}

// NewBridge creates a new Bridge instance with connections to both client and endpoint.
func NewBridge(
	logger polylog.Logger,
	clientWSSConn *websocket.Conn,
	endpointWSSConn *websocket.Conn,
	clientMessageHandler WebSocketMessageHandler,
	endpointMessageHandler WebSocketMessageHandler,
	messageSuccessChan chan<- struct{},
	messageErrorChan chan<- error,
) (*bridge, error) {
	logger = logger.With("component", "websocket_bridge")

	// Create a context that can be canceled from either connection
	ctx, cancelCtx := context.WithCancel(context.Background())

	// Create a channel to pass messages between the Client and Endpoint
	msgChan := make(chan message)

	// Create bridge instance
	b := &bridge{
		logger: logger,
		ctx:    ctx,

		msgChan:                msgChan,
		clientMessageHandler:   clientMessageHandler,
		endpointMessageHandler: endpointMessageHandler,
		messageSuccessChan:     messageSuccessChan,
		messageErrorChan:       messageErrorChan,
	}
	if err := b.validateComponents(); err != nil {
		cancelCtx() // Cancel context to prevent leak
		return nil, fmt.Errorf("❌ invalid bridge components: %w", err)
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
func (b *bridge) validateComponents() error {
	switch {
	case b.messageSuccessChan == nil:
		return fmt.Errorf("messageSuccessChan is nil")
	case b.messageErrorChan == nil:
		return fmt.Errorf("messageErrorChan is nil")
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
func (b *bridge) StartAsync() {
	b.logger.Info().Msg("🏗️ Websocket bridge operation started successfully")

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

// TODO_TECHDEBT(@adshmh): Add observations for client messages.
// This is needed to track e.g. whether the client closed the connection.
//
// handleClientMessage processes a message from the Client and sends it to the endpoint.
func (b *bridge) handleClientMessage(msg message) {
	// Process the message through the client message handler
	processedData, err := b.clientMessageHandler.HandleMessage(msg.data)
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
// The bridge notifies the gateway about message processing results through channels,
// allowing the gateway to handle observations without the bridge knowing about them.
func (b *bridge) handleEndpointMessage(msg message) {
	// Process the message through the endpoint message handler
	processedData, err := b.endpointMessageHandler.HandleMessage(msg.data)
	if err != nil {
		// Notify gateway about the error
		select {
		case b.messageErrorChan <- err:
		default:
			// Channel is full, log but don't block
			b.logger.Warn().Msg("messageErrorChan is full, dropping error notification")
		}
		b.endpointConn.handleDisconnect(fmt.Errorf("handleEndpointMessage: %w", err))
		return
	}

	// Send the processed message to the client
	if err := b.clientConn.WriteMessage(msg.messageType, processedData); err != nil {
		// NOTE: On session rollover, the Endpoint will disconnect the Endpoint connection, which will trigger this
		// error. This is expected and the Client is expected to handle the reconnection in their connection logic.

		// Notify gateway about the write error
		select {
		case b.messageErrorChan <- fmt.Errorf("error writing to client: %w", err):
		default:
			// Channel is full, log but don't block
			b.logger.Warn().Msg("messageErrorChan is full, dropping error notification")
		}
		b.clientConn.handleDisconnect(fmt.Errorf("handleEndpointMessage: error writing to client: %w", err))
		return
	}

	// Notify gateway about successful message processing
	select {
	case b.messageSuccessChan <- struct{}{}:
	default:
		// Channel is full, log but don't block
		b.logger.Warn().Msg("messageSuccessChan is full, dropping success notification")
	}
}
