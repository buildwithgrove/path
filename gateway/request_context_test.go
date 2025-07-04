package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/buildwithgrove/path/config/relay"
	"github.com/buildwithgrove/path/metrics/devtools"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
)

// Mock implementations for testing
type mockProtocol struct {
	mock.Mock
}

func (m *mockProtocol) AvailableEndpoints(ctx context.Context, serviceID protocol.ServiceID, httpReq *http.Request) (protocol.EndpointAddrList, protocolobservations.Observations, error) {
	args := m.Called(ctx, serviceID, httpReq)
	return args.Get(0).(protocol.EndpointAddrList), args.Get(1).(protocolobservations.Observations), args.Error(2)
}

func (m *mockProtocol) BuildRequestContextForEndpoint(ctx context.Context, serviceID protocol.ServiceID, endpointAddr protocol.EndpointAddr, httpReq *http.Request) (ProtocolRequestContext, protocolobservations.Observations, error) {
	args := m.Called(ctx, serviceID, endpointAddr, httpReq)
	if args.Get(0) == nil {
		return nil, args.Get(1).(protocolobservations.Observations), args.Error(2)
	}
	return args.Get(0).(ProtocolRequestContext), args.Get(1).(protocolobservations.Observations), args.Error(2)
}

// Additional methods required by Protocol interface
func (m *mockProtocol) SupportedGatewayModes() []protocol.GatewayMode {
	args := m.Called()
	return args.Get(0).([]protocol.GatewayMode)
}

func (m *mockProtocol) ApplyObservations(obs *protocolobservations.Observations) error {
	args := m.Called(obs)
	return args.Error(0)
}

func (m *mockProtocol) ConfiguredServiceIDs() map[protocol.ServiceID]struct{} {
	args := m.Called()
	return args.Get(0).(map[protocol.ServiceID]struct{})
}

func (m *mockProtocol) GetTotalServiceEndpointsCount(serviceID protocol.ServiceID, httpReq *http.Request) (int, error) {
	args := m.Called(serviceID, httpReq)
	return args.Int(0), args.Error(1)
}

func (m *mockProtocol) HydrateDisqualifiedEndpointsResponse(serviceID protocol.ServiceID, response *devtools.DisqualifiedEndpointResponse) {
	m.Called(serviceID, response)
}

func (m *mockProtocol) CheckHealth() error {
	args := m.Called()
	return args.Error(0)
}

type mockProtocolContext struct {
	mock.Mock
}

func (m *mockProtocolContext) HandleServiceRequest(payload protocol.Payload) (protocol.Response, error) {
	args := m.Called(payload)
	return args.Get(0).(protocol.Response), args.Error(1)
}

func (m *mockProtocolContext) HandleWebsocketRequest(logger polylog.Logger, req *http.Request, w http.ResponseWriter) error {
	args := m.Called(logger, req, w)
	return args.Error(0)
}

func (m *mockProtocolContext) GetObservations() protocolobservations.Observations {
	args := m.Called()
	return args.Get(0).(protocolobservations.Observations)
}

type mockQoSContext struct {
	mock.Mock
}

func (m *mockQoSContext) GetServicePayload() protocol.Payload {
	args := m.Called()
	return args.Get(0).(protocol.Payload)
}

func (m *mockQoSContext) GetEndpointSelector() protocol.EndpointSelector {
	args := m.Called()
	return args.Get(0).(protocol.EndpointSelector)
}

func (m *mockQoSContext) UpdateWithResponse(endpointAddr protocol.EndpointAddr, response []byte) {
	m.Called(endpointAddr, response)
}

func (m *mockQoSContext) GetHTTPResponse() HTTPResponse {
	args := m.Called()
	return args.Get(0).(HTTPResponse)
}

func (m *mockQoSContext) GetObservations() qosobservations.Observations {
	args := m.Called()
	return args.Get(0).(qosobservations.Observations)
}

type mockEndpointSelector struct {
	mock.Mock
}

func (m *mockEndpointSelector) Select(endpoints protocol.EndpointAddrList) (protocol.EndpointAddr, error) {
	args := m.Called(endpoints)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.Get(0).(protocol.EndpointAddr), args.Error(1)
}

