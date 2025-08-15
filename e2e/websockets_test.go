//go:build e2e

// Package e2e provides WebSocket testing functionality for PATH E2E tests.
// This file contains the WebSocket test client and related functions for testing
// JSON-RPC over WebSocket connections, with full integration into the existing
// E2E test framework and validation logic.
//
// SEPARATION OF CONCERNS:
// - websockets_test.go: Handles all WebSocket-specific functionality
// - vegeta_test.go: Handles only HTTP testing using Vegeta
// - main_test.go: Orchestrates both HTTP and WebSocket tests via runAllServiceTests()
//
// This separation ensures clean boundaries between different transport protocols
// while maintaining shared validation logic in assertions_test.go.
package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// websocketTestClient handles WebSocket connections and JSON-RPC requests.
//
// This client provides a complete WebSocket testing solution that:
// - Converts HTTP/HTTPS gateway URLs to WebSocket (WS/WSS) equivalents
// - Establishes persistent WebSocket connections with proper headers
// - Sends JSON-RPC requests over WebSocket and validates responses
// - Integrates with the existing E2E test validation framework
// - Supports all EVM JSON-RPC methods defined in service_evm_test.go
//
// Usage:
//
//	client := newWebSocketTestClient(gatewayURL, serviceID)
//	err := client.connect(ctx)
//	results, err := client.sendEVMRequestsFromServiceParams(ctx, serviceParams)
//	client.close()
type websocketTestClient struct {
	conn       *websocket.Conn
	serviceID  string
	gatewayURL string
	mutex      sync.RWMutex
	closed     bool
}

// newWebSocketTestClient creates a new WebSocket test client
func newWebSocketTestClient(gatewayURL, serviceID string) *websocketTestClient {
	return &websocketTestClient{
		serviceID:  serviceID,
		gatewayURL: gatewayURL,
	}
}

// connect establishes a WebSocket connection to the PATH gateway
func (c *websocketTestClient) connect(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Convert HTTP/HTTPS URL to WS/WSS URL
	wsURL, err := convertToWebSocketURL(c.gatewayURL)
	if err != nil {
		return fmt.Errorf("failed to convert gateway URL to WebSocket URL: %w", err)
	}

	// Set up headers including the Target-Service-Id
	headers := http.Header{
		"Target-Service-Id": []string{c.serviceID},
	}

	// Create WebSocket dialer with timeout
	dialer := websocket.Dialer{
		HandshakeTimeout: 30 * time.Second,
	}

	// Establish WebSocket connection
	conn, resp, err := dialer.DialContext(ctx, wsURL, headers)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("WebSocket dial failed with status %d: %w", resp.StatusCode, err)
		}
		return fmt.Errorf("WebSocket dial failed: %w", err)
	}

	c.conn = conn
	c.closed = false

	return nil
}

// sendJSONRPCRequest sends a JSON-RPC request over the WebSocket connection
func (c *websocketTestClient) sendJSONRPCRequest(ctx context.Context, req jsonrpc.Request) (*jsonrpc.Response, error) {
	c.mutex.RLock()
	if c.closed || c.conn == nil {
		c.mutex.RUnlock()
		return nil, fmt.Errorf("WebSocket connection is not open")
	}
	conn := c.conn
	c.mutex.RUnlock()

	// Marshal the JSON-RPC request
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON-RPC request: %w", err)
	}

	// Set write deadline
	if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return nil, fmt.Errorf("failed to set write deadline: %w", err)
	}

	// Send the request
	if err := conn.WriteMessage(websocket.TextMessage, reqBytes); err != nil {
		return nil, fmt.Errorf("failed to send WebSocket message: %w", err)
	}

	// Set read deadline
	if err := conn.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Read the response
	_, respBytes, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to read WebSocket response: %w", err)
	}

	// Parse the JSON-RPC response
	var resp jsonrpc.Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON-RPC response: %w", err)
	}

	return &resp, nil
}

// sendEVMRequest sends an EVM JSON-RPC request with the specified method and parameters
func (c *websocketTestClient) sendEVMRequest(ctx context.Context, method string, params jsonrpc.Params) (*jsonrpc.Response, error) {
	req := jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(1),
		Method:  jsonrpc.Method(method),
		Params:  params,
	}

	return c.sendJSONRPCRequest(ctx, req)
}

// sendEVMRequestsFromServiceParams sends multiple EVM requests based on service parameters
func (c *websocketTestClient) sendEVMRequestsFromServiceParams(ctx context.Context, sp ServiceParams) (map[string]*jsonrpc.Response, error) {
	methods := getEVMTestMethods()
	results := make(map[string]*jsonrpc.Response)

	for _, method := range methods {
		params := createEVMJsonRPCParams(jsonrpc.Method(method), sp)

		resp, err := c.sendEVMRequest(ctx, method, params)
		if err != nil {
			return nil, fmt.Errorf("failed to send %s request: %w", method, err)
		}

		results[method] = resp
	}

	return results, nil
}

