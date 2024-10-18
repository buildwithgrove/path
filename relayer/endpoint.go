package relayer

// Endpoint represents an entity which serves relay requests.
type Endpoint interface {
	// Addr is used to uniquely identify an endpoint
	Addr() EndpointAddr
	// PublicURL is the URL to which relay requests can be sent.
	PublicURL() string
}

// EndpointSelector defines the functionality that the user of a relayer needs to provide,
// i.e. selecting an endpoint, from the list of available ones, to which the relay is to be sent.
type EndpointSelector interface {
	Select([]Endpoint) (EndpointAddr, error)
}

// EndpointAddr is used as the unique identifier for an endpoint.
type EndpointAddr string
