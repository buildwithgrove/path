//go:build auth_server

package db

import (
	"context"

	"github.com/buildwithgrove/auth-server/user"
)

// TODO_UPNEXT(@commoddity): Investigate alternative authentication solutions to in-house rolled API key, eg. Clerk, Auth0, etc.
// TODO_UPNEXT(@commoddity): Implement development solution that abstracts away database implementation in favour of Envoy.

// DBDriver is a general purpose interface that must be implemented by any database (e.g. postgres, sqlite, MySQL, etc) driver.
type DBDriver interface {
	// GetGatewayEndpoints returns all GatewayEndpoints in the database and is used to set the in-memory cache.
	GetGatewayEndpoints(ctx context.Context) (map[user.EndpointID]user.GatewayEndpoint, error)
}
