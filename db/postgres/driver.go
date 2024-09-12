package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/buildwithgrove/path/db"
	"github.com/buildwithgrove/path/user"
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

/* ---------- Qu Funcs ---------- */

func (d *postgresDriver) GetGatewayEndpoints(ctx context.Context) (map[user.EndpointID]user.GatewayEndpoint, error) {
	rows, err := d.Queries.SelectGatewayEndpoints(ctx)
	if err != nil {
		return nil, err
	}

	return d.convertToGatewayEndpoints(rows)
}

func (d *postgresDriver) convertToGatewayEndpoints(rows []SelectGatewayEndpointsRow) (map[user.EndpointID]user.GatewayEndpoint, error) {
	gatewayEndpoints := make(map[user.EndpointID]user.GatewayEndpoint, len(rows))

	for _, row := range rows {
		gatewayEndpoint := user.GatewayEndpoint{
			EndpointID: user.EndpointID(row.ID),
			Auth: user.Auth{
				APIKey:         row.ApiKey.String,
				APIKeyRequired: row.ApiKeyRequired.Bool,
			},
			UserAccount: user.UserAccount{
				AccountID: user.AccountID(row.AccountID.String),
				PlanType:  user.PlanType(row.Plan.String),
			},
			RateLimiting: user.RateLimiting{
				ThroughputLimit:     int(row.RateLimitThroughput.Int32),
				CapacityLimit:       int(row.RateLimitCapacity.Int32),
				CapacityLimitPeriod: user.CapacityLimitPeriod(row.RateLimitCapacityPeriod.RateLimitCapacityPeriod),
			},
		}

		gatewayEndpoints[gatewayEndpoint.EndpointID] = gatewayEndpoint
	}

	return gatewayEndpoints, nil
}
