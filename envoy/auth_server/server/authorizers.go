//go:build auth_server

package server

import (
	"fmt"

	"github.com/buildwithgrove/auth-server/user"
)

type ProviderUserIDAuthorizer struct{}

// authorizeRequest checks if the account user ID is authorized to access the endpoint
func (a *ProviderUserIDAuthorizer) authorizeRequest(providerUserID user.ProviderUserID, endpoint user.GatewayEndpoint) error {
	if !endpoint.IsUserAuthorized(providerUserID) {
		return fmt.Errorf("user is not authorized to access this endpoint")
	}
	return nil
}
