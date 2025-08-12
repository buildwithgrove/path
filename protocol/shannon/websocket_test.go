package shannon

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
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
	sdk "github.com/pokt-network/shannon-sdk"
	"github.com/stretchr/testify/require"

	"github.com/buildwithgrove/path/observation"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/websockets"
)

type (
	clientReq         string // clientReq is a JSON RPC request sent from client to endpoint
	endpointResp      string // endpointResp is a JSON RPC response from endpoint to client
	subscriptionEvent string // subscriptionEvent is a subscription push event from endpoint to client
)

var capturedShannonMessages struct {
	sync.Mutex
	clientRequests     map[clientReq]struct{}
	endpointResponses  map[endpointResp]struct{}
	subscriptionEvents map[subscriptionEvent]struct{}
	relayRequests      []servicetypes.RelayRequest
	relayResponses     []servicetypes.RelayResponse
}

func Test_ShannonWebsocketBridge_ProtocolEndpoints(t *testing.T) {
	tests := []struct {
		name               string
		selectedEndpoint   *mockEndpoint
		jsonrpcRequests    map[clientReq]endpointResp
		subscriptionEvents map[subscriptionEvent]struct{}
		expectSigning      bool
		expectValidation   bool
	}{
		{
			name: "should sign client messages and validate endpoint responses for protocol endpoints",
			selectedEndpoint: &mockEndpoint{
				addr:     "protocol-endpoint-1",
				url:      "",
				supplier: "supplier1",
				session: &sessiontypes.Session{
					SessionId: "session1",
					Header: &sessiontypes.SessionHeader{
						ServiceId:          "ethereum-mainnet",
						ApplicationAddress: "app_address_1",
					},
					Application: &apptypes.Application{
						Address: "app_address_1",
					},
				},
				isFallback: false,
			},
			jsonrpcRequests: map[clientReq]endpointResp{
				`{"jsonrpc":"2.0","id":1,"method":"eth_gasPrice"}`:    `{"jsonrpc":"2.0","id":1,"result":"0x337d04a3b"}`,
				`{"jsonrpc":"2.0","id":2,"method":"eth_blockNumber"}`: `{"jsonrpc":"2.0","id":2,"result":"0x12c1b21"}`,
			},
			expectSigning:    true,
			expectValidation: true,
		},
		{
			name: "should handle subscription events for protocol endpoints",
			selectedEndpoint: &mockEndpoint{
				addr:     "protocol-endpoint-2",
				url:      "",
				supplier: "supplier2",
				session: &sessiontypes.Session{
					SessionId: "session2",
					Header: &sessiontypes.SessionHeader{
						ServiceId:          "ethereum-mainnet",
						ApplicationAddress: "app_address_2",
					},
					Application: &apptypes.Application{
						Address: "app_address_2",
					},
				},
				isFallback: false,
			},
			jsonrpcRequests: map[clientReq]endpointResp{
				`{"jsonrpc":"2.0","id":1,"method":"eth_subscribe","params":["newPendingTransactions"]}`: `{"jsonrpc":"2.0","id":1,"result":"0x456"}`,
			},
			subscriptionEvents: map[subscriptionEvent]struct{}{
				`{"jsonrpc":"2.0","method":"eth_subscription","params":{"result":"0x123","subscription":"0x456"}}`: {},
				`{"jsonrpc":"2.0","method":"eth_subscription","params":{"result":"0x789","subscription":"0x456"}}`: {},
			},
			expectSigning:    true,
			expectValidation: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			// Reset captured messages
			capturedShannonMessages.clientRequests = make(map[clientReq]struct{})
			capturedShannonMessages.endpointResponses = make(map[endpointResp]struct{})
			capturedShannonMessages.subscriptionEvents = make(map[subscriptionEvent]struct{})
			capturedShannonMessages.relayRequests = []servicetypes.RelayRequest{}
			capturedShannonMessages.relayResponses = []servicetypes.RelayResponse{}

			// Create mock request context
			rc := &requestContext{
				logger:             polyzero.NewLogger(),
				selectedEndpoint:   test.selectedEndpoint,
				relayRequestSigner: &mockRelayRequestSigner{},
				fullNode:           &mockFullNode{},
				serviceID:          "ethereum-mainnet",
			}

			// Create test WebSocket connections directly
			clientConn, endpointConn := createTestWebSocketConnections(t, test.jsonrpcRequests, test.subscriptionEvents, false)
			test.selectedEndpoint.url = "ws://test-endpoint"
			test.selectedEndpoint.websocketURL = "ws://test-endpoint"

			// Create Shannon-specific message handlers
			clientHandler := &shannonClientMessageHandler{
				logger:             polyzero.NewLogger(),
				selectedEndpoint:   test.selectedEndpoint,
				relayRequestSigner: rc.relayRequestSigner,
				serviceID:          rc.serviceID,
			}
			endpointHandler := &shannonEndpointMessageHandler{
				logger:           polyzero.NewLogger(),
				selectedEndpoint: test.selectedEndpoint,
				fullNode:         rc.fullNode,
				serviceID:        rc.serviceID,
			}
			observationPublisher := &shannonObservationPublisher{
				serviceID:            rc.serviceID,
				protocolObservations: buildWebsocketBridgeEndpointObservation(rc.logger, rc.serviceID, test.selectedEndpoint),
			}

			// Create the bridge directly using the generic websocket bridge
			bridge, err := websockets.NewBridge(
				polyzero.NewLogger(),
				clientConn,
				endpointConn,
				clientHandler,
				endpointHandler,
				observationPublisher,
			)
			c.NoError(err)

			// Start the bridge
			mockObservations := &observation.GatewayObservations{}
			go bridge.StartAsync(mockObservations, nil)

			// Wait for processing
			time.Sleep(2 * time.Second)

			// Verify message flow
			for clientReq, endpointResp := range test.jsonrpcRequests {
				_, exists := capturedShannonMessages.clientRequests[clientReq]
				c.True(exists, "Client request not captured: %s", clientReq)
				_, exists = capturedShannonMessages.endpointResponses[endpointResp]
				c.True(exists, "Endpoint response not captured: %s", endpointResp)
			}

			// Verify protocol-specific behavior
			if test.expectSigning {
				c.True(len(capturedShannonMessages.relayRequests) > 0, "Expected relay requests to be signed")
			}

			if test.expectValidation {
				c.True(len(capturedShannonMessages.relayResponses) > 0, "Expected relay responses to be validated")
			}
		})
	}
}

