package auth

import (
	"fmt"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
)

const (
	reqHeaderAPIKey    = "authorization" // Standard header for API keys
	reqHeaderJWTUserID = "x-jwt-user-id" // Defined in envoy.yaml
)

// errUnauthorized is returned when a request is not authorized.
// It is left intentionally vague to avoid leaking information to the client.
var errUnauthorized = fmt.Errorf("unauthorized")

// The Authorizer interface performs requests authorization, for example using
// API key authentication to ensures a downstream (client) request is authorized.
type Authorizer interface {
	authorizeRequest(headers map[string]string, endpoint *proto.GatewayEndpoint) error
}

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
	if endpoint.GetAuth().GetApiKey().GetApiKey() != apiKey {
		return errUnauthorized
	}
	return nil
}

// JWTAuthorizer is an Authorizer that ensures the request is authorized
// by checking if the account user ID is in the GatewayEndpoint's authorized users.
type JWTAuthorizer struct{}

// Enforce that the JWTAuthorizer implements the Authorizer interface.
var _ Authorizer = &JWTAuthorizer{}

// authorizeRequest checks if the account user ID is authorized to access the endpoint
func (a *JWTAuthorizer) authorizeRequest(headers map[string]string, endpoint *proto.GatewayEndpoint) error {
	providerUserID, ok := headers[reqHeaderJWTUserID]
	if !ok || providerUserID == "" {
		return errUnauthorized
	}
	if _, ok := endpoint.GetAuth().GetJwt().GetAuthorizedUsers()[providerUserID]; !ok {
		return errUnauthorized
	}
	return nil
}
