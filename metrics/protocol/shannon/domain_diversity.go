package shannon

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/buildwithgrove/path/protocol"
)

// GetEndpointTLDs extracts TLD information from endpoint addresses
func GetEndpointTLDs(endpoints protocol.EndpointAddrList) map[protocol.EndpointAddr]string {
	endpointTLDs := make(map[protocol.EndpointAddr]string)

	for _, endpointAddr := range endpoints {
		endpointTLDs[endpointAddr] = ExtractTLDFromEndpointAddr(string(endpointAddr))
	}

	return endpointTLDs
}

// ExtractTLDFromEndpointAddr extracts effective TLD+1 from endpoint address.
// Returns an empty string if the TLD cannot be determined.
func ExtractTLDFromEndpointAddr(addr string) string {
	// Try direct URL parsing first
	if etld, err := ExtractDomainOrHost(addr); err == nil {
		return etld
	}

	// Handle embedded URLs (e.g., "supplier-https://example.com")
	if idx := strings.Index(addr, "http"); idx != -1 {
		if etld, err := ExtractDomainOrHost(addr[idx:]); err == nil {
			return etld
		}
	}

	// Fallback: try adding https:// prefix for domain-like strings
	parts := strings.FieldsFunc(addr, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	})

	for _, part := range parts {
		if strings.Contains(part, ".") && !strings.HasPrefix(part, "http") {
			if etld, err := ExtractDomainOrHost("https://" + part); err == nil {
				return etld
			}
		}
	}

	return ""
}

// SelectEndpointWithDifferentTLD attempts to select an endpoint with a TLD that hasn't been used yet
func SelectEndpointWithDifferentTLD(
	availableEndpoints protocol.EndpointAddrList,
	endpointTLDs map[protocol.EndpointAddr]string,
	usedTLDs map[string]bool,
) (protocol.EndpointAddr, error) {
	// Filter endpoints to only those with different TLDs
	var endpointsWithDifferentTLDs protocol.EndpointAddrList

	for _, endpoint := range availableEndpoints {
		if tld, exists := endpointTLDs[endpoint]; exists {
			if !usedTLDs[tld] {
				endpointsWithDifferentTLDs = append(endpointsWithDifferentTLDs, endpoint)
			}
		} else {
			// If we can't determine TLD, include it anyway
			endpointsWithDifferentTLDs = append(endpointsWithDifferentTLDs, endpoint)
		}
	}

	if len(endpointsWithDifferentTLDs) == 0 {
		return "", fmt.Errorf("no endpoints with different TLDs available")
	}

	// Select a random endpoint from the filtered list
	return endpointsWithDifferentTLDs[rand.Intn(len(endpointsWithDifferentTLDs))], nil
}
