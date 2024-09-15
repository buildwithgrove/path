package db

import (
	"context"

	"github.com/buildwithgrove/authorizer-plugin/user"
)

// DBDriver is a general purpose interface that must be implemented by any database (e.g. postgres, sqlite, MySQL, etc) driver.
type DBDriver interface {
	// GetGatewayEndpoints returns all GatewayEndpoints in the database and is used to set the in-memory cache.
	GetGatewayEndpoints(ctx context.Context) (map[user.EndpointID]user.GatewayEndpoint, error)
}
