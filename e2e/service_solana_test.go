//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/buildwithgrove/path/qos/jsonrpc"
	vegeta "github.com/tsenart/vegeta/lib"
)

/* -------------------- Solana JSON-RPC Method Definitions -------------------- */

// Reference for all Solana JSON-RPC methods:
// - https://solana.com/docs/rpc/http
const (
	getEpochInfo            jsonrpc.Method = "getEpochInfo"
	getHealth               jsonrpc.Method = "getHealth"
	getBalance              jsonrpc.Method = "getBalance"
	getGenesisHash          jsonrpc.Method = "getGenesisHash"
	getSignaturesForAddress jsonrpc.Method = "getSignaturesForAddress"
	getSlot                 jsonrpc.Method = "getSlot"
	getTransaction          jsonrpc.Method = "getTransaction"
	getBlock                jsonrpc.Method = "getBlock"
)

// getSolanaTestMethods returns all Solana JSON-RPC methods for a service load test.
func getSolanaTestMethods() []jsonrpc.Method {
	return []jsonrpc.Method{
		getEpochInfo,
		getHealth,
		getBalance,
		getGenesisHash,
		getSignaturesForAddress,
		getSlot,
		getTransaction,
		getBlock,
	}
}

// createSolanaJsonRPCParams builds RPC params for each Solana method using the provided service parameters.
func createSolanaJsonRPCParams(method jsonrpc.Method, sp ServiceParams) jsonrpc.Params {
	switch method {
	// Methods with empty params
	//   Example: []
	case getEpochInfo, getHealth, getSlot, getGenesisHash:
		return jsonrpc.Params{}

	// Methods that just need the transaction signature
	//   Example: ["5UHzPFpc8Gy7LGtEPcuWWZaHRRLzpZQQiVPZyK7H1yvez5F2FBfmHmRo3WzaWuBzKZgPm4ULxh8H6Ha1YJwvLaza"]
	case getTransaction:
		params, _ := jsonrpc.BuildParamsFromString(sp.TransactionHash)
		return params

	// Methods that need account address and commitment
	//   Example: ["83astBRguLMdt2h5U1Tpdq5tjFoJ6noeGwaY3mDLVcri", {"commitment": "finalized"}]
	case getBalance:
		params, _ := jsonrpc.BuildParamsFromStringAndObject(
			sp.ContractAddress,
			map[string]any{
				"commitment": "finalized",
			},
		)
		return params

		// getSignaturesForAddress needs [address, options]
		//   Example: ["Vote111111111111111111111111111111111111111", {"limit": 10}]
	case getSignaturesForAddress:
		params, _ := jsonrpc.BuildParamsFromStringAndObject(
			sp.ContractAddress,
			map[string]any{
				"limit": 10,
			},
		)
		return params

	// getBlock needs [slot, options]
	//   Example: [430, {"encoding": "json", "transactionDetails": "full", "maxSupportedTransactionVersion": 0}]
	case getBlock:
		slotNumber, err := strconv.ParseUint(sp.blockNumber, 10, 64)
		if err != nil {
			// Use slot 0 if parsing fails
			slotNumber = 0
		}
		params, _ := jsonrpc.BuildParamsFromUint64AndObject(
			slotNumber,
			map[string]any{
				"encoding":                       "json",
				"transactionDetails":             "none",
				"maxSupportedTransactionVersion": 0,
				"rewards":                        false,
			},
		)
		return params

	default:
		return jsonrpc.Params{}
	}
}

func getSolanaVegetaTargets(
	ts *TestService,
	methods []jsonrpc.Method,
	gatewayURL string,
) ([]vegeta.Target, error) {
	headers := getRequestHeaders(ts.ServiceID)

	// Get the appropriate block number for the test
	blockNumber, err := getSolanaBlockNumber(ts, headers, gatewayURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get Solana block number: %w", err)
	}
	ts.ServiceParams.blockNumber = blockNumber

	targets := make([]vegeta.Target, 0, len(methods))
	for _, method := range methods {
		// Create JSON-RPC request with appropriate parameters
		jsonrpcReq := jsonrpc.Request{
			JSONRPC: jsonrpc.Version2,
			ID:      jsonrpc.IDFromInt(1),
			Method:  method,
			Params:  createSolanaJsonRPCParams(method, ts.ServiceParams),
		}

		// Marshal the request body
		body, err := json.Marshal(jsonrpcReq)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON-RPC request for method %s: %w", method, err)
		}

		// Create vegeta target
		target := vegeta.Target{
			Method: http.MethodPost,
			URL:    gatewayURL,
			Body:   body,
			Header: headers,
		}

		targets = append(targets, target)
	}

	return targets, nil
}

