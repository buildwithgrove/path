package cosmos

import (
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
