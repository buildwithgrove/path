package cosmos

import "strings"

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
