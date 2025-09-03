package http

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// ExtractEffectiveTLDPlusOne extracts the "effective TLD+1" (eTLD+1) from a given URL.
// Example: "https://blog.example.co.uk" → "example.co.uk"
// - Parses the URL and validates the host.
// - Uses publicsuffix package to determine the registrable domain.
// - Returns an error if input is malformed or domain is not derivable.
func ExtractEffectiveTLDPlusOne(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err // malformed URL
	}

	host := parsedURL.Hostname()
	if host == "" {
		return "", fmt.Errorf("empty host") // no host in URL
	}

	etld, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return "", err // domain may not be derivable (e.g., IP, localhost)
	}
	return etld, nil
}

// ExtractDomainFromEndpointAddr extracts the eTLD+1 domain from an endpoint address.
// Handles the Shannon endpoint address format: "pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq-https://relayminer.shannon-mainnet.eu.nodefleet.net"
// Returns "unknown" if domain cannot be extracted.
func ExtractDomainFromEndpointAddr(logger polylog.Logger, endpointAddr string) string {
	// Split by dash to separate the address part from the URL part
	parts := strings.Split(endpointAddr, "-")
	if len(parts) < 2 {
		// No dash found, try to extract domain directly from the entire string
		if domain, err := ExtractEffectiveTLDPlusOne(endpointAddr); err == nil {
			return domain
		}
		logger.Debug().Str("endpoint_addr", endpointAddr).Msg("Could not extract domain from endpoint address - no dash separator found")
		return "unknown"
	}

	// Take everything after the first dash as the URL
	urlPart := strings.Join(parts[1:], "-")

	// Try to extract domain from the URL part
	if domain, err := ExtractEffectiveTLDPlusOne(urlPart); err == nil {
		return domain
	}

	logger.Debug().Str("endpoint_addr", endpointAddr).Str("url_part", urlPart).Msg("Could not extract eTLD+1 from URL part")

	// If domain extraction failed, return unknown
	return "unknown"
}
