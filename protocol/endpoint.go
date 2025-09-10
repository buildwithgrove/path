package protocol

import (
	"fmt"
	"strings"
)

// EndpointAddr is used as the unique identifier for a service endpoint.
// operator address and the endpoint's URL, separated by a "-" character.
//
// For example:
//   - "pokt1ggdpwj5stslx2e567qcm50wyntlym5c4n0dst8-https://im.oldgreg.org"
type EndpointAddr string

type EndpointAddrList []EndpointAddr

// Endpoint represents an entity which serves relay requests.
type Endpoint interface {
	// Addr is used to uniquely identify an endpoint.
	// Defining this as an interface allows Shannon to
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

func (e EndpointAddrList) String() string {
	// Converts each EndpointAddr to string and joins them with a comma
	addrs := make([]string, len(e))
	for i, addr := range e {
		addrs[i] = string(addr)
	}
	return strings.Join(addrs, ", ")
}

func (e EndpointAddr) String() string {
	return string(e)
}

// GetURL returns the URL part of the endpoint address.
// For example:
// - Given the endpoint address "pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq-https://relayminer.shannon-mainnet.eu.nodefleet.net"
// - Would return "https://relayminer.shannon-mainnet.eu.nodefleet.net"
func (e EndpointAddr) GetURL() (string, error) {
	// Find the first dash to separate supplier address from URL
	// This handles cases where the URL itself contains dashes
	dashIndex := strings.Index(e.String(), "-")
	if dashIndex == -1 {
		return "", fmt.Errorf("endpoint address %s does not contain a dash separator", e.String())
	}

	// Extract the URL part (everything after the first dash)
	urlPart := e.String()[dashIndex+1:]
	if urlPart == "" {
		return "", fmt.Errorf("endpoint address %s has empty URL part", e.String())
	}

	// Use enhanced domain extraction with fallback handling
	// for edge cases like localhost, IPs, and non-standard domains
	return urlPart, nil
}

// GetAddress returns the address part of the endpoint.
// For example:
// - Given the endpoint address "pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq-https://relayminer.shannon-mainnet.eu.nodefleet.net"
// - Would return "pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq"
func (e EndpointAddr) GetAddress() (string, error) {
	// Find the first dash to separate supplier address from URL
	// This handles cases where the URL itself contains dashes
	dashIndex := strings.Index(e.String(), "-")
	if dashIndex == -1 {
		return "", fmt.Errorf("endpoint address %s does not contain a dash separator", e.String())
	}

	// Extract the supplier address part (everything before the first dash)
	return e.String()[:dashIndex], nil
}
