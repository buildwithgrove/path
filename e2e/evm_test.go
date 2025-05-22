//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cheggaaa/pb/v3"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/evm"
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/buildwithgrove/path/request"
)

// TODO_IN_THIS_PR(@commoddity): create separate "LOAD TEST" mode alongside "MORSE" and "SHANNON" modes to differentiate between the different test modes
//  		- LOAD TEST mode should be used for load testing the gateway using a custom override URL (no docker PATH instance required - uses either local PATH)
//  		- MORSE and SHANNON modes actually spin up PATH in Docker and run the tests against it
//
/*
For full information on the test options, see `opts_test.go`

Example Usage:
- `make test_e2e_evm_morse`                           - Run all EVM tests for Morse
- `make test_e2e_evm_shannon`                         - Run all EVM tests for Shannon
- `make test_e2e_evm_morse SERVICE_ID_OVERRIDE=F021`  - Run only the F021 EVM test for Morse
*/

// -------------------- Test Configuration Initialization --------------------

// Global config
var cfg *Config

// init initializes the test configuration
func init() {
	var err error
	cfg, err = loadE2EConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load E2E config: %v", err))
	}
}

/* -------------------- EVM Load Test Function -------------------- */

// Test_PATH_E2E_EVM runs an E2E load test against the EVM JSON-RPC endpoints
func Test_PATH_E2E_EVM(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setupSIGINTHandler(cancel)

	// Get test cases from config based on `TEST_PROTOCOL` env var
	testCases := cfg.getTestCases()

	// If running in E2E mode, start the PATH instance in Docker
	var gatewayURL string // Docker port only set if we are running in E2E mode
	if cfg.getTestMode() == testModeE2E {
		configFilePath := fmt.Sprintf("./config/.%s.config.yaml", cfg.getTestProtocol())
		pathContainerPort, teardownFn := setupPathInstance(t, configFilePath, cfg.TestConfig.E2EConfig.DockerConfig)
		defer teardownFn()

		gatewayURL = fmt.Sprintf("http://localhost:%s/v1", pathContainerPort)

		waitForHydratorIfNeeded()
	} else {
		gatewayURL = cfg.getGatewayURLForLoadTest()
	}

	// Log test information
	logEVMTestStartInfo(gatewayURL, testCases)

	serviceSummaries := make(map[protocol.ServiceID]*serviceSummary)

	for _, tc := range testCases {
		fmt.Printf("\n🛠️  Running EVM test: %s%s%s\n\n", BOLD_BLUE, tc.Name, RESET)

		serviceGatewayURL := gatewayURL // Make a copy to avoid appending to the original

		// If specifying the service ID in the subdomain, set the subdomain in the gateway URL.
		//
		// TODO_TECHDEBT(@commoddity): Remove this once PATH in production supports service in headers
		//   - Issue: https://github.com/buildwithgrove/infrastructure/issues/91
		if cfg.useServiceSubdomain() {
			serviceGatewayURL = setServiceIDInGatewayURLSubdomain(serviceGatewayURL, tc.ServiceID)
		}

		// Set up service parameters
		serviceParams := evmServiceParameters{
			contractAddress:    tc.ServiceParams.ContractAddress,
			contractStartBlock: tc.ServiceParams.ContractStartBlock,
			transactionHash:    tc.ServiceParams.TransactionHash,
			callData:           tc.ServiceParams.CallData,
		}

		// Get request headers
		headers := getRequestHeaders(tc.ServiceID)

		// Determine if test is archival
		isArchival := tc.Archival
		if !isArchival {
			serviceParams.blockNumber = "latest"
		} else {
			serviceParams.blockNumber = setTestBlockNumber(
				t,
				serviceGatewayURL,
				headers,
				serviceParams.contractStartBlock,
			)
		}

		// Create method configs using default configuration in the config file
		methodConfigs := createMethodConfigs(tc)

		// Create summary for this service
		serviceSummaries[tc.ServiceID] = &serviceSummary{
			serviceID:     tc.ServiceID,
			methodConfigs: methodConfigs,
			methodErrors:  make(map[jsonrpc.Method]map[string]int),
			methodCount:   len(methodConfigs),
			totalErrors:   0,
		}

		// Run the service test
		serviceTestFailed := runEVMServiceTest(
			t,
			ctx,
			tc.Name,
			headers,
			serviceParams,
			tc.LatencyMultiplier,
			methodConfigs,
			serviceGatewayURL,
			serviceSummaries[tc.ServiceID],
		)

		if serviceTestFailed {
			fmt.Printf("\n%s❌ TEST FAILED: Service %s failed assertions%s\n", RED, tc.ServiceID, RESET)
			printServiceSummaries(serviceSummaries)
			t.FailNow()
		} else {
			fmt.Printf("\n%s✅ Service %s test passed%s\n", GREEN, tc.ServiceID, RESET)
		}
	}

	fmt.Printf("\n%s✅ EVM E2E Test: All %d services passed%s\n", GREEN, len(testCases), RESET)
	printServiceSummaries(serviceSummaries)
}

