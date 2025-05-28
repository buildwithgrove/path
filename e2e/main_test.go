//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// -----------------------------------------------------------------------------
// Vegeta E2E & Load Tests
// -----------------------------------------------------------------------------
//
// Documentation:
//   https://path.grove.city/develop/path/e2e_tests
//
// Example Usage - E2E tests:
//   - make e2e_test_all    # Run all E2E tests for all services
//   - make e2e_test <service IDs>  # Run all E2E tests for the specified services
//
// Example Usage - Load tests:
//   - make load_test_all    # Run all load tests for all services
//   - make load_test <service IDs>  # Run all load tests for the specified services
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

// -------------------- Test Function --------------------

func Test_PATH_E2E(t *testing.T) {
	// Initialize test context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setupSIGINTHandler(cancel)

	// Get gateway URL and optional teardown function for E2E mode
	gatewayURL, teardownFn := getGatewayURLForTestMode(t, cfg)
	if teardownFn != nil {
		defer teardownFn()
	}

	// Get test cases from config based on `TEST_PROTOCOL` env var
	testServices, err := cfg.getTestServices()
	if err != nil {
		t.Fatalf("❌ Failed to get test cases: %v", err)
	}

	// Log general test information
	logTestStartInfo(gatewayURL, testServices)

	// Initialize service summaries map (will be logged out at the end of the test)
	serviceSummaries := make(map[protocol.ServiceID]*serviceSummary)

	// Loop through each test case
	for _, ts := range testServices {
		// Make a copy to avoid appending to the original URL.
		serviceGatewayURL := gatewayURL

		// If specifying the service ID in the subdomain, set the subdomain in the gateway URL.
		//
		// TODO_TECHDEBT(@commoddity): Remove this once PATH in production supports service in headers
		//   - Issue: https://github.com/buildwithgrove/infrastructure/issues/91
		if cfg.getTestMode() == testModeLoad && cfg.useServiceSubdomain() {
			serviceGatewayURL = setServiceIDInGatewayURLSubdomain(serviceGatewayURL, ts.ServiceID)
		}

		// Get methods to test
		methodsToTest := ts.getTestMethods()
		methodCount := len(methodsToTest)

		// Generate targets for this service
		targets, err := ts.getVegetaTargets(methodsToTest, serviceGatewayURL)
		if err != nil {
			t.Fatalf("❌ Failed to get vegeta targets: %v", err)
		}

		// Get test config (either use default or test case override)
		serviceConfig := cfg.DefaultServiceConfig
		if override, exists := cfg.ServiceConfigOverrides[ts.ServiceID]; exists {
			serviceConfig.applyOverride(&override)
		}

		// Create summary for this service
		serviceSummaries[ts.ServiceID] = &serviceSummary{
			serviceID:     ts.ServiceID,
			serviceConfig: serviceConfig,
			methodsToTest: methodsToTest,
			methodErrors:  make(map[jsonrpc.Method]map[string]int),
			methodCount:   methodCount,
			totalErrors:   0,
		}

		// Log service specific info
		logTestServiceInfo(ts, serviceGatewayURL, serviceConfig)

		// Run the service test
		serviceTestFailed := runServiceTest(
			t,
			ctx,
			targets,
			serviceConfig,
			methodCount,
			serviceSummaries[ts.ServiceID],
		)

		if serviceTestFailed {
			fmt.Printf("\n%s❌ TEST FAILED: Service %s failed assertions%s\n", RED, ts.ServiceID, RESET)
			printServiceSummaries(serviceSummaries)
			t.FailNow()
		} else {
			fmt.Printf("\n%s✅ Service %s test passed%s\n", GREEN, ts.ServiceID, RESET)
		}
	}

	fmt.Printf("\n%s✅ Test Success: All %d services passed%s\n", GREEN, len(testServices), RESET)
	printServiceSummaries(serviceSummaries)
}

// TODO_TECHDEBT(@commoddity): Separate E2E and Load test modes into separate files and `Test_PATH_E2E` and `Test_PATH_Load` functions.
//
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

// logTestStartInfo logs the test start information for the user.
func logTestStartInfo(gatewayURL string, testServices []TestService) {
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

	fmt.Printf("\n⛓️  Running tests for service IDs:\n")
	for _, ts := range testServices {
		if ts.Archival {
			fmt.Printf("  🔗 %s%s%s (Archival)\n", GREEN, ts.ServiceID, RESET)
		} else {
			fmt.Printf("  🔗 %s%s%s (Non-archival)\n", GREEN, ts.ServiceID, RESET)
		}
	}
}

func logTestServiceInfo(ts TestService, serviceGatewayURL string, serviceConfig ServiceConfig) {
	fmt.Printf("\n🛠️  Running test: %s%s%s\n", BOLD_BLUE, ts.Name, RESET)
	fmt.Printf("  🖥️  Service Gateway URL: %s%s%s\n", BLUE, serviceGatewayURL, RESET)
	fmt.Printf("  🏎️  Global Requests per Second: %s%d%s\n", GREEN, serviceConfig.GlobalRPS, RESET)
	fmt.Printf("  🚗 Total Requests per Method: %s%d%s\n\n", GREEN, serviceConfig.RequestsPerMethod, RESET)
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

// calculateServiceSummary validates method results, aggregates summary metrics, and updates the service summary.
func calculateServiceSummary(
	t *testing.T,
	methodConfigs map[jsonrpc.Method]ServiceConfig,
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

		// Convert ServiceConfig to methodTestConfig for validation
		methodDef := methodConfigs[method]
		serviceConfig := ServiceConfig{
			RequestsPerMethod: methodDef.RequestsPerMethod,
			GlobalRPS:         methodDef.GlobalRPS,
			SuccessRate:       methodDef.SuccessRate,
			MaxP50LatencyMS:   methodDef.MaxP50LatencyMS,
			MaxP95LatencyMS:   methodDef.MaxP95LatencyMS,
			MaxP99LatencyMS:   methodDef.MaxP99LatencyMS,
		}

		validateResults(t, methodMetrics, serviceConfig)

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
