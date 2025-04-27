package framework

// endpointQuery represents a raw communication attempt with an endpoint.
// Instantiated by: RequestQoSContext.
// Used in EndpointQueryResultContext.
type endpointQuery struct {
	// request is the JSONRPC request that was sent.
	request *jsonrpc.Request

	// endpointAddr identifies the endpoint
	endpointAddr protocol.EndpointAddr

	// receivedData is the raw response data received from the endpoint (may be nil)
	receivedData []byte

	// JSONRPC response, parsed from the data received from the endpoint.
	parsedResponse *jsonrpc.Response

	// the result of processing the endpoint query.
	result *EndpointQueryResult
}
