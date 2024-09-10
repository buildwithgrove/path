package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/stretchr/testify/require"
)

// localdev.me is a hosted domain that resolves to 127.0.0.1 (localhost).
// This allows a subdomain to be specified without modifying /etc/hosts.
// It is hosted by AWS. See https://codeengineered.com/blog/2022/localdev-me/
const localdevMe = "localdev.me"

// When the ephemeral PATH Docker container is running it exposes a dynamically
// assigned port. This global variable is used to capture the port number.
var pathPort string

func TestMain(m *testing.M) {
	// Initialize the ephemeral PATH Docker container
	pool, resource, containerPort := setupPathDocker()

	// Assign the port the container is listening on to the global variable
	pathPort = containerPort

	// Run PATH E2E Shannon relay tests
	exitCode := m.Run()

	// Cleanup the ephemeral PATH Docker container
	cleanupPathDocker(m, pool, resource)

	// Exit with the test result
	os.Exit(exitCode)
}

// TODO_IMPROVE: use gocuke (github.com/regen-network/gocuke) for defining and running E2E tests.
func Test_ShannonRelay(t *testing.T) {
	tests := []struct {
		name         string
		reqMethod    string
		reqPath      string
		serviceID    string
		serviceAlias string
		relayID      string
		body         string
	}{
		{
			name:         "should successfully relay eth_blockNumber for eth-mainnet (0021)",
			reqMethod:    http.MethodPost,
			reqPath:      "/v1",
			serviceID:    "gatewaye2e",
			serviceAlias: "test-service",
			relayID:      "1001",
			body:         `{"jsonrpc": "2.0", "id": "1001", "method": "eth_blockNumber"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			// eg. fullURL = "http://test-service.localdev.me:55006/v1"
			fullURL := fmt.Sprintf("http://%s.%s:%s%s", test.serviceAlias, localdevMe, pathPort, test.reqPath)

			client := &http.Client{}
			req, err := http.NewRequest(test.reqMethod, fullURL, bytes.NewBuffer([]byte(test.body)))
			c.NoError(err)
			req.Header.Set("Content-Type", "application/json")

			var success bool
			var allErrors []error
			for i := 0; i < 10; i++ {
				resp, err := client.Do(req)
				if err != nil {
					allErrors = append(allErrors, fmt.Errorf("request error: %v", err))
					continue
				}
				defer resp.Body.Close()

				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil {
					allErrors = append(allErrors, fmt.Errorf("response read error: %v", err))
					continue
				}

				err = validateJsonRpcResponse(test.relayID, bodyBytes)
				if err != nil {
					allErrors = append(allErrors, fmt.Errorf("validation error: %v --- %s", err, string(bodyBytes)))
					continue
				}

				success = true
				break
			}

			if !success {
				for _, err := range allErrors {
					fmt.Println(err)
				}
			}

			// Assert that one relay request was successful.
			c.True(success)
		})
	}
}

// TODO_TECHDEBT: delete (NOT MOVE) this function and implement a proper JSONRPC validator in the service package.
//
// DO NOT use this function either directly or as a base/guide for general JSONRPC validation.
// The sole purpose of this function is to check whether the relay response received from an endpoint
// looks like a valid JSONRPC response.
// This is a very rudimentary validatior that can only be used when the outgoing
// JSONRPC request is limited to a few special cases, e.g. in the E2E tests.
func validateJsonRpcResponse(expectedID string, response []byte) error {
	type jsonRpcResponse struct {
		JsonRpc string `json:"jsonrpc"`
		// TODO_TECHDEBT: ID field can contain other values. We are using a string here because
		// the E2E tests use a string ID for relays that are sent.
		// Proper JSONRPC validation requires referencing the ID field against the relay request on both type and value.
		ID     string `json:"id"`
		Result string `json:"result"`
	}

	var parsedResponse jsonRpcResponse
	if err := json.Unmarshal(response, &parsedResponse); err != nil {
		return err
	}

	if parsedResponse.JsonRpc != "2.0" {
		return fmt.Errorf("invalid JSONRPC field, expected %q, got %q", "2.0", parsedResponse.JsonRpc)
	}

	if parsedResponse.ID != expectedID {
		return fmt.Errorf("expected ID %q, got %q", expectedID, parsedResponse.ID)
	}

	if len(parsedResponse.Result) == 0 {
		return errors.New("empty Result field")
	}

	return nil
}

/* -------------------- Dockertest Ephemeral PATH Container Setup -------------------- */

const (
	containerName        = "path"
	internalPathPort     = "3000"
	dockerfilePath       = "../Dockerfile"
	configFilePath       = "./.config.test.yaml"
	configMountPoint     = ":/app/.config.yaml"
	containerEnvImageTag = "IMAGE_TAG=test"
	containerExtraHost   = "host.docker.internal:host-gateway"
	timeoutSeconds       = 120
)

var (
	containerConfigMount     = filepath.Join(os.Getenv("PWD"), configFilePath) + configMountPoint
	containerPortAndProtocol = internalPathPort + "/tcp"
)

func setupPathDocker() (*dockertest.Pool, *dockertest.Resource, string) {
	opts := &dockertest.RunOptions{
		Name:         containerName,
		Mounts:       []string{containerConfigMount},
		Env:          []string{containerEnvImageTag},
		ExposedPorts: []string{containerPortAndProtocol},
		ExtraHosts:   []string{containerExtraHost},
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		fmt.Printf("Could not construct pool: %s", err)
		os.Exit(1)
	}
	resource, err := pool.BuildAndRunWithOptions(dockerfilePath, opts, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		fmt.Printf("Could not start resource: %s", err)
		os.Exit(1)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			fmt.Printf("exit signal %d received\n", sig)
			if err := pool.Purge(resource); err != nil {
				fmt.Printf("could not purge resource: %s", err)
			}
		}
	}()

	if err := resource.Expire(timeoutSeconds); err != nil {
		fmt.Printf("[ERROR] Failed to set expiration on docker container: %v", err)
		os.Exit(1)
	}

	healthCheckURL := fmt.Sprintf("http://%s/healthz", resource.GetHostPort(containerPortAndProtocol))

	poolRetryChan := make(chan struct{}, 1)
	retryConnectFn := func() error {
		resp, err := http.Get(healthCheckURL)
		if err != nil {
			return fmt.Errorf("unable to connect to health check endpoint: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("health check endpoint returned non-200 status: %d", resp.StatusCode)
		}

		poolRetryChan <- struct{}{}
		return nil
	}
	if err = pool.Retry(retryConnectFn); err != nil {
		fmt.Printf("could not connect to docker: %s", err)
		os.Exit(1)
	}

	<-poolRetryChan

	return pool, resource, resource.GetPort(containerPortAndProtocol)
}

func cleanupPathDocker(_ *testing.M, pool *dockertest.Pool, resource *dockertest.Resource) {
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %s", err)
	}
}
