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

// Global test options
var opts testOptions

// init initializes the test options
func init() {
	opts = gatherTestOptions()
}

// Environment variable names
const (
	// Required environment variables
	envTestProtocol = "TEST_PROTOCOL"

	// Optional environment variables
	envGatewayURLOverride = "GATEWAY_URL_OVERRIDE"
	envDockerLog          = "DOCKER_LOG"
	envDockerForceRebuild = "DOCKER_FORCE_REBUILD"
)

// testOptions contains all configuration options for the E2E tests
type (
	testOptions struct {
		// Protocol to use for testing (morse or shannon)
		// Required environment variable: TEST_PROTOCOL
		testProtocol protocolStr

		// URL for accessing the gateway
		// If not set, default is "http://localhost:%s/v1" where %s is the port of the Docker container
		// If set via GATEWAY_URL_OVERRIDE, the Docker container won't be used and
		// the test will run against the provided URL directly
		gatewayURL string

		// Whether the gateway URL was explicitly set via GATEWAY_URL_OVERRIDE
		// This also indicates that no Docker container should be started
		//
		// If GATEWAY_URL_OVERRIDE is set, we'll use the provided URL directly and skip starting a Docker container,
		// assuming PATH is already running externally at the provided URL.
		gatewayURLOverridden bool

		// Docker-related configuration options
		docker dockerOptions

		// Config file path template
		// Format: "./.%s.config.yaml" where %s is the protocol name
		configPathTemplate string
	}
	// dockerOptions contains configuration for the Docker test container
	dockerOptions struct {
		// Whether to log docker container output
		// Default: false
		// Can be enabled with DOCKER_LOG=true
		logOutput bool

		// Whether to force rebuild of the docker image
		// Default: false
		// Can be enabled with DOCKER_FORCE_REBUILD=true
		forceRebuild bool
	}
)

// protocolStr is a type to determine whether to test PATH with Morse or Shannon
type protocolStr string

const (
	morse   protocolStr = "morse"
	shannon protocolStr = "shannon"
)

func (p protocolStr) isValid() bool {
	return p == morse || p == shannon
}

// gatherTestOptions collects all test configuration options from environment variables
func gatherTestOptions() testOptions {
	// Default values
	options := testOptions{
		gatewayURL:         "http://localhost:%s/v1", // eg. `http://localhost:3069/v1`
		configPathTemplate: "./.%s.config.yaml",      // eg. `./.morse.config.yaml` or `./.shannon.config.yaml`
	}

	// Required environment variables
	if testProtocol := protocolStr(os.Getenv(envTestProtocol)); testProtocol == "" {
		panic(fmt.Sprintf("%s environment variable is not set", envTestProtocol))
	} else if !testProtocol.isValid() {
		panic(fmt.Sprintf("%s environment variable is not set to `morse` or `shannon`", envTestProtocol))
	} else {
		options.testProtocol = testProtocol
	}

	// Optional environment variables
	if gatewayURLOverride := os.Getenv(envGatewayURLOverride); gatewayURLOverride != "" {
		options.gatewayURL = gatewayURLOverride
		options.gatewayURLOverridden = true
	}

	// Docker configuration
	if logValue := os.Getenv(envDockerLog); logValue != "" {
		if logParsed, err := strconv.ParseBool(logValue); err == nil {
			options.docker.logOutput = logParsed
		}
	}

	if rebuildValue := os.Getenv(envDockerForceRebuild); rebuildValue != "" {
		if rebuildParsed, err := strconv.ParseBool(rebuildValue); err == nil {
			options.docker.forceRebuild = rebuildParsed
		}
	}

	return options
}

/* -------------------- Get Test Cases for Protocol -------------------- */

