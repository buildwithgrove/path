//go:build e2e

// Package e2e provides comprehensive End-to-End and Load testing for PATH services.
//
// This package implements a flexible testing framework that supports both HTTP and WebSocket
// protocols, with configurable load testing capabilities using Vegeta and custom WebSocket clients.
//
// PACKAGE ARCHITECTURE:
// - main_test.go: Test orchestration, configuration, and coordination between HTTP/WebSocket tests
// - vegeta_test.go: HTTP load testing using Vegeta library with concurrent request execution
// - websockets_test.go: WebSocket testing using single persistent connections for EVM JSON-RPC
// - assertions_test.go: Shared validation logic for JSON-RPC responses (transport-agnostic)
// - calculations_test.go: Metrics calculation functions for success rates and latency percentiles
// - log_test.go: Progress bars, ANSI colors, and formatted logging utilities
// - config_test.go: Configuration loading, environment variable parsing, and service setup
// - service_test.go: Service definitions, target generation, and protocol-specific implementations
// - service_*_test.go: Protocol-specific request builders (EVM, Cosmos SDK, Solana, Anvil)
// - docker_test.go: Local PATH instance management for E2E testing
//
// TESTING MODES:
// - E2E Mode: Spins up local PATH instance, tests against localhost
// - Load Mode: Tests against remote PATH deployment with configurable RPS and request volumes
// - WebSocket-only Mode: Tests only WebSocket-compatible services using persistent connections
//
// SUPPORTED PROTOCOLS:
// - EVM JSON-RPC (HTTP + WebSocket): Ethereum-compatible blockchain interactions
// - Cosmos SDK REST: RESTful API endpoints for Cosmos-based chains
// - CometBFT JSON-RPC: Tendermint consensus and node status endpoints
// - Solana JSON-RPC: Solana-specific blockchain methods
package e2e

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"testing"
	"time"

	"github.com/buildwithgrove/path/protocol"
)

// -----------------------------------------------------------------------------
// Vegeta E2E & Load Tests
// -----------------------------------------------------------------------------
//
// Documentation:
//   https://path.grove.city/develop/path/e2e_tests
//
// Example Usage - E2E tests:
//   - make e2e_test_all             # Run all HTTP E2E tests for all services
//   - make e2e_test <service IDs>   # Run HTTP E2E tests for the specified services
//   - make e2e_test_websocket_all   # Run WebSocket E2E tests for all compatible services
//   - make e2e_test_websocket <IDs> # Run WebSocket E2E tests for specified services
//   - make e2e_test_eth_fallback <URL> # Run E2E test with ETH fallback URL enabled
//
// Example Usage - Load tests:
//   - make load_test_all            # Run all HTTP load tests for all services
//   - make load_test <service IDs>  # Run all HTTP load tests for the specified services
//   - make load_test_websocket_all  # Run all WebSocket load tests for all compatible services
//   - make load_test_websocket <IDs> # Run all WebSocket load tests for specified services
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
	logTestStartInfo(gatewayURL)

	// Initialize service summaries map (will be logged out at the end of the test)
	serviceSummaries := make(map[protocol.ServiceID]*serviceSummary)

	// Assign test service configs to each test service
	testServiceConfigs := setTestServiceConfigs(testServices)

	// Filter test services for WebSocket-only mode if enabled
	if cfg.isWebSocketsOnly() {
		filteredServices := make([]*TestService, 0)
		for _, ts := range testServices {
			if ts.supportsEVMWebSockets() {
				filteredServices = append(filteredServices, ts)
			}
		}
		testServices = filteredServices

		if len(testServices) == 0 {
			fmt.Printf("%s⚠️  No services support WebSocket tests in WebSocket-only mode%s\n", YELLOW, RESET)
			return
		}

		fmt.Printf("\n%s🔌 WebSocket-only mode enabled - filtered to %d WebSocket-compatible service(s)%s\n",
			BOLD_CYAN, len(testServices), RESET)
	}

	// Log the test service IDs
	logTestServiceIDs(testServices)

	for _, ts := range testServices {
		serviceConfig, exists := testServiceConfigs[ts.ServiceID]
		if !exists {
			t.Fatalf("❌ Failed to get test service config for service ID: %s", ts.ServiceID)
		}

		// Make a copy to avoid appending to the original URL.
		serviceGatewayURL := gatewayURL

		// If specifying the service ID in the subdomain, set the subdomain in the gateway URL.
		//
		// TODO_TECHDEBT(@commoddity): Remove this once PATH in production supports service in headers
		//   - Issue: https://github.com/buildwithgrove/infrastructure/issues/91
		if cfg.getTestMode() == testModeLoad && cfg.useServiceSubdomain() {
			serviceGatewayURL = setServiceIDInGatewayURLSubdomain(
				serviceGatewayURL,
				ts.ServiceID,
				ts.Alias,
			)
		}

		// Generate targets for this service
		targets, err := ts.getVegetaTargets(serviceGatewayURL)
		if err != nil {
			t.Fatalf("❌ Failed to get vegeta targets: %v", err)
		}

		// Create summary for this service
		serviceSummaries[ts.ServiceID] = newServiceSummary(ts.ServiceID, serviceConfig, ts.testMethodsMap)

		// Assign all relevant fields to the test service
		ts.hydrate(serviceConfig, ts.ServiceType, targets, serviceSummaries[ts.ServiceID])

		// Log service specific info
		logTestServiceInfo(ts, serviceGatewayURL, serviceConfig)

		// Run the service tests (HTTP and/or WebSocket based on configuration)
		runAllServiceTests(t, ctx, ts)
	}

	printServiceSummaries(serviceSummaries)
}

