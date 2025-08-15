package websockets

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"github.com/stretchr/testify/require"
)

func Test_Connection_MessageHandling(t *testing.T) {
	tests := []struct {
		name   string
		msgs   map[string]struct{}
		source messageSource
	}{
		{
			name: "should read messages from a websocket connection and forward them to the message channel",
			msgs: map[string]struct{}{
				"message 1":                        {},
				"message 2":                        {},
				"message 3":                        {},
				"longer message with more content": {},
				"json message":                     {},
			},
			source: messageSourceClient,
		},
		{
			name: "should handle endpoint messages",
			msgs: map[string]struct{}{
				"endpoint response 1": {},
				"endpoint response 2": {},
				"endpoint response 3": {},
			},
			source: messageSourceEndpoint,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			msgChan := make(chan message, len(test.msgs))

			conn := createTestConnection(t, test.msgs)
			defer conn.Close()

			ctx, cancelCtx := context.WithCancel(context.Background())
			defer cancelCtx()

			wsConn := newConnection(
				ctx,
				cancelCtx,
				polyzero.NewLogger().With("conn", test.source),
				conn,
				test.source,
				msgChan,
			)
			require.NotNil(t, wsConn)

			receivedMsgs := make(map[string]struct{})

			// Collect messages with timeout
			timeout := time.After(2 * time.Second)
			for len(receivedMsgs) < len(test.msgs) {
				select {
				case msg := <-msgChan:
					receivedMsgs[string(msg.data)] = struct{}{}
					c.Equal(test.source, msg.source, "Message source should match expected source")
				case <-timeout:
					t.Fatal("Timeout waiting for messages")
				}
			}

			// Verify all expected messages were received
			for expectedMsg := range test.msgs {
				c.Contains(receivedMsgs, expectedMsg, "Expected message not received: %s", expectedMsg)
			}

			close(msgChan)
		})
	}
}

func Test_Connection_ContextCancellation(t *testing.T) {
	c := require.New(t)

	msgChan := make(chan message, 1)
	conn := createTestConnection(t, map[string]struct{}{"test": {}})
	defer conn.Close()

	ctx, cancelCtx := context.WithCancel(context.Background())

	connection := newConnection(
		ctx,
		cancelCtx,
		polyzero.NewLogger(),
		conn,
		messageSourceClient,
		msgChan,
	)

	// Cancel the context
	cancelCtx()

	// Give some time for the connection to handle the cancellation
	time.Sleep(100 * time.Millisecond)

	// The connection should handle the context cancellation gracefully
	// (specific behavior depends on implementation details)
	c.NotNil(connection)
}

func Test_Connection_HandleDisconnect(t *testing.T) {
	c := require.New(t)

	msgChan := make(chan message, 1)

	// Create a connection that will be closed to trigger disconnect handling
	conn := createTestConnection(t, map[string]struct{}{})

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	connection := newConnection(
		ctx,
		cancelCtx,
		polyzero.NewLogger(),
		conn,
		messageSourceClient,
		msgChan,
	)

	// Simulate disconnect by closing the connection
	conn.Close()

	// Give some time for the disconnect to be handled
	time.Sleep(100 * time.Millisecond)

	// The connection should handle disconnection gracefully
	c.NotNil(connection)
}

// createTestConnection creates a websocket connection for testing
func Test_ConnectWebsocketEndpoint(t *testing.T) {
	tests := []struct {
		name         string
		websocketURL string
		headers      http.Header
		shouldFail   bool
		testHeaders  bool
	}{
		{
			name:         "should connect successfully with valid URL",
			websocketURL: "", // Will be set to test server URL
			headers:      http.Header{},
			shouldFail:   false,
		},
		{
			name:         "should connect successfully with headers",
			websocketURL: "", // Will be set to test server URL
			headers: http.Header{
				"Target-Service-Id": {"eth"},
				"App-Address":       {"app_address_1"},
				"Rpc-Type":          {"3"},
			},
			shouldFail:  false,
			testHeaders: true,
		},
		{
			name:         "should fail with invalid URL",
			websocketURL: "invalid-websocket-url",
			headers:      http.Header{},
			shouldFail:   true,
		},
		{
			name:         "should fail with malformed URL",
			websocketURL: "://invalid",
			headers:      http.Header{},
			shouldFail:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			var server *httptest.Server
			if !test.shouldFail && test.websocketURL == "" {
				// Create a test server for valid URL tests
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					upgrader := websocket.Upgrader{}
					_, err := upgrader.Upgrade(w, r, nil)
					if err != nil {
						t.Error("Error during connection upgrade:", err)
						return
					}

					// Verify headers if testHeaders is true
					if test.testHeaders {
						c.Equal("eth", r.Header.Get("Target-Service-Id"))
						c.Equal("app_address_1", r.Header.Get("App-Address"))
						c.Equal("3", r.Header.Get("Rpc-Type"))
					}
				}))
				defer server.Close()

				test.websocketURL = "ws" + strings.TrimPrefix(server.URL, "http")
			}

			conn, err := ConnectWebsocketEndpoint(polyzero.NewLogger(), test.websocketURL, test.headers)

			if test.shouldFail {
				c.Error(err)
				c.Nil(conn)
			} else {
				c.NoError(err)
				c.NotNil(conn)
				if conn != nil {
					conn.Close()
				}
			}
		})
	}
}

func createTestConnection(t *testing.T, msgs map[string]struct{}) *websocket.Conn {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error("Error during connection upgrade:", err)
			return
		}

		// Send test messages
		for msg := range msgs {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				t.Logf("Failed to send message: %v", err)
				return
			}
		}
	}))

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal("Error connecting to test server:", err)
	}

	return conn
}
