package cosmos

import (
	"time"

	"github.com/buildwithgrove/path/protocol"
)

// endpoint maintains QoS data on a single endpoint for a CosmosSDK-based blockchain service.
type endpoint struct {
	// TODO_TECHDEBT(@adshmh): endpointAddr is here but the address is also the key to the `map[protocol.EndpointAddr]endpoint` in the endpointStore.
	// Consider removing it from the struct to avoid duplication.
	endpointAddr protocol.EndpointAddr

	// Response validation tracking
	hasReturnedEmptyResponse    bool
	hasReturnedInvalidResponse  bool
	invalidResponseLastObserved *time.Time

	// CosmosSDK-specific checks
	checkStatus endpointCheckStatus // Checks chain ID, catching up status, and latest block height via /status
	checkHealth endpointCheckHealth // Checks node health via /health
}