// createMethodConfigs creates the method configurations for a test case
func createMethodConfigs(tc TestCase) map[jsonrpc.Method]MethodConfig {
	// Initialize the method config map
	methodConfigs := make(map[jsonrpc.Method]MethodConfig)

	// Get the list of methods to test
	var methodsToTest []jsonrpc.Method
	if len(tc.TestCaseMethodOverride) > 0 {
		// Use the specified methods
		for _, method := range tc.TestCaseMethodOverride {
			methodsToTest = append(methodsToTest, jsonrpc.Method(method))
		}
	} else {
		methodsToTest = allEVMTestMethods()
	}

	// Create config for each method
	for _, method := range methodsToTest {
		// Start with the default method config
		methodConfig := cfg.DefaultMethodConfig

		// Override with test case specific config if provided
		if tc.TestCaseConfigOverride != nil {
			// Apply all overrides from test case
			if tc.TestCaseConfigOverride.TotalRequests != 0 {
				methodConfig.TotalRequests = tc.TestCaseConfigOverride.TotalRequests
			}
			if tc.TestCaseConfigOverride.RPS != 0 {
				methodConfig.RPS = tc.TestCaseConfigOverride.RPS
			}
			if tc.TestCaseConfigOverride.SuccessRate != 0 {
				methodConfig.SuccessRate = tc.TestCaseConfigOverride.SuccessRate
			}
			if tc.TestCaseConfigOverride.MaxP50LatencyMS != 0 {
				methodConfig.MaxP50LatencyMS = tc.TestCaseConfigOverride.MaxP50LatencyMS
			}
			if tc.TestCaseConfigOverride.MaxP95LatencyMS != 0 {
				methodConfig.MaxP95LatencyMS = tc.TestCaseConfigOverride.MaxP95LatencyMS
			}
			if tc.TestCaseConfigOverride.MaxP99LatencyMS != 0 {
				methodConfig.MaxP99LatencyMS = tc.TestCaseConfigOverride.MaxP99LatencyMS
			}
		}

		// Add to the map
		methodConfigs[method] = methodConfig
	}

	return methodConfigs
}

// setupSIGINTHandler sets up a signal handler for SIGINT to cancel the test context.
func setupSIGINTHandler(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		fmt.Println("🛑 Received SIGINT, cancelling test...")
		cancel()

		// Give a short time for cleanup to happen in the other handlers
		// but don't hang forever
		timer := time.NewTimer(5 * time.Second)
		select {
		case <-timer.C:
			fmt.Println("Cleanup timed out, forcing exit...")
			os.Exit(1)
		}
	}()
}

// logEVMTestStartInfo logs the test start information for the user.
func logEVMTestStartInfo(gatewayURL string, testCases []TestCase) {
	if cfg.getTestMode() == testModeLoad {
		fmt.Println("\n🔥 Starting Vegeta load test ...")
	} else {
		fmt.Println("\n🌿 Starting PATH E2E test ...")
	}
	fmt.Printf("  🧬 Gateway URL: %s\n", gatewayURL)

	if cfg.TestConfig.LoadTestConfig != nil {
		if cfg.TestConfig.LoadTestConfig.PortalApplicationIDOverride != "" {
			fmt.Printf("  🌀 Portal Application ID: %s\n", cfg.TestConfig.LoadTestConfig.PortalApplicationIDOverride)
		}
		if cfg.TestConfig.LoadTestConfig.PortalAPIKeyOverride != "" {
			fmt.Printf("  🔑 Portal API Key: %s\n", cfg.TestConfig.LoadTestConfig.PortalAPIKeyOverride)
		}
	}

	fmt.Printf("  📡 Test protocol: %s\n", cfg.getTestProtocol())
	var serviceIDs []string
	for _, tc := range testCases {
		serviceIDs = append(serviceIDs, string(tc.ServiceID))
	}
	fmt.Printf("  ⛓️  Running tests for all service IDs: %s\n", strings.Join(serviceIDs, ", "))
}

