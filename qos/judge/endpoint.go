package judge

import (
	"sync"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_TECHDEBT(@adshmh): Persist this state (which may include sanctions) across restarts to maintain endpoint exclusions.
// TODO_MVP(@adshmh): add support for removing expired query results.
//
// Endpoint represents a service endpoint with its associated attributes.
// - Read-only for client code
// - All attributes are set internally by the framework
type Endpoint struct {
	logger polylog.Logger

	// queryResults maps keys to query results for this endpoint.
	// The map key is the method of the JSONRPC request for which the query result was built.
	// Examples:
	// - "eth_blockNumber": &EndpointQueryResult{IntValues: {"blockNumber": 0x1234}}
	// - "eth_getBalance": &EndpointQueryResult{
	//     StrValues: {"address": "0x8d97..."},
	//     IntValues: {"balance": 133456789},
	//   }
	queryResults map[jsonrpc.Method]*EndpointQueryResult

	// mutex for query results
	resultsMu sync.RWMutex
}

// GetStrResult retrieves a string attribute of a result by key.
// DEV_NOTE: This design pattern:
// - Prevents map leaking and unauthorized modifications through pointers
// - Avoids expensive struct cloning
// - Maintains proper encapsulation
func (e *Endpoint) GetStrResult(resultKey jsonrpc.Method, valueKey string) (string, bool) {
	e.resultsMu.RLock()
	defer e.resultsMu.RUnlock()

	result, exists := e.queryResults[resultKey]
	if !exists || result == nil {
		return "", false
	}

	strValue, found := result.StrValues[valueKey]
	return strValue, found
}

// GetIntResult retrieves an integer attribute of a result by key.
// See the comment on GetStrResult for notes on this pattern.
func (e *Endpoint) GetIntResult(resultKey jsonrpc.Method, valueKey string) (int, bool) {
	e.resultsMu.RLock()
	defer e.resultsMu.RUnlock()

	result, exists := e.queryResults[resultKey]
	if !exists || result == nil {
		return 0, false
	}

	intValue, found := result.IntValues[valueKey]
	return intValue, found
}

// TODO_IN_THIS_PR: implement.
func (e *Endpoint) GetActiveSanction() (Sanction, bool) {
	return Sanction{}, false
}

// ApplyQueryResult updates the endpoint's attributes with attributes from the query result.
// It merges the EndpointAttributes from the query result into the endpoint's attributes map.
func (e *Endpoint) applyQueryResults(endpointQueryResults []*EndpointQueryResult) {
	e.resultsMu.Lock()
	defer e.resultsMu.Unlock()

	// Initialize the results map if nil.
	if e.queryResults == nil {
		e.queryResults = make(map[jsonrpc.Method]*EndpointQueryResult)
	}

	// Add or update attributes from the query result
	for _, endpointQueryResult := range endpointQueryResults {
		jsonrpcRequestMethod := endpointQueryResult.getJSONRPCRequestMethod()

		if jsonrpcRequestMethod == "" {
			e.logger.Warn().Msg("Endpoint received query result with no JSONRPC method set: skipping update.")
			return
		}

		// Update the endpoint result matching the JSONRPC request.
		e.queryResults[jsonrpcRequestMethod] = endpointQueryResult

		e.logger.With("jsonrpc_request_method", jsonrpcRequestMethod).Debug().Msg("Updated endpoint with query result.")
	}
}
