//go:build e2e

package e2e

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	// morseConfigFile is the name of the config file under "e2e" directory, that contains
	// the config for a PATH instance that uses Shannon as the relaying protocol.
	morseConfigFile = ".morse.config.yaml"
)

func Test_MorseRelay(t *testing.T) {
	// Start an instance of PATH using the E2E config file for Shannon.
	pathContainerPort, teardownFn := setupPathInstance(t, morseConfigFile)
	defer teardownFn()

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
			name:         "should successfully relay eth_chainId for eth-mainnet (0021)",
			reqMethod:    http.MethodPost,
			reqPath:      "/v1",
			serviceAlias: "eth-mainnet",
			relayID:      "1201",
			body:         `{"jsonrpc": "2.0", "id": "1201", "method": "eth_chainId"}`,
		},
		{
			name:         "should successfully relay eth_blockNumber for eth-mainnet (0021)",
			reqMethod:    http.MethodPost,
			reqPath:      "/v1",
			serviceAlias: "eth-mainnet",
			relayID:      "1202",
			body:         `{"jsonrpc": "2.0", "id": "1202", "method": "eth_blockNumber"}`,
		},

		// TODO_UPNEXT(@adshmh): add more test cases with valid and invalid jsonrpc request payloads.
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			// eg. fullURL = "http://test-service.localdev.me:55006/v1"
			fullURL := fmt.Sprintf("http://%s.%s:%s%s", test.serviceAlias, localdevMe, pathContainerPort, test.reqPath)

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
