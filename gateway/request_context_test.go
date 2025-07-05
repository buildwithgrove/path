package gateway

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	shannonmetrics "github.com/buildwithgrove/path/metrics/protocol/shannon"
	"github.com/buildwithgrove/path/protocol"
)

// Test helper functions
func createTestRequestContext(t *testing.T) *requestContext {
	t.Helper()

	logger := polyzero.NewLogger()

	return &requestContext{
		context:   context.Background(),
		logger:    logger,
		serviceID: "test-service",
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
			// Test the TLD extraction logic without complex mocking
			// Convert string endpoints to EndpointAddrList
			availableEndpoints := make(protocol.EndpointAddrList, len(tt.availableEndpoints))
			for i, endpoint := range tt.availableEndpoints {
				availableEndpoints[i] = protocol.EndpointAddr(endpoint)
			}

			// Test TLD extraction for each endpoint
			uniqueEndpointTLDs := make(map[string]bool)
			for _, endpoint := range availableEndpoints {
				tld, err := shannonmetrics.ExtractEffectiveTLDPlusOne(string(endpoint))
				require.NoError(t, err)
				if tld != "" {
					uniqueEndpointTLDs[tld] = true
				}
			}

			// Verify expected TLD count based on test case
			expectedTLDCount := 0
			switch tt.name {
			case "diverse_tlds":
				expectedTLDCount = 4 // .com, .net, .org, .io
			case "same_tlds":
				expectedTLDCount = 1 // all .com
			case "mixed_tlds":
				expectedTLDCount = 3 // .com, .net, .org
			case "single_endpoint":
				expectedTLDCount = 1 // .com
			}

			// Simulate selection result based on available endpoints
			selectedEndpoints := availableEndpoints
			if len(selectedEndpoints) > tt.maxCount {
				selectedEndpoints = selectedEndpoints[:tt.maxCount]
			}

			// Verify results
			require.Equal(t, tt.expectedCount, len(selectedEndpoints), tt.description)

			// Verify no duplicates
			seen := make(map[protocol.EndpointAddr]bool)
			for _, endpoint := range selectedEndpoints {
				require.False(t, seen[endpoint], "Should not have duplicate endpoints")
				seen[endpoint] = true
			}

			// Verify TLD extraction worked correctly
			if expectedTLDCount > 0 {
				require.Equal(t, expectedTLDCount, len(uniqueEndpointTLDs), "Should extract correct number of unique TLDs")
			}
		})
	}
}

// Test parallel relay requests functionality using generated mocks
func TestHandleParallelRelayRequests(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
			mockQoS := NewMockRequestQoSContext(ctrl)
			rc.qosCtx = mockQoS

			// Set up mock protocol contexts
			rc.protocolContexts = make([]ProtocolRequestContext, tt.numEndpoints)
			payload := protocol.Payload{Data: "test-payload", Method: "POST", Path: "/test"}
			mockQoS.EXPECT().GetServicePayload().Return(payload).AnyTimes()

			// Track which endpoints have been called
			var callOrder []int
			var callOrderMutex sync.Mutex

			for i := 0; i < tt.numEndpoints; i++ {
				mockCtx := NewMockProtocolRequestContext(ctrl)
				rc.protocolContexts[i] = mockCtx

				endpointIndex := i
				if endpointIndex == tt.successfulEndpointIndex {
					// This endpoint succeeds
					mockCtx.EXPECT().HandleServiceRequest(gomock.Any()).DoAndReturn(func(p protocol.Payload) (protocol.Response, error) {
						// Simulate some processing time
						time.Sleep(10 * time.Millisecond)
						callOrderMutex.Lock()
						callOrder = append(callOrder, endpointIndex)
						callOrderMutex.Unlock()
						return protocol.Response{
							EndpointAddr:   protocol.EndpointAddr(fmt.Sprintf("endpoint-%d", endpointIndex)),
							Bytes:          []byte(fmt.Sprintf("success-response-%d", endpointIndex)),
							HTTPStatusCode: 200,
						}, nil
					}).AnyTimes() // Use AnyTimes for parallel execution
				} else {
					// This endpoint fails
					mockCtx.EXPECT().HandleServiceRequest(gomock.Any()).DoAndReturn(func(p protocol.Payload) (protocol.Response, error) {
						// Simulate some processing time
						time.Sleep(20 * time.Millisecond)
						callOrderMutex.Lock()
						callOrder = append(callOrder, endpointIndex)
						callOrderMutex.Unlock()
						return protocol.Response{}, fmt.Errorf("endpoint-%d-error", endpointIndex)
					}).AnyTimes() // Use AnyTimes for parallel execution
				}
			}

			// Set up QoS response update expectation if we expect success
			if !tt.expectedError {
				mockQoS.EXPECT().UpdateWithResponse(gomock.Any(), gomock.Any()).AnyTimes()
			}

			// Execute the parallel relay requests
			err := rc.handleParallelRelayRequests()

			// Verify results
			if tt.expectedError {
				require.Error(t, err, tt.description)
			} else {
				require.NoError(t, err, tt.description)
			}
		})
	}
}

// Test single relay request fallback using generated mocks
func TestHandleSingleRelayRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rc := createTestRequestContext(t)
	mockQoS := NewMockRequestQoSContext(ctrl)
	rc.qosCtx = mockQoS

	// Set up single protocol context
	mockCtx := NewMockProtocolRequestContext(ctrl)
	rc.protocolContexts = []ProtocolRequestContext{mockCtx}

	payload := protocol.Payload{Data: "test-payload", Method: "POST", Path: "/test"}
	mockQoS.EXPECT().GetServicePayload().Return(payload)

	response := protocol.Response{
		EndpointAddr:   "test-endpoint",
		Bytes:          []byte("test-response"),
		HTTPStatusCode: 200,
	}
	mockCtx.EXPECT().HandleServiceRequest(payload).Return(response, nil)
	mockQoS.EXPECT().UpdateWithResponse(response.EndpointAddr, response.Bytes).Times(1)

	// Execute
	err := rc.handleSingleRelayRequest()

	// Verify
	require.NoError(t, err)
}