// -----------------------------------------------------------------------------
// Get Test Block Number helpers - Used for Solana services
//
// TODO_IMPROVE(@commoddity): Add archival support once Solana archival QoS checks are implemented.
// -----------------------------------------------------------------------------

// getEpochInfoResponse represents the response from getEpochInfo
type getEpochInfoResponse struct {
	AbsoluteSlot     uint64 `json:"absoluteSlot"`
	BlockHeight      uint64 `json:"blockHeight"`
	Epoch            uint64 `json:"epoch"`
	SlotIndex        uint64 `json:"slotIndex"`
	SlotsInEpoch     uint64 `json:"slotsInEpoch"`
	TransactionCount uint64 `json:"transactionCount"`
}

func getSolanaBlockNumber(_ *TestService, headers http.Header, gatewayURL string) (string, error) {
	// Get the current slot number
	slotNumber, err := getCurrentSlotNumber(gatewayURL, headers)
	if err != nil {
		return "", fmt.Errorf("Could not get current slot number: %v", err)
	}
	return strconv.FormatUint(slotNumber, 10), nil
}

// getCurrentSlotNumber gets current slot number with consensus from multiple requests.
func getCurrentSlotNumber(gatewayURL string, headers http.Header) (uint64, error) {
	// Track frequency of each slot number seen
	slotNumbers := make(map[uint64]int)
	maxAttempts := 10
	requiredAgreement := 3
	client := &http.Client{Timeout: 5 * time.Second}

	// Make multiple attempts to get consensus
	for i := 0; i < maxAttempts; i++ {
		slotNum, err := fetchSlotNumber(client, gatewayURL, headers)
		if err != nil {
			continue
		}

		// Update consensus tracking
		slotNumbers[slotNum]++
		if slotNumbers[slotNum] >= requiredAgreement {
			return slotNum, nil
		}
	}

	// If we get here, we didn't reach consensus
	return 0, fmt.Errorf("failed to reach consensus on slot number after %d attempts", maxAttempts)
}

// fetchSlotNumber makes a single request to get the current slot number using getEpochInfo.
func fetchSlotNumber(client *http.Client, gatewayURL string, headers http.Header) (uint64, error) {
	// Build and send request
	req, err := buildEpochInfoRequest(gatewayURL, headers)
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

	// Parse JSONRPC response
	var jsonRPC jsonrpc.Response
	if err := json.NewDecoder(resp.Body).Decode(&jsonRPC); err != nil {
		return 0, err
	}

	// Marshal the result back to JSON so we can unmarshal it into our struct
	resultBytes, err := json.Marshal(jsonRPC.Result)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal result: %v", err)
	}

	// Unmarshal into getEpochInfoResponse
	var epochInfo getEpochInfoResponse
	if err := json.Unmarshal(resultBytes, &epochInfo); err != nil {
		return 0, fmt.Errorf("failed to unmarshal epoch info: %v", err)
	}

	return epochInfo.AbsoluteSlot, nil
}

// buildEpochInfoRequest creates a JSON-RPC request for getEpochInfo.
func buildEpochInfoRequest(gatewayURL string, headers http.Header) (*http.Request, error) {
	epochInfoReq := jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(1),
		Method:  getEpochInfo,
	}

	epochInfoReqBytes, err := json.Marshal(epochInfoReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, gatewayURL, bytes.NewReader(epochInfoReqBytes))
	if err != nil {
		return nil, err
	}

	req.Header = headers.Clone()

	return req, nil
}
