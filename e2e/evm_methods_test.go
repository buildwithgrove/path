//go:build e2e

package e2e

import "github.com/buildwithgrove/path/qos/jsonrpc"

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

// createEVMJsonRPCParams builds RPC params for each EVM method using the provided service parameters.
func createEVMJsonRPCParams(method jsonrpc.Method, sp ServiceParams) jsonrpc.Params {
	switch method {

	// Methods with empty params
	case eth_blockNumber, eth_chainId, eth_gasPrice:
		return jsonrpc.Params{}

	// Methods that just need the transaction hash
	//   Example: ["0xfeccd627b5b391d04fe45055873de3b2c0b4302d52e96bd41d5f0019a704165f"]
	case eth_getTransactionReceipt, eth_getTransactionByHash:
		params, _ := jsonrpc.BuildParamsFromString(sp.TransactionHash)
		return params

	// Methods that need [address, blockNumber]
	//   Example: ["0xdAC17F958D2ee523a2206206994597C13D831ec7", "latest"]
	case eth_getBalance, eth_getTransactionCount:
		params, _ := jsonrpc.BuildParamsFromStringArray([2]string{
			sp.ContractAddress,
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
				"to":   sp.ContractAddress,
				"data": sp.CallData,
			},
			sp.blockNumber,
		)
		return params

	default:
		return jsonrpc.Params{}
	}
}