// close closes the WebSocket connection
func (c *websocketTestClient) close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed || c.conn == nil {
		return nil
	}

	c.closed = true

	// Send close message
	err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		// If we can't send close message, just close the connection
		return c.conn.Close()
	}

	// Wait for close confirmation or timeout
	select {
	case <-time.After(time.Second):
		// Timeout waiting for close confirmation
	default:
		// Try to read close message
		c.conn.SetReadDeadline(time.Now().Add(time.Second))
		_, _, _ = c.conn.ReadMessage()
	}

	return c.conn.Close()
}

// isConnected returns true if the WebSocket connection is open
func (c *websocketTestClient) isConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return !c.closed && c.conn != nil
}

// convertToWebSocketURL converts an HTTP/HTTPS URL to WS/WSS URL
func convertToWebSocketURL(httpURL string) (string, error) {
	u, err := url.Parse(httpURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Convert scheme
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	default:
		return "", fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}

	return u.String(), nil
}

// websocketTestResult represents the result of a WebSocket test
type websocketTestResult struct {
	Method   string
	Request  jsonrpc.Request
	Response *jsonrpc.Response
	Error    error
	Duration time.Duration
}

// runWebSocketEVMTest runs a complete EVM test over WebSocket
func runWebSocketEVMTest(ctx context.Context, gatewayURL, serviceID string, serviceParams ServiceParams) ([]*websocketTestResult, error) {
	client := newWebSocketTestClient(gatewayURL, serviceID)

	// Connect to WebSocket
	if err := client.connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	defer client.close()

	methods := getEVMTestMethods()
	results := make([]*websocketTestResult, 0, len(methods))

	for _, method := range methods {
		start := time.Now()

		params := createEVMJsonRPCParams(jsonrpc.Method(method), serviceParams)
		req := jsonrpc.Request{
			JSONRPC: jsonrpc.Version2,
			ID:      jsonrpc.IDFromInt(1),
			Method:  jsonrpc.Method(method),
			Params:  params,
		}

		resp, err := client.sendJSONRPCRequest(ctx, req)
		duration := time.Since(start)

		result := &websocketTestResult{
			Method:   method,
			Request:  req,
			Response: resp,
			Error:    err,
			Duration: duration,
		}

		results = append(results, result)

		// If there was an error, we might want to continue with other methods
		// or break depending on the test strategy
		if err != nil {
			fmt.Printf("Warning: WebSocket request for method %s failed: %v\n", method, err)
		}
	}

	return results, nil
}

// validateWebSocketResults validates WebSocket test results using the existing validation logic
func validateWebSocketResults(results []*websocketTestResult, serviceID string) map[string]*MethodMetrics {
	metrics := make(map[string]*MethodMetrics)

	for _, result := range results {
		// Initialize method metrics if not exists
		if metrics[result.Method] == nil {
			metrics[result.Method] = &MethodMetrics{
				Method:                  result.Method,
				StatusCodes:             make(map[int]int),
				Errors:                  make(map[string]int),
				Results:                 make([]*VegetaResult, 0),
				JSONRPCParseErrors:      make(map[string]int),
				JSONRPCValidationErrors: make(map[string]int),
			}
		}

		methodMetrics := metrics[result.Method]

		// Create a VegetaResult for compatibility with existing metrics collection
		vegetaResult := &VegetaResult{
			Latency: result.Duration,
		}

		if result.Error != nil {
			methodMetrics.Failed++
			methodMetrics.Errors[result.Error.Error()]++
			vegetaResult.Code = 0 // No HTTP status for WebSocket errors
			vegetaResult.Error = result.Error.Error()
			// For WebSocket errors, we don't have a valid response body
			vegetaResult.Body = []byte{}
		} else {
			methodMetrics.Success++
			vegetaResult.Code = 200 // Simulate successful WebSocket connection

			// Validate JSON-RPC response if we have one
			if result.Response != nil {
				respBytes, _ := json.Marshal(result.Response)
				vegetaResult.Body = respBytes

				// Use the existing transport-agnostic JSON-RPC validation function
				_ = validateJSONRPCResponse(respBytes, jsonrpc.IDFromInt(1), methodMetrics)
			} else {
				vegetaResult.Body = []byte{}
			}
		}

		methodMetrics.Results = append(methodMetrics.Results, vegetaResult)
		methodMetrics.StatusCodes[int(vegetaResult.Code)]++
	}

	// Calculate success rates and percentiles for all methods using existing functions
	for _, methodMetrics := range metrics {
		calculateAllSuccessRates(methodMetrics)
		calculatePercentiles(methodMetrics)
	}

	return metrics
}