func Test_ShannonWebsocketBridge_FallbackEndpoints(t *testing.T) {
	c := require.New(t)

	// Reset captured messages
	capturedShannonMessages.clientRequests = make(map[clientReq]struct{})
	capturedShannonMessages.endpointResponses = make(map[endpointResp]struct{})
	capturedShannonMessages.relayRequests = []servicetypes.RelayRequest{}
	capturedShannonMessages.relayResponses = []servicetypes.RelayResponse{}

	fallbackEndpoint := &mockEndpoint{
		addr:     "fallback-endpoint-1",
		url:      "",
		supplier: "fallback",
		session: &sessiontypes.Session{
			SessionId: "fallback-session",
			Header: &sessiontypes.SessionHeader{
				ServiceId:          "ethereum-mainnet",
				ApplicationAddress: "fallback_app_address",
			},
			Application: &apptypes.Application{
				Address: "fallback_app_address",
			},
		},
		isFallback: true,
	}

	jsonrpcRequests := map[clientReq]endpointResp{
		`{"jsonrpc":"2.0","id":1,"method":"eth_gasPrice"}`:    `{"jsonrpc":"2.0","id":1,"result":"0x337d04a3b"}`,
		`{"jsonrpc":"2.0","id":2,"method":"eth_blockNumber"}`: `{"jsonrpc":"2.0","id":2,"result":"0x12c1b21"}`,
	}

	// Create mock request context
	rc := &requestContext{
		logger:             polyzero.NewLogger(),
		selectedEndpoint:   fallbackEndpoint,
		relayRequestSigner: &mockRelayRequestSigner{},
		fullNode:           &mockFullNode{},
		serviceID:          "ethereum-mainnet",
	}

	// Create test WebSocket connections directly for fallback endpoint
	clientConn, endpointConn := createTestWebSocketConnections(t, jsonrpcRequests, nil, true)
	fallbackEndpoint.url = "ws://fallback-endpoint"
	fallbackEndpoint.websocketURL = "ws://fallback-endpoint"

	// Create Shannon-specific message handlers for fallback endpoint
	clientHandler := &shannonClientMessageHandler{
		logger:             polyzero.NewLogger(),
		selectedEndpoint:   fallbackEndpoint,
		relayRequestSigner: rc.relayRequestSigner,
		serviceID:          rc.serviceID,
	}
	endpointHandler := &shannonEndpointMessageHandler{
		logger:           polyzero.NewLogger(),
		selectedEndpoint: fallbackEndpoint,
		fullNode:         rc.fullNode,
		serviceID:        rc.serviceID,
	}
	observationPublisher := &shannonObservationPublisher{
		serviceID:            rc.serviceID,
		protocolObservations: buildWebsocketBridgeEndpointObservation(rc.logger, rc.serviceID, fallbackEndpoint),
	}

	// Create the bridge directly using the generic websocket bridge
	bridge, err := websockets.NewBridge(
		polyzero.NewLogger(),
		clientConn,
		endpointConn,
		clientHandler,
		endpointHandler,
		observationPublisher,
	)
	c.NoError(err)

	// Start the bridge
	go bridge.StartAsync(&observation.GatewayObservations{}, nil)

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Verify message flow
	for clientReq, endpointResp := range jsonrpcRequests {
		_, exists := capturedShannonMessages.clientRequests[clientReq]
		c.True(exists, "Client request not captured: %s", clientReq)
		_, exists = capturedShannonMessages.endpointResponses[endpointResp]
		c.True(exists, "Endpoint response not captured: %s", endpointResp)
	}

	// Verify fallback behavior (no signing/validation)
	c.Equal(0, len(capturedShannonMessages.relayRequests), "Fallback endpoints should not sign requests")
	c.Equal(0, len(capturedShannonMessages.relayResponses), "Fallback endpoints should not validate responses")
}

