//go:build e2e

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
	"github.com/buildwithgrove/path/qos/evm"
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/buildwithgrove/path/request"
)

/* -------------------- Test Configuration Initialization -------------------- */

// protocolStr is a type to determine whether to test PATH with Morse or Shannon
type protocolStr string

const (
	morse   protocolStr = "morse"
	shannon protocolStr = "shannon"
)

func (p protocolStr) isValid() bool {
	return p == morse || p == shannon
}

var (
	// Set default gateway URL
	// Uses port from test Docker container
	// Defaults to port `3069`
	gatewayURL = "http://localhost:%s/v1"

	// Protocol string for the test
	// eg. `morse` or `shannon`
	testProtocol protocolStr
	// Config path for protocol
	// eg. `./.morse.config.yaml` or `./.shannon.config.yaml`
	configPath = "./.%s.config.yaml"
)

// init initializes the gateway URL with an optional override
func init() {
	if gatewayURLOverride := os.Getenv("GATEWAY_URL"); gatewayURLOverride != "" {
		gatewayURL = gatewayURLOverride
	}
	if testProtocol = protocolStr(os.Getenv("TEST_PROTOCOL")); testProtocol == "" {
		panic("TEST_PROTOCOL environment variable is not set")
	}
	if !testProtocol.isValid() {
		panic("TEST_PROTOCOL environment variable is not set to `morse` or `shannon`")
	}
}

/* -------------------- Get Test Cases for Protocol -------------------- */

// testCase represents a single service load test configuration
type testCase struct {
	name          string
	serviceID     protocol.ServiceID // The service ID to test
	archival      bool               // Whether to select a random historical block
	methods       []jsonrpc.Method   // The methods to test for this service
	serviceParams serviceParameters  // Service-specific parameters
}

// getTestCases returns the appropriate test cases based on the protocol
func getTestCases(protocolStr protocolStr) []testCase {
	switch protocolStr {
	case morse:
		return getMorseTestCases()
	case shannon:
		return getShannonTestCases()
	default:
		// This shouldn't happen due to the init check, but just in case
		panic(fmt.Sprintf("Unsupported protocol: %s", protocolStr))
	}
}

// getMorseTestCases returns test cases for Morse protocol
func getMorseTestCases() []testCase {
	return []testCase{
		{
			name:      "F00C (Ethereum) Load Test",
			serviceID: "F00C",
			methods:   runAllMethods(),
			archival:  true, // F00C is an archival service so we should use a random historical block.
			serviceParams: serviceParameters{
				contractAddress: "0xdAC17F958D2ee523a2206206994597C13D831ec7",
				transactionHash: "0xfeccd627b5b391d04fe45055873de3b2c0b4302d52e96bd41d5f0019a704165f",
				callData:        "0x18160ddd",
			},
		},
	}
}

// getShannonTestCases returns test cases for Shannon protocol
func getShannonTestCases() []testCase {
	return []testCase{
		{
			name:      "anvil (Ethereum) Load Test",
			serviceID: "anvil",
			methods:   runAllMethods(),
			serviceParams: serviceParameters{
				contractAddress: "0xdAC17F958D2ee523a2206206994597C13D831ec7",
				transactionHash: "0xfeccd627b5b391d04fe45055873de3b2c0b4302d52e96bd41d5f0019a704165f",
				callData:        "0x18160ddd",
			},
		},
	}
}

/* -------------------- EVM Load Test Function -------------------- */

