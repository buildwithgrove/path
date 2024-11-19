package protocol

// Response is a general purpose struct for capturing the response to a relay, received from an endpoint.
// TODO_FUTURE: It only supports HTTP responses for now.
type Response struct {
	// Bytes is the response to a relay received from an endpoint.
	// This can be a response to any type of RPC(GRPC, HTTP, etc.)
	Bytes []byte
	// HTTPStatusCode is the HTTP status returned by an endpoint in response to a relay request.
	HTTPStatusCode int

	// EndpointAddr is the address of the endpoint which returned the response.
	EndpointAddr
}
