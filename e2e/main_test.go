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
)

// -----------------------------------------------------------------------------
// Vegeta E2E & Load Tests
// -----------------------------------------------------------------------------
//
// Documentation:
//   https://path.grove.city/develop/path/e2e_tests
//
// Example Usage - E2E tests:
//   - make e2e_test_all             # Run all E2E tests for all services
//   - make e2e_test <service IDs>   # Run all E2E tests for the specified services
//
// Example Usage - Load tests:
//   - make load_test_all            # Run all load tests for all services
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
	logTestStartInfo(gatewayURL)

	// Initialize service summaries map (will be logged out at the end of the test)
	serviceSummaries := make(map[protocol.ServiceID]*serviceSummary)

	// Assign test service configs to each test service
	testServiceConfigs := setTestServiceConfigs(testServices)

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

		// Get methods to test
		methodsToTest := ts.getTestMethods()

		// Generate targets for this service
		targets, err := ts.getVegetaTargets(methodsToTest, serviceGatewayURL)
		if err != nil {
			t.Fatalf("❌ Failed to get vegeta targets: %v", err)
		}

		// Create summary for this service
		serviceSummaries[ts.ServiceID] = newServiceSummary(ts.ServiceID, serviceConfig, methodsToTest)

		// Assign all relevant fields to the test service
		ts.hydrate(serviceConfig, ts.ServiceType, targets, serviceSummaries[ts.ServiceID])

		// Log service specific info
		logTestServiceInfo(ts, serviceGatewayURL, serviceConfig)

		// Run the service test
		runServiceTest(t, ctx, ts)
	}

	printServiceSummaries(serviceSummaries)
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
		port, teardownFn = setupPathInstance(t, configFile, cfg.E2ELoadTestConfig.E2EConfig.DockerConfig)
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
		if ts.Archival {
			fmt.Printf("  🗄️  %s%s%s (Archival)\n", GREEN, ts.ServiceID, RESET)
		} else {
			fmt.Printf("  📝  %s%s%s (Non-archival)\n", GREEN, ts.ServiceID, RESET)
		}
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
