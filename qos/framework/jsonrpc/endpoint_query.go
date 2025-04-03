package jsonrpc

// endpointQuery represents a raw communication attempt with an endpoint.
// Instantiated by: RequestQoSContext.
// Used in EndpointQueryResultContext.
type endpointQuery struct {
	serviceName ServiceName

	// request is the JSONRPC request that was sent
	request *jsonrpc.Request

	// endpointAddr identifies the endpoint
	endpointAddr protocol.EndpointAddr

	// receivedData is the raw response data received from the endpoint (may be nil)
	receivedData []byte
}
