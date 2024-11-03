package auth

import (
	"fmt"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
)

// The Authorizer interface performs requests authorization, for example using
// API key authentication to ensures a downstream (client) request is authorized.
type Authorizer interface {
	authorizeRequest(string, *proto.GatewayEndpoint) error
}

// ProviderUserIDAuthorizer is an Authorizer that ensures the request is authorized
// by checking if the account user ID is in the GatewayEndpoint's authorized users.
type ProviderUserIDAuthorizer struct{}

// Enforce that the ProviderUserIDAuthorizer implements the Authorizer interface.
var _ Authorizer = &ProviderUserIDAuthorizer{}

// authorizeRequest checks if the account user ID is authorized to access the endpoint
func (a *ProviderUserIDAuthorizer) authorizeRequest(providerUserID string, endpoint *proto.GatewayEndpoint) error {
	if _, ok := endpoint.GetAuth().GetAuthorizedUsers()[providerUserID]; !ok {
		return fmt.Errorf("user is not authorized to access this endpoint")
	}
	return nil
}
