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
func loadE2EConfig() (*Config, error) {
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

	// Validate the configuration
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

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
func loadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// Config is the top-level E2E test configuration
type (
	Config struct {
		envConfig           envConfig    // envConfig is loaded from environment variables not YAML
		TestConfig          TestConfig   `yaml:"test_config"`
		DefaultMethodConfig MethodConfig `yaml:"default_method_config"`
		TestCases           []TestCase   `yaml:"test_cases"`
	}

	// envConfig for environment configuration (loaded from environment variables)
	envConfig struct {
		testMode     testMode
		testProtocol testProtocol
	}

	// TestConfig for general test settings
	TestConfig struct {
		// E2E test mode configuration
		E2EConfig *E2EConfig `yaml:"e2e_config"`
		// Load test mode configuration
		LoadTestConfig *LoadTestConfig `yaml:"load_test_config"`
	}

	// E2EConfig for E2E test mode configuration
	E2EConfig struct {
		// Seconds to wait for hydrator checks
		WaitForHydrator int `yaml:"wait_for_hydrator"`
		// Docker configuration
		DockerConfig DockerConfig `yaml:"docker_config"`
	}

	// DockerConfig for Docker configuration
	DockerConfig struct {
		// Log Docker container output
		LogToFile bool `yaml:"log_to_file"`
		// Force Docker image rebuild (useful after code changes)
		ForceRebuildImage bool `yaml:"force_rebuild_image"`
	}

	// LoadTestConfig for load test mode configuration
	LoadTestConfig struct {
		// Custom PATH gateway URL
		GatewayURLOverride string `yaml:"gateway_url_override"`
		// Whether to specify the service using the subdomain per-test case
		// TODO_TECHDEBT(@commoddity): Remove this once PATH in production supports service in headers
		//     - Issue: https://github.com/buildwithgrove/infrastructure/issues/91
		UseServiceSubdomain bool `yaml:"use_service_subdomain"`
		// Custom user identifier for the test (eg. portal-application-id)
		PortalApplicationIDOverride string `yaml:"portal_application_id_override"`
		// Custom API key for the test (eg. portal-api-key)
		PortalAPIKeyOverride string `yaml:"portal_api_key_override"`
	}

	// MethodConfig for common test configuration options
	MethodConfig struct {
		// Total number of requests to send for each method
		TotalRequests int `yaml:"total_requests"`
		// Requests per second
		RPS int `yaml:"rps"`
		// Minimum success rate required (0-1)
		SuccessRate float64 `yaml:"success_rate"`
		// Maximum P50 latency in milliseconds
		MaxP50LatencyMS time.Duration `yaml:"max_p50_latency_ms"`
		// Maximum P95 latency in milliseconds
		MaxP95LatencyMS time.Duration `yaml:"max_p95_latency_ms"`
		// Maximum P99 latency in milliseconds
		MaxP99LatencyMS time.Duration `yaml:"max_p99_latency_ms"`
	}

	// TestCase for test case configuration
	TestCase struct {
		// Name of the test case
		Name string `yaml:"name"`
		// Protocol name (morse or shannon)
		Protocol testProtocol `yaml:"protocol"`
		// Service ID to test (identifies the specific blockchain service)
		ServiceID protocol.ServiceID `yaml:"service_id"`
		// Whether this is an archival test (historical data access)
		Archival bool `yaml:"archival,omitempty"`
		// Service-specific parameters for test requests
		ServiceParams ServiceParams `yaml:"service_params"`
		// Multiplier for latency thresholds for this test case
		LatencyMultiplier int `yaml:"latency_multiplier,omitempty"`
		// Override default configuration for this test case
		TestCaseConfigOverride *MethodConfig `yaml:"test_case_config_override,omitempty"`
		// Override methods to test for this test case
		TestCaseMethodOverride []string `yaml:"test_case_method_override,omitempty"`
	}

	// ServiceParams for service-specific parameters
	ServiceParams struct {
		// Contract address for eth calls
		ContractAddress string `yaml:"contract_address,omitempty"`
		// Call data for eth_call
		CallData string `yaml:"call_data,omitempty"`
		// Minimum block number for archival tests
		ContractStartBlock uint64 `yaml:"contract_start_block,omitempty"`
		// Transaction hash for receipt/transaction queries
		TransactionHash string `yaml:"transaction_hash,omitempty"`
	}
)

func (c *Config) getTestMode() testMode {
	return c.envConfig.testMode
}

func (c *Config) getTestProtocol() testProtocol {
	return c.envConfig.testProtocol
}

func (c *Config) useServiceSubdomain() bool {
	return c.TestConfig.LoadTestConfig.UseServiceSubdomain
}

func (c *Config) getGatewayURLForLoadTest() string {
	return c.TestConfig.LoadTestConfig.GatewayURLOverride
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
func (c *Config) getTestCases() []TestCase {
	// Filter test cases by protocol
	var filteredTestCases []TestCase
	for _, tc := range c.TestCases {
		if tc.Protocol == c.getTestProtocol() {
			filteredTestCases = append(filteredTestCases, tc)
		}
	}

	return filteredTestCases
}

// validate performs configuration validation based on schema and runtime requirements
func (c *Config) validate() error {
	// Validate based on test mode
	mode := c.getTestMode()

	// Mode-specific validations
	if mode == testModeLoad {
		if c.TestConfig.LoadTestConfig == nil {
			return fmt.Errorf("load test mode requires LoadTestConfig to be set")
		}

		// Required fields validation for load test mode
		if c.TestConfig.LoadTestConfig.GatewayURLOverride == "" {
			return fmt.Errorf("load test mode requires GatewayURLOverride to be set")
		}

		if c.TestConfig.LoadTestConfig.PortalApplicationIDOverride == "" {
			return fmt.Errorf("load test mode requires PortalApplicationIDOverride to be set")
		}
	} else if mode == testModeE2E {
		if c.TestConfig.E2EConfig == nil {
			return fmt.Errorf("e2e test mode requires E2EConfig to be set")
		}

		// Check for protocol-specific config files in e2e mode
		protocol := c.getTestProtocol()
		var configFile string

		if protocol == protocolMorse {
			configFile = "config/.morse.config.yaml"
		} else if protocol == protocolShannon {
			configFile = "config/.shannon.config.yaml"
		}

		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			return fmt.Errorf("e2e test mode requires %s to exist", configFile)
		}
	}

	// Validate based on protocol
	protocol := c.getTestProtocol()

	// Check for presence of test cases for the specified protocol
	hasMorseCases := false
	hasShannonCases := false

	for _, tc := range c.TestCases {
		if tc.Protocol == protocolMorse {
			hasMorseCases = true
		} else if tc.Protocol == protocolShannon {
			hasShannonCases = true
		}
	}

	if protocol == protocolMorse && !hasMorseCases {
		return fmt.Errorf("no test cases found for Morse protocol")
	}

	if protocol == protocolShannon && !hasShannonCases {
		return fmt.Errorf("no test cases found for Shannon protocol")
	}

	// Validate test cases
	for i, tc := range c.TestCases {
		if err := c.validateTestCase(tc, i); err != nil {
			return err
		}
	}

	// Validate default method config
	if c.DefaultMethodConfig.TotalRequests <= 0 {
		return fmt.Errorf("DefaultMethodConfig.TotalRequests must be greater than 0")
	}

	if c.DefaultMethodConfig.RPS <= 0 {
		return fmt.Errorf("DefaultMethodConfig.RPS must be greater than 0")
	}

	if c.DefaultMethodConfig.SuccessRate < 0 || c.DefaultMethodConfig.SuccessRate > 1 {
		return fmt.Errorf("DefaultMethodConfig.SuccessRate must be between 0 and 1")
	}

	// All validations passed
	return nil
}

// validateTestCase validates an individual test case
func (c *Config) validateTestCase(tc TestCase, index int) error {
	// Validate required fields
	if tc.Name == "" {
		return fmt.Errorf("test case #%d: Name is required", index)
	}

	if tc.Protocol == "" {
		return fmt.Errorf("test case #%d: Protocol is required", index)
	}

	if tc.Protocol != protocolMorse && tc.Protocol != protocolShannon {
		return fmt.Errorf("test case #%d: Protocol must be either 'morse' or 'shannon'", index)
	}

	if tc.ServiceID == "" {
		return fmt.Errorf("test case #%d: ServiceID is required", index)
	}

	// Validate service params based on protocol
	if tc.Protocol == protocolMorse {
		if tc.Archival && tc.ServiceParams.ContractStartBlock == 0 {
			return fmt.Errorf("test case #%d: ContractStartBlock is required for archival Morse tests", index)
		}

		if tc.ServiceParams.ContractAddress == "" {
			return fmt.Errorf("test case #%d: ContractAddress is required for Morse tests", index)
		}

		if tc.ServiceParams.TransactionHash == "" {
			return fmt.Errorf("test case #%d: TransactionHash is required for Morse tests", index)
		}
	}

	if tc.Protocol == protocolShannon {
		if tc.ServiceParams.ContractAddress == "" {
			return fmt.Errorf("test case #%d: ContractAddress is required for Shannon tests", index)
		}
	}

	// Validate test case override config if present
	if tc.TestCaseConfigOverride != nil {
		if tc.TestCaseConfigOverride.TotalRequests <= 0 {
			return fmt.Errorf("test case #%d: TestCaseConfigOverride.TotalRequests must be greater than 0", index)
		}

		if tc.TestCaseConfigOverride.RPS <= 0 {
			return fmt.Errorf("test case #%d: TestCaseConfigOverride.RPS must be greater than 0", index)
		}

		if tc.TestCaseConfigOverride.SuccessRate < 0 || tc.TestCaseConfigOverride.SuccessRate > 1 {
			return fmt.Errorf("test case #%d: TestCaseConfigOverride.SuccessRate must be between 0 and 1", index)
		}
	}

	// Validate latency multiplier if present
	if tc.LatencyMultiplier < 0 {
		return fmt.Errorf("test case #%d: LatencyMultiplier must be greater than or equal to 0", index)
	}

	return nil
}
