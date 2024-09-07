package driver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// The PostgresDriver struct satisfies the Driver interface which defines all database driver methods
type PostgresDriver struct {
	*Queries
	DB *pgxpool.Pool
}

/* ---------- Postgres Connection Funcs ---------- */

/*
NewPostgresDriver
- Creates a pool of connections to a PostgreSQL database using the provided connection string.
- Parses the connection string into a pgx pool configuration object.
- For each acquired connection from the pool, custom enum types are registered.
- Creates an instance of PostgresDriver using the provided pgx connection.
- Returns the created PostgresDriver instance.
*/
func NewPostgresDriver(connectionString string) (*PostgresDriver, func() error, error) {
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, nil, err
	}

	pool, err := createAndConfigurePool(config)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() error {
		pool.Close()
		return nil
	}

	driver := &PostgresDriver{
		Queries: New(pool),
		DB:      pool,
	}

	return driver, cleanup, nil
}

// Configures the connection pool with custom enums
func createAndConfigurePool(config *pgxpool.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.NewWithConfig: %v", err)
	}

	return pool, nil
}

// Ping ensures the database connection is healthy
func (d *PostgresDriver) Ping(ctx context.Context) error {
	return d.DB.Ping(ctx)
}

/* ---------- UserApp Funcs ---------- */

func (d *PostgresDriver) GetUserApps(ctx context.Context) (map[UserAppID]UserApp, error) {
	rows, err := d.Queries.SelectUserApps(ctx)
	if err != nil {
		return nil, err
	}

	return d.convertToUserApps(rows)
}

func (d *PostgresDriver) convertToUserApps(rows []SelectUserAppsRow) (map[UserAppID]UserApp, error) {
	apps := make(map[UserAppID]UserApp, len(rows))

	for _, row := range rows {
		var whitelists map[WhitelistType]map[WhitelistValue]struct{}
		if err := json.Unmarshal(row.Whitelists, &whitelists); err != nil {
			return nil, err
		}

		app := UserApp{
			ID:                UserAppID(row.ID),
			AccountID:         AccountID(row.AccountID.String),
			PlanType:          row.Plan.String,
			SecretKey:         row.SecretKey.String,
			SecretKeyRequired: row.SecretKeyRequired.Bool,
			ThroughputLimit:   row.ThroughputLimit.Int32,
			Whitelists:        whitelists,
		}

		apps[app.ID] = app
	}

	return apps, nil
}
