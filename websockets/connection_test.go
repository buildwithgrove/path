package websockets

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"github.com/stretchr/testify/require"
)

func Test_connectEndpoint(t *testing.T) {
	tests := []struct {
		name          string
		nodeURL       string
		expectedError bool
	}{
		{
			name:          "should connect successfully",
			nodeURL:       "ws://localhost:8080",
			expectedError: false,
		},
		{
			name:          "should fail to connect with invalid URL",
			nodeURL:       "http://invalid-url",
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				upgrader := websocket.Upgrader{}
				_, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					t.Error("Error during connection upgrade:", err)
					return
				}
			}))
			defer server.Close()

			u, _ := url.Parse(test.nodeURL)
			u.Host = strings.TrimPrefix(server.URL, "http://")
			nodeURL := u.String()

			conn, err := connectEndpoint(nodeURL)
			if test.expectedError {
				c.Error(err)
			} else {
				c.NoError(err)
				c.NotNil(conn)
				conn.Close()
			}
		})
	}
}

func Test_connection(t *testing.T) {
	tests := []struct {
		name   string
		msgs   map[string]struct{}
		source messageSource
	}{
		{
			name: "should read messages from a websockets connection and successfully forward them to the message channel",
			msgs: map[string]struct{}{
				`{"jsonrpc":"2.0","id":1,"method":"eth_gasPrice"}`:              {},
				`{"jsonrpc":"2.0","id":1,"method":"eth_blockNumber"}`:           {},
				`{"jsonrpc":"2.0","id":1,"method":"eth_getBalance"}`:            {},
				`{"jsonrpc":"2.0","id":1,"method":"eth_getTransactionCount"}`:   {},
				`{"jsonrpc":"2.0","id":1,"method":"eth_getBlockByNumber"}`:      {},
				`{"jsonrpc":"2.0","id":1,"method":"eth_getBlockByHash"}`:        {},
				`{"jsonrpc":"2.0","id":1,"method":"eth_getTransactionByHash"}`:  {},
				`{"jsonrpc":"2.0","id":1,"method":"eth_getTransactionReceipt"}`: {},
				`{"jsonrpc":"2.0","id":1,"method":"eth_call"}`:                  {},
				`{"jsonrpc":"2.0","id":1,"method":"eth_estimateGas"}`:           {},
			},
			source: messageSourceClient,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			msgChan := make(chan message)
			stopChan := make(chan error)

			conn := testConn(t, test.msgs)
			defer conn.Close()

			_ = newConnection(
				polyzero.NewLogger().With("conn", test.source),
				conn,
				test.source,
				msgChan,
				stopChan,
			)

			receivedMsgs := make(map[string]struct{})
			go func() {
				for msg := range msgChan {
					receivedMsgs[string(msg.data)] = struct{}{}
				}
			}()

			<-time.After(2 * time.Second)

			close(stopChan)
			close(msgChan)

			for msg := range test.msgs {
				c.Contains(receivedMsgs, msg)
			}
		})
	}
}

// testConn creates a new httptest server, upgrades the connection to a websocket
// and sends the messages to the message channel to be read by the connection.
func testConn(t *testing.T, msgs map[string]struct{}) *websocket.Conn {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error("Error during connection upgrade:", err)
			return
		}

		for msg := range msgs {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				t.Fatalf("failed to send message: %v", err)
			}
		}
	}))

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Error(err)
	}

	return conn
}