func Test_ShannonMessageHandlers(t *testing.T) {
	t.Run("ClientMessageHandler", func(t *testing.T) {
		c := require.New(t)

		endpoint := &mockEndpoint{
			addr:     "test-endpoint",
			supplier: "supplier1",
			session: &sessiontypes.Session{
				Header: &sessiontypes.SessionHeader{
					ServiceId:          "ethereum-mainnet",
					ApplicationAddress: "app_address_1",
				},
				Application: &apptypes.Application{
					Address: "app_address_1",
				},
			},
			isFallback: false,
		}

		handler := &shannonClientMessageHandler{
			logger:             polyzero.NewLogger(),
			selectedEndpoint:   endpoint,
			relayRequestSigner: &mockRelayRequestSigner{},
			serviceID:          "ethereum-mainnet",
		}

		msg := websockets.Message{
			Data:        []byte(`{"jsonrpc":"2.0","id":1,"method":"eth_gasPrice"}`),
			MessageType: websocket.TextMessage,
		}

		result, err := handler.HandleMessage(msg)
		c.NoError(err)
		c.NotNil(result)

		// Verify the message was signed (should be different from original)
		c.NotEqual(msg.Data, result)
	})

	t.Run("EndpointMessageHandler", func(t *testing.T) {
		c := require.New(t)

		endpoint := &mockEndpoint{
			addr:       "test-endpoint",
			supplier:   "supplier1",
			isFallback: false,
		}

		handler := &shannonEndpointMessageHandler{
			logger:           polyzero.NewLogger(),
			selectedEndpoint: endpoint,
			fullNode:         &mockFullNode{},
			serviceID:        "ethereum-mainnet",
		}

		// Create a mock relay response
		relayResponse := &servicetypes.RelayResponse{
			Payload: []byte(`{"jsonrpc":"2.0","id":1,"result":"0x337d04a3b"}`),
		}
		responseBytes, _ := relayResponse.Marshal()

		msg := websockets.Message{
			Data:        responseBytes,
			MessageType: websocket.TextMessage,
		}

		result, err := handler.HandleMessage(msg)
		c.NoError(err)
		c.Equal(relayResponse.Payload, result)
	})

	t.Run("ObservationPublisher", func(t *testing.T) {
		c := require.New(t)

		publisher := &shannonObservationPublisher{
			serviceID:            "ethereum-mainnet",
			protocolObservations: &protocolobservations.Observations{},
		}

		// Set observation context
		gatewayObs := &observation.GatewayObservations{}
		mockReporter := &mockDataReporter{}
		publisher.SetObservationContext(gatewayObs, mockReporter)

		// Publish observations
		publisher.PublishObservations()

		c.True(mockReporter.publishCalled, "Publish should have been called")
		c.NotNil(mockReporter.lastObservation, "Observation should have been published")
	})
}

