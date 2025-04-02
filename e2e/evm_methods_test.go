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
	eth_getLogs               jsonrpc.Method = "eth_getLogs"
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
		eth_getLogs,
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
		totalRequests int    // Total number of requests to send
		rps           int    // Requests per second
		workers       uint64 // Number of workers to use
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
			totalRequests: 300,
			rps:           10,
			workers:       20,
		},
		methodSuccessRates: methodSuccessRates{
			successRate:   0.95,
			maxP50Latency: 250 * time.Millisecond,
			maxP95Latency: 750 * time.Millisecond,
			maxP99Latency: 1500 * time.Millisecond,
		},
	},
	eth_call: {
		methodConfig: methodConfig{
			totalRequests: 300,
			rps:           10,
			workers:       20,
		},
		methodSuccessRates: methodSuccessRates{
			successRate:   0.95,
			maxP50Latency: 250 * time.Millisecond,
			maxP95Latency: 750 * time.Millisecond,
			maxP99Latency: 1500 * time.Millisecond,
		},
	},
	eth_getTransactionReceipt: {
		methodConfig: methodConfig{
			totalRequests: 300,
			rps:           10,
			workers:       20,
		},
		methodSuccessRates: methodSuccessRates{
			successRate:   0.95,
			maxP50Latency: 250 * time.Millisecond,
			maxP95Latency: 750 * time.Millisecond,
			maxP99Latency: 1500 * time.Millisecond,
		},
	},
	eth_getBlockByNumber: {
		methodConfig: methodConfig{
			totalRequests: 300,
			rps:           10,
			workers:       20,
		},
		methodSuccessRates: methodSuccessRates{
			successRate:   0.95,
			maxP50Latency: 250 * time.Millisecond,
			maxP95Latency: 750 * time.Millisecond,
			maxP99Latency: 1500 * time.Millisecond,
		},
	},
	eth_getLogs: {
		methodConfig: methodConfig{
			totalRequests: 300,
			rps:           10,
			workers:       20,
		},
		methodSuccessRates: methodSuccessRates{
			successRate:   0.9,
			maxP50Latency: 500 * time.Millisecond,
			maxP95Latency: 1250 * time.Millisecond,
			maxP99Latency: 2000 * time.Millisecond,
		},
	},
	eth_getBalance: {
		methodConfig: methodConfig{
			totalRequests: 300,
			rps:           10,
			workers:       20,
		},
		methodSuccessRates: methodSuccessRates{
			successRate:   0.95,
			maxP50Latency: 250 * time.Millisecond,
			maxP95Latency: 750 * time.Millisecond,
			maxP99Latency: 1500 * time.Millisecond,
		},
	},
	eth_chainId: {
		methodConfig: methodConfig{
			totalRequests: 300,
			rps:           10,
			workers:       20,
		},
		methodSuccessRates: methodSuccessRates{
			successRate:   0.95,
			maxP50Latency: 250 * time.Millisecond,
			maxP95Latency: 750 * time.Millisecond,
			maxP99Latency: 1500 * time.Millisecond,
		},
	},
	eth_getTransactionCount: {
		methodConfig: methodConfig{
			totalRequests: 300,
			rps:           10,
			workers:       20,
		},
		methodSuccessRates: methodSuccessRates{
			successRate:   0.95,
			maxP50Latency: 250 * time.Millisecond,
			maxP95Latency: 750 * time.Millisecond,
			maxP99Latency: 1500 * time.Millisecond,
		},
	},
	eth_getTransactionByHash: {
		methodConfig: methodConfig{
			totalRequests: 300,
			rps:           10,
			workers:       20,
		},
		methodSuccessRates: methodSuccessRates{
			successRate:   0.95,
			maxP50Latency: 250 * time.Millisecond,
			maxP95Latency: 750 * time.Millisecond,
			maxP99Latency: 1500 * time.Millisecond,
		},
	},
	eth_gasPrice: {
		methodConfig: methodConfig{
			totalRequests: 300,
			rps:           10,
			workers:       20,
		},
		methodSuccessRates: methodSuccessRates{
			successRate:   0.95,
			maxP50Latency: 250 * time.Millisecond,
			maxP95Latency: 750 * time.Millisecond,
			maxP99Latency: 1500 * time.Millisecond,
		},
	},
}

// methodParams holds service-specific parameter data for all methods
type methodParams struct {
	// Common parameters
	blockNumber     string
	contractAddress string
	transactionHash string
	callData        string
}

// createParams builds RPC params for each method using the provided service parameters
func createParams(method jsonrpc.Method, p methodParams) []any {
	switch method {
	case eth_blockNumber, eth_chainId, eth_gasPrice:
		// Methods with empty params
		return []any{}

	case eth_call:
		// eth_call needs [{ to: address, data: calldata }, blockNumber]
		// eg. [{"to":"0xdAC17F958D2ee523a2206206994597C13D831ec7","data":"0x18160ddd"}, "latest"]
		return []any{
			map[string]string{
				"to":   p.contractAddress,
				"data": p.callData,
			},
			p.blockNumber,
		}

	case eth_getBlockByNumber:
		// eth_getBlockByNumber needs [blockNumber, <boolean>]
		// eg. ["0x1", false]
		return []any{p.blockNumber, false}

	case eth_getTransactionReceipt, eth_getTransactionByHash:
		// Methods that just need the transaction hash
		// eg. ["0xfeccd627b5b391d04fe45055873de3b2c0b4302d52e96bd41d5f0019a704165f"]
		return []any{p.transactionHash}

	case eth_getBalance, eth_getTransactionCount:
		// Methods that need [address, blockNumber]
		// eg. ["0xdAC17F958D2ee523a2206206994597C13D831ec7", "latest"]
		return []any{p.contractAddress, p.blockNumber}

	case eth_getLogs:
		// eth_getLogs needs [{ fromBlock, toBlock }]
		// eg. [{"fromBlock":"0x1","toBlock":"0x1"}]
		return []any{
			map[string]string{
				"fromBlock": p.blockNumber,
				"toBlock":   p.blockNumber,
			},
		}
	default:
		return []any{}
	}
}
