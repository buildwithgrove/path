package e2e

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

/* -------------------- Gateway URL -------------------- */

// Set default gateway URL
var gatewayURL = "http://localhost:3069/v1"

// init initializes the gateway URL with an optional override
func init() {
	if gatewayURLOverride := os.Getenv("GATEWAY_URL"); gatewayURLOverride != "" {
		gatewayURL = gatewayURLOverride
	}
}

/* -------------------- Load Test Functions -------------------- */

// Test_PATHLoad runs load tests against JSON-RPC endpoints
func Test_PATHLoad(t *testing.T) {
	t.Logf("Starting load test with gateway URL: %s", gatewayURL)

	// Define test cases by service
	testCases := []struct {
		name      string
		serviceID protocol.ServiceID // The service ID to test
		methods   []jsonrpc.Method   // The methods to test for this service
		params    methodParams       // Service-specific parameters
	}{
		{
			name:      "F00C (Ethereum) Load Test",
			serviceID: "F00C",
			methods: []jsonrpc.Method{
				eth_blockNumber,
				eth_call,
			},
			params: methodParams{
				blockNumber:     "latest", //TODO_IMPROVE(@commoddity): select a random block number
				contractAddress: "0xdAC17F958D2ee523a2206206994597C13D831ec7",
				transactionHash: "0xfeccd627b5b391d04fe45055873de3b2c0b4302d52e96bd41d5f0019a704165f",
				callData:        "0x18160ddd",
			},
		},
		// Add more test cases here as needed
	}

	// Create a wait group to wait for all service tests to complete
	var serviceWg sync.WaitGroup
	// Channel to synchronize progress bar output
	barStartCh := make(chan struct{}, 1)

	// Start all service tests concurrently
	for i := range testCases {
		tc := testCases[i] // Capture the test case
		serviceWg.Add(1)

		// Run each service test in a separate goroutine
		go func() {
			defer serviceWg.Done()
			runServiceTest(t, tc, barStartCh)
		}()
	}

	// Wait for all service tests to complete
	serviceWg.Wait()
}

// runServiceTest executes a load test for a specific service
func runServiceTest(t *testing.T, tc struct {
	name      string
	serviceID protocol.ServiceID
	methods   []jsonrpc.Method
	params    methodParams
}, barStartCh chan struct{}) {
	// Use t.Run for proper test reporting
	t.Run(tc.name, func(t *testing.T) {
		t.Logf("Testing service: %s", tc.serviceID)

		// Create results map
		results := make(map[jsonrpc.Method]*MethodMetrics)

		// Log method configurations and validate they exist
		for _, method := range tc.methods {
			methodDef, exists := methodDefinitions[method]
			if !exists {
				t.Fatalf("No definition for method %s", method)
			}

			t.Logf("Method: %s, Total Requests: %d, rps: %d",
				method, methodDef.totalRequests, methodDef.rps)
		}

		// Create and start all progress bars upfront
		progBars, err := newProgressBars(tc.methods, methodDefinitions)
		if err != nil {
			t.Fatalf("Failed to create progress bars: %v", err)
		}

		// Synchronize progress bar display across services
		select {
		case barStartCh <- struct{}{}:
			// We acquired the lock, continue
			defer func() { <-barStartCh }() // Release the lock when done
		case <-barStartCh:
			// Someone else has the lock, wait a moment
			time.Sleep(100 * time.Millisecond)
			barStartCh <- struct{}{}        // Put it back
			defer func() { <-barStartCh }() // Release when done
		}

		// Make sure we stop the progress bars before printing results
		defer func() {
			if err := progBars.finish(); err != nil {
				t.Logf("Error stopping progress bars: %v", err)
			}
		}()

		// Create wait group for methods
		var methodWg sync.WaitGroup

		// Run attack for each method within this service
		for _, method := range tc.methods {
			methodWg.Add(1)

			// Get method configuration
			methodDef := methodDefinitions[method]

			// Create method-specific parameters
			methodParams := createParams(method, tc.params)

			// Run the attack in a goroutine
			go func(method jsonrpc.Method, def methodDefinition, params []any) {
				defer methodWg.Done()
				results[method] = runAttack(
					gatewayURL,
					tc.serviceID,
					method,
					def,
					progBars.get(method),
					params...,
				)
			}(method, methodDef, methodParams)
		}

		// Wait for all method tests within this service to complete
		methodWg.Wait()

		// Make sure progress bars are stopped before printing results
		if err := progBars.finish(); err != nil {
			t.Logf("Error stopping progress bars: %v", err)
		}

		// Add space after progress bars
		fmt.Println()

		// Validate results for each method
		for _, method := range tc.methods {
			validateResults(t, results[method], methodDefinitions[method])
		}
	})
}