func Test_ConnectWebsocketEndpoint(t *testing.T) {
	tests := []struct {
		name          string
		endpoint      *mockEndpoint
		expectedError bool
		testHeaders   bool
	}{
		{
			name: "should connect successfully to protocol endpoint with headers",
			endpoint: &mockEndpoint{
				addr:       "protocol-endpoint",
				supplier:   "supplier1",
				isFallback: false,
				session: &sessiontypes.Session{
					Header: &sessiontypes.SessionHeader{
						ServiceId:          "ethereum-mainnet",
						ApplicationAddress: "app_address_1",
					},
				},
			},
			expectedError: false,
			testHeaders:   true,
		},
		{
			name: "should connect successfully to fallback endpoint without headers",
			endpoint: &mockEndpoint{
				addr:       "fallback-endpoint",
				supplier:   "fallback",
				isFallback: true,
			},
			expectedError: false,
			testHeaders:   false,
		},
		{
			name: "should fail with invalid URL",
			endpoint: &mockEndpoint{
				addr:          "invalid-endpoint",
				websocketURL:  "invalid-url",
				shouldFailURL: true,
			},
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			if !test.endpoint.shouldFailURL {
				// Create a test server
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					upgrader := websocket.Upgrader{}
					_, err := upgrader.Upgrade(w, r, nil)
					if err != nil {
						t.Error("Error during connection upgrade:", err)
						return
					}

					// Verify headers for protocol endpoints
					if test.testHeaders && !test.endpoint.isFallback {
						c.NotEmpty(r.Header.Get("Target-Service-Id"))
						c.NotEmpty(r.Header.Get("App-Address"))
						c.NotEmpty(r.Header.Get("Rpc-Type"))
					}
				}))
				defer server.Close()

				wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
				test.endpoint.websocketURL = wsURL
			}

			conn, err := connectWebsocketEndpoint(polyzero.NewLogger(), test.endpoint)

			if test.expectedError {
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

// Helper functions and mocks

func createTestWebSocketConnections(t *testing.T, jsonrpcRequests map[clientReq]endpointResp, subscriptionEvents map[subscriptionEvent]struct{}, isFallback bool) (*websocket.Conn, *websocket.Conn) {
	// Create client connection
	clientServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error("Error upgrading client connection:", err)
			return
		}

		// Read responses from endpoint and capture them
		go func() {
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					return // Connection closed
				}

				capturedShannonMessages.Lock()
				capturedShannonMessages.endpointResponses[endpointResp(string(message))] = struct{}{}
				capturedShannonMessages.subscriptionEvents[subscriptionEvent(string(message))] = struct{}{}
				capturedShannonMessages.Unlock()
			}
		}()

		// Send test messages to endpoint
		for req := range jsonrpcRequests {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(req)); err != nil {
				t.Error("Error sending client message:", err)
			}
		}
	}))

	// Create endpoint connection
	endpointServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error("Error upgrading endpoint connection:", err)
			return
		}

		// Handle incoming messages from client
		go func() {
			for {
				_, requestBz, err := conn.ReadMessage()
				if err != nil {
					return
				}

				var message []byte
				if isFallback {
					message = requestBz
				} else {
					var relayRequest servicetypes.RelayRequest
					if err := relayRequest.Unmarshal(requestBz); err != nil {
						t.Error("Error unmarshalling relay request:", err)
						return
					}
					capturedShannonMessages.Lock()
					capturedShannonMessages.relayRequests = append(capturedShannonMessages.relayRequests, relayRequest)
					capturedShannonMessages.Unlock()
					message = relayRequest.Payload
				}

				capturedShannonMessages.Lock()
				capturedShannonMessages.clientRequests[clientReq(message)] = struct{}{}
				capturedShannonMessages.Unlock()

				// Send response if exists
				if response, ok := jsonrpcRequests[clientReq(message)]; ok {
					var responseBz []byte
					if isFallback {
						responseBz = []byte(response)
					} else {
						relayResponse := &servicetypes.RelayResponse{
							Payload: []byte(response),
						}
						responseBz, _ = relayResponse.Marshal()
						capturedShannonMessages.Lock()
						capturedShannonMessages.relayResponses = append(capturedShannonMessages.relayResponses, *relayResponse)
						capturedShannonMessages.Unlock()
					}

					if err := conn.WriteMessage(websocket.TextMessage, responseBz); err != nil {
						t.Error("Error sending response:", err)
					}
				}
			}
		}()

		// Send subscription events
		for event := range subscriptionEvents {
			var eventBz []byte
			if isFallback {
				eventBz = []byte(event)
			} else {
				relayResponse := &servicetypes.RelayResponse{
					Payload: []byte(event),
				}
				eventBz, _ = relayResponse.Marshal()
			}

			if err := conn.WriteMessage(websocket.TextMessage, eventBz); err != nil {
				t.Error("Error sending subscription event:", err)
			}

			capturedShannonMessages.Lock()
			capturedShannonMessages.subscriptionEvents[event] = struct{}{}
			capturedShannonMessages.Unlock()
		}
	}))

	// Connect to both servers
	clientWSURL := "ws" + strings.TrimPrefix(clientServer.URL, "http")
	endpointWSURL := "ws" + strings.TrimPrefix(endpointServer.URL, "http")

	clientConn, _, err := websocket.DefaultDialer.Dial(clientWSURL, nil)
	if err != nil {
		t.Fatal("Error connecting to client server:", err)
	}

	endpointConn, _, err := websocket.DefaultDialer.Dial(endpointWSURL, nil)
	if err != nil {
		t.Fatal("Error connecting to endpoint server:", err)
	}

	return clientConn, endpointConn
}

