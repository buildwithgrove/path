//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"github.com/stretchr/testify/require"

	"github.com/pokt-foundation/portal-middleware/config"
	"github.com/pokt-foundation/portal-middleware/relayer"
	"github.com/pokt-foundation/portal-middleware/relayer/shannon"
	"github.com/pokt-foundation/portal-middleware/request"
)

const configPath = ".config.test.yaml"

var testTimeout = 20 * time.Second

// TODO_INCOMPLETE: add an action to the CI for running E2E tests, which at the minimum
// includes using the e2e tag and defining secrets to be used as environment variables, e.g. gateway/signer private key and address.
//
// TODO_IMPROVE: use gocuke (github.com/regen-network/gocuke) for defining and running E2E tests.
func TestShannonRelay(t *testing.T) {
	tests := []struct {
		name      string
		serviceID relayer.ServiceID
		relayID   string
		httpReq   *http.Request
	}{
		{
			name:      "should successfully relay eth_blockNumber for eth-mainnet (0021)",
			serviceID: "gatewaye2e",
			relayID:   "1001",
			httpReq: &http.Request{
				Method: http.MethodPost,
				Host:   "test-service.gateway.pokt.network",
				URL:    &url.URL{Path: "/v1"},
				Header: http.Header{"Content-Type": []string{"application/json"}},
				Body: io.NopCloser(bytes.NewBuffer([]byte(
					`{"jsonrpc": "2.0", "id": "1001", "method": "eth_blockNumber"}`,
				))),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			logger := polyzero.NewLogger()

			// Initialize all required components.
			config, err := config.LoadGatewayConfigFromYAML(configPath)
			c.NoError(err)

			requestParser, err := request.NewParser(config, logger)
			c.NoError(err)

			ctx := test.httpReq.Context()

			serviceID, qosService, err := requestParser.GetQoSService(ctx, test.httpReq)
			c.NoError(err)

			// service ID parsed from alias should match the test service ID
			c.Equal(test.serviceID, serviceID)

			gatewayRelayer, err := getTestRelayer(serviceID, config, logger)
			c.NoError(err)

			// Prepare the relay payload.
			payload, err := qosService.ParseHTTPRequest(ctx, test.httpReq)
			c.NoError(err)

			// Send the relay request to the gateway.
			// The relay is attempted 10 times before failing.
			// TODO_TECHDEBT: improve the error handling and retry mechanism.
			var success bool
			var allErrors []error
			for i := 0; i < 10; i++ {
				response, err := gatewayRelayer.SendRelay(ctx, test.serviceID, payload, randomEndpointSelector{})
				if err != nil {
					allErrors = append(allErrors, err)
					continue
				}

				// TODO_TECHDEBT: use the service package to parse and validate the response.
				err = validateJsonRpcResponse(test.relayID, response.Bytes)
				if err != nil {
					allErrors = append(allErrors, fmt.Errorf("validation error: %v --- %s", err, string(response.Bytes)))
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

// getTestRelayer initializes a shannon fullnode from the config, initializes the protocol,
// and waits for onchain data to be fetched. It will timeout if the protocol does not
// fetch onchain data within the test timeout period to avoid hanging the test.
func getTestRelayer(serviceID relayer.ServiceID, config config.GatewayConfig, logger polylog.Logger) (relayer.Relayer, error) {
	shannonConfig := config.GetShannonConfig()
	if shannonConfig == nil {
		return relayer.Relayer{}, fmt.Errorf("shannon config not found")
	}

	fullNode, err := shannon.NewFullNode(shannonConfig.FullNodeConfig, logger)
	if err != nil {
		return relayer.Relayer{}, err
	}

	protocol, err := shannon.NewProtocol(context.Background(), fullNode)
	if err != nil {
		return relayer.Relayer{}, err
	}

	// Wait for onchain data to be fetched by the initialized shannon protocol.
	startTime := time.Now()
	cacheHasEndpoints := false

	for !cacheHasEndpoints {

		endpoints, err := protocol.Endpoints(serviceID)
		cacheHasEndpoints = len(endpoints) > 0 && err == nil

		if time.Since(startTime) > testTimeout {
			return relayer.Relayer{}, fmt.Errorf("timeout waiting for protocol data")
		}
		<-time.After(500 * time.Millisecond)
	}

	return relayer.Relayer{Protocol: protocol}, nil
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

// randomEndpointSelector is used to fulfill the relayer's requirement for an
// EndpointSelector, through random selection of an endpoint among available ones.
var _ relayer.EndpointSelector = randomEndpointSelector{}

type randomEndpointSelector struct{}

func (r randomEndpointSelector) Select(allEndpoints map[relayer.AppAddr][]relayer.Endpoint) (relayer.AppAddr, relayer.EndpointAddr, error) {
	if len(allEndpoints) == 0 {
		return "", "", fmt.Errorf("endpointSelector: no endpoint available")
	}

	// return the first app from the list, and a random endpoint matching the app.
	for appAddr, endpoints := range allEndpoints {
		if len(endpoints) == 0 {
			return "", "", fmt.Errorf("endpointSelector: no endpoints found for app %s", appAddr)
		}

		return appAddr, endpoints[rand.Intn(len(endpoints))].Addr(), nil
	}

	return "", "", fmt.Errorf("endpointSelector: could not find any endpoints")
}
