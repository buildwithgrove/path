package auth

import (
	"github.com/buildwithgrove/path/envoy/auth_server/proto"
)

const reqHeaderJWTUserID = "jwt-user-id" // Defined in envoy.yaml

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
