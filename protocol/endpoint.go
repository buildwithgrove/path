package protocol

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
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

// GetDomain returns the effective TLD+1 domain of the endpoint address.
// For example:
// - Given the endpoint address "pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq-https://relayminer.shannon-mainnet.eu.nodefleet.net"
// - Would return "nodefleet.net"
// - Given "pokt1d3atlnepcvsa9j5uunpvf64g80eucjqtem77mn-https://rm02-eu.dopokt.com:443"
// - Would return "dopokt.com"
func (e EndpointAddr) GetDomain() (string, error) {
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
	return extractDomainOrHost(urlPart)
}

// GetAddress returns the address of the endpoint.
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

// extractDomainOrHost extracts the effective TLD+1 from a URL.
// It falls back to a reasonable domain extraction for localhost, IP addresses, and other non-standard hosts.
// This is a copy of the logic from path/metrics/protocol/shannon/domain.go to avoid import cycles.
func extractDomainOrHost(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("malformed URL: %w", err)
	}

	host := parsedURL.Hostname()
	if host == "" {
		return "", fmt.Errorf("empty host in URL")
	}

	// Try to get effective TLD+1 first
	etld, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err == nil {
		return etld, nil
	}

	// Fallback cases when publicsuffix fails
	return fallbackDomainExtraction(host)
}

// fallbackDomainExtraction handles cases where publicsuffix.EffectiveTLDPlusOne fails
func fallbackDomainExtraction(host string) (string, error) {
	// Check if it's an IP address
	if ip := net.ParseIP(host); ip != nil {
		return host, nil // Return the IP as-is
	}

	// Check for localhost variants
	if isLocalhost(host) {
		return host, nil
	}

	// Check for private/internal domains (no dots, or .local, etc.)
	if isPrivateOrInternalDomain(host) {
		return host, nil
	}

	// For other cases, try to extract a reasonable domain
	// This handles cases like "relayminer1" or custom internal hostnames
	parts := strings.Split(host, ".")
	if len(parts) == 1 {
		// Single hostname without dots (like "relayminer1")
		return host, nil
	}

	// If it has dots but publicsuffix failed, take the last two parts
	// This is a reasonable fallback for unknown TLDs
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], "."), nil
	}

	return host, nil
}

// isLocalhost checks if the host is a localhost variant
func isLocalhost(host string) bool {
	lowercase := strings.ToLower(host)
	return lowercase == "localhost" ||
		lowercase == "localhost.localdomain" ||
		strings.HasPrefix(lowercase, "localhost.")
}

// isPrivateOrInternalDomain checks for private/internal domain patterns
func isPrivateOrInternalDomain(host string) bool {
	lowercase := strings.ToLower(host)

	// Common internal TLDs
	internalTLDs := []string{".local", ".internal", ".corp", ".home", ".lan"}
	for _, tld := range internalTLDs {
		if strings.HasSuffix(lowercase, tld) {
			return true
		}
	}

	// Single label domains (no dots) are typically internal
	return !strings.Contains(host, ".")
}
