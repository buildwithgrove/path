package websockets

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"

	"github.com/buildwithgrove/path/request"
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

// messageSource is used to identify the source of a message in a bidrectional websocket connection.
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

// connection represents a websocket connection between PATH and:
// - A client
// - An endpoint
type connection struct {
	ctx       context.Context
	cancelCtx context.CancelFunc

	logger polylog.Logger

	*websocket.Conn

	source  messageSource
	msgChan chan<- message
}

// connectEndpoint makes a websocket connection to the websocket Endpoint.
func connectEndpoint(selectedEndpoint SelectedEndpoint) (*websocket.Conn, error) {
	u, err := url.Parse(selectedEndpoint.PublicURL())
	if err != nil {
		return nil, err
	}

	headers := getBridgeRequestHeaders(selectedEndpoint.Session())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// getBridgeRequestHeaders returns the headers that should be sent to the RelayMiner
// when establishing a new websocket connection to the Endpoint.
func getBridgeRequestHeaders(session *sessiontypes.Session) http.Header {
	headers := http.Header{}
	headers.Add(request.HTTPHeaderTargetServiceID, session.Header.ServiceId)
	headers.Add(request.HTTPHeaderAppAddress, session.Header.ApplicationAddress)
	return headers
}

// connectClient initiates a websocket connection to the client.
func newConnection(
	ctx context.Context,
	cancelCtx context.CancelFunc,
	logger polylog.Logger,
	conn *websocket.Conn,
	source messageSource,
	msgChan chan message,
) *connection {
	c := &connection{
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
func (c *connection) connLoop() {
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
func (c *connection) handleDisconnect(err error) {
	c.logger.Info().Err(err).Msgf("handling error in websocket connection")
	c.cancelCtx() // Cancel the context to signal the bridge to handle shutdown
}

// pingLoop sends keep-alive ping messages to the connection and handles pong messages
// This loop is used to keep the connection alive and functions by sending a ping message
// to the connection and waiting for a pong response. If a pong response is not received,
// the connection is considered dead and the stopChan is closed.
// See: https://pkg.go.dev/github.com/gorilla/websocket#hdr-Control_Messages
func (c *connection) pingLoop() {
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
