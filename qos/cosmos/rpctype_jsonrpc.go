package cosmos

import (
	"strings"

	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
)

// ------------------------------------------------------------------------------------------------
// JSON-RPC Service Type Detection - Consolidated Logic
// ------------------------------------------------------------------------------------------------

// detectJSONRPCServiceType determines service type based on JSONRPC method
// This is the main entry point for JSON-RPC service detection
func detectJSONRPCServiceType(method string) sharedtypes.RPCType {
	// Route based on method prefix/name
	if isCometBftJSONRPCMethod(method) {
		return sharedtypes.RPCType_COMET_BFT
	}

	if isEVMJSONRPCMethod(method) {
		return sharedtypes.RPCType_JSON_RPC
	}

	// Unrecognized method → default to EVM JSONRPC
	return sharedtypes.RPCType_JSON_RPC
}

// ------------------------------------------------------------------------------------------------
// EVM JSON-RPC RPC Type Detection
// ------------------------------------------------------------------------------------------------

// EVM JSON-RPC prefixes
// API reference: https://ethereum.org/en/developers/docs/apis/json-rpc/
var evmPrefixes = []string{
	"eth_",      // Ethereum standard methods
	"net_",      // Network information
	"web3_",     // Web3 provider methods
	"debug_",    // Debug methods
	"txpool_",   // Transaction pool methods
	"admin_",    // Admin methods
	"miner_",    // Miner methods
	"personal_", // Personal methods
	"clique_",   // Clique consensus methods
	"les_",      // Light Ethereum Subprotocol methods
}

// isEVMJSONRPCMethod checks if method is an EVM JSONRPC method
func isEVMJSONRPCMethod(method string) bool {
	// EVM methods typically use prefixes
	for _, prefix := range evmPrefixes {
		if strings.HasPrefix(method, prefix) {
			return true
		}
	}
	return false
}
