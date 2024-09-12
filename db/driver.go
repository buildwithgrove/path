package db

import (
	"context"

	"github.com/buildwithgrove/path/user"
)

// DBDriver is a general purpose interface that is expected to be implemented by
// each specific database (e.g. postgres, sqlite, MySQL, etc) driver.
type DBDriver interface {
	GetGatewayEndpoints(ctx context.Context) (map[user.EndpointID]user.GatewayEndpoint, error)
}
