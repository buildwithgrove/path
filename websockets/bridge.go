package websockets

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// bridge routes data between an Endpoint and a Client.
// One bridge represents a single WebSocket connection between a
// Client and a WebSocket Endpoint.
//
// Full data flow: Client <------> PATH Bridge <------> WebSocket Endpoint
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

	// endpointURL is the URL of the WebSocket Endpoint
	// selected by PATH to be used for the WebSocket connection.
	endpointURL string
}

// NewBridge creates a new Bridge instance and a new connection to the Endpoint from the Endpoint URL
func NewBridge(
	logger polylog.Logger,
	endpointURL string,
	clientWSSConn *websocket.Conn,
) (*bridge, error) {
	endpointWSSConn, err := connectEndpoint(endpointURL)
	if err != nil {
		return nil, fmt.Errorf("error establishing connection to endpoint URL %s: %s", endpointURL, err.Error())
	}

	msgChan := make(chan message)
	stopChan := make(chan error)

	logger = logger.With("component", "bridge")

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
		logger:       logger,
		endpointConn: endpointConnection,
		clientConn:   clientConnection,
		msgChan:      msgChan,
		stopChan:     stopChan,
		endpointURL:  endpointURL,
	}, nil
}

// Run starts the bridge and establishes a bidirectional communication
// through PATH between the Client and the selected websocket endpoint.
//
// Full data flow: Client <------> PATH Bridge <------> WebSocket Endpoint
func (b *bridge) Run() {
	// Start goroutine to read messages from message channel
	go b.messageLoop()

	b.logger.Info().Str("endpoint_url", b.endpointURL).Msg("bridge operation started successfully")

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
func (b *bridge) handleClientMessage(msg message) {
	b.logger.Debug().Msgf("received message from client: %s", string(msg.data))

	if err := b.endpointConn.WriteMessage(msg.messageType, msg.data); err != nil {
		b.endpointConn.handleError(err, messageSourceEndpoint)
		return
	}
}

// handleEndpointMessage processes a message from the Endpoint and sends it to the Client
func (b *bridge) handleEndpointMessage(msg message) {
	b.logger.Debug().Msgf("received message from endpoint: %s", string(msg.data))

	if err := b.clientConn.WriteMessage(msg.messageType, msg.data); err != nil {
		b.clientConn.handleError(err, messageSourceClient)
		return
	}
}
