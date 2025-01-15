package websockets

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// MessageSource is used to identify the source of a message
// Possible values are `client` and `endpoint`.
// Full data flow: Client <------> PATH <------> WebSocket Endpoint
type messageSource string

const (
	messageSourceClient   messageSource = "client"
	messageSourceEndpoint messageSource = "endpoint"
)

type message struct {
	data        []byte
	source      messageSource
	messageType int
}

type connection struct {
	*websocket.Conn
	logger   polylog.Logger
	source   messageSource
	msgChan  chan<- message
	stopChan chan error
}

// connectEndpoint makes a websocket connection to the websocket Endpoint.
func connectEndpoint(endpointURL string) (*websocket.Conn, error) {
	u, err := url.Parse(endpointURL)
	if err != nil {
		return nil, err
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{})
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func newConnection(
	logger polylog.Logger,
	conn *websocket.Conn,
	source messageSource,
	msgChan chan message,
	stopChan chan error,
) *connection {
	c := &connection{
		logger:   logger,
		Conn:     conn,
		source:   source,
		msgChan:  msgChan,
		stopChan: stopChan,
	}

	go c.connLoop()
	go c.pingLoop()

	return c
}

// wssLoop reads messages from the websocket connection and sends them to the bridge's msgChan
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

	c.stopChan <- fmt.Errorf("error reading from %s connection: %w", source, err)
}

// pingLoop sends keep-alive ping messages to the connection and handles pong messages
func (c *connection) pingLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()

	initPingLoop := func() {
		// Set initial read deadline
		if err := c.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			c.logger.Error().Err(err).Msg("failed to set initial read deadline")
		}
		// Extend read deadline on pong response
		c.SetPongHandler(func(string) error {
			if err := c.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
				c.logger.Error().Err(err).Msg("failed to set pong handler read deadline")
			}

			return nil
		})
	}

	initPingLoop()

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
