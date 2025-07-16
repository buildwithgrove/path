package cosmos

import "strings"

// ------------------------------------------------------------------------------------------------
// EVM JSON-RPC RPC Type Detection
// ------------------------------------------------------------------------------------------------

// EVM JSON-RPC prefixes
// API reference: https://ethereum.org/en/developers/docs/apis/json-rpc/
var evmPrefixes = []string{
	"eth_",  // Ethereum standard methods
	"net_",  // Network information
	"web3_", // Web3 provider methods
}

// isEVMMethod checks if the JSON-RPC method corresponds to Ethereum/EVM endpoints
// These typically run on port :8545 for EVM-compatible Cosmos chains (like Evmos)
func isEVMMethod(method string) bool {
	for _, prefix := range evmPrefixes {
		if strings.HasPrefix(method, prefix) {
			return true
		}
	}
	return false
}