// runAllServiceTests orchestrates either HTTP or WebSocket tests based on test mode.
// This function runs either HTTP tests OR WebSocket tests, never both.
func runAllServiceTests(t *testing.T, ctx context.Context, ts *TestService) {
	results := make(map[string]*MethodMetrics)
	var resultsMutex sync.Mutex

	// Check if we're in WebSocket-only mode
	websocketsOnlyMode := cfg.isWebSocketsOnly()

	if websocketsOnlyMode {
		// WebSocket-only mode: only run WebSocket tests
		if !ts.supportsEVMWebSockets() {
			t.Errorf("❌ Service %s does not support WebSocket tests but TEST_WEBSOCKETS=true was set", ts.ServiceID)
			return
		}

		websocketTestFailed := runWebSocketServiceTest(t, ctx, ts, results, &resultsMutex)

		// Calculate and validate the WebSocket service summary
		overallTestFailed := calculateServiceSummary(t, ts, results)

		// Mark overall test as failed if any component failed
		if websocketTestFailed || overallTestFailed {
			t.Fail()
		}
	} else {
		// Default mode: only run HTTP tests
		runHTTPServiceTestWithResults(t, ctx, ts, results, &resultsMutex)

		// Calculate and validate the HTTP service summary
		overallTestFailed := calculateServiceSummary(t, ts, results)

		// Mark overall test as failed if HTTP tests failed
		if overallTestFailed {
			t.Fail()
		}
	}
}

// runHTTPServiceTestWithResults runs HTTP tests and populates the shared results map.
// This function is used when running HTTP-only tests.
func runHTTPServiceTestWithResults(t *testing.T, ctx context.Context, ts *TestService, results map[string]*MethodMetrics, resultsMutex *sync.Mutex) {
	httpResults := make(map[string]*MethodMetrics)
	httpResultsMutex := sync.Mutex{}

	// Run the HTTP test with its own results map
	runHTTPServiceTest(t, ctx, ts, httpResults, &httpResultsMutex)

	// Copy HTTP results to the shared results map (no labeling needed since tests are separate)
	resultsMutex.Lock()
	for method, metrics := range httpResults {
		results[method] = metrics
	}
	resultsMutex.Unlock()
}

