package websockets

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"

	"github.com/buildwithgrove/path/request"
)

// TODO_TECHDEBT(@commoddity): remove deprecated `x-` prefix from the header name
// here and in Relay Miner `async.go` file and move const to the request package.
const headerAppAddress = "X-App-Address"

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
	*websocket.Conn

	logger polylog.Logger

	source   messageSource
	msgChan  chan<- message
	stopChan chan error
}

// connectEndpoint makes a websocket connection to the websocket Endpoint.
func connectEndpoint(selectedEndpoint SelectedEndpoint) (*websocket.Conn, error) {
	u, err := url.Parse(selectedEndpoint.PublicURL())
	if err != nil {
		return nil, err
	}

	// TODO_TECHDEBT(@commoddity): Remove this switch once RelayMiner
	// supports setting websocket URLs as `backend_url`.
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
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
	headers.Add(headerAppAddress, session.Header.ApplicationAddress)
	return headers
}

// connectClient initiates a websocket connection to the client.
func newConnection(
	logger polylog.Logger,
	conn *websocket.Conn,
	source messageSource,
	msgChan chan message,
	stopChan chan error,
) *connection {
	c := &connection{
		Conn: conn,

		logger: logger,

		source:   source,
		msgChan:  msgChan,
		stopChan: stopChan,
	}

	go c.connLoop()
	go c.pingLoop()

	return c
}

// connLoop reads messages from the websocket connection and sends them to the bridge's msgChan
func (c *connection) connLoop() {
	for {
		select {
		case err := <-c.stopChan:
			if err := c.cleanup(err); err != nil {
				c.logger.Error().Err(err).Msg("error cleaning up connection")
			}
			return

		default:
			messageType, msg, err := c.ReadMessage()
			if err != nil {
				c.handleError(err, c.source)
				return
			}

			c.msgChan <- message{
				data:        msg,
				source:      c.source,
				messageType: messageType,
			}
		}
	}
}

// cleanup closes the client and gateway connections
func (c *connection) cleanup(err error) error {
	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, err.Error())

	// Close the connection and send a reason for the closure
	if err := c.WriteMessage(websocket.CloseMessage, closeMsg); err != nil {
		c.logger.Error().Err(err).Msg("error writing close message to connection")
		return err
	}
	if err := c.Close(); err != nil {
		c.logger.Error().Err(err).Msg("error closing connection")
		return err
	}

	return nil
}

// handleError handles errors from the websocket connection and sends them to the stopChan if applicable
func (c *connection) handleError(err error, source messageSource) {
	if websocket.IsCloseError(err, websocket.CloseNoStatusReceived) {
		c.logger.Info().Msgf("%s connection closed by peer", source)
	} else {
		c.logger.Error().Err(err).Msgf(" %s error reading from connection", source)
	}

	select {
	case <-c.stopChan:
		// stopChan is already closed, do nothing
	default:
		// stopChan is still open, send the error
		c.stopChan <- fmt.Errorf("error reading from %s connection: %w", source, err)
	}
}

// pingLoop sends keep-alive ping messages to the connection and handles pong messages
// This loop is used to keep the connection alive and functions by sending a ping message
// to the connection and waiting for a pong response. If a pong response is not received,
// the connection is considered dead and the stopChan is closed.
// See: https://pkg.go.dev/github.com/gorilla/websocket#hdr-Control_Messages
func (c *connection) pingLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()

	// Initialize the ping loop by setting the read deadline
	if err := c.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		c.logger.Error().Err(err).Msg("failed to set initial read deadline")
	}

	// Extend read deadline on pong response, ie. when a ping response is received,
	// the loop extends the read deadline for the ping/pong interval to keep
	// the websocket connection alive.
	c.SetPongHandler(func(string) error {
		if err := c.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			c.logger.Error().Err(err).Msg("failed to set pong handler read deadline")
		}

		return nil
	})

	for {
		select {
		case <-c.stopChan:
			return

		case <-ticker.C:
			if err := c.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait)); err != nil {
				c.logger.Error().Err(err).Msg("failed to send ping to connection")
				c.stopChan <- fmt.Errorf("failed to send ping to connection: %w", err)
				return
			}
		}
	}
}
