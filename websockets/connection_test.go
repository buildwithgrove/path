package websockets

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	"github.com/stretchr/testify/require"
)

func Test_connectEndpoint(t *testing.T) {
	tests := []struct {
		name                string
		getSelectedEndpoint func(testServerURL string) *selectedEndpoint
		expectedError       bool
	}{
		{
			name: "should connect successfully",
			getSelectedEndpoint: func(testServerURL string) *selectedEndpoint {
				baseURL := "ws://localhost:8080"
				u, _ := url.Parse(baseURL)
				u.Host = strings.TrimPrefix(testServerURL, "http://")
				nodeURL := u.String()
				return &selectedEndpoint{
					url: nodeURL,
					session: &sessiontypes.Session{
						SessionId: "1",
						Header: &sessiontypes.SessionHeader{
							ServiceId:          "service_id",
							ApplicationAddress: "application_address",
						},
						Application: &apptypes.Application{
							Address: "application_address",
						},
					},
					supplier: "supplier",
				}
			},
			expectedError: false,
		},
		{
			name: "should fail to connect with invalid URL",
			getSelectedEndpoint: func(testServerURL string) *selectedEndpoint {
				return &selectedEndpoint{
					url: "http://invalid-url",
					session: &sessiontypes.Session{
						SessionId: "1",
						Header: &sessiontypes.SessionHeader{
							ServiceId:          "service_id",
							ApplicationAddress: "application_address",
						},
						Application: &apptypes.Application{
							Address: "application_address",
						},
					},
					supplier: "supplier",
				}
			},
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

			selectedEndpoint := test.getSelectedEndpoint(server.URL)

			conn, err := connectWebsocketEndpoint(polyzero.NewLogger(), selectedEndpoint)
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

			conn := testConn(t, test.msgs)
			defer conn.Close()

			ctx, cancelCtx := context.WithCancel(context.Background())

			_ = newConnection(
				ctx,
				cancelCtx,
				polyzero.NewLogger().With("conn", test.source),
				conn,
				test.source,
				msgChan,
			)

			receivedMsgs := make(map[string]struct{})
			go func() {
				for msg := range msgChan {
					receivedMsgs[string(msg.data)] = struct{}{}
				}
			}()

			<-time.After(2 * time.Second)

			close(msgChan)
			cancelCtx()

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
