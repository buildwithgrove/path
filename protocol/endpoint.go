package protocol

import "strings"

// EndpointAddr is used as the unique identifier for a service endpoint.
// Each protocol interface implementation needs to define an endpoint address which uniquely identifies a service endpoint.
// As of writing this comment(#50):
// - Morse (POKT): uses a node's public key as its endpoint address
// - Shannon (POKT): appends the URL of each endpoint configured for a Shannon supplier to its operator address to form endpoint addresses.
type EndpointAddr string

type EndpointAddrList []EndpointAddr

// Endpoint represents an entity which serves relay requests.
type Endpoint interface {
	// Addr is used to uniquely identify an endpoint.
	// Defining this as an interface allows each protocol interface implementation (e.g. Pocket's Morse and Shannon) to
	// define its own service endpoint address scheme.
	// See the comment on EndpointAddr type for more details.
	Addr() EndpointAddr

	// PublicURL is the publically exposed/accessible URL to which relay requests can be sent.
	PublicURL() string

	// WebsocketURL is the URL of the endpoint for websocket RPC type requests.
	// Returns an error if the endpoint does not support websocket RPC type requests.
	WebsocketURL() (string, error)
}

// EndpointSelector defines the functionality that the user of a protocol needs to provide.
// E.g. selecting an endpoint, from the list of available ones, to which the relay will be sent.
type EndpointSelector interface {
	Select(EndpointAddrList) (EndpointAddr, error)
	SelectMultiple(EndpointAddrList, uint) (EndpointAddrList, error)
}

func (e EndpointAddr) String() string {
	return string(e)
}

func (e EndpointAddr) Decompose() (supplierAddress string, endpointURL string) {
	parts := strings.Split(string(e), "-")
	supplierAddress = parts[0]
	endpointURL = parts[1]
	return
}

func (e EndpointAddrList) String() string {
	// Converts each EndpointAddr to string and joins them with a comma
	addrs := make([]string, len(e))
	for i, addr := range e {
		addrs[i] = string(addr)
	}
	return strings.Join(addrs, ", ")
}