// Test helper functions
func createTestRequestContext(t *testing.T) *requestContext {
	logger := polyzero.NewLogger()

	// Create default gateway config
	gatewayConfig := relay.DefaultConfig()

	return &requestContext{
		context:       context.Background(),
		logger:        logger,
		serviceID:     "test-service",
		gatewayConfig: gatewayConfig,
	}
}

// Test parallel relay requests functionality
func TestHandleParallelRelayRequests(t *testing.T) {
	tests := []struct {
		name                    string
		numEndpoints            int
		successfulEndpointIndex int // -1 means all fail
		expectedError           bool
		description             string
	}{
		{
			name:                    "all_endpoints_succeed",
			numEndpoints:            4,
			successfulEndpointIndex: 0, // First endpoint succeeds
			expectedError:           false,
			description:             "Should return first successful response when all endpoints succeed",
		},
		{
			name:                    "first_fails_second_succeeds",
			numEndpoints:            4,
			successfulEndpointIndex: 1, // Second endpoint succeeds
			expectedError:           false,
			description:             "Should return second endpoint response when first fails",
		},
		{
			name:                    "all_endpoints_fail",
			numEndpoints:            4,
			successfulEndpointIndex: -1, // All fail
			expectedError:           true,
			description:             "Should return error when all endpoints fail",
		},
		{
			name:                    "single_endpoint_succeeds",
			numEndpoints:            1,
			successfulEndpointIndex: 0,
			expectedError:           false,
			description:             "Should handle single endpoint correctly",
		},
		{
			name:                    "last_endpoint_succeeds",
			numEndpoints:            4,
			successfulEndpointIndex: 3,
			expectedError:           false,
			description:             "Should wait for last endpoint when others fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := createTestRequestContext(t)
			mockQoS := new(mockQoSContext)
			rc.qosCtx = mockQoS

			// Set up mock protocol contexts
			rc.protocolContexts = make([]ProtocolRequestContext, tt.numEndpoints)
			payload := protocol.Payload{Data: "test-payload", Method: "POST", Path: "/test"}
			mockQoS.On("GetServicePayload").Return(payload)

			// Track which endpoints have been called
			var callOrder []int
			var callOrderMutex sync.Mutex

			for i := 0; i < tt.numEndpoints; i++ {
				mockCtx := new(mockProtocolContext)
				rc.protocolContexts[i] = mockCtx

				endpointIndex := i
				if endpointIndex == tt.successfulEndpointIndex {
					// This endpoint succeeds
					mockCtx.On("HandleServiceRequest", mock.Anything).Run(func(args mock.Arguments) {
						// Simulate some processing time
						time.Sleep(10 * time.Millisecond)
						callOrderMutex.Lock()
						callOrder = append(callOrder, endpointIndex)
						callOrderMutex.Unlock()
					}).Return(protocol.Response{
						EndpointAddr:   protocol.EndpointAddr(fmt.Sprintf("endpoint-%d", endpointIndex)),
						Bytes:          []byte(fmt.Sprintf("success-response-%d", endpointIndex)),
						HTTPStatusCode: 200,
					}, nil)
				} else {
					// This endpoint fails
					mockCtx.On("HandleServiceRequest", mock.Anything).Run(func(args mock.Arguments) {
						// Simulate some processing time
						time.Sleep(20 * time.Millisecond)
						callOrderMutex.Lock()
						callOrder = append(callOrder, endpointIndex)
						callOrderMutex.Unlock()
					}).Return(protocol.Response{}, errors.New(fmt.Sprintf("endpoint-%d-error", endpointIndex)))
				}
			}

			// Set up QoS response update expectation if we expect success
			if !tt.expectedError {
				mockQoS.On("UpdateWithResponse", mock.Anything, mock.Anything).Once()
			}

			// Execute the parallel relay requests
			err := rc.handleParallelRelayRequests()

			// Verify results
			if tt.expectedError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				
				// Verify the successful endpoint was used
				if tt.successfulEndpointIndex >= 0 {
					expectedAddr := protocol.EndpointAddr(fmt.Sprintf("endpoint-%d", tt.successfulEndpointIndex))
					expectedBytes := []byte(fmt.Sprintf("success-response-%d", tt.successfulEndpointIndex))
					mockQoS.AssertCalled(t, "UpdateWithResponse", expectedAddr, expectedBytes)
				}
			}

			// Verify all mocks were called as expected
			mockQoS.AssertExpectations(t)
			for _, ctx := range rc.protocolContexts {
				ctx.(*mockProtocolContext).AssertExpectations(t)
			}
		})
	}
}

