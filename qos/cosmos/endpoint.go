package cosmos

import (
	"time"
)

// endpoint maintains QoS data on a single endpoint for a CosmosSDK-based blockchain service.
type endpoint struct {
	// Response validation tracking
	hasReturnedEmptyResponse     bool
	hasReturnedInvalidResponse   bool
	hasReturnedUnmarshalingError bool
	invalidResponseLastObserved  *time.Time

	// CosmosSDK-specific checks
	checkCometBFTStatus endpointCheckCometBFTStatus // Checks chain ID, catching up status, and latest block height via /status
	checkCometBFTHealth endpointCheckCometBFTHealth // Checks node health via /health
	checkCosmosStatus   endpointCheckCosmosStatus   // Checks Cosmos SDK status via /status
	checkEVMChainID     endpointCheckEVMChainID     // Checks EVM chain ID via eth_chainId
}
