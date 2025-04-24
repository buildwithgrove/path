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

/*
For full information on the test options, see `opts_test.go`

Example Usage:
- `make test_e2e_evm_morse`                           - Run all EVM tests for Morse
- `make test_e2e_evm_shannon`                         - Run all EVM tests for Shannon
- `make test_e2e_evm_morse SERVICE_ID_OVERRIDE=F021`  - Run only the F021 EVM test for Morse
- `make test_e2e_evm_morse DOCKER_FORCE_REBUILD=true` - Force a rebuild of the Docker image for the EVM tests
- `make test_e2e_evm_morse DOCKER_LOG=true`           - Log the output of the Docker container for the EVM tests
- `make test_e2e_evm_morse WAIT_FOR_HYDRATOR=30`      - Wait for 30 seconds before starting tests to allow several rounds of hydrator checks to complete.
*/

// -------------------- Test Configuration Initialization --------------------

// Global test options
var opts testOptions

// init initializes the test options
func init() {
	opts = gatherTestOptions()
}

// -------------------- Get Test Cases for Protocol --------------------

// testCase represents a single service load test configuration
//
// Fields:
// - name:              Descriptive name for the test case
// - serviceID:         The service ID to test
// - archival:          Whether to select a random historical block
// - methods:           The methods to test for this service
// - serviceParams:     Service-specific parameters
// - latencyMultiplier: Multiplier for latency expectations
type testCase struct {
	name          string
	serviceID     protocol.ServiceID
	archival      bool
	methods       []jsonrpc.Method
	serviceParams evmServiceParameters
	// latencyMultiplier is particularly important for dev/test chains that are slower than mainnet.
	// For integration tests, we need complete reliability and avoid false positives.
	latencyMultiplier int
	methodConfigs     map[jsonrpc.Method]methodTestConfig
}

// getTestCases returns the appropriate test cases based on the protocol.
//
// - Filters for a specific service ID if provided.
// - Panics if the service ID override is not found.
func getTestCases(t *testing.T, protocolStr protocolStr, serviceIDOverride protocol.ServiceID) []testCase {
	var testCases []testCase

	// Select test cases based on protocol
	switch protocolStr {
	case morse:
		testCases = morseTestCases
	case shannon:
		testCases = shannonTestCases
	default:
		t.Fatalf("Unsupported protocol: %s", protocolStr)
	}

	// Filter by serviceIDOverride if provided
	if serviceIDOverride != "" {
		for _, tc := range testCases {
			if tc.serviceID == serviceIDOverride {
				// Return single matching test case in a slice
				return []testCase{tc}
			}
		}
		panic(fmt.Sprintf("Service ID override %s not found", serviceIDOverride))
	}

	return testCases
}

