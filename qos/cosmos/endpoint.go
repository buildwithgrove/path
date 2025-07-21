package cosmos

import (
	"time"
)

// endpoint maintains QoS data on a single endpoint for a CosmosSDK-based blockchain service.
type endpoint struct {
	// Response validation tracking
	hasReturnedEmptyResponse    bool
	hasReturnedInvalidResponse  bool
	invalidResponseLastObserved *time.Time

	// CosmosSDK-specific checks
	checkCometbftStatus endpointCheckStatus       // Checks chain ID, catching up status, and latest block height via /status
	checkCometbftHealth endpointCheckHealth       // Checks node health via /health
	checkCosmosStatus   endpointCheckCosmosStatus // Checks latest block height via /cosmos/base/node/v1beta1/status
}
