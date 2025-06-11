package websockets

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"
	"github.com/stretchr/testify/require"
)

type (
	clientReq         string // clientReq is a single JSON RPC request sent from the client to the endpoint over the websocket
	endpointResp      string // endpointResp is a single JSON RPC response sent from the endpoint to the client over the websocket
	subscriptionEvent string // subscriptionEvent is a single subscription push event sent from the endpoint to the client over the websocket
)

var capturedMessages struct {
	sync.Mutex
	clientRequests     map[clientReq]struct{}         // clientRequests is a map of client requests sent to the endpoint
	endpointResponses  map[endpointResp]struct{}      // endpointResponses is a map of endpoint responses sent to the client
	subscriptionEvents map[subscriptionEvent]struct{} // subscriptionEvents is a map of subscription events sent to the client
}

func Test_Bridge_Run(t *testing.T) {
	tests := []struct {
		name               string
		selectedEndpoint   *selectedEndpoint
		jsonrpcRequests    map[clientReq]endpointResp
		subscriptionEvents map[subscriptionEvent]struct{}
	}{
		{
			name: "should forward regular JSON RPC messages from Client to Endpoint and receive response",
			selectedEndpoint: &selectedEndpoint{
				url: "", // Assigned in test to the value of the `url` returned by `testEndpointConnURL`
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
			},
			jsonrpcRequests: map[clientReq]endpointResp{
				`{"jsonrpc":"2.0","id":1,"method":"eth_gasPrice"}`:                                      `{"jsonrpc":"2.0","id":1,"result":"0x337d04a3b"}`,
				`{"jsonrpc":"2.0","id":1,"method":"eth_blockNumber"}`:                                   `{"jsonrpc":"2.0","id":1,"result":"0x12c1b21"}`,
				`{"jsonrpc":"2.0","id":1,"method":"eth_subscribe","params":["newPendingTransactions"]}`: `{"id":1,"result":"0xf13f7073ddef66a8c1b0c9c9f0e543c3","jsonrpc":"2.0"}`,
			},
		},
		{
			name: "should forward subscription push events from the Endpoint to the Client",
			selectedEndpoint: &selectedEndpoint{
				url: "", // Assigned in test to the value of the `url` returned by `testEndpointConnURL`
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
			},
			subscriptionEvents: map[subscriptionEvent]struct{}{
				`{"jsonrpc":"2.0","method":"eth_subscription","params":{"result":"0x35f48044467e5ec65fd536665cd7dffe0664ff14d47d0ca4cd8c5618712bd550","subscription":"0x995f694478fb6d1e56bba87e9bb4405a"}}`: {},
				`{"jsonrpc":"2.0","method":"eth_subscription","params":{"result":"0xf819e6c2b499e26ad305c1e4ed342ba16fb43593353d00cd11f42555b187df48","subscription":"0x995f694478fb6d1e56bba87e9bb4405a"}}`: {},
				`{"jsonrpc":"2.0","method":"eth_subscription","params":{"result":"0x6232a3964ae5cf4df035b7e43cf6be8ac44cfd142a26eeccb27ef59f6621b384","subscription":"0x995f694478fb6d1e56bba87e9bb4405a"}}`: {},
				`{"jsonrpc":"2.0","method":"eth_subscription","params":{"result":"0x54978b9b5bc5de6c70e7ba9bfd4ff1255bf35a2fc2862322752b53b90b367037","subscription":"0x995f694478fb6d1e56bba87e9bb4405a"}}`: {},
				`{"jsonrpc":"2.0","method":"eth_subscription","params":{"result":"0x74576327a742af7c9bd4d2f9fa5b75f9911667322c557bda5b2c8df714cafde5","subscription":"0x995f694478fb6d1e56bba87e9bb4405a"}}`: {},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			// Reset captured messages before each test
			capturedMessages.clientRequests = make(map[clientReq]struct{})
			capturedMessages.endpointResponses = make(map[endpointResp]struct{})
			capturedMessages.subscriptionEvents = make(map[subscriptionEvent]struct{})

			// Create an HTTP test server with a websocket handler to represent a Client connection
			// Return the websocket connection to pass to NewBridge
			clientConn := testClientConn(t, test.jsonrpcRequests)

			// Create an HTTP test server with a websocket handler to represent an Endpoint connection
			// Return the URL of the test server to pass to NewBridge
			testEndpointConnURL := testEndpointConnURL(t, test.jsonrpcRequests, test.subscriptionEvents)
			test.selectedEndpoint.url = testEndpointConnURL

			// Call NewBridge with the clientConn and testEndpointConnURL
			bridge, err := NewBridge(
				polyzero.NewLogger(),
				clientConn,
				test.selectedEndpoint,
				&relayRequestSigner{},
				&fullNode{},
			)
			c.NoError(err)

			// Start the bridge in a goroutine
			go bridge.Run()

			// Wait for a short duration for test requests and events to get sent
			<-time.After(2 * time.Second)

			// Assert that the Client sent the expected requests and the Endpoint returned the expected responses
			for clientReq, endpointResp := range test.jsonrpcRequests {
				_, exists := capturedMessages.clientRequests[clientReq]
				c.True(exists, "Client did not send expected request: %s", clientReq)
				_, exists = capturedMessages.endpointResponses[endpointResp]
				c.True(exists, "Endpoint did not send expected response: %s", endpointResp)
			}

			// Assert that the expected subscription push events were sent by the Endpoint and received by the Client
			for event := range test.subscriptionEvents {
				_, exists := capturedMessages.subscriptionEvents[event]
				c.True(exists, "Endpoint did not send expected subscription event: %s", event)
			}
		})
	}
}