// Shannon network test cases
var (
	shannonTestCases = []testCase{
		{
			name:      "anvil (local Ethereum) Load Test",
			serviceID: "anvil",
			methods:   shannonBetaTestNetMethods,
			serviceParams: evmServiceParameters{
				contractAddress: "0xdAC17F958D2ee523a2206206994597C13D831ec7",
				callData:        "0x18160ddd",
			},
			latencyMultiplier: 10,
			methodConfigs:     shannonBetaTestNetMethodConfigs,
		},
	}

	shannonBetaTestNetMethods = []jsonrpc.Method{
		eth_blockNumber,
		eth_call,
		eth_getBlockByNumber,
		eth_getBalance,
		eth_chainId,
		eth_getTransactionCount,
		eth_gasPrice,
	}

	// TODO_TECHDEBT: Iterate on these tests to make sure the anvil node can handle more load.

	// defaultRequestLoadConfig contains the default configuration for a method.
	shannonBetaTestNetRequestLoadConfig = requestLoadConfig{
		totalRequests: 3,
		rps:           1,
	}

	// defaultSuccessCriteria contains the default success rates and latency requirements for a method.
	shannonBetaTestNetSuccessCriteria = successCriteria{
		successRate:   0.75,
		maxP50Latency: 5_000 * time.Millisecond,  // 5 seconds
		maxP95Latency: 10_000 * time.Millisecond, // 10 seconds
		maxP99Latency: 20_000 * time.Millisecond, // 30 seconds
	}

	shannonBetaTestNetMethodConfigs = map[jsonrpc.Method]methodTestConfig{
		eth_blockNumber: {
			requestLoadConfig: shannonBetaTestNetRequestLoadConfig,
			successCriteria:   shannonBetaTestNetSuccessCriteria,
		},
		eth_call: {
			requestLoadConfig: shannonBetaTestNetRequestLoadConfig,
			successCriteria:   shannonBetaTestNetSuccessCriteria,
		},
		eth_getBlockByNumber: {
			requestLoadConfig: shannonBetaTestNetRequestLoadConfig,
			successCriteria:   shannonBetaTestNetSuccessCriteria,
		},
		eth_getBalance: {
			requestLoadConfig: shannonBetaTestNetRequestLoadConfig,
			successCriteria:   shannonBetaTestNetSuccessCriteria,
		},
		eth_chainId: {
			requestLoadConfig: shannonBetaTestNetRequestLoadConfig,
			successCriteria:   shannonBetaTestNetSuccessCriteria,
		},
		eth_getTransactionCount: {
			requestLoadConfig: shannonBetaTestNetRequestLoadConfig,
			successCriteria:   shannonBetaTestNetSuccessCriteria,
		},
		eth_gasPrice: {
			requestLoadConfig: shannonBetaTestNetRequestLoadConfig,
			successCriteria:   shannonBetaTestNetSuccessCriteria,
		},
	}
)

// Morse network test cases
var morseTestCases = []testCase{
	{
		name:      "F00C (Ethereum) Load Test",
		serviceID: "F00C",
		methods:   allEVMTestMethods(),
		archival:  true, // Use random historical block for archival service
		serviceParams: evmServiceParameters{
			// https://etherscan.io/address/0x28C6c06298d514Db089934071355E5743bf21d60
			contractAddress:    "0x28C6c06298d514Db089934071355E5743bf21d60",
			contractStartBlock: 12_300_000,
			transactionHash:    "0xfeccd627b5b391d04fe45055873de3b2c0b4302d52e96bd41d5f0019a704165f",
			callData:           "0x18160ddd",
		},
		methodConfigs: defaultTestConfigAllMethods,
	},
	{
		name:      "F021 (Polygon) Load Test",
		serviceID: "F021",
		methods:   allEVMTestMethods(),
		archival:  true, // Use random historical block for archival service
		serviceParams: evmServiceParameters{
			// https://polygonscan.com/address/0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270
			contractAddress:    "0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270",
			contractStartBlock: 5_000_000,
			transactionHash:    "0xb4f33e8516656d513df5d827323003c7ad1dcbb5bc46dff57c9bebad676fefe4",
			callData:           "0x18160ddd",
		},
		methodConfigs: defaultTestConfigAllMethods,
	},
	{
		name:      "F01C (Oasys) Load Test",
		serviceID: "F01C",
		methods:   allEVMTestMethods(),
		archival:  true, // Use random historical block for archival service
		serviceParams: evmServiceParameters{
			// https://explorer.oasys.games/address/0xf89d7b9c864f589bbF53a82105107622B35EaA40
			contractAddress:    "0xf89d7b9c864f589bbF53a82105107622B35EaA40",
			contractStartBlock: 424_300,
			transactionHash:    "0x7e5904f6f566577718aa3ddfe589bb6d553daaeb183e2bdc63f5bf838fede8ee",
			callData:           "0x18160ddd",
		},
		methodConfigs: defaultTestConfigAllMethods,
	},
	{
		name:      "F036 (XRPL EVM Testnet) Load Test",
		serviceID: "F036",
		methods:   allEVMTestMethods(),
		archival:  true, // Use random historical block for archival service
		serviceParams: evmServiceParameters{
			// https://explorer.testnet.xrplevm.org/address/0xc29e2583eD5C77df8792067989Baf9E4CCD4D7fc
			contractAddress:    "0xc29e2583eD5C77df8792067989Baf9E4CCD4D7fc",
			contractStartBlock: 368_266,
			transactionHash:    "0xa59fde70cac38068dfd87adb1d7eb40200421ebf7075911f83bcdde810e94058",
			callData:           "0x18160ddd",
		},
		methodConfigs: defaultTestConfigAllMethods,
	},
}

