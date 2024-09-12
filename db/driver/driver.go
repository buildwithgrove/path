package driver

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/buildwithgrove/path/db"
	"github.com/buildwithgrove/path/user"
)

// The postgresDriver struct satisfies the db.Driver interface defined in the cache in the db package.
type postgresDriver struct {
	*Queries
	DB *pgxpool.Pool
}

var _ db.Driver = &postgresDriver{}

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

/* ---------- UserApp Funcs ---------- */

func (d *postgresDriver) GetUserApps(ctx context.Context) (map[user.UserAppID]user.UserApp, error) {
	rows, err := d.Queries.SelectUserApps(ctx)
	if err != nil {
		return nil, err
	}

	return d.convertToUserApps(rows)
}

func (d *postgresDriver) convertToUserApps(rows []SelectUserAppsRow) (map[user.UserAppID]user.UserApp, error) {
	apps := make(map[user.UserAppID]user.UserApp, len(rows))

	for _, row := range rows {
		app := user.UserApp{
			ID:                  user.UserAppID(row.ID),
			AccountID:           user.AccountID(row.AccountID.String),
			PlanType:            user.PlanType(row.Plan.String),
			SecretKey:           row.SecretKey.String,
			SecretKeyRequired:   row.SecretKeyRequired.Bool,
			RateLimitThroughput: int(row.RateLimitThroughput.Int32),
		}

		apps[app.ID] = app
	}

	return apps, nil
}
