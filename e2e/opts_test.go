//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"strconv"

	"github.com/buildwithgrove/path/protocol"
)

// NOTE: The types and constants in this file are also referenced in evm_test.go.
// Since they are in the same package, this causes linter warnings but not Go compilation errors.
// This file contains all the configuration options and their documentation,
// while evm_test.go only keeps a reference to the global options variable.

/* -------------------- Test Configuration Options -------------------- */

// Environment variable names
const (
	// Required environment variables
	envTestProtocol = "TEST_PROTOCOL"

	// Optional environment variables
	envGatewayURLOverride = "GATEWAY_URL_OVERRIDE"
	envServiceIDOverride  = "SERVICE_ID_OVERRIDE"
	envDockerLog          = "DOCKER_LOG"
	envDockerForceRebuild = "DOCKER_FORCE_REBUILD"
	envWaitForHydrator    = "WAIT_FOR_HYDRATOR"
)

// protocolStr is a type to determine whether to test PATH with Morse or Shannon
type protocolStr string

const (
	morse   protocolStr = "morse"
	shannon protocolStr = "shannon"
)

func (p protocolStr) isValid() bool {
	return p == morse || p == shannon
}

// testOptions contains all configuration options for the E2E tests
type (
	testOptions struct {
		// Protocol to use for testing (morse or shannon)
		// Required environment variable: TEST_PROTOCOL
		testProtocol protocolStr

		// URL for accessing the gateway
		// If not set, default is "http://localhost:%s/v1" where %s is the port of the Docker container
		// If set via GATEWAY_URL_OVERRIDE, the Docker container won't be used and
		// the test will run against the provided URL directly
		gatewayURL string

		// Whether the gateway URL was explicitly set via GATEWAY_URL_OVERRIDE
		// This also indicates that no Docker container should be started
		//
		// If GATEWAY_URL_OVERRIDE is set, we'll use the provided URL directly and skip starting a Docker container,
		// assuming PATH is already running externally at the provided URL.
		gatewayURLOverridden bool

		// The specific service ID override to test
		// If not set, the test will run for all service IDs for the protocol
		serviceIDOverride protocol.ServiceID

		// Wait time in seconds for hydrator checks to complete
		// If not set, default is 0 (no wait)
		// Can be set via WAIT_FOR_HYDRATOR env var
		waitForHydrator int

		// Docker-related configuration options
		docker dockerOptions

		// Config file path template
		// Format: "./.%s.config.yaml" where %s is the protocol name
		configPathTemplate string
	}
	// dockerOptions contains configuration for the Docker test container
	dockerOptions struct {
		// Whether to log docker container output
		// Default: false
		// Can be enabled with DOCKER_LOG=true
		logOutput bool

		// Whether to force rebuild of the docker image
		// Default: false
		// Can be enabled with DOCKER_FORCE_REBUILD=true
		forceRebuild bool
	}
)

// gatherTestOptions collects all test configuration options from environment variables
func gatherTestOptions() testOptions {
	// Default values
	options := testOptions{
		gatewayURL:         "http://localhost:%s/v1", // eg. `http://localhost:3069/v1`
		configPathTemplate: "./.%s.config.yaml",      // eg. `./.morse.config.yaml` or `./.shannon.config.yaml`
		waitForHydrator:    0,                        // Default: no wait
	}

	// Required environment variables
	if testProtocol := protocolStr(os.Getenv(envTestProtocol)); testProtocol == "" {
		panic(fmt.Sprintf("%s environment variable is not set", envTestProtocol))
	} else if !testProtocol.isValid() {
		panic(fmt.Sprintf("%s environment variable is not set to `morse` or `shannon`", envTestProtocol))
	} else {
		options.testProtocol = testProtocol
	}

	// Optional environment variables
	if gatewayURLOverride := os.Getenv(envGatewayURLOverride); gatewayURLOverride != "" {
		options.gatewayURL = gatewayURLOverride
		options.gatewayURLOverridden = true
	}

	// Optional environment variable to override the service ID to test
	if serviceIDOverride := os.Getenv(envServiceIDOverride); serviceIDOverride != "" {
		options.serviceIDOverride = protocol.ServiceID(serviceIDOverride)
	}

	// Optional environment variable for hydrator wait time
	if waitTimeStr := os.Getenv(envWaitForHydrator); waitTimeStr != "" {
		if waitTime, err := strconv.Atoi(waitTimeStr); err == nil {
			options.waitForHydrator = waitTime
		}
	}

	// Docker configuration
	if logValue := os.Getenv(envDockerLog); logValue != "" {
		if logParsed, err := strconv.ParseBool(logValue); err == nil {
			options.docker.logOutput = logParsed
		}
	}

	if rebuildValue := os.Getenv(envDockerForceRebuild); rebuildValue != "" {
		if rebuildParsed, err := strconv.ParseBool(rebuildValue); err == nil {
			options.docker.forceRebuild = rebuildParsed
		}
	}

	return options
}
