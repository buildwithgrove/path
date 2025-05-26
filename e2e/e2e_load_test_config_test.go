//go:build e2e

package e2e

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/buildwithgrove/path/protocol"
)

// -----------------------------------------------------------------------------
// Environment Variables
// -----------------------------------------------------------------------------

// Environment variable names (must be set for tests)
const (
	envTestMode     = "TEST_MODE"     // The test mode to run (e2e or load)
	envTestProtocol = "TEST_PROTOCOL" // The protocol to test (morse or shannon)
)

// getEnvConfig fetches and validates environment config from environment variables
func getEnvConfig() (EnvConfig, error) {
	mode := testMode(os.Getenv(envTestMode))
	if err := mode.isValid(); err != nil {
		return EnvConfig{}, err
	}

	protocol := testProtocol(os.Getenv(envTestProtocol))
	if err := protocol.isValid(); err != nil {
		return EnvConfig{}, err
	}

	return EnvConfig{
		TestMode:     mode,
		TestProtocol: protocol,
	}, nil
}

// -----------------------------------------------------------------------------
// Enums
// -----------------------------------------------------------------------------

type testMode string

const (
	testModeE2E  testMode = "e2e"  // Run E2E tests
	testModeLoad testMode = "load" // Run load tests
)

// isValid checks if testMode is valid and set
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
// Valid values: "morse" or "shannon"
type testProtocol string

const (
	protocolMorse   testProtocol = "morse"   // Run tests against Morse
	protocolShannon testProtocol = "shannon" // Run tests against Shannon
)

