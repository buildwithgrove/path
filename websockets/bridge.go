package websockets

import (
	"context"
	"fmt"
	"net/http"

	"github.com/buildwithgrove/path/observation"
	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// WebSocketMessageHandler handles websocket messages.
// It can, for example, transform the message data before forwarding it.
type WebsocketMessageProcessor interface {
	// ProcessClientWebsocketMessage processes a message from the client.
	ProcessClientWebsocketMessage([]byte) ([]byte, error)

	// ProcessEndpointWebsocketMessage processes a message from the endpoint.
	ProcessEndpointWebsocketMessage([]byte) ([]byte, *observation.RequestResponseObservations, error)
}

// bridge routes data between an Endpoint and a Client.
// One bridge represents a single WebSocket connection
// between a Client and a WebSocket Endpoint.
//
// This is a generic websocket bridge that handles the websocket protocol
// and message routing, while delegating Gateway-level message processing to the
// provided message handler.
//
// Architecture:
// - Protocol-agnostic: Works with any protocol (Shannon, future protocols)
// - Message handler: Gateway-level processing of messages. Orchestrates protocol and QoS-level message processing.
// - Notification channels: Gateway-level notifications to trigger sending Observations.
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
	// It is an internal channel used only by the bridge and not exposed to any other package.
	msgChan chan message

	// websocketMessageProcessor processes messages from the client and endpoint.
	websocketMessageProcessor WebsocketMessageProcessor

	// messageObservationsChan receives message observations from the websocketMessageProcessor.
	messageObservationsChan chan<- *observation.RequestResponseObservations
}

// StartBridge creates a new Bridge instance with connections to both client and endpoint
// and starts the bridge in a goroutine to avoid blocking the main thread.
func StartBridge(
	logger polylog.Logger,
	req *http.Request,
	w http.ResponseWriter,
	websocketURL string,
	headers http.Header,
	websocketMessageProcessor WebsocketMessageProcessor,
	messageObservationsChan chan<- *observation.RequestResponseObservations,
) error {
	logger = logger.With("component", "websocket_bridge")

	// Create a context that can be canceled from either connection
	ctx, cancelCtx := context.WithCancel(context.Background())

	// Upgrade HTTP request from client to websocket connection.
	clientConn, err := upgradeClientWebsocketConnection(logger, req, w)
	if err != nil {
		return fmt.Errorf("createWebsocketBridge: %s", err.Error())
	}

	// Connect to the Relay Miner endpoint
	endpointConn, err := connectWebsocketEndpoint(logger, websocketURL, headers)
	if err != nil {
		return fmt.Errorf("createWebsocketBridge: %s", err.Error())
	}

	// Create a channel to pass messages between the Client and Endpoint
	msgChan := make(chan message)

	// Create bridge instance
	b := &bridge{
		logger: logger,
		ctx:    ctx,

		msgChan:                   msgChan,
		websocketMessageProcessor: websocketMessageProcessor,

		messageObservationsChan: messageObservationsChan,
	}
	if err := b.validateComponents(); err != nil {
		cancelCtx() // Cancel context to prevent leak
		return fmt.Errorf("❌ invalid bridge components: %w", err)
	}

	// Initialize connections with context and cancel function
	b.endpointConn = newConnection(
		b.ctx,
		cancelCtx,
		logger.With("conn", "endpoint"),
		endpointConn,
		messageSourceEndpoint,
		msgChan,
	)
	b.clientConn = newConnection(
		b.ctx,
		cancelCtx,
		logger.With("conn", "client"),
		clientConn,
		messageSourceClient,
		msgChan,
	)

	// Start the bridge in a goroutine to avoid blocking the main thread.
	go b.start()
	return nil
}

// validateComponents ensures the Bridge is not created with nil components.
// This is done to avoid panics and to make the Bridge's behavior more predictable.
func (b *bridge) validateComponents() error {
	switch {
	case b.messageObservationsChan == nil:
		return fmt.Errorf("messageObservationsChan is nil")
	case b.websocketMessageProcessor == nil:
		return fmt.Errorf("websocketMessageProcessor is nil")
	}
	return nil
}

// Start starts the bridge and establishes a bidirectional communication
// through PATH between the Client and the selected websocket endpoint.
//
// This method implements the gateway.WebsocketsBridge interface.
//
// Full data flow: Client <---clientConn---> PATH Bridge <---endpointConn---> Relay Miner Bridge <------> Endpoint
func (b *bridge) start() {
	b.logger.Info().Msg("🏗️ Websocket bridge operation started successfully")

	// Listen for the context to be canceled and shut down the bridge
	go func() {
		<-b.ctx.Done()
		b.shutdown(fmt.Errorf("context canceled"))
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
func (b *bridge) shutdown(err error) {
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
	processedData, err := b.websocketMessageProcessor.ProcessClientWebsocketMessage(msg.data)
	if err != nil {
		b.logger.Error().Err(err).Msg("❌ error processing client message, disconnecting client connection")

		b.clientConn.handleDisconnect(fmt.Errorf("handleClientMessage: %w", err))
		return
	}

	b.logger.Debug().Msgf("🔗 client message successfully processed, sending message to endpoint: %s", string(processedData))

	// Send the processed message to the endpoint
	if err := b.endpointConn.WriteMessage(msg.messageType, processedData); err != nil {
		b.logger.Error().Err(err).Msg("❌ error writing client message to endpoint, disconnecting endpoint connection")

		b.endpointConn.handleDisconnect(fmt.Errorf("handleClientMessage: error writing to endpoint: %w", err))
		return
	}
}

// TODO_IN_THIS_PR(@commoddity): Clean up message channel sending in the below method.

// handleEndpointMessage processes a message from the Endpoint and sends it to the Client.
// The bridge notifies the gateway about message processing results through channels,
// allowing the gateway to handle observations without the bridge knowing about them.
func (b *bridge) handleEndpointMessage(msg message) {
	// Process the message through the endpoint message handler
	processedData, msgObservations, err := b.websocketMessageProcessor.ProcessEndpointWebsocketMessage(msg.data)

	// Notify gateway about message processing results
	defer b.sendMessageObservations(msgObservations)

	if err != nil {
		b.logger.Error().Err(err).Msg("❌ error processing endpoint message, disconnecting endpoint connection")
		b.endpointConn.handleDisconnect(fmt.Errorf("handleEndpointMessage: %w", err))
		return
	}

	// Send the processed message to the client
	// NOTE: On session rollover, the Endpoint will disconnect the Endpoint connection, which will trigger this
	// error. This is expected and the Client is expected to handle the reconnection in their connection logic.
	if err := b.clientConn.WriteMessage(msg.messageType, processedData); err != nil {
		b.logger.Error().Err(err).Msg("❌ error writing endpoint message to client, disconnecting client connection")
		b.clientConn.handleDisconnect(fmt.Errorf("handleEndpointMessage: error writing to client: %w", err))
		return
	}

	b.logger.Debug().Msgf("🔗 endpoint message successfully processed, sending message to client: %s", string(processedData))

}

// sendMessageObservations sends message observations to the gateway.
// If the channel is full, the message observations are dropped.
// This is done to avoid blocking the main thread.
func (b *bridge) sendMessageObservations(msgObservations *observation.RequestResponseObservations) {
	select {
	case b.messageObservationsChan <- msgObservations:
	default:
		// Channel is full, log but don't block
		b.logger.Warn().Msg("messageObservationsChan is full, dropping message observations")
	}
}
