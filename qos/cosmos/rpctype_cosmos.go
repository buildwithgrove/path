package cosmos

import (
	"strings"

	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
)

// ------------------------------------------------------------------------------------------------
// REST API Type Detection - Consolidated Logic
// ------------------------------------------------------------------------------------------------

// determineRESTRPCType determines the specific RPC type for REST requests
func determineRESTRPCType(path string) sharedtypes.RPCType {
	// CometBFT REST-style endpoints should be tagged as COMET_BFT
	if isCometBftRpc(path) {
		return sharedtypes.RPCType_COMET_BFT
	}

	// Everything else is regular REST
	return sharedtypes.RPCType_REST
}

// isValidRESTPath validates that the path matches REST API patterns
func isValidRESTPath(path string) bool {
	// CosmosSDK REST API patterns
	if isCosmosRestAPI(path) {
		return true
	}

	// CometBFT REST-style endpoints
	if isCometBftRpc(path) {
		return true
	}

	// Allow some additional common REST patterns
	additionalValidPaths := []string{
		"/txs",          // Legacy transaction endpoint
		"/tx",           // Transaction endpoints
		"/node_info",    // Node information
		"/syncing",      // Sync status
		"/latest_block", // Latest block info
	}

	for _, validPath := range additionalValidPaths {
		if path == validPath || (len(path) > len(validPath) && path[:len(validPath)+1] == validPath+"/") {
			return true
		}
	}

	return false
}

// ------------------------------------------------------------------------------------------------
// Cosmos SDK RPC Type Detection
// ------------------------------------------------------------------------------------------------

// Cosmos REST API prefixes
// API reference: https://docs.cosmos.network/api
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
