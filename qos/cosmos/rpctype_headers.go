package cosmos

import (
	"strconv"

	"github.com/pokt-network/poktroll/pkg/relayer/proxy"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_DOCUMENT(@commodity): Document this in the official docs while keeping it simple.
// Use tables and mermaid diagram flows to show what happens depending on the port number.

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
