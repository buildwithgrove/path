package auth

import (
	"github.com/buildwithgrove/path/envoy/auth_server/proto"
)

const reqHeaderAPIKey = "authorization" // Standard header for API keys

// APIKeyAuthorizer is an Authorizer that ensures the request is authorized
// by checking if the API key matches the GatewayEndpoint's API key.
type APIKeyAuthorizer struct{}

// Enforce that the APIKeyAuthorizer implements the Authorizer interface.
var _ Authorizer = &APIKeyAuthorizer{}

// authorizeRequest checks if the API key is valid for the endpoint
func (a *APIKeyAuthorizer) authorizeRequest(headers map[string]string, endpoint *proto.GatewayEndpoint) error {
	apiKey, ok := headers[reqHeaderAPIKey]
	if !ok || apiKey == "" {
		return errUnauthorized
	}
	if endpoint.GetAuth().GetStaticApiKey().GetApiKey() != apiKey {
		return errUnauthorized
	}
	return nil
}
