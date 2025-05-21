//go:build e2e

package e2e

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// Config file paths relative to the e2e directory
	defaultConfigFile = "config/e2econfig.tmpl.yaml"
	customConfigFile  = "config/.e2econfig.yaml"
)

// LoadE2EConfig loads the E2E configuration in the following order:
//  1. Custom config in e2e/config/.e2econfig.yaml
//  2. Default config in e2e/config/e2econfig.tmpl.yaml
func LoadE2EConfig() (*Config, error) {
	// Check if custom config exists
	if _, err := os.Stat(customConfigFile); err == nil {
		return loadConfig(customConfigFile)
	}

	// Fall back to default config
	return loadConfig(defaultConfigFile)
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
		TestConfig          TestConfig      `yaml:"test_config"`
		DockerConfig        DockerConfig    `yaml:"docker_config"`
		DefaultMethodConfig MethodConfig    `yaml:"default_method_config"`
		MorseConfig         *ProtocolConfig `yaml:"morse_config,omitempty"`
		ShannonConfig       *ProtocolConfig `yaml:"shannon_config,omitempty"`
	}

	// TestConfig for general test settings
	TestConfig struct {
		// Custom PATH gateway URL (useful for local dev)
		GatewayURLOverride string `yaml:"gateway_url_override"`
		// Test only a specific service ID
		ServiceIDOverride string `yaml:"service_id_override"`
		// Seconds to wait for hydrator checks
		WaitForHydrator int `yaml:"wait_for_hydrator"`
	}

	// DockerConfig for Docker-related settings
	DockerConfig struct {
		// Log Docker container output
		DockerLog bool `yaml:"docker_log"`
		// Force Docker image rebuild (useful after code changes)
		DockerForceRebuild bool `yaml:"docker_force_rebuild"`
	}

	// TestConfig for common test configuration options
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

	// ProtocolConfig for protocol-specific configuration (shared by Morse and Shannon)
	ProtocolConfig struct {
		// Array of test cases to run
		TestCases []TestCase `yaml:"test_cases"`
	}

	// TestCase for test case configuration
	TestCase struct {
		// Name of the test case
		Name string `yaml:"name"`
		// Service ID to test (identifies the specific blockchain service)
		ServiceID string `yaml:"service_id"`
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