// validateWebSocketMethod validates a single WebSocket method using existing assertion logic
func validateWebSocketMethod(
	t *testing.T,
	serviceID string,
	method string,
	results []*websocketTestResult,
	config ServiceConfig,
) bool {
	// Filter results for this specific method
	methodResults := make([]*websocketTestResult, 0)
	for _, result := range results {
		if result.Method == method {
			methodResults = append(methodResults, result)
		}
	}

	if len(methodResults) == 0 {
		return true // No results to validate
	}

	// Convert to metrics using existing validation logic
	allMetrics := validateWebSocketResults(methodResults, serviceID)
	methodMetrics, exists := allMetrics[method]
	if !exists {
		return true // No metrics found
	}

	// Use the existing validateMethodResults function from assertions_test.go
	return validateMethodResults(t, protocol.ServiceID(serviceID), methodMetrics, config)
}

// validateAllWebSocketMethods validates all WebSocket methods using existing assertion logic
func validateAllWebSocketMethods(
	t *testing.T,
	serviceID string,
	results []*websocketTestResult,
	config ServiceConfig,
) bool {
	// Group results by method
	methodResults := make(map[string][]*websocketTestResult)
	for _, result := range results {
		methodResults[result.Method] = append(methodResults[result.Method], result)
	}

	allPassed := true
	for method, methodRes := range methodResults {
		if !validateWebSocketMethod(t, serviceID, method, methodRes, config) {
			allPassed = false
		}
	}

	return allPassed
}

// runWebSocketServiceTest runs WebSocket-based E2E tests for a single service.
// This function is called from the main test flow and handles all WebSocket testing logic.
func runWebSocketServiceTest(t *testing.T, ctx context.Context, ts *TestService, results map[string]*MethodMetrics, resultsMutex *sync.Mutex) (serviceTestFailed bool) {
	// Get the gateway URL from the first method target (they all use the same URL)
	var gatewayURL string
	for _, methodConfig := range ts.testMethodsMap {
		gatewayURL = methodConfig.target.URL
		break
	}

	if gatewayURL == "" {
		t.Errorf("❌ Failed to get gateway URL for WebSocket tests")
		return true
	}

	fmt.Printf("\n%s🔌 Starting WebSocket tests for %s%s\n", BOLD_CYAN, ts.ServiceID, RESET)

	// Get the service configuration from any method (they share the same config)
	var serviceConfig ServiceConfig
	for _, methodConfig := range ts.testMethodsMap {
		serviceConfig = methodConfig.serviceConfig
		break
	}

	// Run the WebSocket EVM test
	websocketResults, err := runWebSocketEVMTest(ctx, gatewayURL, string(ts.ServiceID), ts.ServiceParams)
	if err != nil {
		t.Errorf("❌ WebSocket test failed for service %s: %v", ts.ServiceID, err)
		return true
	}

	// Convert WebSocket results to method metrics and validate
	websocketMetrics := validateWebSocketResults(websocketResults, string(ts.ServiceID))

	// Add WebSocket results to the combined results map
	resultsMutex.Lock()
	for method, metrics := range websocketMetrics {
		// Append "(WebSocket)" to method name to distinguish from HTTP results
		websocketMethodKey := method + " (WebSocket)"
		results[websocketMethodKey] = metrics

		// Validate individual WebSocket method results using existing assertion logic
		if !validateMethodResults(t, ts.ServiceID, metrics, serviceConfig) {
			serviceTestFailed = true
		}
	}
	resultsMutex.Unlock()

	if serviceTestFailed {
		fmt.Printf("%s❌ WebSocket tests failed for service %s%s\n", RED, ts.ServiceID, RESET)
	} else {
		fmt.Printf("%s✅ WebSocket tests passed for service %s%s\n", GREEN, ts.ServiceID, RESET)
	}

	return serviceTestFailed
}

// Example usage function showing how to run WebSocket tests with existing validation
func runWebSocketTestExample(
	t *testing.T,
	gatewayURL string,
	serviceID string,
	serviceParams ServiceParams,
	config ServiceConfig,
) bool {
	ctx := context.Background()

	// Run the WebSocket EVM test
	results, err := runWebSocketEVMTest(ctx, gatewayURL, serviceID, serviceParams)
	if err != nil {
		t.Errorf("WebSocket test failed: %v", err)
		return false
	}

	// Validate all methods using the existing assertion infrastructure
	return validateAllWebSocketMethods(t, serviceID, results, config)
}
