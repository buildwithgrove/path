//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/buildwithgrove/path/request"
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
		name      string
		reqMethod string
		serviceID protocol.ServiceID
		relayID   string
		body      string
	}{
		{
			name:      "should successfully relay eth_chainId for eth (F00C)",
			reqMethod: http.MethodPost,
			serviceID: "F00C",
			relayID:   "1201",
			body:      `{"jsonrpc": "2.0", "id": "1201", "method": "eth_chainId"}`,
		},
		{
			name:      "should successfully relay eth_blockNumber for eth (F00C)",
			reqMethod: http.MethodPost,
			serviceID: "F00C",
			relayID:   "1202",
			body:      `{"jsonrpc": "2.0", "id": "1202", "method": "eth_blockNumber"}`,
		},

		// TODO_MVP(@adshmh): add more test cases with valid and invalid jsonrpc request payloads.
		// TODO_MVP(@adshmh): add the following test cases for Solana:
		//	1. Valid GetEpochInfo request
		//	2. Invalid GetEpochInfo request
		//      3. Valid generic request
		//      4. Invalid generic request
	}

	// Request path is arbitrary, as it is not current used by PATH.
	// It is here only to ensure all paths following the `/v1` segment are valid.
	reqPath := "/v1/abcdef12"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			// eg. fullURL = "http://localdev.me:55006/v1/abcdef12"
			fullURL := fmt.Sprintf("http://%s:%s%s", localdevMe, pathContainerPort, reqPath)

			client := &http.Client{}

			// Send a service request to the PATH container running in Docker.
			req, err := http.NewRequest(test.reqMethod, fullURL, bytes.NewBuffer([]byte(test.body)))
			c.NoError(err)
			req.Header.Set("Content-Type", "application/json")

			// Assign the target service ID to the request header.
			req.Header.Set(request.HTTPHeaderTargetServiceID, string(test.serviceID))

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

				var parsedResponse jsonrpc.Response
				if err := json.Unmarshal(bodyBytes, &parsedResponse); err != nil {
					allErrors = append(allErrors, fmt.Errorf("response unmarshal error: %v --- %s", err, string(bodyBytes)))
					continue
				}

				if err := parsedResponse.Validate(jsonrpc.IDFromStr(test.relayID)); err != nil {
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