// Test_PATH_E2E_EVM runs an E2E load test against the EVM JSON-RPC endpoints
func Test_PATH_E2E_EVM(t *testing.T) {
	fmt.Println("Setting up PATH instance...")

	var pathContainerPort string

	// Start an instance of PATH using the E2E config file for Shannon.
	configFilePath := fmt.Sprintf(configPath, testProtocol)
	pathContainerPort, teardownFn := setupPathInstance(t, configFilePath)
	defer teardownFn()

	if pathContainerPort == "" {
		pathContainerPort = "3069"
	}

	gatewayURL = fmt.Sprintf(gatewayURL, pathContainerPort)

	fmt.Printf("🌿 Starting PATH E2E EVM test.\n")
	fmt.Printf("  🧬 Gateway URL: %s\n", gatewayURL)
	fmt.Printf("  📡 Test protocol: %s\n", testProtocol)

	// // In the CI environment, we need to wait for 2 minutes to allow the hydrator checks to complete.
	// if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
	// 	fmt.Println("⏰ Waiting for 2 minutes before starting tests to allow several rounds of hydrator checks to complete...")
	// 	time.Sleep(2 * time.Minute)
	// } else {
	// 	secondsToWait := 40
	// 	fmt.Printf("⏰ Waiting for %d seconds before starting tests to allow several rounds of hydrator checks to complete...\n", secondsToWait)
	// 	showWaitBar(secondsToWait)
	// }

	// Get test cases based on protocol
	testCases := getTestCases(testProtocol)

	for i := range testCases {
		// If archival is true then we will use a random historical block for the test.
		if testCases[i].archival {
			testCases[i].serviceParams.blockNumber = setTestBlockNumber(t, gatewayURL, testCases[i].serviceID)
		} else {
			testCases[i].serviceParams.blockNumber = "latest"
		}

		fmt.Printf("🛠️  Testing service %d of %d\n", i+1, len(testCases))
		fmt.Printf("  ⛓️  Service ID: %s\n", testCases[i].serviceID)
		fmt.Printf("  📡 Block number: %s\n", testCases[i].serviceParams.blockNumber)

		// Use t.Run for proper test reporting
		t.Run(testCases[i].name, func(t *testing.T) {
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

				// Run the attack in a goroutine
				go func(method jsonrpc.Method, def methodDefinition) {
					defer methodWg.Done()

					// Create the JSON-RPC request
					jsonrpcReq := jsonrpc.Request{
						JSONRPC: jsonrpc.Version2,
						ID:      jsonrpc.IDFromInt(1),
						Method:  method,
						Params: createParams(
							method,
							testCases[i].serviceParams,
						),
					}

					// Run the attack
					metrics := runAttack(
						gatewayURL,
						testCases[i].serviceID,
						method,
						def,
						progBars.get(method),
						jsonrpcReq,
					)

					// Safely store the results
					resultsMutex.Lock()
					results[method] = metrics
					resultsMutex.Unlock()
				}(method, methodDef)
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
func setTestBlockNumber(t *testing.T, gatewayURL string, serviceID protocol.ServiceID) string {
	// Get current block height - fail test if this doesn't work
	currentBlock, err := getCurrentBlockNumber(gatewayURL, serviceID)
	if err != nil {
		t.Fatalf("FATAL: Could not get current block height: %v", err)
	}

	// Get random historical block number
	return getBlockNumber(currentBlock)
}

// getCurrentBlockNumber gets current block height with consensus from multiple requests
func getCurrentBlockNumber(gatewayURL string, serviceID protocol.ServiceID) (uint64, error) {
	// Track frequency of each block height seen
	blockHeights := make(map[uint64]int)
	maxAttempts, requiredAgreement := 5, 3
	client := &http.Client{Timeout: 5 * time.Second}

	// Make multiple attempts to get consensus
	for range maxAttempts {
		blockNum, err := fetchBlockNumber(client, gatewayURL, serviceID)
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
func fetchBlockNumber(client *http.Client, gatewayURL string, serviceID protocol.ServiceID) (uint64, error) {
	// Build and send request
	req, err := buildBlockNumberRequest(gatewayURL, serviceID)
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
func buildBlockNumberRequest(gatewayURL string, serviceID protocol.ServiceID) (*http.Request, error) {
	blockNumberReq := jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(1),
		Method:  jsonrpc.Method(eth_blockNumber),
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
	if maxBlock > evm.DefaultEVMArchivalThreshold {
		maxBlock -= evm.DefaultEVMArchivalThreshold
	}

	if minBlock >= maxBlock {
		minBlock, maxBlock = 100, 1100 // Fallback
	}

	// Generate random block number
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomBlock := minBlock + r.Uint64()%(maxBlock-minBlock+1)

	return fmt.Sprintf("0x%x", randomBlock)
}