// testCase represents a single service load test configuration
type testCase struct {
	name              string
	serviceID         protocol.ServiceID // The service ID to test
	archival          bool               // Whether to select a random historical block
	methods           []jsonrpc.Method   // The methods to test for this service
	serviceParams     serviceParameters  // Service-specific parameters
	latencyMultiplier int                // Multiplier for latency expectations
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
				// https://etherscan.io/address/0x28C6c06298d514Db089934071355E5743bf21d60
				contractAddress:    "0x28C6c06298d514Db089934071355E5743bf21d60",
				contractStartBlock: 12_300_000,
				transactionHash:    "0xfeccd627b5b391d04fe45055873de3b2c0b4302d52e96bd41d5f0019a704165f",
				callData:           "0x18160ddd",
			},
		},
		{
			name:      "F021 (Polygon) Load Test",
			serviceID: "F021",
			methods:   runAllMethods(),
			archival:  true, // F021 is an archival service so we should use a random historical block.
			serviceParams: serviceParameters{
				// https://polygonscan.com/address/0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270
				contractAddress:    "0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270",
				contractStartBlock: 5_000_000,
				transactionHash:    "0xb4f33e8516656d513df5d827323003c7ad1dcbb5bc46dff57c9bebad676fefe4",
				callData:           "0x18160ddd",
			},
		},
		{
			name:      "F01C (Oasys) Load Test",
			serviceID: "F01C",
			methods:   runAllMethods(),
			archival:  true, // F01C is an archival service so we should use a random historical block.
			serviceParams: serviceParameters{
				// https://explorer.oasys.games/address/0xf89d7b9c864f589bbF53a82105107622B35EaA40
				contractAddress:    "0xf89d7b9c864f589bbF53a82105107622B35EaA40",
				contractStartBlock: 424_300,
				transactionHash:    "0x7e5904f6f566577718aa3ddfe589bb6d553daaeb183e2bdc63f5bf838fede8ee",
				callData:           "0x18160ddd",
			},
		},
		{
			name:      "F036 (XRPL EVM Testnet) Load Test",
			serviceID: "F036",
			methods:   runAllMethods(),
			archival:  true, // F036 is an archival service so we should use a random historical block.
			serviceParams: serviceParameters{
				// https://explorer.testnet.xrplevm.org/address/0xc29e2583eD5C77df8792067989Baf9E4CCD4D7fc
				contractAddress:    "0xc29e2583eD5C77df8792067989Baf9E4CCD4D7fc",
				contractStartBlock: 368_266,
				transactionHash:    "0xa59fde70cac38068dfd87adb1d7eb40200421ebf7075911f83bcdde810e94058",
				callData:           "0x18160ddd",
			},
		},
	}
}

// getShannonTestCases returns test cases for Shannon protocol
func getShannonTestCases() []testCase {
	return []testCase{
		{
			name:      "anvil (local Ethereum) Load Test",
			serviceID: "anvil",
			// anvil is an ephemeral test chain so we don't test
			// `eth_getTransactionReceipt` and `eth_getTransactionByHash`
			methods: []jsonrpc.Method{
				eth_blockNumber,
				eth_call,
				eth_getBlockByNumber,
				eth_getBalance,
				eth_chainId,
				eth_getTransactionCount,
				eth_gasPrice,
			},
			serviceParams: serviceParameters{
				contractAddress: "0xdAC17F958D2ee523a2206206994597C13D831ec7",
				callData:        "0x18160ddd",
			},
			// TODO_MVP(@commoddity): This is a temporary solution to account for
			// the fact that anvil is slower due to being a test/development chain.
			latencyMultiplier: 2,
		},
	}
}

/* -------------------- EVM Load Test Function -------------------- */

