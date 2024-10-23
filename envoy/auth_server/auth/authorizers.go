package auth

import (
	"fmt"

	"github.com/buildwithgrove/auth-server/proto"
)

type ProviderUserIDAuthorizer struct{}

// authorizeRequest checks if the account user ID is authorized to access the endpoint
func (a *ProviderUserIDAuthorizer) authorizeRequest(providerUserID string, endpoint *proto.GatewayEndpoint) error {
	if _, ok := endpoint.GetAuth().GetAuthorizedUsers()[providerUserID]; !ok {
		return fmt.Errorf("user is not authorized to access this endpoint")
	}
	return nil
}
