package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/buildwithgrove/path/request"
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

/* -------------------- EVM Load Test Function -------------------- */

// Test_PATHLoad runs load tests against JSON-RPC endpoints
func Test_PATHLoad(t *testing.T) {
	fmt.Printf("Starting load test with gateway URL: %s\n", gatewayURL)

	// Define test cases by service
	testCases := []struct {
		name      string
		serviceID protocol.ServiceID // The service ID to test
		archival  bool               // Whether to select a random historical block
		methods   []jsonrpc.Method   // The methods to test for this service
		params    methodParams       // Service-specific parameters
	}{
		{
			name:      "F00C (Ethereum) Load Test",
			serviceID: "F00C",
			methods:   runAllMethods(),
			archival:  true, // F00C is an archival service so we should use a random historical block.
			params: methodParams{
				contractAddress: "0xdAC17F958D2ee523a2206206994597C13D831ec7",
				transactionHash: "0xfeccd627b5b391d04fe45055873de3b2c0b4302d52e96bd41d5f0019a704165f",
				callData:        "0x18160ddd",
			},
		},
	}

	for i := range testCases {
		// If archival is true then we will use a random historical block for the test.
		if testCases[i].archival {
			testCases[i].params.blockNumber = setTestBlockNumber(t, testCases[i].serviceID)
		} else {
			testCases[i].params.blockNumber = "latest"
		}

		// Use t.Run for proper test reporting
		t.Run(testCases[i].name, func(t *testing.T) {
			fmt.Printf("Testing service: %s with block: %s\n", testCases[i].serviceID, testCases[i].params.blockNumber)

			// Create results map with a mutex to protect concurrent access
			results := make(map[jsonrpc.Method]*MethodMetrics)
			var resultsMutex sync.Mutex

			// Validate that all methods have a definition
			for _, method := range testCases[i].methods {
				if _, exists := methodDefinitions[method]; !exists {
					t.Fatalf("No definition for method %s", method)
				}
			}

			// Create and start all progress bars upfront
			progBars, err := newProgressBars(testCases[i].methods, methodDefinitions)
			if err != nil {
				t.Fatalf("Failed to create progress bars: %v", err)
			}

			// Make sure we stop the progress bars before printing results
			defer func() {
				if err := progBars.finish(); err != nil {
					t.Logf("Error stopping progress bars: %v", err)
				}
			}()

			// Create wait group for methods
			var methodWg sync.WaitGroup

			// Run attack for each method concurrently
			for _, method := range testCases[i].methods {
				methodWg.Add(1)

				// Get method configuration
				methodDef := methodDefinitions[method]

				// Create method-specific parameters
				methodParams := createParams(method, testCases[i].params)

				// Run the attack in a goroutine
				go func(method jsonrpc.Method, def methodDefinition, params []any) {
					defer methodWg.Done()

					// Run the attack
					metrics := runAttack(
						gatewayURL,
						testCases[i].serviceID,
						method,
						def,
						progBars.get(method),
						params...,
					)

					// Safely store the results
					resultsMutex.Lock()
					results[method] = metrics
					resultsMutex.Unlock()
				}(method, methodDef, methodParams)
			}

			// Wait for all method tests to complete
			methodWg.Wait()

			// Make sure progress bars are stopped before printing results
			if err := progBars.finish(); err != nil {
				t.Logf("Error stopping progress bars: %v", err)
			}

			// Add space after progress bars
			fmt.Println()

			// Validate results for each method
			for _, method := range testCases[i].methods {
				validateResults(t, results[method], methodDefinitions[method])
			}
		})
	}
}

/* -------------------- Get Test Block Number -------------------- */

// setTestBlockNumber gets a block number for testing or fails the test
func setTestBlockNumber(t *testing.T, serviceID protocol.ServiceID) string {
	// Get current block height - fail test if this doesn't work
	currentBlock, err := getCurrentBlockNumber(serviceID)
	if err != nil {
		t.Fatalf("FATAL: Could not get current block height: %v", err)
	}

	// Get random historical block number
	return getBlockNumber(currentBlock)
}

// getCurrentBlockNumber gets current block height with consensus from multiple requests
func getCurrentBlockNumber(serviceID protocol.ServiceID) (uint64, error) {
	// Track frequency of each block height seen
	blockHeights := make(map[uint64]int)
	maxAttempts, requiredAgreement := 5, 3
	client := &http.Client{Timeout: 5 * time.Second}

	// Make multiple attempts to get consensus
	for i := 0; i < maxAttempts; i++ {
		blockNum, err := fetchBlockNumber(client, serviceID)
		if err != nil {
			continue
		}

		// Update consensus tracking
		blockHeights[blockNum]++
		if blockHeights[blockNum] >= requiredAgreement {
			return blockNum, nil
		}
	}

	// If we get here, we didn't reach consensus
	return 0, fmt.Errorf("failed to reach consensus on block height after %d attempts", maxAttempts)
}

// fetchBlockNumber makes a single request to get the current block number
func fetchBlockNumber(client *http.Client, serviceID protocol.ServiceID) (uint64, error) {
	// Build and send request
	req, err := buildBlockNumberRequest(serviceID)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	// Parse response
	var jsonRPC jsonrpc.Response
	if err := json.NewDecoder(resp.Body).Decode(&jsonRPC); err != nil {
		return 0, err
	}

	// Process hex string result
	hexString, ok := jsonRPC.Result.(string)
	if !ok {
		return 0, fmt.Errorf("expected string result, got %T", jsonRPC.Result)
	}

	// Parse hex (remove "0x" prefix if present)
	hexStr := strings.TrimPrefix(hexString, "0x")
	blockNum, err := strconv.ParseUint(hexStr, 16, 64)
	if err != nil {
		return 0, err
	}

	return blockNum, nil
}

// Helper to build a block number request
func buildBlockNumberRequest(serviceID protocol.ServiceID) (*http.Request, error) {
	blockNumberReq, err := buildJSONRPCReq(1, eth_blockNumber)
	if err != nil {
		return nil, err
	}

	blockNumberReqBytes, err := json.Marshal(blockNumberReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, gatewayURL, bytes.NewReader(blockNumberReqBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(request.HTTPHeaderTargetServiceID, string(serviceID))

	return req, nil
}

// getBlockNumber selects a random block number in a safe range
func getBlockNumber(currentBlock uint64) string {
	// Define safe range with fallbacks
	minBlock := uint64(100)
	maxBlock := currentBlock

	// Ensure the block selected is for archival EVM data
	if maxBlock > 128 {
		maxBlock -= 128
	}

	if minBlock >= maxBlock {
		minBlock, maxBlock = 100, 1100 // Fallback
	}

	// Generate random block number
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomBlock := minBlock + r.Uint64()%(maxBlock-minBlock+1)

	return fmt.Sprintf("0x%x", randomBlock)
}