// Test_PATH_E2E_EVM runs an E2E load test against the EVM JSON-RPC endpoints
func Test_PATH_E2E_EVM(t *testing.T) {
	fmt.Println("🚀 Setting up PATH instance...")

	// Config YAML file, eg. `./.morse.config.yaml` or `./.shannon.config.yaml`
	configFilePath := fmt.Sprintf(opts.configPathTemplate, opts.testProtocol)

	// Default port for PATH instance
	// If using Docker, the port will be dynamically assigned
	// and overridden by the value returned from `setupPathInstance`.
	port := "3069"

	// If GATEWAY_URL_OVERRIDE is not set, we will start an instance of PATH in Docker using `dockertest`.
	// This is configured in the file `docker_test.go` and is the default behavior.
	//
	// If GATEWAY_URL_OVERRIDE is set, we'll use the provided URL directly and skip starting a Docker container,
	// assuming PATH is already running externally at the provided URL.
	if !opts.gatewayURLOverridden {
		pathContainerPort, teardownFn := setupPathInstance(t, configFilePath, opts.docker)
		defer teardownFn()

		port = pathContainerPort

		// Format the gateway URL with the dynamically assigned port
		opts.gatewayURL = fmt.Sprintf(opts.gatewayURL, port)
	}

	fmt.Printf("🌿 Starting PATH E2E EVM test.\n")
	fmt.Printf("  🧬 Gateway URL: %s\n", opts.gatewayURL)
	fmt.Printf("  📡 Test protocol: %s\n", opts.testProtocol)

	// TODO_NEXT: This arbitrary wait is a temporary hacky solution and will be removed once PR #202 is merged:
	// 		See: https://github.com/buildwithgrove/path/pull/202
	//
	// Wait for several rounds of hydrator checks to complete to ensure invalid endpoints are sanctioned.
	// ie.for returning empty responses, etc.
	secondsToWait := 30
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		secondsToWait = secondsToWait * 2
		fmt.Printf("⏰ Waiting for %d seconds before starting tests to allow several rounds of hydrator checks to complete...\n", secondsToWait)
		<-time.After(time.Duration(secondsToWait) * time.Second) // Wait for double the default time in CI
	} else {
		fmt.Printf("⏰ Waiting for %d seconds before starting tests to allow several rounds of hydrator checks to complete...\n", secondsToWait)
		showWaitBar(secondsToWait) // In local environment, show progress bar to indicate we're waiting.
	}

	// Get test cases based on protocol
	testCases := getTestCases(opts.testProtocol)

	// Initialize map to store service summaries
	serviceSummaries := make(map[protocol.ServiceID]*serviceSummary)

	for i := range testCases {
		// If archival is true then we will use a random historical block for the test.
		if testCases[i].archival {
			testCases[i].serviceParams.blockNumber = setTestBlockNumber(
				t,
				opts.gatewayURL,
				testCases[i].serviceID,
				testCases[i].serviceParams.contractStartBlock,
			)
		} else {
			testCases[i].serviceParams.blockNumber = "latest"
		}

		fmt.Printf("🛠️  Testing service %d of %d\n", i+1, len(testCases))
		fmt.Printf("  ⛓️  Service ID: %s\n", testCases[i].serviceID)
		fmt.Printf("  📡 Block number: %s\n", testCases[i].serviceParams.blockNumber)

		// Initialize service summary
		serviceSummaries[testCases[i].serviceID] = &serviceSummary{
			serviceID:    testCases[i].serviceID,
			methodErrors: make(map[jsonrpc.Method]map[string]int),
			methodCount:  len(testCases[i].methods),
			totalErrors:  0,
		}

		// Use t.Run for proper test reporting
		serviceTestFailed := false
		t.Run(testCases[i].name, func(t *testing.T) {
			// Create results map with a mutex to protect concurrent access
			results := make(map[jsonrpc.Method]*methodMetrics)
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
						opts.gatewayURL,
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

			// Adjust latency expectations for slow chain if latency multiplier is set.
			if testCases[i].latencyMultiplier != 0 {
				fmt.Printf("%s⚠️  Adjusting latency expectations for %s by %dx to account for slower than average chain.%s\n",
					YELLOW, testCases[i].name, testCases[i].latencyMultiplier, RESET,
				)
				methodDefinitions = adjustLatencyForTestCase(methodDefinitions, testCases[i].latencyMultiplier)
			}

			// Calculate service summary metrics
			summary := serviceSummaries[testCases[i].serviceID]

			var totalLatency time.Duration
			var totalP90Latency time.Duration
			var totalSuccessRate float64
			var methodsWithResults int

			// Validate results for each method and collect summary data
			for _, method := range testCases[i].methods {
				methodMetrics := results[method]

				// Skip methods with no data
				if methodMetrics == nil || len(methodMetrics.results) == 0 {
					continue
				}

				validateResults(t, methodMetrics, methodDefinitions[method])

				// If the test has failed after validation, set the service failure flag
				if t.Failed() {
					serviceTestFailed = true
				}

				// Extract latencies for P90 calculation
				var latencies []time.Duration
				for _, res := range methodMetrics.results {
					latencies = append(latencies, res.Latency)
				}

				// Calculate P90 for this method
				p90 := calculateP90(latencies)
				avgLatency := calculateAvgLatency(latencies)

				// Add to summary totals
				totalLatency += avgLatency
				totalP90Latency += p90
				totalSuccessRate += methodMetrics.successRate
				methodsWithResults++

				// Collect errors for the summary
				if len(methodMetrics.errors) > 0 {
					// Initialize method errors map if not already created
					if summary.methodErrors[method] == nil {
						summary.methodErrors[method] = make(map[string]int)
					}

					// Copy errors to summary
					for errMsg, count := range methodMetrics.errors {
						summary.methodErrors[method][errMsg] = count
						summary.totalErrors += count
					}
				}
			}

			// Calculate averages if we have methods with results
			if methodsWithResults > 0 {
				summary.avgLatency = time.Duration(int64(totalLatency) / int64(methodsWithResults))
				summary.avgP90Latency = time.Duration(int64(totalP90Latency) / int64(methodsWithResults))
				summary.avgSuccessRate = totalSuccessRate / float64(methodsWithResults)
			}
		})

		// If this service test failed, fail the overall test immediately
		if serviceTestFailed {
			fmt.Printf("\n%s❌ TEST FAILED: Service %s failed assertions%s\n", RED, testCases[i].serviceID, RESET)

			// Print summary before failing
			printServiceSummaries(serviceSummaries)

			t.FailNow() // This will exit the test immediately
		} else {
			fmt.Printf("\n%s✅ Service %s test passed%s\n", GREEN, testCases[i].serviceID, RESET)
		}
	}

	// If execution reaches here, all services have passed
	fmt.Printf("\n%s✅ EVM E2E Test: All %d services passed%s\n", GREEN, len(testCases), RESET)

	// Print summary after all tests are complete
	printServiceSummaries(serviceSummaries)
}

