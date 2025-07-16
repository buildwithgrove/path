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
	checkStatus endpointCheckStatus // Checks chain ID, catching up status, and latest block height via /status
	checkHealth endpointCheckHealth // Checks node health via /health
}
