//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// Anvil JSON-RPC method definitions.
// These are standard Ethereum methods that require no parameters.
var anvilExpectedID = jsonrpc.IDFromInt(1)

const (
	anvilBlockNumber = "eth_blockNumber"
	anvilChainId     = "eth_chainId"
	anvilNetVersion  = "net_version"
)

// getAnvilTestMethods returns Anvil JSON-RPC methods for load testing.
// All methods use empty parameters for simplicity.
func getAnvilTestMethods() []string {
	return []string{
		anvilBlockNumber,
		anvilChainId,
		anvilNetVersion,
	}
}

// getAnvilVegetaTargets creates HTTP targets for Anvil load testing.
// Performs basic health check by fetching current block number.
func getAnvilVegetaTargets(
	ts *TestService,
	methods []string,
	gatewayURL string,
) (map[string]vegeta.Target, error) {
	headers := getRequestHeaders(ts.ServiceID)

	// Verify Anvil is responding and get current block number
	blockNumber, err := getAnvilBlockNumber(ts, headers, gatewayURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get Anvil block number: %w", err)
	}
	ts.ServiceParams.blockNumber = blockNumber

	targets := make(map[string]vegeta.Target)
	for _, method := range methods {
		// Create JSON-RPC request - Anvil test methods use no parameters
		jsonrpcReq := jsonrpc.Request{
			JSONRPC: jsonrpc.Version2,
			ID:      jsonrpc.IDFromInt(1),
			Method:  jsonrpc.Method(method),
			// Params intentionally omitted - all Anvil test methods use empty params
		}

		body, err := json.Marshal(jsonrpcReq)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON-RPC request for method %s: %w", method, err)
		}

		target := vegeta.Target{
			Method: http.MethodPost,
			URL:    gatewayURL,
			Body:   body,
			Header: headers,
		}

		targets[method] = target
	}

	return targets, nil
}

// getAnvilBlockNumber fetches current block number from Anvil.
// Serves as basic health check to verify server is responding.
func getAnvilBlockNumber(_ *TestService, headers http.Header, gatewayURL string) (string, error) {
	blockNumber, err := fetchAnvilBlockNumber(gatewayURL, headers)
	if err != nil {
		return "", fmt.Errorf("could not get current block number: %v", err)
	}
	return strconv.FormatUint(blockNumber, 10), nil
}

// fetchAnvilBlockNumber makes single request to get current block number.
// Returns error if Anvil is not responding or returns invalid data.
func fetchAnvilBlockNumber(gatewayURL string, headers http.Header) (uint64, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := buildAnvilBlockNumberRequest(gatewayURL, headers)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	var jsonRPC jsonrpc.Response
	if err := json.NewDecoder(resp.Body).Decode(&jsonRPC); err != nil {
		return 0, err
	}

	if jsonRPC.Error != nil {
		return 0, fmt.Errorf("JSON-RPC error: %v", jsonRPC.Error)
	}

	// Parse hex block number (e.g., "0x1a" -> 26)
	blockHex, ok := jsonRPC.Result.(string)
	if !ok {
		return 0, fmt.Errorf("unexpected result type: %T", jsonRPC.Result)
	}

	blockNum, err := strconv.ParseUint(blockHex[2:], 16, 64) // Remove "0x" prefix
	if err != nil {
		return 0, fmt.Errorf("failed to parse block number %s: %w", blockHex, err)
	}

	return blockNum, nil
}

// buildAnvilBlockNumberRequest creates JSON-RPC request for eth_blockNumber.
func buildAnvilBlockNumberRequest(gatewayURL string, headers http.Header) (*http.Request, error) {
	blockNumberReq := jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(1),
		Method:  anvilBlockNumber,
	}

	blockNumberReqBytes, err := json.Marshal(blockNumberReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, gatewayURL, bytes.NewReader(blockNumberReqBytes))
	if err != nil {
		return nil, err
	}

	req.Header = headers.Clone()
	return req, nil
}

// parseAnvilBlockNumber safely converts block number string to uint64.
// Returns 0 for empty strings or "latest".
func parseAnvilBlockNumber(blockStr string) uint64 {
	if blockStr == "" || blockStr == "latest" {
		return 0
	}

	blockNum, err := strconv.ParseUint(blockStr, 10, 64)
	if err != nil {
		return 0
	}

	return blockNum
}
