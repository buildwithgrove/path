package config

import (
	"fmt"
	"net/url"

	"github.com/buildwithgrove/path/protocol"
)

// FallbackURLs is a map of service IDs to fallback URLs.
// It is unmarshaled from the YAML field `fallback_urls`.
//
// Example YAML:
//
//	fallback_urls:
//	  eth: https://eth.rpc.grove.city/v1/1a2b3c4d
type FallbackURLs map[protocol.ServiceID]string

// URLsFromStrings returns a map of service IDs to fallback URLs.
func (f FallbackURLs) URLsFromStrings() map[protocol.ServiceID]*url.URL {
	fallbackURLs := make(map[protocol.ServiceID]*url.URL, len(f))

	for serviceID, urlString := range f {
		// URL strings are validated in the validate method
		// so we can safely ignore the error here.
		parsedURL, _ := url.Parse(urlString)
		fallbackURLs[serviceID] = parsedURL
	}

	return fallbackURLs
}

// validate ensures that all fallback URL strings in the config YAML are valid URLs.
func (f FallbackURLs) validate() error {
	for serviceID, urlString := range f {
		_, err := url.Parse(urlString)
		if err != nil {
			return fmt.Errorf("invalid fallback URL for service ID %s: %w", serviceID, err)
		}
	}
	return nil
}