// testClientConn creates an HTTP test server with a websocket handler to represent a Client connection
// It returns the websocket connection to pass to NewBridge, which handles dialing the Client URL to create the Client connection.
func testClientConn(t *testing.T, jsonrpcRequests map[clientReq]endpointResp) *websocket.Conn {
	// clientSocketHandler is the handler for the Client connection, which:
	// - upgrades the HTTP connection to a websocket connection
	// - starts a goroutine to read responses and subscription push events from the Endpoint connection
	// - captures the messages in `capturedMessages`
	// - sends test JSON RPC requests to the Endpoint
	clientSocketHandler := func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error("Error during connection upgradation:", err)
			return
		}

		// Start a goroutine to read messages from the Endpoint
		// websocket connection and record them in `capturedMessages`
		go func() {
			for {
				_, endpointMessage, err := conn.ReadMessage()
				if err != nil {
					fmt.Println("error reading response", err)
					t.Error("Error reading response:", err)
					return
				}

				capturedMessages.Lock()
				capturedMessages.endpointResponses[endpointResp(string(endpointMessage))] = struct{}{}
				capturedMessages.subscriptionEvents[subscriptionEvent(string(endpointMessage))] = struct{}{}
				capturedMessages.Unlock()
			}
		}()

		// Send the test JSON RPC requests to the Endpoint
		for req := range jsonrpcRequests {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(req)); err != nil {
				t.Error("Error sending response:", err)
			}
		}
	}

	s := httptest.NewServer(http.HandlerFunc(clientSocketHandler))

	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Error(err)
	}

	return conn
}

// testEndpointConn creates an HTTP test server with a websocket handler to represent an Endpoint connection
// It returns the URL of the test server to pass to NewBridge.
func testEndpointConnURL(t *testing.T, jsonrpcRequests map[clientReq]endpointResp, subscriptionEvents map[subscriptionEvent]struct{}) string {
	// endpointSocketHandler is the handler for the Endpoint connection, which:
	// - upgrades the HTTP connection to a websocket connection
	// - starts a goroutine to read requests from the Client connection
	// - captures the requests in `capturedMessages`
	// - sends subscription push events to the Client
	endpointSocketHandler := func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error("Error during connection upgradation:", err)
			return
		}

		// Start a goroutine to read requests from the Client
		// websocket connection and record them in `capturedMessages`
		go func() {
			for {
				_, signedRelayRequestBz, err := conn.ReadMessage()
				if err != nil {
					t.Error("Error reading message:", err)
					return
				}

				var signedRelayRequest servicetypes.RelayRequest
				if err := signedRelayRequest.Unmarshal(signedRelayRequestBz); err != nil {
					t.Error("Error unmarshalling message:", err)
					return
				}

				message := signedRelayRequest.Payload

				if response, ok := jsonrpcRequests[clientReq(message)]; ok {
					relayResponseBz, err := getRelayResponseBz(response)
					if err != nil {
						t.Error("Error getting relay response from endpoint response:", err)
						return
					}

					if err := conn.WriteMessage(websocket.TextMessage, relayResponseBz); err != nil {
						t.Error("Error sending response:", err)
					}
				}

				capturedMessages.Lock()
				capturedMessages.clientRequests[clientReq(message)] = struct{}{}
				capturedMessages.Unlock()
			}
		}()

		// Send the test subscription push events to the Client
		for event := range subscriptionEvents {
			relayResponseBz, err := getRelayResponseBz(endpointResp(event))
			if err != nil {
				t.Error("Error getting relay response from subscription event:", err)
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, relayResponseBz); err != nil {
				t.Error("Error sending response:", err)
			}
		}
	}

	s := httptest.NewServer(http.HandlerFunc(endpointSocketHandler))

	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")

	return wsURL
}

func getRelayResponseBz(endpointResp endpointResp) ([]byte, error) {
	relayResponse := &servicetypes.RelayResponse{
		Meta: servicetypes.RelayResponseMetadata{SessionHeader: &sessiontypes.SessionHeader{
			ServiceId:          "service_id",
			ApplicationAddress: "application_address",
		}},
		Payload: []byte(endpointResp),
	}

	relayResponseBz, err := relayResponse.Marshal()
	if err != nil {
		return nil, err
	}

	return relayResponseBz, nil
}

/* Mock structs to satisfy the Bridge interfaces */
// Note these mocks do not test the unerlying logic provided by the `protocol/shannon` structs; they will always return the "happy path"

type selectedEndpoint struct {
	url      string // Assigned to the value of the `url` returned by `testEndpointConnURL`
	session  *sessiontypes.Session
	supplier string
}

func (e *selectedEndpoint) PublicURL() string {
	return e.url
}

func (e *selectedEndpoint) Supplier() string {
	return e.supplier
}

func (e *selectedEndpoint) Session() *sessiontypes.Session {
	return e.session
}

type relayRequestSigner struct{}

func (r *relayRequestSigner) SignRelayRequest(ctx context.Context, req *servicetypes.RelayRequest, app apptypes.Application) (*servicetypes.RelayRequest, error) {
	return req, nil
}

type fullNode struct{}

func (f *fullNode) ValidateRelayResponse(supplierAddr sdk.SupplierAddress, responseBz []byte) (*servicetypes.RelayResponse, error) {
	relayResponse := &servicetypes.RelayResponse{}
	if err := relayResponse.Unmarshal(responseBz); err != nil {
		return nil, err
	}
	return relayResponse, nil
}
