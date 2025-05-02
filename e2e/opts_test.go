//go:build e2e

/* -------------------- Test Configuration Options -------------------- */

package e2e

import (
	"fmt"
	"os"
	"strconv"

	"github.com/buildwithgrove/path/protocol"
)

// -------------------- Environment Variable Names --------------------

// Required environment variables
const (
	envTestProtocol = "TEST_PROTOCOL" // must be set to "morse" or "shannon"

	// Optional environment variables
	envGatewayURLOverride = "GATEWAY_URL_OVERRIDE"
	envServiceIDOverride  = "SERVICE_ID_OVERRIDE"
	envDockerLog          = "DOCKER_LOG"
	envDockerForceRebuild = "DOCKER_FORCE_REBUILD"
	envWaitForHydrator    = "WAIT_FOR_HYDRATOR"
)

// -------------------- Protocol String Type --------------------

// protocolStr determines whether to test PATH with Morse or Shannon
type protocolStr string

const (
	morse   protocolStr = "morse"
	shannon protocolStr = "shannon"
)

// isValid returns true if the protocol is either morse or shannon
func (p protocolStr) isValid() bool {
	return p == morse || p == shannon
}

// -------------------- Test Options Structs --------------------

// testOptions contains all configuration options for the E2E tests
type testOptions struct {
	// Protocol to use for testing ("morse" or "shannon")
	//   - Required: set via TEST_PROTOCOL
	testProtocol protocolStr

	// URL for accessing the gateway
	//   - Default: "http://localhost:%s/v1" (%s = Docker container port)
	//   - If set via GATEWAY_URL_OVERRIDE, Docker is skipped and test runs against the provided URL directly
	gatewayURL string

	// True if gatewayURL was set via GATEWAY_URL_OVERRIDE (i.e., Docker is skipped)
	gatewayURLOverridden bool

	// Service ID override to test
	//   - If empty, test runs for all service IDs for the protocol
	serviceIDOverride protocol.ServiceID

	// Wait time in seconds for hydrator checks to complete
	//   - Default: 0 (no wait)
	//   - Set via WAIT_FOR_HYDRATOR
	waitForHydrator int

	// Docker-related configuration options
	docker dockerOptions

	// Config file path template
	//   - Format: "./.%s.config.yaml" (%s = protocol name)
	configPathTemplate string
}

// dockerOptions contains configuration for the Docker test container
type dockerOptions struct {
	// Log docker container output
	//   - Default: false
	//   - Enable with DOCKER_LOG=true
	logOutput bool

	// Force rebuild of the docker image
	//   - Default: false
	//   - Enable with DOCKER_FORCE_REBUILD=true
	forceRebuild bool
}

// -------------------- Gather Test Options --------------------

// gatherTestOptions collects all test configuration options from environment variables
func gatherTestOptions() testOptions {
	// Set default values
	options := testOptions{
		gatewayURL:         "http://localhost:%s/v1", // e.g., "http://localhost:3069/v1"
		configPathTemplate: "./.%s.config.yaml",      // e.g., "./.morse.config.yaml"
		waitForHydrator:    0,
	}

	// --- Required: TEST_PROTOCOL ---
	testProtocol := protocolStr(os.Getenv(envTestProtocol))
	switch {
	case testProtocol == "":
		panic(fmt.Sprintf("%s environment variable is not set", envTestProtocol))
	case !testProtocol.isValid():
		panic(fmt.Sprintf("%s environment variable is not set to `morse` or `shannon`", envTestProtocol))
	default:
		options.testProtocol = testProtocol
	}

	// --- Optional: GATEWAY_URL_OVERRIDE ---
	if gatewayURLOverride := os.Getenv(envGatewayURLOverride); gatewayURLOverride != "" {
		options.gatewayURL = gatewayURLOverride
		options.gatewayURLOverridden = true
	}

	// --- Optional: SERVICE_ID_OVERRIDE ---
	if serviceIDOverride := os.Getenv(envServiceIDOverride); serviceIDOverride != "" {
		options.serviceIDOverride = protocol.ServiceID(serviceIDOverride)
	}

	// --- Optional: WAIT_FOR_HYDRATOR ---
	if waitTimeStr := os.Getenv(envWaitForHydrator); waitTimeStr != "" {
		if waitTime, err := strconv.Atoi(waitTimeStr); err == nil {
			options.waitForHydrator = waitTime
		}
	}

	// --- Docker: DOCKER_LOG ---
	if logValue := os.Getenv(envDockerLog); logValue != "" {
		if logParsed, err := strconv.ParseBool(logValue); err == nil {
			options.docker.logOutput = logParsed
		}
	}

	// --- Docker: DOCKER_FORCE_REBUILD ---
	if rebuildValue := os.Getenv(envDockerForceRebuild); rebuildValue != "" {
		if rebuildParsed, err := strconv.ParseBool(rebuildValue); err == nil {
			options.docker.forceRebuild = rebuildParsed
		}
	}

	return options
}
