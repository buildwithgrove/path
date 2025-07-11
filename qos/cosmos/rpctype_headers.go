package cosmos

import (
	"strconv"
	"strings"

	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/pokt-network/poktroll/pkg/relayer/proxy"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
)

// ------------------------------------------------------------------------------------------------
// RPC Type Header Determination
// ------------------------------------------------------------------------------------------------

// getRPCTypeHeaders determines the appropriate RPC type header based on the URL path,
// HTTP method, request body, and optional JSON-RPC request for routing Cosmos SDK based blockchain requests.

// getRPCTypeHeaders determines the appropriate RPC type header based on the URL path,
// HTTP method, request body, and optional JSON-RPC request for routing Cosmos SDK based blockchain requests.
//
// Routing Strategy:
// 1. Cosmos REST API (:1317) -> RPCType_REST
// 2. Tendermint RPC (:26657) -> RPCType_JSON_RPC (for JSON-RPC) or RPCType_REST (for REST)
// 3. Ethereum JSON-RPC (:8545) -> RPCType_JSON_RPC
//
// Parameters:
// - urlPath: The URL path of the request
// - httpMethod: HTTP method (GET, POST, etc.) - can be empty if not available
// - requestBody: Raw request body - can be nil if not available
// - jsonrpcReq: Parsed JSON-RPC request - can be nil if not available or not JSON-RPC
func getRPCTypeHeaders(
	urlPath string,
	jsonrpcReq *jsonrpc.Request,
) map[string]string {
	var rpcType sharedtypes.RPCType

	switch {
	// Priority 1: If we have a JSON-RPC request object, use method-based detection
	case jsonrpcReq != nil && jsonrpcReq.Method != "":
		return getRPCTypeFromJSONRPCMethod(string(jsonrpcReq.Method))

	// Priority 2: Check for Cosmos REST API patterns
	case isCosmosRestAPI(urlPath):
		rpcType = sharedtypes.RPCType_REST

	// Priority 3: Check for CometBFT RPC patterns (supports both REST + JSON-RPC)
	case isCometBftRpc(urlPath):
		rpcType = sharedtypes.RPCType_COMET_BFT

	// Default case - no specific pattern matches
	default:
		// Return empty headers to allow request to proceed without specific RPC type routing
		return map[string]string{}
	}

	return map[string]string{
		proxy.RPCTypeHeader: strconv.Itoa(int(rpcType)),
	}
}

// getRPCTypeFromJSONRPCMethod determines RPC type based on JSON-RPC method name
func getRPCTypeFromJSONRPCMethod(jsonRPCMethod string) map[string]string {
	var rpcType sharedtypes.RPCType

	switch {
	// Check for Ethereum/EVM JSON-RPC methods
	case isEVMMethod(jsonRPCMethod):
		rpcType = sharedtypes.RPCType_JSON_RPC

	// Check for CometBFT JSON-RPC methods (supports both REST + JSON-RPC)
	case isCometBftMethod(jsonRPCMethod):
		rpcType = sharedtypes.RPCType_COMET_BFT

	// Unknown JSON-RPC method - default to JSON-RPC type
	default:
		rpcType = sharedtypes.RPCType_JSON_RPC
	}

	return map[string]string{
		proxy.RPCTypeHeader: strconv.Itoa(int(rpcType)),
	}
}

// ------------------------------------------------------------------------------------------------
// RPC Type Detection
// ------------------------------------------------------------------------------------------------

// Cosmos REST API prefixes
var cosmosRestPrefixes = []string{
	"/cosmos/",       // - /cosmos/* (Cosmos SDK modules)
	"/ibc/",          // - /ibc/* (IBC protocol)
	"/staking/",      // - /staking/* (staking module)
	"/auth/",         // - /auth/* (auth module)
	"/bank/",         // - /bank/* (bank module)
	"/txs/",          // - /txs/* (transaction endpoints)
	"/gov/",          // - /gov/* (governance module)
	"/distribution/", // - /distribution/* (distribution module)
	"/slashing/",     // - /slashing/* (slashing module)
	"/mint/",         // - /mint/* (mint module)
	"/upgrade/",      // - /upgrade/* (upgrade module)
	"/evidence/",     // - /evidence/* (evidence module)
}

// isCosmosRestAPI checks if the URL path corresponds to Cosmos SDK REST API endpoints
// These typically run on port :1317 and include paths like:
func isCosmosRestAPI(urlPath string) bool {
	for _, prefix := range cosmosRestPrefixes {
		if strings.HasPrefix(urlPath, prefix) {
			return true
		}
	}

	return false
}

// CometBFT RPC paths
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

// EVM JSON-RPC prefixes
var evmPrefixes = []string{
	"eth_",      // Ethereum standard methods
	"net_",      // Network information
	"web3_",     // Web3 provider methods
	"txpool_",   // Transaction pool
	"debug_",    // Debug namespace
	"trace_",    // Tracing methods
	"engine_",   // Engine API (for consensus)
	"personal_", // Personal namespace
	"admin_",    // Admin namespace
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

// CometBFT JSON-RPC prefixes
var cometBftPrefixes = []string{
	"abci_",        // Application Blockchain Interface
	"broadcast_tx", // Transaction broadcasting (broadcast_tx_sync, broadcast_tx_async, etc.)
}

// Direct method matches for common CometBFT RPC methods
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

	for _, tmMethod := range cometBftMethods {
		if method == tmMethod {
			return true
		}
	}

	return false
}
