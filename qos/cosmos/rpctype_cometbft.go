package cosmos

import (
	"slices"
	"strings"
)

// ------------------------------------------------------------------------------------------------
// CometBFT RPC Type Detection
// ------------------------------------------------------------------------------------------------

// CometBFT JSON-RPC prefixes
// API reference: https://docs.cometbft.com/v1.0/rpc/
var cometBftPrefixes = []string{
	"abci_",        // Application Blockchain Interface
	"broadcast_tx", // Transaction broadcasting (broadcast_tx_sync, broadcast_tx_async, etc.)
}

// Direct method matches for common CometBFT RPC methods
// API reference: https://docs.cometbft.com/v1.0/rpc/
var cometBftMethods = []string{
	"status",               // Node status
	"block",                // Block information
	"commit",               // Block commit information
	"validators",           // Validator set
	"genesis",              // Genesis document
	"health",               // Health check
	"net_info",             // Network information
	"consensus_state",      // Consensus state
	"dump_consensus_state", // Dump consensus state
	"num_unconfirmed_txs",  // Number of unconfirmed transactions
	"tx",                   // Transaction information
	"tx_search",            // Transaction search
	"block_search",         // Block search
	"consensus_params",     // Consensus parameters
	"unconfirmed_txs",      // Unconfirmed transactions
	"block_results",        // Block results
	"header",               // Block header
	"header_by_hash",       // Block header by hash
	"subscribe",            // Event subscription
	"unsubscribe",          // Event unsubscription
	"unsubscribe_all",      // Unsubscribe from all events
}

// isCometBftMethod checks if the JSON-RPC method corresponds to CometBFT RPC
// These are consensus layer methods typically exposed on port :26657
func isCometBftMethod(method string) bool {
	for _, prefix := range cometBftPrefixes {
		if strings.HasPrefix(method, prefix) {
			return true
		}
	}

	return slices.Contains(cometBftMethods, method)
}

// CometBFT RPC paths
// API reference: https://docs.cometbft.com/v1.0/rpc/
var cometBftPaths = []string{
	"/status",               // - /status (node status)
	"/broadcast_tx_sync",    // - /broadcast_tx_sync (transaction broadcast)
	"/broadcast_tx_async",   // - /broadcast_tx_async (transaction broadcast)
	"/broadcast_tx_commit",  // - /broadcast_tx_commit (transaction broadcast)
	"/block",                // - /block (block queries)
	"/commit",               // - /commit (commit queries)
	"/validators",           // - /validators (validator set queries)
	"/genesis",              // - /genesis (genesis document)
	"/health",               // - /health (health check)
	"/abci_info",            // - /abci_info (ABCI info)
	"/abci_query",           // - /abci_query (ABCI query)
	"/consensus_state",      // - /consensus_state (consensus state)
	"/dump_consensus_state", // - /dump_consensus_state (dump consensus state)
	"/net_info",             // - /net_info (network information)
	"/num_unconfirmed_txs",  // - /num_unconfirmed_txs (number of unconfirmed transactions)
	"/tx",                   // - /tx (transaction information)
	"/tx_search",            // - /tx_search (transaction search)
	"/block_search",         // - /block_search (block search)
	"/consensus_params",     // - /consensus_params (consensus parameters)
}

// isCometBftRpc checks if the URL path corresponds to CometBFT RPC endpoints
// These typically run on port :26657
func isCometBftRpc(urlPath string) bool {
	for _, path := range cometBftPaths {
		if strings.HasPrefix(urlPath, path) {
			return true
		}
	}
	return false
}