// Mock implementations

type mockEndpoint struct {
	addr          string
	url           string
	websocketURL  string
	supplier      string
	session       *sessiontypes.Session
	isFallback    bool
	shouldFailURL bool
}

func (e *mockEndpoint) Addr() protocol.EndpointAddr {
	return protocol.EndpointAddr(e.addr)
}

func (e *mockEndpoint) PublicURL() string {
	return e.url
}

func (e *mockEndpoint) WebsocketURL() (string, error) {
	if e.shouldFailURL {
		return "", fmt.Errorf("invalid websocket URL")
	}
	return e.websocketURL, nil
}

func (e *mockEndpoint) Supplier() string {
	return e.supplier
}

func (e *mockEndpoint) Session() *sessiontypes.Session {
	return e.session
}

func (e *mockEndpoint) IsFallback() bool {
	return e.isFallback
}

func (e *mockEndpoint) FallbackURL(rpcType sharedtypes.RPCType) string {
	return e.websocketURL
}

type mockRelayRequestSigner struct{}

func (m *mockRelayRequestSigner) SignRelayRequest(req *servicetypes.RelayRequest, app apptypes.Application) (*servicetypes.RelayRequest, error) {
	// Return the request as-is (mock signing)
	return req, nil
}

type mockFullNode struct{}

func (m *mockFullNode) GetApp(ctx context.Context, appAddr string) (*apptypes.Application, error) {
	return &apptypes.Application{Address: appAddr}, nil
}

func (m *mockFullNode) GetSession(ctx context.Context, serviceID protocol.ServiceID, appAddr string) (sessiontypes.Session, error) {
	return sessiontypes.Session{}, nil
}

func (m *mockFullNode) GetSessionWithExtendedValidity(ctx context.Context, serviceID protocol.ServiceID, appAddr string) (sessiontypes.Session, error) {
	return sessiontypes.Session{}, nil
}

func (m *mockFullNode) GetSharedParams(ctx context.Context) (*sharedtypes.Params, error) {
	return &sharedtypes.Params{}, nil
}

func (m *mockFullNode) GetCurrentBlockHeight(ctx context.Context) (int64, error) {
	return 100, nil
}

func (m *mockFullNode) ValidateRelayResponse(supplierAddr sdk.SupplierAddress, responseBz []byte) (*servicetypes.RelayResponse, error) {
	var relayResponse servicetypes.RelayResponse
	if err := relayResponse.Unmarshal(responseBz); err != nil {
		return nil, err
	}
	return &relayResponse, nil
}

func (m *mockFullNode) IsHealthy() bool {
	return true
}

func (m *mockFullNode) GetAccountClient() *sdk.AccountClient {
	return nil // Mock implementation
}

type mockDataReporter struct {
	publishCalled   bool
	lastObservation *observation.RequestResponseObservations
}

func (m *mockDataReporter) Publish(obs *observation.RequestResponseObservations) {
	m.publishCalled = true
	m.lastObservation = obs
}
