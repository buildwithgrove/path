package noop

import (
	"errors"
	"math/rand"

	"github.com/buildwithgrove/path/protocol"
)

// RandomEndpointSelector provides the functionality defined by the protocol.EndpointSelector interface.
var _ protocol.EndpointSelector = RandomEndpointSelector{}

// RandomEndpointSelector returns a randomly selected endpoint from the set of available ones, when its
// Select method is called.
// It has no fields, since the endpoint selection is random, i.e. no data is kept on the endpoints.
type RandomEndpointSelector struct{}

// Select returns a randomly selected endpoint from the set of supplied endpoints.
// This method fulfills the protocol.EndpointSelector interface.
func (RandomEndpointSelector) Select(endpoints protocol.EndpointAddrList) (protocol.EndpointAddr, error) {
	if len(endpoints) == 0 {
		return protocol.EndpointAddr(""), errors.New("RandomEndpointSelector: an empty endpoint list was supplied to the selector")
	}

	selectedEndpointAddr := endpoints[rand.Intn(len(endpoints))]
	return selectedEndpointAddr, nil
}

// SelectMultiple returns multiple randomly selected endpoints from the set of supplied endpoints.
// This method fulfills the protocol.EndpointSelector interface.
func (RandomEndpointSelector) SelectMultiple(endpoints protocol.EndpointAddrList, maxCount int) ([]protocol.EndpointAddr, error) {
	if len(endpoints) == 0 {
		return nil, errors.New("RandomEndpointSelector: an empty endpoint list was supplied to the selector")
	}

	if maxCount <= 0 {
		maxCount = 1
	}

	// Select up to maxCount endpoints (or all if less available)
	countToSelect := maxCount
	if countToSelect > len(endpoints) {
		countToSelect = len(endpoints)
	}

	// Create a copy to avoid modifying the original slice
	endpointsCopy := make(protocol.EndpointAddrList, len(endpoints))
	copy(endpointsCopy, endpoints)

	// Fisher-Yates shuffle for random selection without replacement
	var selectedEndpoints []protocol.EndpointAddr
	for i := 0; i < countToSelect; i++ {
		j := rand.Intn(len(endpointsCopy) - i) + i
		endpointsCopy[i], endpointsCopy[j] = endpointsCopy[j], endpointsCopy[i]
		selectedEndpoints = append(selectedEndpoints, endpointsCopy[i])
	}

	return selectedEndpoints, nil
}
