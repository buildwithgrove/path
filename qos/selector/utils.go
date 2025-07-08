package selector

import (
	"math/rand"

	"github.com/buildwithgrove/path/protocol"
)

// RandomSelectMultiple performs Fisher-Yates shuffle for random selection without replacement.
// This is a shared utility to avoid code duplication across QoS services.
//
// Parameters:
// - endpoints: The list of endpoints to select from
// - numEndpoints: The number of endpoints to select
//
// Returns a new slice containing randomly selected endpoints.
// If numEndpoints is greater than len(endpoints), returns all endpoints.
func RandomSelectMultiple(endpoints protocol.EndpointAddrList, numEndpoints int) protocol.EndpointAddrList {
	if numEndpoints <= 0 {
		return nil
	}

	if numEndpoints >= len(endpoints) {
		// Return a copy of all endpoints
		result := make(protocol.EndpointAddrList, len(endpoints))
		copy(result, endpoints)
		return result
	}

	// Create a copy to avoid modifying the original slice
	endpointsCopy := make(protocol.EndpointAddrList, len(endpoints))
	copy(endpointsCopy, endpoints)

	// Fisher-Yates shuffle for random selection without replacement
	selectedEndpoints := make(protocol.EndpointAddrList, 0, numEndpoints)
	for i := 0; i < numEndpoints; i++ {
		j := rand.Intn(len(endpointsCopy)-i) + i
		endpointsCopy[i], endpointsCopy[j] = endpointsCopy[j], endpointsCopy[i]
		selectedEndpoints = append(selectedEndpoints, endpointsCopy[i])
	}

	return selectedEndpoints
}