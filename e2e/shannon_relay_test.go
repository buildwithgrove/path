//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

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
			// gatewaye2e is a service created for e2e tests: it is supported by a
			// single endpoint, maintained by Grove.
			name:         "should successfully relay eth_blockNumber for gatewaye2e",
			reqMethod:    http.MethodPost,
			reqPath:      "/v1",
			serviceAlias: "test-service",
			relayID:      "1001",
			body:         `{"jsonrpc": "2.0", "id": "1001", "method": "eth_blockNumber"}`,
		},
		{
			name:         "should successfully relay eth_chainId for gatewaye2e",
			reqMethod:    http.MethodPost,
			reqPath:      "/v1",
			serviceAlias: "test-service",
			relayID:      "1002",
			body:         `{"jsonrpc": "2.0", "id": "1002", "method": "eth_chainId"}`,
		},
		{
			name:         "should successfully relay eth_blockNumber for eth-mainnet (0021)",
			reqMethod:    http.MethodPost,
			reqPath:      "/v1",
			serviceAlias: "etherium-mainnet",
			relayID:      "1101",
			body:         `{"jsonrpc": "2.0", "id": "1101", "method": "eth_blockNumber"}`,
		},

		// TODO_UPNEXT(@adshmh): add more test cases with valid and invalid jsonrpc request payloads.
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			// eg. fullURL = "http://test-service.localdev.me:55006/v1"
			fullURL := fmt.Sprintf("http://%s.%s:%s%s", test.serviceAlias, localdevMe, pathPort, test.reqPath)

			client := &http.Client{}

			// Send a service request to the PATH container running in Docker.
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
