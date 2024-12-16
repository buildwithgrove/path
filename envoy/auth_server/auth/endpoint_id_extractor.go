package auth

import (
	"fmt"
	"strings"

	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
)

// EndpointIDExtractorType specifies the type of endpoint ID extractor to use.
type EndpointIDExtractorType string

const (
	// EndpointIDExtractorTypeURLPath specifies that the endpoint ID is extracted from the URL path.
	// Example: http://eth.path.grove.city/v1/1a2b3c4d -> endpointID = "1a2b3c4d"
	EndpointIDExtractorTypeURLPath EndpointIDExtractorType = "url_path"

	// EndpointIDExtractorTypeHeader specifies that the endpoint ID is extracted from the HTTP headers.
	// Example: Header = "endpoint-id: 1a2b3c4d" -> endpointID = "1a2b3c4d"
	EndpointIDExtractorTypeHeader EndpointIDExtractorType = "header"
)

// IsValid ensure the endpoint ID extractor type is supported.
func (e EndpointIDExtractorType) IsValid() bool {
	switch e {
	case EndpointIDExtractorTypeURLPath, EndpointIDExtractorTypeHeader:
		return true
	default:
		return false
	}
}

// EndpointIDExtractor defines an interface for extracting an endpoint ID from a given source.
// This could be a URL path, HTTP header, etc...
type EndpointIDExtractor interface {
	// extractGatewayEndpointID extracts the endpoint ID from the check request.
	extractGatewayEndpointID(req *envoy_auth.AttributeContext_HttpRequest) (string, error)
}

// URLPathExtractor satisfies the EndpointIDExtractor interface.
var _ EndpointIDExtractor = &URLPathExtractor{}

// URLPathExtractor is an implementation of the EndpointIDExtractor interface
// that extracts the endpoint ID from the URL path.
type URLPathExtractor struct{}

// extractGatewayEndpointID extracts the endpoint ID from the URL path.
// The endpoint ID is expected to be the first segment of the path after the pathPrefix (/v1/)
//
// eg. http://eth.path.grove.city/v1/1a2b3c4d -> endpointID = "1a2b3c4d"
func (p *URLPathExtractor) extractGatewayEndpointID(req *envoy_auth.AttributeContext_HttpRequest) (string, error) {
	path := req.GetPath()

	if strings.HasPrefix(path, pathPrefix) {
		segments := strings.Split(strings.TrimPrefix(path, pathPrefix), "/")
		if len(segments) > 0 && segments[0] != "" {
			return segments[0], nil
		}
	}

	return "", fmt.Errorf("endpoint ID not provided")
}

// HeaderExtractor satisfies the EndpointIDExtractor interface.
var _ EndpointIDExtractor = &HeaderExtractor{}

// HeaderExtractor is an implementation of the EndpointIDExtractor interface
// that extracts the endpoint ID from the HTTP headers.
type HeaderExtractor struct{}

// ExtractGatewayEndpointID extracts the endpoint ID from the HTTP headers.
// The endpoint ID is expected to be in the "endpoint-id" header.
//
// eg. Header = "endpoint-id: 1a2b3c4d" -> endpointID = "1a2b3c4d"
func (h *HeaderExtractor) extractGatewayEndpointID(req *envoy_auth.AttributeContext_HttpRequest) (string, error) {
	headers := req.GetHeaders()

	endpointID, ok := headers[reqHeaderEndpointID]
	if !ok {
		return "", fmt.Errorf("endpoint ID header not found")
	}
	if endpointID == "" {
		return "", fmt.Errorf("endpoint ID not provided")
	}

	return endpointID, nil
}
