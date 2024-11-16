package auth

import (
	"fmt"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
)

const (
	apiKeyHeader = "Authorization"
)

// The Authorizer interface performs requests authorization, for example using
// API key authentication to ensures a downstream (client) request is authorized.
type Authorizer interface {
	authorizeRequest(map[string]string, *proto.GatewayEndpoint) error
}

// NoAuthAuthorizer is an Authorizer that ensures no authorization is required.
type NoAuthAuthorizer struct{}

// Enforce that the NoAuthAuthorizer implements the Authorizer interface.
var _ Authorizer = &NoAuthAuthorizer{}

// authorizeRequest always returns nil, as no authorization is required.
func (a *NoAuthAuthorizer) authorizeRequest(headers map[string]string, endpoint *proto.GatewayEndpoint) error {
	return nil
}

// APIKeyAuthorizer is an Authorizer that ensures the request is authorized
// by checking if the API key matches the GatewayEndpoint's API key.
type APIKeyAuthorizer struct{}

// Enforce that the APIKeyAuthorizer implements the Authorizer interface.
var _ Authorizer = &APIKeyAuthorizer{}

// authorizeRequest checks if the API key is valid for the endpoint
func (a *APIKeyAuthorizer) authorizeRequest(headers map[string]string, endpoint *proto.GatewayEndpoint) error {
	apiKey, ok := headers[apiKeyHeader]
	if !ok || apiKey == "" {
		return fmt.Errorf("unauthorized")
	}
	if endpoint.GetAuth().GetApiKey().GetApiKey() != apiKey {
		return fmt.Errorf("invalid API key")
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
	providerUserID, ok := headers[reqHeaderAccountUserID]
	if !ok || providerUserID == "" {
		return fmt.Errorf("unauthorized")
	}
	if _, ok := endpoint.GetAuth().GetJwt().GetAuthorizedUsers()[providerUserID]; !ok {
		return fmt.Errorf("user is not authorized to access this endpoint")
	}
	return nil
}
