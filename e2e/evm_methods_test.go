//go:build e2e

package e2e

import (
	"time"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

/* -------------------- EVM JSON-RPC Method Definitions -------------------- */

// Reference for all EVM JSON-RPC methods:
// - https://ethereum.org/en/developers/docs/apis/json-rpc/
const (
	eth_blockNumber           jsonrpc.Method = "eth_blockNumber"
	eth_call                  jsonrpc.Method = "eth_call"
	eth_getTransactionReceipt jsonrpc.Method = "eth_getTransactionReceipt"
	eth_getBlockByNumber      jsonrpc.Method = "eth_getBlockByNumber"
	eth_getBalance            jsonrpc.Method = "eth_getBalance"
	eth_chainId               jsonrpc.Method = "eth_chainId"
	eth_getTransactionCount   jsonrpc.Method = "eth_getTransactionCount"
	eth_getTransactionByHash  jsonrpc.Method = "eth_getTransactionByHash"
	eth_gasPrice              jsonrpc.Method = "eth_gasPrice"
)

// allEVMTestMethods returns all EVM JSON-RPC methods for a service load test.
func allEVMTestMethods() []jsonrpc.Method {
	return []jsonrpc.Method{
		eth_blockNumber,
		eth_call,
		eth_getTransactionReceipt,
		eth_getBlockByNumber,
		eth_getBalance,
		eth_chainId,
		eth_getTransactionCount,
		eth_getTransactionByHash,
		eth_gasPrice,
	}
}

type (
	// methodTestConfig contains all configuration and test requirements for a single method.
	methodTestConfig struct {
		requestLoadConfig
		successCriteria
	}

	// requestLoadConfig specifies the load configuration for a method to be tested.
	// - totalRequests: Total number of requests to send
	// - rps:           Requests per second
	requestLoadConfig struct {
		totalRequests int
		rps           int
	}

	// successCriteria contains the minimum success rate and maximum
	// latency requirements for a method to pass the load test.
	// - successRate:   Minimum success rate (0-1)
	// - maxP50Latency: Maximum P50 latency
	// - maxP95Latency: Maximum P95 latency
	// - maxP99Latency: Maximum P99 latency
	successCriteria struct {
		successRate   float64
		maxP50Latency time.Duration
		maxP95Latency time.Duration
		maxP99Latency time.Duration
	}
)

var (
	// defaultRequestLoadConfig contains the default configuration for a method.
	defaultRequestLoadConfig = requestLoadConfig{
		totalRequests: 100,
		rps:           10,
	}

	// defaultSuccessCriteria contains the default success rates and latency requirements for a method.
	defaultSuccessCriteria = successCriteria{
		successRate:   0.90,
		maxP50Latency: 1_000 * time.Millisecond, //
		maxP95Latency: 5_000 * time.Millisecond,
		maxP99Latency: 10_000 * time.Millisecond,
	}

	defaultTestConfigAllMethods = map[jsonrpc.Method]methodTestConfig{
		eth_blockNumber: {
			requestLoadConfig: defaultRequestLoadConfig,
			successCriteria:   defaultSuccessCriteria,
		},
		eth_call: {
			requestLoadConfig: defaultRequestLoadConfig,
			successCriteria:   defaultSuccessCriteria,
		},
		eth_getTransactionReceipt: {
			requestLoadConfig: defaultRequestLoadConfig,
			successCriteria:   defaultSuccessCriteria,
		},
		eth_getBlockByNumber: {
			requestLoadConfig: defaultRequestLoadConfig,
			successCriteria:   defaultSuccessCriteria,
		},
		eth_getBalance: {
			requestLoadConfig: defaultRequestLoadConfig,
			successCriteria:   defaultSuccessCriteria,
		},
		eth_chainId: {
			requestLoadConfig: defaultRequestLoadConfig,
			successCriteria:   defaultSuccessCriteria,
		},
		eth_getTransactionCount: {
			requestLoadConfig: defaultRequestLoadConfig,
			successCriteria:   defaultSuccessCriteria,
		},
		eth_getTransactionByHash: {
			requestLoadConfig: defaultRequestLoadConfig,
			successCriteria:   defaultSuccessCriteria,
		},
		eth_gasPrice: {
			requestLoadConfig: defaultRequestLoadConfig,
			successCriteria:   defaultSuccessCriteria,
		},
	}
)

// evmServiceParameters holds service-specific test data for all methods.
// Allows testing specific requests that require parameters.
//
// TODO_IMPROVE(@commoddity): allow reading this configuration from a YAML file
// Per-method configs are now embedded in each testCase in evm_test.go as methodConfigs.
type evmServiceParameters struct {
	// For eth_getBalance, eth_getTransactionCount, eth_getTransactionReceipt
	blockNumber string

	// For eth_getBalance, eth_getTransactionCount, eth_getTransactionReceipt
	//
	// `contractAddress` address should match the `evmArchivalCheckConfig.contractAddress`
	// value in `config/service_qos_config.go`
	contractAddress string

	// The minimum block number to use for archival tests.
	// Ensures we are not fetching a block where the contract address has no balance or transactions.
	//
	// `contractStartBlock` should match the `evmArchivalCheckConfig.contractStartBlock`
	// value in `config/service_qos_config.go`
	contractStartBlock uint64

	// For eth_getTransactionReceipt and eth_getTransactionByHash
	transactionHash string

	// For eth_call
	callData string
}

// createEVMJsonRPCParams builds RPC params for each EVM method using the provided service parameters.
func createEVMJsonRPCParams(
	method jsonrpc.Method,
	sp evmServiceParameters,
) jsonrpc.Params {
	switch method {

	// Methods with empty params
	case eth_blockNumber, eth_chainId, eth_gasPrice:
		return jsonrpc.Params{}

	// Methods that just need the transaction hash
	//   Example: ["0xfeccd627b5b391d04fe45055873de3b2c0b4302d52e96bd41d5f0019a704165f"]
	case eth_getTransactionReceipt, eth_getTransactionByHash:
		params, _ := jsonrpc.BuildParamsFromString(sp.transactionHash)
		return params

	// Methods that need [address, blockNumber]
	//   Example: ["0xdAC17F958D2ee523a2206206994597C13D831ec7", "latest"]
	case eth_getBalance, eth_getTransactionCount:
		params, _ := jsonrpc.BuildParamsFromStringArray([2]string{
			sp.contractAddress,
			sp.blockNumber,
		})
		return params

	// eth_getBlockByNumber needs [blockNumber, <boolean>]
	//   Example: ["0xe71e1d", false]
	case eth_getBlockByNumber:
		params, _ := jsonrpc.BuildParamsFromStringAndBool(
			sp.blockNumber,
			false,
		)
		return params

	// eth_call needs [{ to: address, data: calldata }, blockNumber]
	//   Example: [{"to":"0xdAC17F958D2ee523a2206206994597C13D831ec7","data":"0x18160ddd"}, "latest"]
	case eth_call:
		params, _ := jsonrpc.BuildParamsFromObjectAndString(
			map[string]string{
				"to":   sp.contractAddress,
				"data": sp.callData,
			},
			sp.blockNumber,
		)
		return params

	default:
		return jsonrpc.Params{}
	}
}
