//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/buildwithgrove/path/protocol"
	"gopkg.in/yaml.v3"
)

// -------------------- Environment Variables --------------------

// Both environment variables must be set
const (
	envTestMode     = "TEST_MODE"     // The test mode to run (e2e or load)
	envTestProtocol = "TEST_PROTOCOL" // The protocol to test (morse or shannon)
)

// getEnvConfig returns the environment configuration
func getEnvConfig() (envConfig, error) {
	mode := testMode(os.Getenv(envTestMode))
	if err := mode.isValid(); err != nil {
		return envConfig{}, err
	}

	protocol := testProtocol(os.Getenv(envTestProtocol))
	if err := protocol.isValid(); err != nil {
		return envConfig{}, err
	}

	return envConfig{
		testMode:     mode,
		testProtocol: protocol,
	}, nil
}

// -------------------- Enums --------------------

type testMode string

const (
	testModeE2E  testMode = "e2e"
	testModeLoad testMode = "load"
)

func (t testMode) isValid() error {
	if t == "" {
		return fmt.Errorf("[REQUIRED] %s environment variable is not set", envTestMode)
	}
	if t != testModeE2E && t != testModeLoad {
		return fmt.Errorf("invalid test mode %s", t)
	}
	return nil
}

// testProtocol determines whether to test PATH with Morse or Shannon
type testProtocol string

const (
	protocolMorse   testProtocol = "morse"
	protocolShannon testProtocol = "shannon"
)

// isValid returns true if the protocol is either morse or shannon
func (p testProtocol) isValid() error {
	if p == "" {
		return fmt.Errorf("[REQUIRED] %s environment variable is not set", envTestProtocol)
	}
	if p != protocolMorse && p != protocolShannon {
		return fmt.Errorf("invalid protocol %s", p)
	}
	return nil
}

// -------------------- Config Files --------------------

// Config file paths relative to the e2e directory
const (
	customConfigFile  = "config/.e2econfig.yaml"     // Custom config file (loaded if it exists)
	defaultConfigFile = "config/e2econfig.tmpl.yaml" // Default config file (used if custom config file is not found)
)

// loadE2EConfig loads the E2E configuration in the following order:
//  1. Custom config in e2e/config/.e2econfig.yaml
//  2. Default config in e2e/config/e2econfig.tmpl.yaml
func loadE2EConfig() (*config, error) {
	envConfig, err := getEnvConfig()
	if err != nil {
		return nil, err
	}

	var cfgPath string
	// Check if custom config exists
	if _, err := os.Stat(customConfigFile); err == nil {
		cfgPath = customConfigFile
	} else {
		cfgPath = defaultConfigFile
	}

	cfg, err := loadConfig(cfgPath)
	if err != nil {
		return nil, err
	}
	cfg.envConfig = envConfig

	return cfg, nil
}

func PrettyLog(args ...interface{}) {
	for _, arg := range args {
		var prettyJSON bytes.Buffer
		jsonArg, _ := json.Marshal(arg)
		str := string(jsonArg)
		_ = json.Indent(&prettyJSON, []byte(str), "", "    ")
		output := prettyJSON.String()

		fmt.Println(output)
	}
}

