package jsonrpc

import (
	"sync"

	"github.com/buildwithgrove/path/protocol"
)

// endpointStore maintains data on the set of available endpoints.
// It is package-private and not meant to be used directly by any entity outside the jsonrpc package.
type endpointStore struct {
	endpointsMu sync.RWMutex
	endpoints   map[protocol.EndpointAddr]Endpoint
}

func (es *endpointStore) updateStoredEndpoints(endpointQueries []*endpointQuery) []Endpoint {
	es.endpointsMu.Lock()
	defer es.endpointsMu.Unlock()

	endpoints := make([]Endpoint, len(endpointQueries))
	for index, endpointQuery := range endpointQueries {
		endpoint := es.endpoints[endpointQuery.endpointAddr]
		if 
	}
}

// storeEndpoint stores or updates an endpoint in the store.
func (es *endpointStore) storeEndpoint(addr protocol.EndpointAddr, endpoint Endpoint) {
	es.endpointsMu.Lock()
	defer es.endpointsMu.Unlock()

	if es.endpoints == nil {
		es.endpoints = make(map[protocol.EndpointAddr]Endpoint)
	}

	es.endpoints[addr] = endpoint
}

// getEndpoint retrieves an endpoint by its address.
func (es *endpointStore) getEndpoint(addr protocol.EndpointAddr) (Endpoint, bool) {
	es.endpointsMu.RLock()
	defer es.endpointsMu.RUnlock()

	if es.endpoints == nil {
		return Endpoint{}, false
	}

	endpoint, found := es.endpoints[addr]
	return endpoint, found
}
