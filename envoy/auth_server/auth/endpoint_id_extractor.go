package auth

import (
	"fmt"
	"strings"

	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
)

type EndpointIDExtractorType string

const (
	EndpointIDExtractorTypeURLPath EndpointIDExtractorType = "url_path"
	EndpointIDExtractorTypeHeader  EndpointIDExtractorType = "header"
)

func (e EndpointIDExtractorType) IsValid() bool {
	switch e {
	case EndpointIDExtractorTypeURLPath, EndpointIDExtractorTypeHeader:
		return true
	default:
		return false
	}
}

// EndpointIDExtractor defines an interface for extracting an endpoint ID
// from a given source, which could be a URL path or an HTTP header.
type EndpointIDExtractor interface {
	// Extract extracts the endpoint ID from the provided source.
	// The sourceType parameter specifies whether the source is a "path" or "header".
	extractGatewayEndpointID(req *envoy_auth.AttributeContext_HttpRequest) (string, error)
}

// URLPathExtractor satisfies the EndpointIDExtractor interface.
var _ EndpointIDExtractor = &URLPathExtractor{}

// URLPathExtractor is an implementation of the EndpointIDExtractor interface
// that extracts the endpoint ID from the URL path.
type URLPathExtractor struct{}

// ExtractGatewayEndpointID extracts the endpoint ID from the URL path.
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
// The endpoint ID is expected to be in the "x-endpoint-id" header.
//
// eg. Header = "x-endpoint-id: 1a2b3c4d" -> endpointID = "1a2b3c4d"
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