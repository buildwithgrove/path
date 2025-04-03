package jsonrpc

// EndpointQueryResult captures information extracted from a service endpoint query.
type EndpointQueryResult struct {
	// The endpointQuery from which this result was built.
	// It can be used, e.g. to retrieve the JSONRPC request and its method.
	*endpointQuery

	// The set of endpoint attributes, extracted from the endpoint query and the endpoint's parsed JSONRPC response.
	// e.g. for a Solana `getEpochInfo` request, the custom service could derive two endpoint attributes as follows:
	// - "BlockHeight": "0x1234"
	// - "Epoch": "5"
	// It could also include sanctions:
	// e.g. for a returned value of "invalid" to an EVM `eth_blockNumber` request, the custom service could set:
	// - "eth_blockNumber": EndpointAtribute{error: EndpointError:{RecommendedSanction: {Duration: 5 * time.Minute}}}
	EndpointAttributes map[string]EndpointAttribute

	// TODO_FUTURE(@adshmh): add a JSONRPCErrorResponse to allow a result builder to supply its custom JSONRPC response.
}
