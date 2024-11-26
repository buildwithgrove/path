package protocol

// Response is a general purpose struct for capturing the response to a relay, received from an endpoint.
// TODO_FUTURE(@adshmh): It only supports HTTP responses for now; add support for others.
type Response struct {
	// Bytes is the response to a relay received from an endpoint.
	// An endpoint is the backend server servicing an onchain service.
	// This can be the serialized response to any type of RPC (gRPC, HTTP, etc.)
	Bytes []byte
	// HTTPStatusCode is the HTTP status returned by an endpoint in response to a relay request.
	HTTPStatusCode int

	// EndpointAddr is the address of the endpoint which returned the response.
	EndpointAddr
}