// TODO_MVP(@commoddity): This is a temporary solution.
//
// adjustLatencyForTestCase increases the latency expectations by the multiplier
// for all methods in the test case to account for a slower than average chain.
func adjustLatencyForTestCase(defs map[jsonrpc.Method]methodDefinition, multiplier int) map[jsonrpc.Method]methodDefinition {
	// Create a new map to avoid modifying the original
	adjustedDefs := make(map[jsonrpc.Method]methodDefinition, len(defs))

	for method, def := range defs {
		adjustedDef := def

		adjustedDef.maxP50Latency = def.maxP50Latency * time.Duration(multiplier)
		adjustedDef.maxP95Latency = def.maxP95Latency * time.Duration(multiplier)
		adjustedDef.maxP99Latency = def.maxP99Latency * time.Duration(multiplier)

		adjustedDefs[method] = adjustedDef
	}

	return adjustedDefs
}

/* -------------------- Get Test Block Number -------------------- */

// setTestBlockNumber gets a block number for testing or fails the test
func setTestBlockNumber(
	t *testing.T,
	gatewayURL string,
	serviceID protocol.ServiceID,
	contractStartBlock uint64,
) string {
	// Get current block height - fail test if this doesn't work
	currentBlock, err := getCurrentBlockNumber(gatewayURL, serviceID)
	if err != nil {
		t.Fatalf("FATAL: Could not get current block height: %v", err)
	}

	// Get random historical block number
	return calculateArchivalBlockNumber(currentBlock, contractStartBlock)
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

func calculateArchivalBlockNumber(currentBlock, contractStartBlock uint64) string {
	var blockNumHex string
	// Case 1: Block number is below or equal to the archival threshold
	if currentBlock <= evm.DefaultEVMArchivalThreshold {
		blockNumHex = blockNumberToHex(1)
	} else {
		// Case 2: Block number is above the archival threshold
		maxBlockNumber := currentBlock - evm.DefaultEVMArchivalThreshold

		// Ensure we don't go below the minimum archival block
		if maxBlockNumber < contractStartBlock {
			blockNumHex = blockNumberToHex(contractStartBlock)
		} else {
			// Generate a random block number within valid range
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			rangeSize := maxBlockNumber - contractStartBlock + 1
			blockNumHex = blockNumberToHex(contractStartBlock + (r.Uint64() % rangeSize))
		}
	}

	return blockNumHex
}

func blockNumberToHex(blockNumber uint64) string {
	return fmt.Sprintf("0x%x", blockNumber)
}
