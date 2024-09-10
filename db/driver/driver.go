package driver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/buildwithgrove/path/user"
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

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, nil, fmt.Errorf("pgxpool.NewWithConfig: %v", err)
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

/* ---------- UserApp Funcs ---------- */

func (d *PostgresDriver) GetUserApps(ctx context.Context) (map[user.UserAppID]user.UserApp, error) {
	rows, err := d.Queries.SelectUserApps(ctx)
	if err != nil {
		return nil, err
	}

	return d.convertToUserApps(rows)
}

func (d *PostgresDriver) convertToUserApps(rows []SelectUserAppsRow) (map[user.UserAppID]user.UserApp, error) {
	apps := make(map[user.UserAppID]user.UserApp, len(rows))

	for _, row := range rows {
		var allowlists map[user.AllowlistType]map[string]struct{}
		if err := json.Unmarshal(row.Allowlists, &allowlists); err != nil {
			return nil, err
		}

		app := user.UserApp{
			ID:                  user.UserAppID(row.ID),
			AccountID:           user.AccountID(row.AccountID.String),
			PlanType:            user.PlanType(row.Plan.String),
			SecretKey:           row.SecretKey.String,
			SecretKeyRequired:   row.SecretKeyRequired.Bool,
			RateLimitThroughput: int(row.RateLimitThroughput.Int32),
			Allowlists:          allowlists,
		}

		apps[app.ID] = app
	}

	return apps, nil
}
