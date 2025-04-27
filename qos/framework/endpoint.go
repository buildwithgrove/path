package jsonrpc

// TODO_TECHDEBT(@adshmh): Persist this state (which may include sanctions) across restarts to maintain endpoint exclusions.
// TODO_MVP(@adshmh): add support for removing expired query results.
//
// Endpoint represents a service endpoint with its associated attributes.
// - Read-only for client code
// - All attributes are set internally by the framework
type Endpoint struct {
	// queryResults maps keys to query results for this endpoint.
	// Keys are defined by the QoS service implementation (typically JSONRPC method names).
	// Examples:
	// - "eth_blockNumber": &EndpointQueryResult{IntValues: {"blockNumber": 0x1234}}
	// - "eth_getBalance": &EndpointQueryResult{
	//     StringValues: {"address": "0x8D97..."},
	//     IntValues: {"balance": 133456789},
	//   }
	queryResults map[string]*EndpointQueryResult

	// mutex for query results
	resultsMu sync.Mutex
}

// GetQueryResultStringValue retrieves a string attribute of a result by key.
// DEV_NOTE: This design pattern:
// - Prevents map leaking and unauthorized modifications through pointers
// - Avoids expensive struct cloning
// - Maintains proper encapsulation
func (e *Endpoint) GetQueryResultStringValue(resultKey, valueKey string) (string, bool) {
	result, exists := e.queryResults[resultKey]
	if !exists || result == nil {
		return "", false
	}

	return result.StringValues[valueKey]
}

// GetQueryResultStringValue retrieves an integer attribute of a result by key.
// See the comment on GetQueryResultStringValue for notes on this pattern.
func (e *Endpoint) GetQueryResultIntValue(resultKey, valueKey string) (int, bool) {
	result, exists := e.queryResults[resultKey]
	if !exists || result == nil {
		return "", false
	}

	return result.IntValues[valueKey]
}

// TODO_IN_THIS_PR: implement.
func (e *Endpoint) HasActiveSanction() (Sanction, bool) {

}

// ApplyQueryResult updates the endpoint's attributes with attributes from the query result.
// It merges the EndpointAttributes from the query result into the endpoint's attributes map.
func (e *Endpoint) ApplyQueryResults(result map[string]EndpointQueryResult) {
	// Initialize the results map if nil.
	if e.queryResults == nil {
		e.queryResults = make(map[string]*EndpointQueryResult)
	}

	// Add or update attributes from the query result
	for key, result := range results {
		e.queryResult[key] = result
	}
}
