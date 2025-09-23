package http

import (
	"errors"
	"net/url"

	"golang.org/x/net/publicsuffix"
)

const (
	EmptyHostDomain = "empty host"
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
		return "", errors.New(EmptyHostDomain)
	}

	etld, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return "", err // domain may not be derivable (e.g., IP, localhost)
	}
	return etld, nil
}
