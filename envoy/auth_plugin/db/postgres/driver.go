//go:build auth_plugin

package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/buildwithgrove/auth-plugin/db"
	"github.com/buildwithgrove/auth-plugin/types"
)

// The postgresDriver struct satisfies the db.Driver interface defined in the db package.
type postgresDriver struct {
	*Queries
	DB *pgxpool.Pool
}

var _ db.DBDriver = &postgresDriver{}

/* ---------- Postgres Connection Funcs ---------- */

/*
NewPostgresDriver
- Creates a pool of connections to a PostgreSQL database using the provided connection string.
- Parses the connection string into a pgx pool configuration object.
- For each acquired connection from the pool, custom enum types are registered.
- Creates an instance of PostgresDriver using the provided pgx connection.
- Returns the created PostgresDriver instance.
*/
func NewPostgresDriver(connectionString string) (*postgresDriver, func() error, error) {
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, nil, err
	}

	// Enforce that connections are read-only, as PATH does not make any modifications to the database.
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY")
		return err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, nil, fmt.Errorf("pgxpool.NewWithConfig: %v", err)
	}

	cleanup := func() error {
		pool.Close()
		return nil
	}

	driver := &postgresDriver{
		Queries: New(pool),
		DB:      pool,
	}

	return driver, cleanup, nil
}

/* ---------- Query Funcs ---------- */

// GetGatewayEndpoints retrieves all GatewayEndpoints from the database and returns them as a map.
func (d *postgresDriver) GetGatewayEndpoints(ctx context.Context) (map[types.EndpointID]types.GatewayEndpoint, error) {
	rows, err := d.Queries.SelectGatewayEndpoints(ctx)
	if err != nil {
		return nil, err
	}

	return d.convertToGatewayEndpoints(rows)
}

// convertToGatewayEndpoints converts a slice of the SelectGatewayEndpointsRow struct fetched from
// the database to a map of the types.GatewayEndpoint struct that is used throughout the repo.
func (d *postgresDriver) convertToGatewayEndpoints(rows []SelectGatewayEndpointsRow) (map[types.EndpointID]types.GatewayEndpoint, error) {
	gatewayEndpoints := make(map[types.EndpointID]types.GatewayEndpoint, len(rows))

	for _, row := range rows {
		gatewayEndpoint := types.GatewayEndpoint{
			EndpointID: types.EndpointID(row.ID),
			Auth: types.Auth{
				APIKey:         row.ApiKey.String,
				APIKeyRequired: row.ApiKeyRequired.Bool,
			},
			UserAccount: types.UserAccount{
				AccountID: types.AccountID(row.AccountID.String),
				PlanType:  types.PlanType(row.Plan.String),
			},
			RateLimiting: types.RateLimiting{
				ThroughputLimit:     int(row.RateLimitThroughput.Int32),
				CapacityLimit:       int(row.RateLimitCapacity.Int32),
				CapacityLimitPeriod: types.CapacityLimitPeriod(row.RateLimitCapacityPeriod.RateLimitCapacityPeriod),
			},
		}

		gatewayEndpoints[gatewayEndpoint.EndpointID] = gatewayEndpoint
	}

	return gatewayEndpoints, nil
}