/* -------------------- EVM Load Test Function -------------------- */

// Test_PATH_E2E_EVM runs an E2E load test against the EVM JSON-RPC endpoints
func Test_PATH_E2E_EVM(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setupSIGINTHandler(ctx, cancel, t)

	configFilePath := fmt.Sprintf(opts.configPathTemplate, opts.testProtocol)
	if !opts.gatewayURLOverridden {
		pathContainerPort, teardownFn := setupPathInstance(t, configFilePath, opts.docker)
		defer teardownFn()
		opts.gatewayURL = fmt.Sprintf(opts.gatewayURL, pathContainerPort)
	}

	logEVMTestStartInfo(t, opts)
	waitForHydratorIfNeeded(opts, t)

	testCases := getTestCases(t, opts.testProtocol, opts.serviceIDOverride)
	serviceSummaries := make(map[protocol.ServiceID]*serviceSummary)

	for _, tc := range testCases {
		isArchival := tc.archival
		if !isArchival {
			tc.serviceParams.blockNumber = "latest"
		} else {
			tc.serviceParams.blockNumber = setTestBlockNumber(
				t,
				opts.gatewayURL,
				tc.serviceID,
				tc.serviceParams.contractStartBlock,
			)
		}

		serviceSummaries[tc.serviceID] = &serviceSummary{
			serviceID:    tc.serviceID,
			methodErrors: make(map[jsonrpc.Method]map[string]int),
			methodCount:  len(tc.methods),
			totalErrors:  0,
		}

		serviceTestFailed := runEVMServiceTest(t, ctx, tc, opts, serviceSummaries[tc.serviceID])
		if serviceTestFailed {
			fmt.Printf("\n%s❌ TEST FAILED: Service %s failed assertions%s\n", RED, tc.serviceID, RESET)
			printServiceSummaries(serviceSummaries)
			t.FailNow()
		} else {
			fmt.Printf("\n%s✅ Service %s test passed%s\n", GREEN, tc.serviceID, RESET)
		}
	}

	fmt.Printf("\n%s✅ EVM E2E Test: All %d services passed%s\n", GREEN, len(testCases), RESET)
	printServiceSummaries(serviceSummaries)
}

// setupSIGINTHandler sets up a signal handler for SIGINT to cancel the test context.
func setupSIGINTHandler(ctx context.Context, cancel context.CancelFunc, t *testing.T) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		fmt.Print("Received SIGINT, cancelling test...")
		cancel()
	}()
}

// logEVMTestStartInfo logs the test start information for the user.
func logEVMTestStartInfo(t *testing.T, opts testOptions) {
	fmt.Printf("\n 🌿 Starting PATH E2E EVM test.\n")
	fmt.Printf(" 🧬 Gateway URL: %s\n", opts.gatewayURL)
	fmt.Printf(" 📡 Test protocol: %s\n", opts.testProtocol)
	if opts.serviceIDOverride != "" {
		fmt.Printf(" ⛓️  Running tests for service ID: %s\n", opts.serviceIDOverride)
	} else {
		fmt.Printf(" ⛓️  Running tests for all service IDs\n\n")
	}
}

// waitForHydratorIfNeeded waits for several rounds of hydrator checks if configured.
func waitForHydratorIfNeeded(opts testOptions, t *testing.T) {
	if opts.waitForHydrator > 0 {
		fmt.Printf("⏰ Waiting for %d seconds before starting tests to allow several rounds of hydrator checks to complete...\n", opts.waitForHydrator)
		if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
			<-time.After(time.Duration(opts.waitForHydrator) * time.Second)
		} else {
			showWaitBar(opts.waitForHydrator)
		}
	}
}

