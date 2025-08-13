package websockets

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog"
)

const (
	// Time allowed (in seconds) to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed (in seconds) to read the next pong message from the peer.
	pongWait = 30 * time.Second

	// Send pings to peer with this period (in seconds).
	// Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// messageSource is used to identify the source of a message in a bidirectional websocket connection.
// Possible values are `client` and `endpoint`.
//
// Full data flow: Client <------> PATH <------> WebSocket Endpoint
type messageSource string

const (
	messageSourceClient   messageSource = "client"
	messageSourceEndpoint messageSource = "endpoint"
)

// message represents a websocket message that can be:
// - Client request
// - Endpoint response
// - Subscription push event (e.g. eth_subscribe)
type message struct {
	// data is the message payload
	data []byte

	// source may be either `client` or `endpoint`
	source messageSource

	// messageType is an int returned by the gorilla/websocket package
	messageType int
}

// websocketConnection represents a websocket connection between PATH and:
// - A client
// - An endpoint
type websocketConnection struct {
	*websocket.Conn

	ctx       context.Context
	cancelCtx context.CancelFunc

	logger polylog.Logger

	source  messageSource
	msgChan chan<- message
}

// UpgradeClientWebsocketConnection upgrades an HTTP connection to WebSocket.
// Used to upgrade a Client's HTTP request to a WebSocket connection.
//
// Note: This function uses a permissive CheckOrigin policy (always returns true),
// eliminating origin-based rejections as a potential cause of upgrade failures.
//
// See: https://pkg.go.dev/github.com/gorilla/websocket#hdr-Overview
func UpgradeClientWebsocketConnection(
	wsLogger polylog.Logger,
	req *http.Request,
	w http.ResponseWriter,
) (*websocket.Conn, error) {
	upgrader := websocket.Upgrader{
		// Allow all origins.
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	clientConn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		// Upgrade errors are often client-side protocol violations, but can also
		// indicate server resource issues or network problems. The specific error
		// message will help distinguish between client and server-side causes.
		wsLogger.Error().Err(err).Msg("Error upgrading websocket connection request")
		return nil, err
	}

	return clientConn, nil
}

// ConnectWebsocketEndpoint makes a websocket connection to the websocket Endpoint.
func ConnectWebsocketEndpoint(logger polylog.Logger, websocketURL string, headers http.Header) (*websocket.Conn, error) {
	logger.Info().Msgf("🔗 Connecting to websocket endpoint: %s", websocketURL)

	// Ensure the websocket URL is valid.
	u, err := url.Parse(websocketURL)
	if err != nil {
		logger.Error().Err(err).Msgf("❌ Error parsing endpoint URL: %s", websocketURL)
		return nil, err
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

// newConnection creates a new websocket connection wrapper.
func newConnection(
	ctx context.Context,
	cancelCtx context.CancelFunc,
	logger polylog.Logger,
	conn *websocket.Conn,
	source messageSource,
	msgChan chan message,
) *websocketConnection {
	c := &websocketConnection{
		ctx:       ctx,
		cancelCtx: cancelCtx,

		logger: logger.With("connection", source),

		Conn: conn,

		source:  source,
		msgChan: msgChan,
	}

	go c.connLoop()
	go c.pingLoop()

	return c
}

// connLoop reads messages from the websocket connection and sends them to the bridge's msgChan
func (c *websocketConnection) connLoop() {
	for {
		messageType, msg, err := c.ReadMessage()
		if err != nil {
			c.handleDisconnect(err)
			return
		}

		c.msgChan <- message{
			data:        msg,
			source:      c.source,
			messageType: messageType,
		}
	}
}

// handleDisconnect handles any disconnection issues from the websocket connection.
// This includes both:
// - Expected disconnections (e.g. when the RelayMiner disconnects on session rollover)
// - Unexpected disconnections (e.g. when the connection is lost due to network issues)
//
// This function will cancel the context to signal the bridge to handle shutdown.
func (c *websocketConnection) handleDisconnect(err error) {
	c.logger.Warn().Err(err).Msgf("🔌 Handling websocket disconnection")
	c.cancelCtx() // Cancel the context to signal the bridge to handle shutdown
}

// pingLoop sends keep-alive ping messages to the connection and handles pong messages
// This loop is used to keep the connection alive and functions by sending a ping message
// to the connection and waiting for a pong response. If a pong response is not received,
// the connection is considered dead and the stopChan is closed.
// See: https://pkg.go.dev/github.com/gorilla/websocket#hdr-Control_Messages
func (c *websocketConnection) pingLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	if err := c.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		c.logger.Error().Err(err).Msg("failed to set initial read deadline")
	}

	c.SetPongHandler(func(string) error {
		if err := c.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			c.logger.Error().Err(err).Msg("failed to set pong handler read deadline")
		}
		return nil
	})

	for {
		select {
		case <-ticker.C:
			if err := c.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait)); err != nil {
				c.logger.Error().Err(err).Msg("failed to send ping to connection")
				c.handleDisconnect(fmt.Errorf("failed to send ping: %w", err))
				return
			}

		case <-c.ctx.Done():
			c.logger.Info().Msg("pingLoop stopped due to context cancellation")
			return
		}
	}
}
