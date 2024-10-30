package relayer

// EndpointAddr is used as the unique identifier for a service endpoint.
// Each protocol interface implementation needs to define an endpoint address which uniquely identifies a service endpoint.
// As of writing this comment(#50):
// - Morse (POKT): uses a node's public key as its endpoint address
// - Shannon (POKT): appends the URL of each endpoint configured for a Shannon supplier to its operator address to form endpoint addresses.
type EndpointAddr string

// Endpoint represents an entity which serves relay requests.
type Endpoint interface {
	// Addr is used to uniquely identify an endpoint.
	// Defining this as an interface allows each protocl interface implementation (e.g. Pocket's Morse and Shannon) to
	// define its own service endpoint address scheme.
	// See the comment on EndpointAddr type for more details.
	Addr() EndpointAddr
	// PublicURL is the URL to which relay requests can be sent.
	PublicURL() string
}

// EndpointSelector defines the functionality that the user of a relayer needs to provide.
// E.g. selecting an endpoint, from the list of available ones, to which the relay will be sent.
type EndpointSelector interface {
	Select([]Endpoint) (EndpointAddr, error)
}
