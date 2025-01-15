package websockets

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 30 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// Bridge routes data between Client and PATH.
// One bridge represents exactly one WebSocket connection between the Client and a WebSocket Endpoint.
// Full data flow: Client <------> PATH <------> WebSocket Endpoint
type bridge struct {
	logger       polylog.Logger
	endpointConn *connection
	clientConn   *connection
	msgChan      <-chan message
	stopChan     chan error
}

// NewBridge creates a new Bridge instance and a new connection to the Endpoint from the Endpoint URL
func NewBridge(logger polylog.Logger, endpointURL string, clientWSSConn *websocket.Conn) (*bridge, error) {
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
		endpointConn: endpointConnection,
		clientConn:   clientConnection,
		msgChan:      msgChan,
		stopChan:     stopChan,

		logger: logger,
	}, nil
}

/* ---------- Public method - Run Bridge ---------- */

// Run starts the bridge and establishes a bidirectional communication
// through PATH between the Client and the selected websocket endpoint.
//
// Full data flow: Client <------> PATH <------> WebSocket Endpoint
func (b *bridge) Run() {
	// Start goroutine to read messages from message channel
	go b.messageLoop()

	b.logger.Info().Msg("bridge operation started successfully")

	// If close signal is received, stop the bridge and close both connections
	<-b.stopChan
}

/* ---------- Private methods - Message loop ---------- */

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
	if err := b.endpointConn.WriteMessage(msg.messageType, msg.data); err != nil {
		b.endpointConn.handleError(err, messageSourceEndpoint)
		return
	}
}

// handleEndpointMessage processes a message from the Endpoint and sends it to the Client
func (b *bridge) handleEndpointMessage(msg message) {
	if err := b.clientConn.WriteMessage(msg.messageType, msg.data); err != nil {
		b.clientConn.handleError(err, messageSourceClient)
		return
	}
}