// waitForHydratorIfNeeded waits for several rounds of hydrator checks if configured.
func waitForHydratorIfNeeded() {
	if waitSeconds := cfg.TestConfig.E2EConfig.WaitForHydrator; waitSeconds > 0 {
		fmt.Printf("\n⏰ Waiting for %d seconds before starting tests to allow several rounds of hydrator checks to complete...\n",
			waitSeconds,
		)
		if isCIEnv() {
			<-time.After(time.Duration(waitSeconds) * time.Second)
		} else {
			showWaitBar(waitSeconds)
		}
	}
}

// runEVMServiceTest runs the E2E test for a single EVM service in a test case.
func runEVMServiceTest(
	t *testing.T,
	ctx context.Context,
	testName string,
	headers http.Header,
	serviceParams evmServiceParameters,
	latencyMultiplier int,
	methodConfigs map[jsonrpc.Method]MethodConfig,
	gatewayURL string,
	summary *serviceSummary,
) (serviceTestFailed bool) {
	results := make(map[jsonrpc.Method]*methodMetrics)
	var resultsMutex sync.Mutex

	// Validate that all methods have a definition
	for method := range methodConfigs {
		if _, exists := methodConfigs[method]; !exists {
			t.Fatalf("No definition for method %s", method)
		}
	}

	var methods []jsonrpc.Method
	for method := range methodConfigs {
		methods = append(methods, method)
	}

	progBars, err := newProgressBars(methods, methodConfigs)
	if err != nil {
		t.Fatalf("Failed to create progress bars: %v", err)
	}
	defer func() {
		if err := progBars.finish(); err != nil {
			fmt.Printf("Error stopping progress bars: %v", err)
		}
	}()

	var methodWg sync.WaitGroup
	for method := range methodConfigs {
		methodWg.Add(1)

		methodDef := methodConfigs[method]

		go func(ctx context.Context, method jsonrpc.Method, methodDef MethodConfig) {
			defer methodWg.Done()

			metrics := runMethodAttack(
				ctx,
				method,
				methodDef,
				headers,
				serviceParams,
				gatewayURL,
				progBars.get(method),
			)

			resultsMutex.Lock()
			results[method] = metrics
			resultsMutex.Unlock()

		}(ctx, method, methodDef)
	}
	methodWg.Wait()

	if err := progBars.finish(); err != nil {
		fmt.Printf("Error stopping progress bars: %v", err)
	}

	if latencyMultiplier != 0 {
		fmt.Printf("%s⚠️  Adjusting latency expectations for %s by %dx to account for slower than average chain.%s ⚠️\n",
			YELLOW, testName, latencyMultiplier, RESET,
		)
		methodConfigs = adjustLatencyForTestCase(methodConfigs, latencyMultiplier)
	}

	calculateServiceSummary(t, methodConfigs, results, summary, &serviceTestFailed)
	return serviceTestFailed
}

// runMethodAttack executes the attack for a single JSON-RPC method and returns metrics.
func runMethodAttack(
	ctx context.Context,
	method jsonrpc.Method,
	methodConfig MethodConfig,
	headers http.Header,
	serviceParams evmServiceParameters,
	gatewayURL string,
	progBar *pb.ProgressBar,
) *methodMetrics {
	select {
	case <-ctx.Done():
		fmt.Printf("Method %s cancelled", method)
		return nil
	default:
	}

	jsonrpcReq := jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(1),
		Method:  method,
		Params:  createEVMJsonRPCParams(method, serviceParams),
	}

	metrics := runAttack(
		ctx,
		gatewayURL,
		method,
		methodConfig,
		progBar,
		jsonrpcReq,
		headers,
	)

	return metrics
}

func getRequestHeaders(serviceID protocol.ServiceID) http.Header {
	headers := http.Header{
		"Content-Type":                    []string{"application/json"},
		request.HTTPHeaderTargetServiceID: []string{string(serviceID)},
	}

	if cfg.getTestMode() == testModeLoad {
		// Portal App ID is required for load tests
		headers.Set("Portal-Application-ID", cfg.TestConfig.LoadTestConfig.PortalApplicationIDOverride)
		// Portal API Key is optional for load tests
		if cfg.TestConfig.LoadTestConfig.PortalAPIKeyOverride != "" {
			headers.Set("Authorization", cfg.TestConfig.LoadTestConfig.PortalAPIKeyOverride)
		}
	}

	return headers
}

