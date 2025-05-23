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

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/evm"
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/buildwithgrove/path/request"
)

// -----------------------------------------------------------------------------
// Vegeta E2E & Load Tests
// -----------------------------------------------------------------------------
//
// Documentation:
//   https://path.grove.city/develop/path/e2e_tests
//
// Example Usage - E2E tests:
//   - make test_e2e_evm_morse    # Run all EVM E2E tests for Morse
//   - make test_e2e_evm_shannon  # Run all EVM E2E tests for Shannon
//
// Example Usage - Load tests:
//   - make test_load_evm_morse   # Run all EVM load tests for Morse
//   - make test_load_evm_shannon # Run all EVM load tests for Shannon
// -----------------------------------------------------------------------------

// -------------------- Test Configuration Initialization --------------------

// Global config for this test package.
var cfg *Config

// init initializes the test configuration.
func init() {
	var err error
	cfg, err = loadE2ELoadTestConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load E2E config: %v", err))
	}
}

// -------------------- EVM Load Test Function --------------------

// Test_PATH_E2E_EVM runs an E2E load test against the EVM JSON-RPC endpoints.
func Test_PATH_E2E_EVM(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setupSIGINTHandler(cancel)

	// Get gateway URL and optional teardown function for E2E mode
	gatewayURL, teardownFn := getGatewayURLForTestMode(t, cfg)
	if teardownFn != nil {
		defer teardownFn()
	}

	// Get test cases from config based on `TEST_PROTOCOL` env var
	testCases := cfg.getTestCases()

	// Log test information
	logEVMTestStartInfo(gatewayURL, testCases)

	// Initialize service summaries map (will be logged out at the end of the test)
	serviceSummaries := make(map[protocol.ServiceID]*serviceSummary)

	// Loop through each test case
	for _, tc := range testCases {
		// Make a copy to avoid appending to the original URL.
		serviceGatewayURL := gatewayURL

		// If specifying the service ID in the subdomain, set the subdomain in the gateway URL.
		//
		// TODO_TECHDEBT(@commoddity): Remove this once PATH in production supports service in headers
		//   - Issue: https://github.com/buildwithgrove/infrastructure/issues/91
		if cfg.getTestMode() == testModeLoad && cfg.useServiceSubdomain() {
			serviceGatewayURL = setServiceIDInGatewayURLSubdomain(serviceGatewayURL, tc.ServiceID)
		}

		// Get request headers
		headers := getRequestHeaders(tc.ServiceID)

		// Determine if test is archival
		isArchival := tc.Archival
		if !isArchival {
			tc.ServiceParams.blockNumber = "latest"
		} else {
			tc.ServiceParams.blockNumber = setTestBlockNumber(
				t,
				serviceGatewayURL,
				headers,
				tc.ServiceParams.ContractStartBlock,
			)
		}

		// Get methods to test
		methodsToTest := getMethodsToTest(tc)
		methodCount := len(methodsToTest)

		// Get test config (either use default or test case override)
		testConfig := cfg.DefaultTestConfig
		if tc.TestCaseConfigOverride != nil {
			testConfig.MergeNonZero(tc.TestCaseConfigOverride)
		}

		fmt.Printf("\n🛠️  Running EVM test: %s%s%s\n", BOLD_BLUE, tc.Name, RESET)
		fmt.Printf("  🖥️  Service Gateway URL: %s%s%s\n", BLUE, serviceGatewayURL, RESET)
		fmt.Printf("  🏎️  Global Requests per Second: %s%d%s\n", GREEN, testConfig.GlobalRPS, RESET)
		fmt.Printf("  🚗 Total Requests per Method: %s%d%s\n\n", GREEN, testConfig.RequestsPerMethod, RESET)

		// Create summary for this service
		serviceSummaries[tc.ServiceID] = &serviceSummary{
			serviceID:     tc.ServiceID,
			testConfig:    testConfig,
			methodsToTest: methodsToTest,
			methodErrors:  make(map[jsonrpc.Method]map[string]int),
			methodCount:   methodCount,
			totalErrors:   0,
		}

		// Run the service test
		serviceTestFailed := runEVMServiceTest(
			t,
			ctx,
			tc.Name,
			headers,
			tc.ServiceParams,
			testConfig,
			methodCount,
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

// getMethodsToTest determines which methods to test for a test case.
func getMethodsToTest(tc TestCase) []jsonrpc.Method {
	// If no override, use all methods
	if len(tc.TestCaseMethodOverride) == 0 {
		return allEVMTestMethods()
	}

	// Otherwise, use only the methods specified in the override
	var methodsToTest []jsonrpc.Method
	for _, method := range tc.TestCaseMethodOverride {
		methodsToTest = append(methodsToTest, jsonrpc.Method(method))
	}

	return methodsToTest
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
		fmt.Println("\n🔥 Starting Vegeta Load test ...")
	} else {
		fmt.Println("\n🌿 Starting PATH E2E test ...")
	}
	fmt.Printf("  📡 Test protocol: %s%s%s\n", BOLD_CYAN, cfg.getTestProtocol(), RESET)
	fmt.Printf("  🧬 Gateway URL: %s%s%s\n", BLUE, gatewayURL, RESET)

	if cfg.E2ELoadTestConfig.LoadTestConfig != nil {
		if cfg.E2ELoadTestConfig.LoadTestConfig.PortalApplicationID != "" {
			fmt.Printf("  🌀 Portal Application ID: %s%s%s\n", CYAN, cfg.E2ELoadTestConfig.LoadTestConfig.PortalApplicationID, RESET)
		}
		if cfg.E2ELoadTestConfig.LoadTestConfig.PortalAPIKey != "" {
			fmt.Printf("  🔑 Portal API Key: %s%s%s\n", CYAN, cfg.E2ELoadTestConfig.LoadTestConfig.PortalAPIKey, RESET)
		}
	}

	var serviceIDs []string
	for _, tc := range testCases {
		serviceIDs = append(serviceIDs, string(tc.ServiceID))
	}
	fmt.Printf("  ⛓️  Running tests for service IDs: %s%s%s\n", GREEN, strings.Join(serviceIDs, ", "), RESET)
}

// waitForHydratorIfNeeded waits for several rounds of hydrator checks if configured.
func waitForHydratorIfNeeded() {
	if waitSeconds := cfg.E2ELoadTestConfig.E2EConfig.WaitForHydrator; waitSeconds > 0 {
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
	serviceParams ServiceParams,
	testConfig TestConfig,
	methodCount int,
	gatewayURL string,
	summary *serviceSummary,
) (serviceTestFailed bool) {
	results := make(map[jsonrpc.Method]*methodMetrics)
	var resultsMutex sync.Mutex

	// Get methods to test from the summary
	methods := summary.methodsToTest

	// Create a map for progress bars and summary calculation (with the same config for all methods)
	methodConfigMap := make(map[jsonrpc.Method]TestConfig)
	for _, method := range methods {
		methodConfigMap[method] = testConfig
	}

	progBars, err := newProgressBars(methods, methodConfigMap)
	if err != nil {
		t.Fatalf("Failed to create progress bars: %v", err)
	}
	defer func() {
		if err := progBars.finish(); err != nil {
			fmt.Printf("Error stopping progress bars: %v", err)
		}
	}()

	var methodWg sync.WaitGroup
	for _, method := range methods {
		methodWg.Add(1)

		go func(ctx context.Context, method jsonrpc.Method, config TestConfig) {
			defer methodWg.Done()

			metrics := runMethodAttack(
				ctx,
				method,
				config,
				methodCount,
				headers,
				serviceParams,
				gatewayURL,
				progBars.get(method),
			)

			resultsMutex.Lock()
			results[method] = metrics
			resultsMutex.Unlock()

		}(ctx, method, testConfig)
	}
	methodWg.Wait()

	if err := progBars.finish(); err != nil {
		fmt.Printf("Error stopping progress bars: %v", err)
	}

	// We use the same methodConfigMap we created earlier for the summary calculation

	calculateServiceSummary(t, methodConfigMap, results, summary, &serviceTestFailed)
	return serviceTestFailed
}

// runMethodAttack executes the attack for a single JSON-RPC method and returns metrics.
func runMethodAttack(
	ctx context.Context,
	method jsonrpc.Method,
	methodConfig TestConfig,
	methodCount int,
	headers http.Header,
	serviceParams ServiceParams,
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
		methodCount,
		progBar,
		jsonrpcReq,
		headers,
	)

	return metrics
}

// getRequestHeaders returns the HTTP headers for a given service ID, including Portal credentials if in load test mode.
func getRequestHeaders(serviceID protocol.ServiceID) http.Header {
	headers := http.Header{
		"Content-Type":                    []string{"application/json"},
		request.HTTPHeaderTargetServiceID: []string{string(serviceID)},
	}

	if cfg.getTestMode() == testModeLoad {
		// Portal App ID is required for load tests
		headers.Set(gateway.HttpHeaderPortalAppID, cfg.E2ELoadTestConfig.LoadTestConfig.PortalApplicationID)

		// Portal API Key is optional for load tests
		if cfg.E2ELoadTestConfig.LoadTestConfig.PortalAPIKey != "" {
			headers.Set(gateway.HttpHeaderAuthorization, cfg.E2ELoadTestConfig.LoadTestConfig.PortalAPIKey)
		}
	}

	return headers
}

// calculateServiceSummary validates method results, aggregates summary metrics, and updates the service summary.
func calculateServiceSummary(
	t *testing.T,
	methodConfigs map[jsonrpc.Method]TestConfig,
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

		// Convert TestConfig to methodTestConfig for validation
		methodDef := methodConfigs[method]
		testConfig := TestConfig{
			RequestsPerMethod: methodDef.RequestsPerMethod,
			GlobalRPS:         methodDef.GlobalRPS,
			SuccessRate:       methodDef.SuccessRate,
			MaxP50LatencyMS:   methodDef.MaxP50LatencyMS,
			MaxP95LatencyMS:   methodDef.MaxP95LatencyMS,
			MaxP99LatencyMS:   methodDef.MaxP99LatencyMS,
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

// -----------------------------------------------------------------------------
// Get Test Block Number helpers
// -----------------------------------------------------------------------------

// setTestBlockNumber gets an archival block number for testing or fails the test.
// Selected by picking a random block number between the current block and the contract start block.
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

// getCurrentBlockNumber gets current block height with consensus from multiple requests.
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

// fetchBlockNumber makes a single request to get the current block number.
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

// buildBlockNumberRequest creates a JSON-RPC request for the current block number.
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

// calculateArchivalBlockNumber picks a random historical block number for archival tests.
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

// blockNumberToHex converts a block number to a hex string.
func blockNumberToHex(blockNumber uint64) string {
	return fmt.Sprintf("0x%x", blockNumber)
}

// getGatewayURLForTestMode returns the gateway URL based on the current test mode.
// It also performs any necessary setup for E2E mode (e.g., Docker container startup) and calls waitForHydratorIfNeeded.
//
// Note: If E2E mode, the caller is responsible for deferring the returned teardown function, if not nil.
func getGatewayURLForTestMode(t *testing.T, cfg *Config) (gatewayURL string, teardownFn func()) {
	switch cfg.getTestMode() {

	case testModeE2E:
		configFilePath := fmt.Sprintf("./config/.%s.config.yaml", cfg.getTestProtocol())
		var port string
		port, teardownFn = setupPathInstance(t, configFilePath, cfg.E2ELoadTestConfig.E2EConfig.DockerConfig)
		gatewayURL = fmt.Sprintf("http://localhost:%s/v1", port)
		waitForHydratorIfNeeded()
		return gatewayURL, teardownFn

	case testModeLoad:
		gatewayURL = cfg.getGatewayURLForLoadTest()
		return gatewayURL, nil

	default:
		t.Fatalf("Invalid test mode: %s", cfg.getTestMode())
		return "", nil
	}
}
