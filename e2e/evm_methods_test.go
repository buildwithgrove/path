//go:build e2e

package e2e

import (
	"time"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

/* -------------------- EVM JSON-RPC Method Definitions -------------------- */

// Docs reference for all methods:
// https://ethereum.org/en/developers/docs/apis/json-rpc/
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

// runAllMethods returns all EVM JSON-RPC methods for a service load test
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
	// methodDefinition contains all configuration and test requirements for a method
	methodDefinition struct {
		methodConfig
		methodSuccessRates
	}

	// methodConfig contains the configuration for a method to be tested.
	// This includes the total number of requests to send, the requests per second,
	// and the number of workers to use.
	methodConfig struct {
		totalRequests int // Total number of requests to send
		rps           int // Requests per second
	}

	// methodSuccessRates contains the minimum success rate and maximum
	// latency requirements for a method to pass the load test.
	methodSuccessRates struct {
		successRate   float64       // Minimum success rate (0-1)
		maxP50Latency time.Duration // Maximum P50 latency
		maxP95Latency time.Duration // Maximum P95 latency
		maxP99Latency time.Duration // Maximum P99 latency
	}
)

// methodDefinitions contains all method definitions for a service load test.
// this allows customizing the configuration for each method as desired.
var methodDefinitions = map[jsonrpc.Method]methodDefinition{
	eth_blockNumber: {
		methodConfig: methodConfig{
			totalRequests: 200,
			rps:           10,
		},
		methodSuccessRates: methodSuccessRates{
			successRate:   0.95,
			maxP50Latency: 350 * time.Millisecond,
			maxP95Latency: 900 * time.Millisecond,
			maxP99Latency: 3_000 * time.Millisecond,
		},
	},
	eth_call: {
		methodConfig: methodConfig{
			totalRequests: 200,
			rps:           10,
		},
		methodSuccessRates: methodSuccessRates{
			successRate:   0.90,
			maxP50Latency: 350 * time.Millisecond,
			maxP95Latency: 900 * time.Millisecond,
			maxP99Latency: 3_000 * time.Millisecond,
		},
	},
	eth_getTransactionReceipt: {
		methodConfig: methodConfig{
			totalRequests: 200,
			rps:           10,
		},
		methodSuccessRates: methodSuccessRates{
			successRate:   0.95,
			maxP50Latency: 350 * time.Millisecond,
			maxP95Latency: 900 * time.Millisecond,
			maxP99Latency: 3_000 * time.Millisecond,
		},
	},
	eth_getBlockByNumber: {
		methodConfig: methodConfig{
			totalRequests: 200,
			rps:           10,
		},
		methodSuccessRates: methodSuccessRates{
			successRate:   0.95,
			maxP50Latency: 350 * time.Millisecond,
			maxP95Latency: 900 * time.Millisecond,
			maxP99Latency: 3_000 * time.Millisecond,
		},
	},
	eth_getBalance: {
		methodConfig: methodConfig{
			totalRequests: 200,
			rps:           10,
		},
		methodSuccessRates: methodSuccessRates{
			successRate:   0.95,
			maxP50Latency: 350 * time.Millisecond,
			maxP95Latency: 900 * time.Millisecond,
			maxP99Latency: 3_000 * time.Millisecond,
		},
	},
	eth_chainId: {
		methodConfig: methodConfig{
			totalRequests: 200,
			rps:           10,
		},
		methodSuccessRates: methodSuccessRates{
			successRate:   0.95,
			maxP50Latency: 350 * time.Millisecond,
			maxP95Latency: 900 * time.Millisecond,
			maxP99Latency: 3_000 * time.Millisecond,
		},
	},
	eth_getTransactionCount: {
		methodConfig: methodConfig{
			totalRequests: 200,
			rps:           10,
		},
		methodSuccessRates: methodSuccessRates{
			successRate:   0.95,
			maxP50Latency: 350 * time.Millisecond,
			maxP95Latency: 900 * time.Millisecond,
			maxP99Latency: 3_000 * time.Millisecond,
		},
	},
	eth_getTransactionByHash: {
		methodConfig: methodConfig{
			totalRequests: 200,
			rps:           10,
		},
		methodSuccessRates: methodSuccessRates{
			successRate:   0.95,
			maxP50Latency: 350 * time.Millisecond,
			maxP95Latency: 900 * time.Millisecond,
			maxP99Latency: 3_000 * time.Millisecond,
		},
	},
	eth_gasPrice: {
		methodConfig: methodConfig{
			totalRequests: 200,
			rps:           10,
		},
		methodSuccessRates: methodSuccessRates{
			successRate:   0.95,
			maxP50Latency: 350 * time.Millisecond,
			maxP95Latency: 900 * time.Millisecond,
			maxP99Latency: 3_000 * time.Millisecond,
		},
	},
}

// serviceParameters holds service-specific test data for all methods.
// to allow testing specific requests that require parameters.
type serviceParameters struct {
	// Used for eth_getBalance, eth_getTransactionCount, and eth_getTransactionReceipt
	blockNumber string
	// Used for eth_getBalance, eth_getTransactionCount, and eth_getTransactionReceipt
	//
	// `contractAddress` address should match the `evmArchivalCheckConfig.contractAddress`
	// value in `config/service_qos_config.go`
	contractAddress string
	// The minimum block number to use for the test; this is to ensure we are not
	// trying to fetch a block where the  contract address has no balance or transactions.
	//
	// `contractStartBlock` should match the `evmArchivalCheckConfig.contractStartBlock`
	// value in `config/service_qos_config.go`
	contractStartBlock uint64
	// Used for eth_getTransactionReceipt and eth_getTransactionByHash
	transactionHash string
	// Used for eth_call
	callData string
}

// createParams builds RPC params for each method using the provided service parameters
func createParams(method jsonrpc.Method, p serviceParameters) jsonrpc.Params {
	switch method {
	// Methods with empty params
	case eth_blockNumber, eth_chainId, eth_gasPrice:
		return jsonrpc.Params{}

	// Methods that just need the transaction hash
	// eg. ["0xfeccd627b5b391d04fe45055873de3b2c0b4302d52e96bd41d5f0019a704165f"]
	case eth_getTransactionReceipt, eth_getTransactionByHash:
		params, _ := jsonrpc.BuildParamsFromString(p.transactionHash)
		return params

	// Methods that need [address, blockNumber]
	// eg. ["0xdAC17F958D2ee523a2206206994597C13D831ec7", "latest"]
	case eth_getBalance, eth_getTransactionCount:
		params, _ := jsonrpc.BuildParamsFromStringArray([2]string{
			p.contractAddress,
			p.blockNumber,
		})
		return params

	// eth_getBlockByNumber needs [blockNumber, <boolean>]
	// eg. ["0xe71e1d", false]
	case eth_getBlockByNumber:
		params, _ := jsonrpc.BuildParamsFromStringAndBool(
			p.blockNumber,
			false,
		)
		return params

	// eth_call needs [{ to: address, data: calldata }, blockNumber]
	// eg. [{"to":"0xdAC17F958D2ee523a2206206994597C13D831ec7","data":"0x18160ddd"}, "latest"]
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
