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
	endpointConn *connection
	clientConn   *connection
	msgChan      <-chan message
	stopChan     chan error

	log polylog.Logger
}

// NewBridge creates a new Bridge instance and a new connection to the Endpoint from the Endpoint URL
func NewBridge(endpointURL string, clientWSSConn *websocket.Conn, log polylog.Logger) (*bridge, error) {
	endpointWSSConn, err := connectEndpoint(endpointURL)
	if err != nil {
		return nil, fmt.Errorf("error establishing connection to endpoint URL %s: %s", endpointURL, err.Error())
	}

	msgChan := make(chan message)
	stopChan := make(chan error)

	log = log.With("component", "bridge")

	endpointConnection := newConnection(
		endpointWSSConn,
		messageSourceEndpoint,
		msgChan,
		stopChan,
		log.With("conn", "endpoint"),
	)
	clientConnection := newConnection(
		clientWSSConn,
		messageSourceClient,
		msgChan,
		stopChan,
		log.With("conn", "client"),
	)

	return &bridge{
		endpointConn: endpointConnection,
		clientConn:   clientConnection,
		msgChan:      msgChan,
		stopChan:     stopChan,

		log: log,
	}, nil
}

/* ---------- Public method - Run Bridge ---------- */

// Run starts the bridge and establishes a bidirectional communication between the wss manager and server
func (b *bridge) Run() {
	// Start goroutine to read messages from message channel
	go b.messageLoop()

	b.log.Info().Msg("bridge operation started successfully")

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
		b.log.Error().Err(err).Msg("error writing to endpoint websocket")
		b.stopChan <- fmt.Errorf("error writing to endpoint websocket: %w", err)
		return
	}
}

// handleEndpointMessage processes a message from the Endpoint and sends it to the Client
func (b *bridge) handleEndpointMessage(msg message) {
	if err := b.clientConn.WriteMessage(msg.messageType, msg.data); err != nil {
		b.log.Error().Err(err).Msg("error writing to client websocket")
		b.stopChan <- fmt.Errorf("error writing to client websocket: %w", err)
		return
	}
}
