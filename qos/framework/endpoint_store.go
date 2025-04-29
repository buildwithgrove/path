package framework

import (
	"sync"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
)

// endpointStore maintains data on the set of available endpoints.
// It is package-private and not meant to be used directly by any entity outside the jsonrpc package.
type endpointStore struct {
	logger      polylog.Logger
	endpointsMu sync.RWMutex
	endpoints   map[protocol.EndpointAddr]*Endpoint
}

func (es *endpointStore) updateStoredEndpoints(endpointQueryResults []*EndpointQueryResult) []*Endpoint {
	es.endpointsMu.Lock()
	defer es.endpointsMu.Unlock()

	groupedEndpointResults := groupResultsByEndpointAddr(endpointQueryResults)

	// Track the updated endpoints
	var updatedEndpoints []Endpoint
	// Loop over query results, grouped by endpoint address, and update the corresponding stored endpoint.
	for endpointAddr, queryResults := range groupedEndpointResults {
		endpoint, found := es.endpoints[endpointQueryResult.endpointAddr]
		if !found {
			endpoint = &Endpoint{}
		}

		endpoint.applyQueryResults(queryResults)

		// Store the updated endpoint
		es.endpoints[endpointQueryResult.endpointAddr]

		// Add the updated endpoint to the list to be returned.
		updatedEndpoints = append(updatedEndpoints, endpoint)
	}

	return updatedEndpoints
}

func groupResultsByEndpointAddr(endpointQueryResults []*EndpointQueryResult) map[protocol.EndpointAddr][]*EndpointQueryResult {
	resultsByEndpoint := make(map[protocol.EndpointAddr][]*EndpointQueryResult)

	for _, queryResult := range endpointQueryResults {
		resultsByEndpoint[queryResult.endpointAddr] = append(resultsByEndpoint[queryResult.endpointAddr], queryResult)
	}

	return resultsByEndpoint
}

// storeEndpoint stores or updates an endpoint in the store.
func (es *endpointStore) storeEndpoint(addr protocol.EndpointAddr, endpoint Endpoint) {
	es.endpointsMu.Lock()
	defer es.endpointsMu.Unlock()

	if es.endpoints == nil {
		es.endpoints = make(map[protocol.EndpointAddr]Endpoint)
	}

	es.endpoints[addr] = &endpoint
}

// getEndpoint retrieves an endpoint by its address.
func (es *endpointStore) getEndpoint(addr protocol.EndpointAddr) (*Endpoint, bool) {
	es.endpointsMu.RLock()
	defer es.endpointsMu.RUnlock()

	if es.endpoints == nil {
		return Endpoint{}, false
	}

	endpoint, found := es.endpoints[addr]
	return endpoint, found
}