// calculateServiceSummary validates method results, aggregates summary metrics, and updates the service summary.
func calculateServiceSummary(
	t *testing.T,
	methodConfigs map[jsonrpc.Method]MethodConfig,
	results map[jsonrpc.Method]*methodMetrics,
	summary *serviceSummary,
	serviceTestFailed *bool,
) {
	var totalLatency time.Duration
	var totalP90Latency time.Duration
	var totalSuccessRate float64
	var methodsWithResults int

	// Track service totals
	summary.totalRequests = 0
	summary.totalSuccess = 0
	summary.totalFailure = 0

	// Validate results for each method and collect summary data
	for method := range methodConfigs {
		methodMetrics := results[method]

		// Skip methods with no data
		if methodMetrics == nil || len(methodMetrics.results) == 0 {
			continue
		}

		// Convert MethodConfig to methodTestConfig for validation
		methodDef := methodConfigs[method]
		testConfig := MethodConfig{
			TotalRequests:   methodDef.TotalRequests,
			RPS:             methodDef.RPS,
			SuccessRate:     methodDef.SuccessRate,
			MaxP50LatencyMS: methodDef.MaxP50LatencyMS,
			MaxP95LatencyMS: methodDef.MaxP95LatencyMS,
			MaxP99LatencyMS: methodDef.MaxP99LatencyMS,
		}

		validateResults(t, methodMetrics, testConfig)

		// If the test has failed after validation, set the service failure flag
		if t.Failed() {
			*serviceTestFailed = true
		}

		// Accumulate totals for the service summary
		summary.totalRequests += methodMetrics.requestCount
		summary.totalSuccess += methodMetrics.success
		summary.totalFailure += methodMetrics.failed

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
}

// adjustLatencyForTestCase increases the latency expectations by the multiplier
// for all methods in the test case to account for a slower than average service providers (e.g. dev/test environments)
func adjustLatencyForTestCase(
	testConfig map[jsonrpc.Method]MethodConfig,
	latencyMultiplier int,
) map[jsonrpc.Method]MethodConfig {
	// Create a new map to avoid modifying the original test config
	adjustedDefs := make(map[jsonrpc.Method]MethodConfig, len(testConfig))

	for method, methodDef := range testConfig {
		adjustedDef := methodDef
		adjustedDef.MaxP50LatencyMS = methodDef.MaxP50LatencyMS * time.Duration(latencyMultiplier)
		adjustedDef.MaxP95LatencyMS = methodDef.MaxP95LatencyMS * time.Duration(latencyMultiplier)
		adjustedDef.MaxP99LatencyMS = methodDef.MaxP99LatencyMS * time.Duration(latencyMultiplier)
		adjustedDefs[method] = adjustedDef
	}

	return adjustedDefs
}

/* -------------------- Get Test Block Number -------------------- */

// setTestBlockNumber gets a block number for testing or fails the test
func setTestBlockNumber(
	t *testing.T,
	gatewayURL string,
	headers http.Header,
	contractStartBlock uint64,
) string {
	// Get current block height - fail test if this doesn't work
	currentBlock, err := getCurrentBlockNumber(gatewayURL, headers)
	if err != nil {
		t.Fatalf("FATAL: Could not get current block height: %v", err)
	}

	// Get random historical block number
	return calculateArchivalBlockNumber(currentBlock, contractStartBlock)
}

// getCurrentBlockNumber gets current block height with consensus from multiple requests
func getCurrentBlockNumber(gatewayURL string, headers http.Header) (uint64, error) {
	// Track frequency of each block height seen
	blockHeights := make(map[uint64]int)
	maxAttempts := 10
	requiredAgreement := 3
	client := &http.Client{Timeout: 5 * time.Second}

	// Make multiple attempts to get consensus
	for range maxAttempts {
		blockNum, err := fetchBlockNumber(client, gatewayURL, headers)
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
func fetchBlockNumber(client *http.Client, gatewayURL string, headers http.Header) (uint64, error) {
	// Build and send request
	req, err := buildBlockNumberRequest(gatewayURL, headers)
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
func buildBlockNumberRequest(gatewayURL string, headers http.Header) (*http.Request, error) {
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

	req.Header = headers.Clone()

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