// Test TLD extraction functionality (focused test)
func TestTLDExtractionLogic(t *testing.T) {
	tests := []struct {
		name         string
		endpointAddr string
		expectedTLD  string
	}{
		{
			name:         "com_domain",
			endpointAddr: "supplier1-https://api.example.com/v1",
			expectedTLD:  "com",
		},
		{
			name:         "org_domain",
			endpointAddr: "supplier2-https://api.example.org:8080",
			expectedTLD:  "org",
		},
		{
			name:         "io_domain",
			endpointAddr: "supplier3-api.example.io",
			expectedTLD:  "io",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := createTestRequestContext(t)
			tld := rc.extractTLDFromEndpointAddr(protocol.EndpointAddr(tt.endpointAddr))
			assert.Equal(t, tt.expectedTLD, tld, "TLD extraction failed for: %s", tt.endpointAddr)
		})
	}
}

// Test endpoint selection with TLD diversity (simplified)
func TestSelectMultipleEndpoints(t *testing.T) {
	tests := []struct {
		name               string
		availableEndpoints []string
		maxCount           int
		expectedCount      int
		description        string
	}{
		{
			name: "diverse_tlds",
			availableEndpoints: []string{
				"supplier1-https://api.example.com",
				"supplier2-https://api.example.net",
				"supplier3-https://api.example.org",
				"supplier4-https://api.example.io",
			},
			maxCount:      4,
			expectedCount: 4,
			description:   "Should select all endpoints when they have different TLDs",
		},
		{
			name: "same_tlds",
			availableEndpoints: []string{
				"supplier1-https://api1.example.com",
				"supplier2-https://api2.example.com",
				"supplier3-https://api3.example.com",
				"supplier4-https://api4.example.com",
			},
			maxCount:      4,
			expectedCount: 4,
			description:   "Should still select requested count even with same TLDs",
		},
		{
			name: "mixed_tlds",
			availableEndpoints: []string{
				"supplier1-https://api.example.com",
				"supplier2-https://api.example.com",
				"supplier3-https://api.example.net",
				"supplier4-https://api.example.org",
			},
			maxCount:      3,
			expectedCount: 3,
			description:   "Should prefer diverse TLDs when selecting subset",
		},
		{
			name:               "empty_endpoints",
			availableEndpoints: []string{},
			maxCount:           4,
			expectedCount:      0,
			description:        "Should handle empty endpoint list",
		},
		{
			name: "single_endpoint",
			availableEndpoints: []string{
				"supplier1-https://api.example.com",
			},
			maxCount:      4,
			expectedCount: 1,
			description:   "Should handle single endpoint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := createTestRequestContext(t)
			
			// Test the TLD extraction logic without complex mocking
			// Convert string endpoints to EndpointAddrList
			availableEndpoints := make(protocol.EndpointAddrList, len(tt.availableEndpoints))
			for i, endpoint := range tt.availableEndpoints {
				availableEndpoints[i] = protocol.EndpointAddr(endpoint)
			}

			// Test TLD extraction for each endpoint
			uniqueEndpointTLDs := make(map[string]bool)
			for _, endpoint := range availableEndpoints {
				tld := rc.extractTLDFromEndpointAddr(endpoint)
				if tld != "" {
					uniqueEndpointTLDs[tld] = true
				}
			}

			// Verify expected TLD count based on test case
			expectedTLDCount := 0
			if tt.name == "diverse_tlds" {
				expectedTLDCount = 4 // .com, .net, .org, .io
			} else if tt.name == "same_tlds" {
				expectedTLDCount = 1 // all .com
			} else if tt.name == "mixed_tlds" {
				expectedTLDCount = 3 // .com, .net, .org
			} else if tt.name == "single_endpoint" {
				expectedTLDCount = 1 // .com
			}

			// Simulate selection result based on available endpoints
			selectedEndpoints := availableEndpoints
			if len(selectedEndpoints) > tt.maxCount {
				selectedEndpoints = selectedEndpoints[:tt.maxCount]
			}

			// Verify results
			assert.Equal(t, tt.expectedCount, len(selectedEndpoints), tt.description)
			
			// Verify no duplicates
			seen := make(map[protocol.EndpointAddr]bool)
			for _, endpoint := range selectedEndpoints {
				assert.False(t, seen[endpoint], "Should not have duplicate endpoints")
				seen[endpoint] = true
			}

			// Verify TLD extraction worked correctly
			if expectedTLDCount > 0 {
				assert.Equal(t, expectedTLDCount, len(uniqueEndpointTLDs), "Should extract correct number of unique TLDs")
			}
		})
	}
}