// -------------------- Helper Functions --------------------

// TODO_TECHDEBT(@commoddity): Separate E2E and Load test modes into separate files and `Test_PATH_E2E` and `Test_PATH_Load` functions.
//
// getGatewayURLForTestMode returns the gateway URL based on the current test mode.
// It also performs any necessary setup for E2E mode (e.g., Docker container startup) and calls waitForHydratorIfNeeded.
//
// Note: If E2E mode, the caller is responsible for deferring the returned teardown function, if not nil.
func getGatewayURLForTestMode(t *testing.T, cfg *Config) (gatewayURL string, teardownFn func()) {
	switch cfg.getTestMode() {

	case testModeE2E:
		var port string
		port, teardownFn = setupPathInstance(t, shannonConfigFile, cfg.E2ELoadTestConfig.E2EConfig.DockerConfig)
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

// setTestServiceConfigs sets the test service configs for the test services.
func setTestServiceConfigs(testServices []*TestService) map[protocol.ServiceID]ServiceConfig {
	serviceConfigs := make(map[protocol.ServiceID]ServiceConfig)
	for _, ts := range testServices {
		serviceConfig := cfg.DefaultServiceConfig
		if override, exists := cfg.ServiceConfigOverrides[ts.ServiceID]; exists {
			serviceConfig.applyOverride(ts, &override)
		}
		serviceConfigs[ts.ServiceID] = serviceConfig
	}
	return serviceConfigs
}

// -------------------- Log Functions --------------------

// logTestStartInfo logs the test start information for the user.
func logTestStartInfo(gatewayURL string) {
	if cfg.getTestMode() == testModeLoad {
		fmt.Println("\n🔥 Starting Vegeta Load test ...")
	} else {
		fmt.Println("\n🌿 Starting PATH E2E test ...")
	}
	fmt.Printf("  📡 Test protocol: %sShannon%s\n", BOLD_CYAN, RESET)
	fmt.Printf("  🧬 Gateway URL: %s%s%s\n", BLUE, gatewayURL, RESET)

	if cfg.E2ELoadTestConfig.LoadTestConfig != nil {
		if cfg.E2ELoadTestConfig.LoadTestConfig.PortalApplicationID != "" {
			fmt.Printf("  🌀 Portal Application ID: %s%s%s\n", CYAN, cfg.E2ELoadTestConfig.LoadTestConfig.PortalApplicationID, RESET)
		}
		if cfg.E2ELoadTestConfig.LoadTestConfig.PortalAPIKey != "" {
			fmt.Printf("  🔑 Portal API Key: %s%s%s\n", CYAN, cfg.E2ELoadTestConfig.LoadTestConfig.PortalAPIKey, RESET)
		}
	}
}

func logTestServiceIDs(testServices []*TestService) {
	fmt.Printf("\n\n=======================================================\n")
	fmt.Printf("⛓️  Will be running tests for service IDs:\n")
	for _, ts := range testServices {
		var typeStr string
		var icon string

		if ts.Archival && ts.supportsEVMWebSockets() {
			typeStr = "(Archival + WebSocket)"
			icon = "🗄️🔌"
		} else if ts.Archival {
			typeStr = "(Archival)"
			icon = "🗄️"
		} else if ts.supportsEVMWebSockets() {
			typeStr = "(WebSocket)"
			icon = "🔌"
		} else {
			typeStr = "(Non-archival)"
			icon = "📝"
		}

		fmt.Printf("  %s  %s%s%s %s\n", icon, GREEN, ts.ServiceID, RESET, typeStr)
	}
}

func logTestServiceInfo(ts *TestService, serviceGatewayURL string, serviceConfig ServiceConfig) {
	fmt.Printf("\n\n=======================================================\n")
	fmt.Printf("🛠️  Starting test for : %s%s%s\n", BOLD_BLUE, ts.Name, RESET)
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
		fmt.Println("🛑 Received SIGINT, canceling test...")
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
