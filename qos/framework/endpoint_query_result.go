package framework

import (
	"errors"
	"time"
)

// TODO_IMPROVE(@adshmh): Enhance EndpointQueryResult to support data types commonly stored for endpoints.
//
// EndpointQueryResult captures data extracted from an endpoint query.
// - Stores one or more string/integer values.
// - Contains error/sanction information on endpoint error.
type EndpointQueryResult struct {
	// The endpointQuery from which this result was built.
	// It can be used, e.g. to retrieve the JSONRPC request and its method.
	*endpointQuery

	// TODO_IN_THIS_PR: verify this is set by all result builders.

	// The JSONRPC response to be returned to the client.
	// MUST be set.
	clientResponse *jsonrpc.Response

	// The set of values/attributes extracted from the endpoint query and the endpoint's parsed JSONRPC response.
	// e.g. for a Solana `getEpochInfo` request, the custom service could derive two endpoint attributes as follows:
	// - "BlockHeight": 0x1234
	// - "Epoch": 5
	StringValues map[string]string
	IntValues    map[string]int

	// Captures the queried endpoint's error.
	// Only set if the query result indicates an endpoint error.
	// It could also include sanctions:
	// e.g. for an invalid value returned for an EVM `eth_blockNumber` request, the custom service could set:
	// Error:
	// - Description: "invalid response to eth_blockNumber"
	// - RecommendedSanction: {Duration: 5 * time.Minute}
	Error *EndpointError

	// The time at which the query result is expired.
	// Expired results will be ignored, including in:
	// - endpoint selection, e.g. sanctions.
	// - state update: e.g. archival state of the QoS service.
	ExpiryTime time.Time

	// TODO_FUTURE(@adshmh): add a JSONRPCErrorResponse to allow a result builder to supply its custom JSONRPC response.
}