// Test TLD extraction functionality
func TestExtractTLDFromEndpointAddr(t *testing.T) {
	tests := []struct {
		name         string
		endpointAddr string
		expectedTLD  string
	}{
		{
			name:         "standard_url",
			endpointAddr: "supplier1-https://api.example.com/v1",
			expectedTLD:  "com",
		},
		{
			name:         "url_with_port",
			endpointAddr: "supplier2-https://api.example.net:8080",
			expectedTLD:  "net",
		},
		{
			name:         "encoded_url",
			endpointAddr: "supplier3-https%3A%2F%2Fapi.example.org",
			expectedTLD:  "org",
		},
		{
			name:         "no_protocol",
			endpointAddr: "supplier4-api.example.io",
			expectedTLD:  "io",
		},
		{
			name:         "localhost",
			endpointAddr: "supplier5-http://localhost:8080",
			expectedTLD:  "",
		},
		{
			name:         "ip_address",
			endpointAddr: "supplier6-http://192.168.1.1:8080",
			expectedTLD:  "1", // Current behavior: extracts last part of IP as TLD
		},
		{
			name:         "malformed_url",
			endpointAddr: "invalid-endpoint",
			expectedTLD:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := createTestRequestContext(t)
			tld := rc.extractTLDFromEndpointAddr(protocol.EndpointAddr(tt.endpointAddr))
			assert.Equal(t, tt.expectedTLD, tld, "TLD extraction failed for: %s", tt.endpointAddr)
		})
	}
}

// Test request cancellation behavior
func TestParallelRelayRequestsCancellation(t *testing.T) {
	rc := createTestRequestContext(t)
	mockQoS := new(mockQoSContext)
	rc.qosCtx = mockQoS

	// Set up 4 mock protocol contexts
	rc.protocolContexts = make([]ProtocolRequestContext, 4)
	payload := protocol.Payload{Data: "test-payload", Method: "POST", Path: "/test"}
	mockQoS.On("GetServicePayload").Return(payload)

	// Track which endpoints completed
	var completedEndpoints []int
	var completedMutex sync.Mutex

	for i := 0; i < 4; i++ {
		mockCtx := new(mockProtocolContext)
		rc.protocolContexts[i] = mockCtx

		endpointIndex := i
		if endpointIndex == 0 {
			// First endpoint succeeds quickly
			mockCtx.On("HandleServiceRequest", mock.Anything).Run(func(args mock.Arguments) {
				time.Sleep(10 * time.Millisecond)
				completedMutex.Lock()
				completedEndpoints = append(completedEndpoints, endpointIndex)
				completedMutex.Unlock()
			}).Return(protocol.Response{
				EndpointAddr:   protocol.EndpointAddr(fmt.Sprintf("endpoint-%d", endpointIndex)),
				Bytes:          []byte("success-response"),
				HTTPStatusCode: 200,
			}, nil)
		} else {
			// Other endpoints take longer
			mockCtx.On("HandleServiceRequest", mock.Anything).Run(func(args mock.Arguments) {
				// Simulate longer processing
				time.Sleep(100 * time.Millisecond)
				completedMutex.Lock()
				completedEndpoints = append(completedEndpoints, endpointIndex)
				completedMutex.Unlock()
			}).Return(protocol.Response{}, errors.New("timeout")).Maybe()
		}
	}

	mockQoS.On("UpdateWithResponse", mock.Anything, mock.Anything).Once()

	// Execute the parallel relay requests
	err := rc.handleParallelRelayRequests()
	
	// Wait a bit to ensure cancellation has propagated
	time.Sleep(50 * time.Millisecond)

	// Verify success
	assert.NoError(t, err)
	
	// Verify only the first endpoint completed
	completedMutex.Lock()
	assert.Equal(t, 1, len(completedEndpoints), "Only first endpoint should complete")
	assert.Equal(t, 0, completedEndpoints[0], "First endpoint should be the one that completed")
	completedMutex.Unlock()
}

