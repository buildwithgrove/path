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

// runAllMethods returns all EVM JSON-RPC methods for a service load test.
func runAllMethods() []jsonrpc.Method {
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
	// methodDefinition contains all configuration and test requirements for a single method.
	methodDefinition struct {
		methodConfig
		methodSuccessRates
	}

	// methodConfig specifies the configuration for a method to be tested.
	// - totalRequests: Total number of requests to send
	// - rps:           Requests per second
	methodConfig struct {
		totalRequests int
		rps           int
	}

	// methodSuccessRates contains the minimum success rate and maximum
	// latency requirements for a method to pass the load test.
	// - successRate:   Minimum success rate (0-1)
	// - maxP50Latency: Maximum P50 latency
	// - maxP95Latency: Maximum P95 latency
	// - maxP99Latency: Maximum P99 latency
	methodSuccessRates struct {
		successRate   float64
		maxP50Latency time.Duration
		maxP95Latency time.Duration
		maxP99Latency time.Duration
	}
)

var (
	// defaultMethodConfig contains the default configuration for a method.
	defaultMethodConfig = methodConfig{
		totalRequests: 10,
		rps:           1,
	}

	// defaultMethodSuccessRates contains the default success rates and latency requirements for a method.
	defaultMethodSuccessRates = methodSuccessRates{
		successRate:   0.95,
		maxP50Latency: 3_500 * time.Millisecond,
		maxP95Latency: 9_000 * time.Millisecond,
		maxP99Latency: 30_000 * time.Millisecond,
	}
)

// TODO_IMPROVE(@commoddity): allow reading this configuration from a YAML file
//
// methodDefinitions contains all method definitions for a service load test.
// Allows customizing the configuration for each method as desired.
var methodDefinitions = map[jsonrpc.Method]methodDefinition{
	eth_blockNumber: {
		methodConfig:       defaultMethodConfig,
		methodSuccessRates: defaultMethodSuccessRates,
	},
	eth_call: {
		methodConfig:       defaultMethodConfig,
		methodSuccessRates: defaultMethodSuccessRates,
	},
	eth_getTransactionReceipt: {
		methodConfig:       defaultMethodConfig,
		methodSuccessRates: defaultMethodSuccessRates,
	},
	eth_getBlockByNumber: {
		methodConfig:       defaultMethodConfig,
		methodSuccessRates: defaultMethodSuccessRates,
	},
	eth_getBalance: {
		methodConfig:       defaultMethodConfig,
		methodSuccessRates: defaultMethodSuccessRates,
	},
	eth_chainId: {
		methodConfig:       defaultMethodConfig,
		methodSuccessRates: defaultMethodSuccessRates,
	},
	eth_getTransactionCount: {
		methodConfig:       defaultMethodConfig,
		methodSuccessRates: defaultMethodSuccessRates,
	},
	eth_getTransactionByHash: {
		methodConfig:       defaultMethodConfig,
		methodSuccessRates: defaultMethodSuccessRates,
	},
	eth_gasPrice: {
		methodConfig:       defaultMethodConfig,
		methodSuccessRates: defaultMethodSuccessRates,
	},
}

// serviceParameters holds service-specific test data for all methods.
// Allows testing specific requests that require parameters.
type serviceParameters struct {
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

// createParams builds RPC params for each method using the provided service parameters.
func createParams(method jsonrpc.Method, p serviceParameters) jsonrpc.Params {
	switch method {

	// Methods with empty params
	case eth_blockNumber, eth_chainId, eth_gasPrice:
		return jsonrpc.Params{}

	// Methods that just need the transaction hash
	//   Example: ["0xfeccd627b5b391d04fe45055873de3b2c0b4302d52e96bd41d5f0019a704165f"]
	case eth_getTransactionReceipt, eth_getTransactionByHash:
		params, _ := jsonrpc.BuildParamsFromString(p.transactionHash)
		return params

	// Methods that need [address, blockNumber]
	//   Example: ["0xdAC17F958D2ee523a2206206994597C13D831ec7", "latest"]
	case eth_getBalance, eth_getTransactionCount:
		params, _ := jsonrpc.BuildParamsFromStringArray([2]string{
			p.contractAddress,
			p.blockNumber,
		})
		return params

	// eth_getBlockByNumber needs [blockNumber, <boolean>]
	//   Example: ["0xe71e1d", false]
	case eth_getBlockByNumber:
		params, _ := jsonrpc.BuildParamsFromStringAndBool(
			p.blockNumber,
			false,
		)
		return params

	// eth_call needs [{ to: address, data: calldata }, blockNumber]
	//   Example: [{"to":"0xdAC17F958D2ee523a2206206994597C13D831ec7","data":"0x18160ddd"}, "latest"]
	case eth_call:
		params, _ := jsonrpc.BuildParamsFromObjectAndString(
			map[string]string{
				"to":   p.contractAddress,
				"data": p.callData,
			},
			p.blockNumber,
		)
		return params

	default:
		return jsonrpc.Params{}
	}
}
