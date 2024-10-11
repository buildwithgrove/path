//go:build auth_server

package server

import (
	"fmt"

	"github.com/buildwithgrove/auth-server/user"
)

type AccountUserIDAuthorizer struct{}

// authorizeRequest checks if the account user ID is authorized to access the endpoint
func (a *AccountUserIDAuthorizer) authorizeRequest(accountUserID user.AccountUserID, endpoint user.GatewayEndpoint) error {
	if !endpoint.IsUserAuthorized(accountUserID) {
		return fmt.Errorf("user is not authorized to access this endpoint")
	}
	return nil
}