// runEVMServiceTest runs the E2E test for a single EVM service in a test case.
func runEVMServiceTest(
	t *testing.T,
	ctx context.Context,
	tc testCase,
	opts testOptions,
	summary *serviceSummary,
) (serviceTestFailed bool) {
	results := make(map[jsonrpc.Method]*methodMetrics)
	var resultsMutex sync.Mutex

	// Validate that all methods have a definition
	for _, method := range tc.methods {
		if _, exists := tc.methodConfigs[method]; !exists {
			t.Fatalf("No definition for method %s", method)
		}
	}

	progBars, err := newProgressBars(tc.methods, tc.methodConfigs)
	if err != nil {
		t.Fatalf("Failed to create progress bars: %v", err)
	}
	defer func() {
		if err := progBars.finish(); err != nil {
			fmt.Printf("Error stopping progress bars: %v", err)
		}
	}()

	var methodWg sync.WaitGroup
	for _, method := range tc.methods {
		methodWg.Add(1)
		methodDef := tc.methodConfigs[method]
		go func(ctx context.Context, method jsonrpc.Method, def methodTestConfig) {
			defer methodWg.Done()
			metrics := runMethodAttack(ctx, t, method, def, tc, opts, progBars.get(method))
			resultsMutex.Lock()
			results[method] = metrics
			resultsMutex.Unlock()
		}(ctx, method, methodDef)
	}
	methodWg.Wait()

	if err := progBars.finish(); err != nil {
		fmt.Printf("Error stopping progress bars: %v", err)
	}
	fmt.Println()

	if tc.latencyMultiplier != 0 {
		fmt.Printf("%s⚠️  Adjusting latency expectations for %s by %dx to account for slower than average chain.%s ⚠️\n",
			YELLOW, tc.name, tc.latencyMultiplier, RESET,
		)
		tc.methodConfigs = adjustLatencyForTestCase(tc.methodConfigs, tc.latencyMultiplier)
	}

	calculateServiceSummary(t, tc, results, summary, &serviceTestFailed)
	return serviceTestFailed
}

// runMethodAttack executes the attack for a single JSON-RPC method and returns metrics.
func runMethodAttack(
	ctx context.Context,
	t *testing.T,
	method jsonrpc.Method,
	def methodTestConfig,
	tc testCase,
	opts testOptions,
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
		Params: createEVMJsonRPCParams(
			method,
			tc.serviceParams,
		),
	}
	metrics := runAttack(
		ctx,
		opts.gatewayURL,
		tc.serviceID,
		method,
		def,
		progBar,
		jsonrpcReq,
	)
	return metrics
}

// calculateServiceSummary validates method results, aggregates summary metrics, and updates the service summary.
func calculateServiceSummary(
	t *testing.T,
	tc testCase,
	results map[jsonrpc.Method]*methodMetrics,
	summary *serviceSummary,
	serviceTestFailed *bool,
) {
	var totalLatency time.Duration
	var totalP90Latency time.Duration
	var totalSuccessRate float64
	var methodsWithResults int

	// Validate results for each method and collect summary data
	for _, method := range tc.methods {
		methodMetrics := results[method]

		// Skip methods with no data
		if methodMetrics == nil || len(methodMetrics.results) == 0 {
			continue
		}

		validateResults(t, methodMetrics, tc.methodConfigs[method])

		// If the test has failed after validation, set the service failure flag
		if t.Failed() {
			*serviceTestFailed = true
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
}

// adjustLatencyForTestCase increases the latency expectations by the multiplier
// for all methods in the test case to account for a slower than average service providers (e.g. dev/test environments)
func adjustLatencyForTestCase(
	testConfig map[jsonrpc.Method]methodTestConfig,
	latencyMultiplier int,
) map[jsonrpc.Method]methodTestConfig {
	// Create a new map to avoid modifying the original test config
	adjustedDefs := make(map[jsonrpc.Method]methodTestConfig, len(testConfig))

	for method, def := range testConfig {
		adjustedDef := def
		adjustedDef.maxP50Latency = def.maxP50Latency * time.Duration(latencyMultiplier)
		adjustedDef.maxP95Latency = def.maxP95Latency * time.Duration(latencyMultiplier)
		adjustedDef.maxP99Latency = def.maxP99Latency * time.Duration(latencyMultiplier)
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
