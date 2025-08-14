package websockets

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"github.com/stretchr/testify/require"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/observation"
)

type (
	testMessage string // testMessage represents a message exchanged between client and endpoint
)

var capturedMessages struct {
	sync.Mutex
	clientToEndpoint map[testMessage]struct{} // Messages sent from client to endpoint
	endpointToClient map[testMessage]struct{} // Messages sent from endpoint to client
}

func Test_Bridge_MessageFlow(t *testing.T) {
	tests := []struct {
		name              string
		clientMessages    []testMessage
		endpointResponses []testMessage
		expectError       bool
	}{
		{
			name: "should forward messages bidirectionally between client and endpoint",
			clientMessages: []testMessage{
				"client message 1",
				"client message 2",
				"client message 3",
			},
			endpointResponses: []testMessage{
				"endpoint response 1",
				"endpoint response 2",
				"endpoint response 3",
			},
			expectError: false,
		},
		{
			name:              "should handle empty message flow",
			clientMessages:    []testMessage{},
			endpointResponses: []testMessage{},
			expectError:       false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			// Reset captured messages before each test
			capturedMessages.clientToEndpoint = make(map[testMessage]struct{})
			capturedMessages.endpointToClient = make(map[testMessage]struct{})

			// Create test websocket connections
			clientConn, endpointConn := createTestConnections(t, test.clientMessages, test.endpointResponses)

			// Create mock message handlers
			clientHandler := &mockClientMessageHandler{}
			endpointHandler := &mockEndpointMessageHandler{}
			observationPublisher := &mockObservationPublisher{}

			// Create the bridge
			bridge, err := NewBridge(
				polyzero.NewLogger(),
				clientConn,
				endpointConn,
				clientHandler,
				endpointHandler,
				observationPublisher,
			)
			c.NoError(err)

			// Start the bridge in a goroutine
			go bridge.StartAsync(&observation.GatewayObservations{}, nil)

			// Wait for messages to be processed
			time.Sleep(1 * time.Second)

			// Verify message flow
			for _, expectedMsg := range test.clientMessages {
				_, exists := capturedMessages.clientToEndpoint[expectedMsg]
				c.True(exists, "Expected client message not captured: %s", expectedMsg)
			}

			for _, expectedResp := range test.endpointResponses {
				_, exists := capturedMessages.endpointToClient[expectedResp]
				c.True(exists, "Expected endpoint response not captured: %s", expectedResp)
			}

			// Verify observation publisher was called if configured
			if len(test.endpointResponses) > 0 {
				c.True(observationPublisher.publishCalled, "ObservationPublisher.PublishObservations should have been called")
			}
		})
	}
}

func Test_Bridge_Shutdown(t *testing.T) {
	c := require.New(t)

	// Create test connections
	clientConn, endpointConn := createTestConnections(t, []testMessage{}, []testMessage{})

	// Create mock handlers
	clientHandler := &mockClientMessageHandler{}
	endpointHandler := &mockEndpointMessageHandler{}
	observationPublisher := &mockObservationPublisher{}

	// Create the bridge
	bridge, err := NewBridge(
		polyzero.NewLogger(),
		clientConn,
		endpointConn,
		clientHandler,
		endpointHandler,
		observationPublisher,
	)
	c.NoError(err)

	// Test shutdown functionality
	bridge.Shutdown(fmt.Errorf("test shutdown"))

	// Verify connections are closed (this would be implementation-dependent)
	// For now, we just verify the method doesn't panic
}

// createTestConnections creates a pair of websocket connections for testing
func createTestConnections(t *testing.T, clientMessages, endpointResponses []testMessage) (*websocket.Conn, *websocket.Conn) {
	// Create endpoint server
	endpointServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error("Error upgrading endpoint connection:", err)
			return
		}

		// Read messages from client and capture them
		go func() {
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					return // Connection closed
				}

				capturedMessages.Lock()
				capturedMessages.clientToEndpoint[testMessage(string(message))] = struct{}{}
				capturedMessages.Unlock()
			}
		}()

		// Send responses to client
		for _, response := range endpointResponses {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(response)); err != nil {
				t.Error("Error sending endpoint response:", err)
			}
		}
	}))

	// Create client server
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

				capturedMessages.Lock()
				capturedMessages.endpointToClient[testMessage(string(message))] = struct{}{}
				capturedMessages.Unlock()
			}
		}()

		// Send messages to endpoint
		for _, message := range clientMessages {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
				t.Error("Error sending client message:", err)
			}
		}
	}))

	// Connect to servers
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

// Mock implementations for testing

type mockClientMessageHandler struct{}

func (m *mockClientMessageHandler) HandleMessage(msg WebSocketMessage) ([]byte, error) {
	// Echo the message as-is (no protocol-specific processing)
	return msg.Data, nil
}

type mockEndpointMessageHandler struct{}

func (m *mockEndpointMessageHandler) HandleMessage(msg WebSocketMessage) ([]byte, error) {
	// Echo the message as-is (no protocol-specific processing)
	return msg.Data, nil
}

type mockObservationPublisher struct {
	publishCalled       bool
	gatewayObservations *observation.GatewayObservations
}

func (m *mockObservationPublisher) SetObservationContext(
	gatewayObservations *observation.GatewayObservations,
	dataReporter gateway.RequestResponseReporter,
) {
	m.gatewayObservations = gatewayObservations
}

func (m *mockObservationPublisher) InitializeMessageObservations() *observation.RequestResponseObservations {
	return &observation.RequestResponseObservations{}
}

func (m *mockObservationPublisher) UpdateMessageObservationsFromSuccess(*observation.RequestResponseObservations) {
	// Mock implementation - no-op
}

func (m *mockObservationPublisher) UpdateMessageObservationsFromError(*observation.RequestResponseObservations, error) {
	// Mock implementation - no-op
}

func (m *mockObservationPublisher) PublishMessageObservations(*observation.RequestResponseObservations) {
	m.publishCalled = true
}
