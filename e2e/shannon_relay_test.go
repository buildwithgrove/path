//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	// shannonConfigFile is the name of the config file under "e2e" directory, that contains
	// the config for a PATH instance that uses Shannon as the relaying protocol.
	shannonConfigFile = ".shannon.config.yaml"
)

func Test_ShannonRelay(t *testing.T) {
	// Start an instance of PATH using the E2E config file for Shannon.
	pathContainerPort, teardownFn := setupPathInstance(t, shannonConfigFile)
	defer teardownFn()

	tests := []struct {
		name      string
		reqMethod string
		serviceID string
		relayID   string
		body      string
	}{
		{
			// anvil is a service created for e2e tests: it is supported by a
			// single endpoint, maintained by Grove.
			name:      "should successfully relay eth_blockNumber for anvil",
			reqMethod: http.MethodPost,
			relayID:   "1001",
			body:      `{"jsonrpc": "2.0", "id": "1001", "method": "eth_blockNumber"}`,
		},
		{
			name:      "should successfully relay eth_chainId for anvil",
			reqMethod: http.MethodPost,
			relayID:   "1002",
			body:      `{"jsonrpc": "2.0", "id": "1002", "method": "eth_chainId"}`,
		},
		// TODO_UPNEXT(@adshmh): add more test cases with valid and invalid jsonrpc request payloads.
	}

	reqPath := "/v1/abcdef12"
	serviceAlias := "anvil"
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			// eg. fullURL = "http://anvil.localdev.me:55006/v1/abcdef12"
			fullURL := fmt.Sprintf("http://%s.%s:%s%s", serviceAlias, localdevMe, pathContainerPort, reqPath)

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