// isValid checks if testProtocol is valid and set
func (p testProtocol) isValid() error {
	if p == "" {
		return fmt.Errorf("[REQUIRED] %s environment variable is not set", envTestProtocol)
	}
	if p != protocolMorse && p != protocolShannon {
		return fmt.Errorf("invalid protocol %s", p)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Config Files
// -----------------------------------------------------------------------------

// TODO_TECHDEBT(@commoddity): Making this configurable via a flag or env var.
// Config file paths relative to the e2e directory
const (
	// Expected name and location of custom config file
	customConfigFile = "config/.e2e_load_test.config.yaml"

	// Default config file (used if custom config file is not found)
	defaultConfigFile = "config/e2e_load_test.config.tmpl.yaml"
)

// loadE2ELoadTestConfig loads the E2E configuration in the following order:
//  1. Custom config in e2e/config/.e2e_load_test.config.yaml
//  2. Default config in e2e/config/e2e_load_test.config.tmpl.yaml
func loadE2ELoadTestConfig() (*Config, error) {
	envConfig, err := getEnvConfig()
	if err != nil {
		return nil, err
	}

	var cfgPath string
	// Prefer custom config if present, otherwise fall back to default
	if _, err := os.Stat(customConfigFile); err == nil {
		fmt.Printf("⚠️ Using custom config file: e2e/%s\n", customConfigFile)
		cfgPath = customConfigFile
	} else {
		fmt.Printf("⚠️ Using default config file: e2e/%s\n", defaultConfigFile)
		cfgPath = defaultConfigFile
	}

	// Load the config
	cfg, err := loadConfig(cfgPath)
	if err != nil {
		return nil, err
	}
	cfg.EnvConfig = envConfig

	// Validate the configuration
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// loadConfig loads the E2E configuration from the specified YAML file
func loadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// -----------------------------------------------------------------------------
// Config Structs
// -----------------------------------------------------------------------------

// DEV_NOTE: All structs and `yaml:` tagged fields must be public to allow for unmarshalling using `gopkg.in/yaml`
type (
	// Config is the top-level E2E test configuration
	Config struct {
		EnvConfig         EnvConfig         // Loaded from environment variables, not YAML
		E2ELoadTestConfig E2ELoadTestConfig `yaml:"e2e_load_test_config"`
		DefaultTestConfig TestConfig        `yaml:"default_test_config"`
		TestCases         []TestCase        `yaml:"test_cases"`
	}

	// EnvConfig for environment configuration (loaded from environment variables, not YAML)
	EnvConfig struct {
		TestMode     testMode
		TestProtocol testProtocol
	}

	// E2ELoadTestConfig for test mode configuration
	E2ELoadTestConfig struct {
		E2EConfig      E2EConfig       `yaml:"e2e_config"`       // E2E test mode configuration
		LoadTestConfig *LoadTestConfig `yaml:"load_test_config"` // Load test mode configuration (pointer, may be nil)
	}

	// E2EConfig for E2E test mode configuration
	E2EConfig struct {
		WaitForHydrator int          `yaml:"wait_for_hydrator"` // Seconds to wait for hydrator checks
		DockerConfig    DockerConfig `yaml:"docker_config"`     // Docker configuration
	}

	// DockerConfig for Docker configuration
	DockerConfig struct {
		LogToFile         bool `yaml:"log_to_file"`         // Log Docker container output
		ForceRebuildImage bool `yaml:"force_rebuild_image"` // Force Docker image rebuild (useful after code changes)
	}

	// LoadTestConfig for load test mode configuration
	LoadTestConfig struct {
		GatewayURLOverride  string `yaml:"gateway_url_override"`  // [REQUIRED] Custom PATH gateway URL
		UseServiceSubdomain bool   `yaml:"use_service_subdomain"` // Whether to specify the service using the subdomain per-test case
		PortalApplicationID string `yaml:"portal_application_id"` // [REQUIRED] Portal Application ID for the test
		PortalAPIKey        string `yaml:"portal_api_key"`        // [OPTIONAL] Portal API key for the test
	}

	// TestConfig for common test configuration options
	TestConfig struct {
		GlobalRPS         int           `yaml:"global_rps"`          // Requests per second (shared by all methods)
		RequestsPerMethod int           `yaml:"requests_per_method"` // Total number of requests to send for each method
		SuccessRate       float64       `yaml:"success_rate"`        // Minimum success rate required (0-1)
		MaxP50LatencyMS   time.Duration `yaml:"max_p50_latency_ms"`  // Maximum P50 latency in milliseconds
		MaxP95LatencyMS   time.Duration `yaml:"max_p95_latency_ms"`  // Maximum P95 latency in milliseconds
		MaxP99LatencyMS   time.Duration `yaml:"max_p99_latency_ms"`  // Maximum P99 latency in milliseconds
	}

	// TestCase for test case configuration
	TestCase struct {
		Name                   string             `yaml:"name"`                                // Name of the test case
		Protocol               testProtocol       `yaml:"protocol"`                            // Protocol name (morse or shannon)
		ServiceID              protocol.ServiceID `yaml:"service_id"`                          // Service ID to test (identifies the specific blockchain service)
		Archival               bool               `yaml:"archival,omitempty"`                  // Whether this is an archival test (historical data access)
		ServiceParams          ServiceParams      `yaml:"service_params"`                      // Service-specific parameters for test requests
		TestCaseConfigOverride *TestConfig        `yaml:"test_case_config_override,omitempty"` // Override default configuration for this test case
		TestCaseMethodOverride []string           `yaml:"test_case_method_override,omitempty"` // Override methods to test for this test case
	}

	// ServiceParams holds service-specific test data for all methods.
	ServiceParams struct {
		ContractAddress    string `yaml:"contract_address,omitempty"`     // EVM contract address (should match service_qos_config.go)
		CallData           string `yaml:"call_data,omitempty"`            // Call data for eth_call
		ContractStartBlock uint64 `yaml:"contract_start_block,omitempty"` // Minimum block number to use for archival tests
		TransactionHash    string `yaml:"transaction_hash,omitempty"`     // Transaction hash for receipt/transaction queries
		blockNumber        string // Not marshaled; set in test case. Can be "latest" or an archival block number
	}
)

// -----------------------------------------------------------------------------
// TestConfig Methods
// -----------------------------------------------------------------------------

// MergeNonZero merges non-zero values from the override config into this config.
// This ensures that only fields that are explicitly set in the override config are merged,
// while preserving the default values for the other fields.
func (tc *TestConfig) MergeNonZero(override *TestConfig) {
	if override == nil {
		return
	}

	if override.GlobalRPS != 0 {
		tc.GlobalRPS = override.GlobalRPS
	}
	if override.RequestsPerMethod != 0 {
		tc.RequestsPerMethod = override.RequestsPerMethod
	}
	if override.SuccessRate != 0 {
		tc.SuccessRate = override.SuccessRate
	}
	if override.MaxP50LatencyMS != 0 {
		tc.MaxP50LatencyMS = override.MaxP50LatencyMS
	}
	if override.MaxP95LatencyMS != 0 {
		tc.MaxP95LatencyMS = override.MaxP95LatencyMS
	}
	if override.MaxP99LatencyMS != 0 {
		tc.MaxP99LatencyMS = override.MaxP99LatencyMS
	}
}

// -----------------------------------------------------------------------------
// Config Accessors
// -----------------------------------------------------------------------------

// TODO_TECHDEBT(@commoddity): Refactor EVM Tests to avoid `if cfg.getTestMode() == ` checks.
// Separate out load tests and E2E tests into different files.
func (c *Config) getTestMode() testMode {
	return c.EnvConfig.TestMode
}

func (c *Config) getTestProtocol() testProtocol {
	return c.EnvConfig.TestProtocol
}

func (c *Config) useServiceSubdomain() bool {
	return c.E2ELoadTestConfig.LoadTestConfig.UseServiceSubdomain
}

func (c *Config) getGatewayURLForLoadTest() string {
	return c.E2ELoadTestConfig.LoadTestConfig.GatewayURLOverride
}

// -----------------------------------------------------------------------------
// Utility Functions
// -----------------------------------------------------------------------------

// setServiceIDInGatewayURLSubdomain inserts the service ID as a subdomain in the gateway URL.
//   - https://rpc.grove.city/v1 → https://F00C.rpc.grove.city/v1
//   - http://localhost:3091/v1 → http://F00C.localhost:3091/v1
//   - https://api.example.com/path?query=param → https://F00C.api.example.com/path?query=param
//
// TODO_TECHDEBT(@commoddity): Remove this once PATH in production supports service in headers
//   - Issue: https://github.com/buildwithgrove/infrastructure/issues/91
func setServiceIDInGatewayURLSubdomain(gatewayURL string, serviceID protocol.ServiceID) string {
	parsedURL, err := url.Parse(gatewayURL)
	if err != nil {
		// If parsing fails, fall back to simple string insertion
		return gatewayURL
	}
	parsedURL.Host = fmt.Sprintf("%s.%s", serviceID, parsedURL.Host)
	return parsedURL.String()
}

// getTestCases returns test cases filtered by protocol specified in environment
func (c *Config) getTestCases() []TestCase {
	var filteredTestCases []TestCase
	for _, tc := range c.TestCases {
		if tc.Protocol == c.getTestProtocol() {
			filteredTestCases = append(filteredTestCases, tc)
		}
	}
	return filteredTestCases
}

// -----------------------------------------------------------------------------
// Validation
// -----------------------------------------------------------------------------

// validate performs configuration validation based on schema and runtime requirements
func (c *Config) validate() error {
	mode := c.getTestMode()

	// Validate load test mode
	if mode == testModeLoad {
		if c.E2ELoadTestConfig.LoadTestConfig == nil {
			return fmt.Errorf("load test mode requires loadTestConfig to be set")
		}
		if c.E2ELoadTestConfig.LoadTestConfig.GatewayURLOverride == "" {
			return fmt.Errorf("load test mode requires GatewayURLOverride to be set")
		}
		if c.E2ELoadTestConfig.LoadTestConfig.PortalApplicationID == "" {
			return fmt.Errorf("load test mode requires PortalApplicationID to be set")
		}
	}

	// Validate e2e test mode
	if mode == testModeE2E {
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

	// Protocol-specific test case presence
	protocol := c.getTestProtocol()
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

	// Validate all test cases
	for i, tc := range c.TestCases {
		if err := c.validateTestCase(tc, i); err != nil {
			return err
		}
	}

	// Validate default test config
	if c.DefaultTestConfig.RequestsPerMethod <= 0 {
		return fmt.Errorf("DefaultTestConfig.RequestsPerMethod must be greater than 0")
	}
	if c.DefaultTestConfig.GlobalRPS <= 0 {
		return fmt.Errorf("DefaultTestConfig.GlobalRPS must be greater than 0")
	}
	if c.DefaultTestConfig.SuccessRate < 0 || c.DefaultTestConfig.SuccessRate > 1 {
		return fmt.Errorf("DefaultTestConfig.SuccessRate must be between 0 and 1")
	}

	return nil
}

// validateTestCase validates an individual test case and its config
func (c *Config) validateTestCase(tc TestCase, index int) error {
	// Validate common fields
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

	// Validate Morse-specific params
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

	// Validate Shannon-specific params
	if tc.Protocol == protocolShannon {
		if tc.ServiceParams.ContractAddress == "" {
			return fmt.Errorf("test case #%d: ContractAddress is required for Shannon tests", index)
		}
	}

	return nil
}