// loadConfig loads the E2E configuration from the specified file path
func loadConfig(filePath string) (*config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// Config is the top-level E2E test configuration
type (
	config struct {
		envConfig           envConfig
		testConfig          testConfig   `yaml:"test_config"`
		defaultMethodConfig methodConfig `yaml:"default_method_config"`
		testCases           []testCase   `yaml:"test_cases"`
	}

	envConfig struct {
		testMode     testMode
		testProtocol testProtocol
	}

	// testConfig for general test settings
	testConfig struct {
		// E2E test mode configuration
		e2eConfig *e2eConfig `yaml:"e2e_config"`
		// Load test mode configuration
		loadTestConfig *loadTestConfig `yaml:"load_test_config"`
	}

	// e2eConfig for E2E test mode configuration
	e2eConfig struct {
		// Seconds to wait for hydrator checks
		waitForHydrator int `yaml:"wait_for_hydrator"`
		// Docker configuration
		dockerConfig dockerConfig `yaml:"docker_config"`
	}

	// dockerConfig for Docker configuration
	dockerConfig struct {
		// Log Docker container output
		logToFile bool `yaml:"log_to_file"`
		// Force Docker image rebuild (useful after code changes)
		forceRebuildImage bool `yaml:"force_rebuild_image"`
	}

	// loadTestConfig for load test mode configuration
	loadTestConfig struct {
		// Custom PATH gateway URL
		gatewayURLOverride string `yaml:"gateway_url_override"`
		// Whether to specify the service using the subdomain per-test case
		// TODO_TECHDEBT(@commoddity): Remove this once PATH in production supports service in headers
		//     - Issue: https://github.com/buildwithgrove/infrastructure/issues/91
		useServiceSubdomain bool `yaml:"use_service_subdomain"`
		// Custom user identifier for the test (eg. portal-application-id)
		portalApplicationIDOverride string `yaml:"portal_application_id_override"`
		// Custom API key for the test (eg. portal-api-key)
		portalAPIKeyOverride string `yaml:"portal_api_key_override"`
	}

	// methodConfig for common test configuration options
	methodConfig struct {
		// Total number of requests to send for each method
		totalRequests int `yaml:"total_requests"`
		// Requests per second
		rps int `yaml:"rps"`
		// Minimum success rate required (0-1)
		successRate float64 `yaml:"success_rate"`
		// Maximum P50 latency in milliseconds
		maxP50LatencyMS time.Duration `yaml:"max_p50_latency_ms"`
		// Maximum P95 latency in milliseconds
		maxP95LatencyMS time.Duration `yaml:"max_p95_latency_ms"`
		// Maximum P99 latency in milliseconds
		maxP99LatencyMS time.Duration `yaml:"max_p99_latency_ms"`
	}

	// testCase for test case configuration
	testCase struct {
		// Name of the test case
		name string `yaml:"name"`
		// Protocol name (morse or shannon)
		protocol testProtocol `yaml:"protocol"`
		// Service ID to test (identifies the specific blockchain service)
		serviceID string `yaml:"service_id"`
		// Whether this is an archival test (historical data access)
		archival bool `yaml:"archival,omitempty"`
		// Service-specific parameters for test requests
		serviceParams serviceParams `yaml:"service_params"`
		// Multiplier for latency thresholds for this test case
		latencyMultiplier int `yaml:"latency_multiplier,omitempty"`
		// Override default configuration for this test case
		testCaseConfigOverride *methodConfig `yaml:"test_case_config_override,omitempty"`
		// Override methods to test for this test case
		testCaseMethodOverride []string `yaml:"test_case_method_override,omitempty"`
	}

	// serviceParams for service-specific parameters
	serviceParams struct {
		// Contract address for eth calls
		contractAddress string `yaml:"contract_address,omitempty"`
		// Call data for eth_call
		callData string `yaml:"call_data,omitempty"`
		// Minimum block number for archival tests
		contractStartBlock uint64 `yaml:"contract_start_block,omitempty"`
		// Transaction hash for receipt/transaction queries
		transactionHash string `yaml:"transaction_hash,omitempty"`
	}
)

func (c *config) getTestMode() testMode {
	return c.envConfig.testMode
}

func (c *config) getTestProtocol() testProtocol {
	return c.envConfig.testProtocol
}

func (c *config) useServiceSubdomain() bool {
	return c.testConfig.loadTestConfig.useServiceSubdomain
}

// getGatewayURL returns the gateway URL based on the test mode
//   - In load test mode, the gateway URL is specified in the config file
//   - In E2E test mode, the gateway URL is the Docker container URL
func (c *config) getGatewayURL(dockerPort string) string {
	if c.getTestMode() == testModeLoad {
		return c.testConfig.loadTestConfig.gatewayURLOverride
	}
	return fmt.Sprintf("http://localhost:%s/v1", dockerPort)
}

// setServiceIDInGatewayURLSubdomain inserts the service ID as a subdomain in the gateway URL
// Examples:
//   - https://rpc.grove.city/v1 → https://F00C.rpc.grove.city/v1
//   - http://localhost:3091/v1 → http://F00C.localhost:3091/v1
//   - https://api.example.com/path?query=param → https://F00C.api.example.com/path?query=param
//
// TODO_TECHDEBT(@commoddity): Remove this once PATH in production supports service in headers
//   - Issue: https://github.com/buildwithgrove/infrastructure/issues/91
func setServiceIDInGatewayURLSubdomain(gatewayURL string, serviceID protocol.ServiceID) string {
	// Parse the URL to extract protocol, host, and path
	parsedURL, err := url.Parse(gatewayURL)
	if err != nil {
		// If parsing fails, fall back to simple string insertion
		return gatewayURL
	}

	// Insert service ID as subdomain before the host
	parsedURL.Host = fmt.Sprintf("%s.%s", serviceID, parsedURL.Host)

	// Return the modified URL
	return parsedURL.String()
}

// getTestCases returns test cases filtered by protocol if specified in environment
// If no protocol is specified, returns all test cases
func (c *config) getTestCases() []testCase {
	// Filter test cases by protocol
	var filteredTestCases []testCase
	for _, tc := range c.testCases {
		if tc.protocol == c.getTestProtocol() {
			filteredTestCases = append(filteredTestCases, tc)
		}
	}

	return filteredTestCases
}
