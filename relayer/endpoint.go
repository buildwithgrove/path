package relayer

// EndpointAddr is used as the unique identifier for a service endpoint.
// Each protocol interface implementation needs to define an endpoint address which uniquely identifies a service endpoint.
// As of now, the Morse-based protocol interface implementation, under relayer/morse package, uses a Morse node's public key as its endpoint address.
// The Shannon-based protocol interface implementation, under relayer/shannon package, appends the URL of each endpoint configured for a Shannon supplier to its operator address to form endpoint addresses.
type EndpointAddr string

// Endpoint represents an entity which serves relay requests.
type Endpoint interface {
	// Addr is used to uniquely identify an endpoint. Defining this as an interface allows each protocl interface implementation (Morse and Shannon as of now), to
	// define its own service endpoint address scheme. See the comment on EndpointAddr type for more details.
	Addr() EndpointAddr
	// PublicURL is the URL to which relay requests can be sent.
	PublicURL() string
}

// EndpointSelector defines the functionality that the user of a relayer needs to provide,
// i.e. selecting an endpoint, from the list of available ones, to which the relay is to be sent.
type EndpointSelector interface {
	Select([]Endpoint) (EndpointAddr, error)
}