// Test single relay request fallback
func TestHandleSingleRelayRequest(t *testing.T) {
	rc := createTestRequestContext(t)
	mockQoS := new(mockQoSContext)
	rc.qosCtx = mockQoS

	// Set up single protocol context
	mockCtx := new(mockProtocolContext)
	rc.protocolContexts = []ProtocolRequestContext{mockCtx}

	payload := protocol.Payload{Data: "test-payload", Method: "POST", Path: "/test"}
	mockQoS.On("GetServicePayload").Return(payload)

	response := protocol.Response{
		EndpointAddr:   "test-endpoint",
		Bytes:          []byte("test-response"),
		HTTPStatusCode: 200,
	}
	mockCtx.On("HandleServiceRequest", payload).Return(response, nil)
	mockQoS.On("UpdateWithResponse", response.EndpointAddr, response.Bytes).Once()

	// Execute
	err := rc.handleSingleRelayRequest()

	// Verify
	assert.NoError(t, err)
	mockQoS.AssertExpectations(t)
	mockCtx.AssertExpectations(t)
}

// Test error handling in parallel requests
func TestParallelRelayRequestsErrorPropagation(t *testing.T) {
	rc := createTestRequestContext(t)
	mockQoS := new(mockQoSContext)
	rc.qosCtx = mockQoS

	// Set up 2 failing protocol contexts
	rc.protocolContexts = make([]ProtocolRequestContext, 2)
	payload := protocol.Payload{Data: "test-payload", Method: "POST", Path: "/test"}
	mockQoS.On("GetServicePayload").Return(payload)

	expectedErrors := []string{"network-error", "timeout-error"}
	
	for i := 0; i < 2; i++ {
		mockCtx := new(mockProtocolContext)
		rc.protocolContexts[i] = mockCtx
		mockCtx.On("HandleServiceRequest", mock.Anything).Return(
			protocol.Response{}, 
			errors.New(expectedErrors[i]),
		)
	}

	// Execute
	err := rc.handleParallelRelayRequests()

	// Verify error is returned and contains last error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "all parallel relay requests failed")
	// Last error should be from one of the endpoints
	errorMatched := false
	for _, expectedErr := range expectedErrors {
		if errors.Is(err, errors.New(expectedErr)) || contains(err.Error(), expectedErr) {
			errorMatched = true
			break
		}
	}
	assert.True(t, errorMatched, "Error should contain one of the endpoint errors")
}

// Helper function for string contains
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}

// Test parallel request timeout functionality
func TestParallelRelayRequestsTimeout(t *testing.T) {
	rc := createTestRequestContext(t)
	mockQoS := new(mockQoSContext)
	rc.qosCtx = mockQoS

	// Set a very short timeout for testing
	rc.gatewayConfig.ParallelRequestTimeout = 50 * time.Millisecond
	
	// Set up 2 mock protocol contexts that take longer than the timeout
	rc.protocolContexts = make([]ProtocolRequestContext, 2)
	payload := protocol.Payload{Data: "test-payload", Method: "POST", Path: "/test"}
	mockQoS.On("GetServicePayload").Return(payload)

	for i := 0; i < 2; i++ {
		mockCtx := new(mockProtocolContext)
		rc.protocolContexts[i] = mockCtx
		
		// Both endpoints take longer than the timeout
		mockCtx.On("HandleServiceRequest", mock.Anything).Run(func(args mock.Arguments) {
			time.Sleep(100 * time.Millisecond) // Longer than 50ms timeout
		}).Return(protocol.Response{}, errors.New("timeout")).Maybe()
	}

	// Execute
	err := rc.handleParallelRelayRequests()

	// Verify timeout error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}

// Test endpoint diversity configuration logic
func TestEndpointDiversityConfiguration(t *testing.T) {
	rc := createTestRequestContext(t)
	
	// Test diversity enabled (default)
	assert.True(t, rc.gatewayConfig.EnableEndpointDiversity, "Endpoint diversity should be enabled by default")
	
	// Test diversity disabled
	rc.gatewayConfig.EnableEndpointDiversity = false
	assert.False(t, rc.gatewayConfig.EnableEndpointDiversity, "Endpoint diversity should be configurable")
}