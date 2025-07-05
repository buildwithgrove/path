package shannon

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
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

// ExtractDomainFromURL extracts the domain from a URL for metrics labeling
func ExtractDomainFromURL(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "unknown"
	}

	// Extract hostname and remove port if present
	hostname := parsedURL.Hostname()
	if hostname == "" {
		return "unknown"
	}

	// For IP addresses or localhost, return as-is
	if strings.Contains(hostname, "127.0.0.1") || strings.Contains(hostname, "localhost") {
		return "localhost"
	}

	// For domain names, try to extract TLD+1 (simplified)
	parts := strings.Split(hostname, ".")
	if len(parts) >= 2 {
		// Return last two parts (domain.tld)
		return strings.Join(parts[len(parts)-2:], ".")
	}

	return hostname
}
