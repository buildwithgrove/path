package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"

	"github.com/buildwithgrove/path/user"
)

var connectionString string
var dbConn *pgx.Conn

func TestMain(m *testing.M) {
	// Initialize the ephemeral postgres docker container
	pool, resource, databaseURL := setupPostgresDocker()
	connectionString = databaseURL

	// Run DB integration test
	exitCode := m.Run()

	// Cleanup the ephemeral postgres docker container
	cleanupPostgresDocker(m, pool, resource)
	os.Exit(exitCode)
}

func Test_Integration_GetUserApps(t *testing.T) {
	tests := []struct {
		name     string
		expected map[user.UserAppID]user.UserApp
	}{
		{
			name: "should retrieve all user apps correctly",
			expected: map[user.UserAppID]user.UserApp{
				"user_app_1": {
					ID:                  "user_app_1",
					AccountID:           "account_1",
					PlanType:            "PLAN_FREE",
					SecretKey:           "secret_key_1",
					SecretKeyRequired:   true,
					RateLimitThroughput: 30,
					RateLimitCapacity:   100_000,
					Allowlists: map[user.AllowlistType]map[string]struct{}{
						user.AllowlistTypeServices: {"service_1": {}},
					},
				},
				"user_app_2": {
					ID:                "user_app_2",
					AccountID:         "account_2",
					PlanType:          "PLAN_UNLIMITED",
					SecretKey:         "secret_key_2",
					SecretKeyRequired: true,
					Allowlists: map[user.AllowlistType]map[string]struct{}{
						user.AllowlistTypeContracts: {"contract_1": {}},
					},
				},
				"user_app_3": {
					ID:                  "user_app_3",
					AccountID:           "account_3",
					PlanType:            "PLAN_FREE",
					SecretKey:           "secret_key_3",
					SecretKeyRequired:   true,
					RateLimitThroughput: 30,
					RateLimitCapacity:   100_000,
					Allowlists: map[user.AllowlistType]map[string]struct{}{
						user.AllowlistTypeMethods: {"method_1": {}},
					},
				},
				"user_app_4": {
					ID:                  "user_app_4",
					AccountID:           "account_1",
					PlanType:            "PLAN_FREE",
					RateLimitThroughput: 30,
					RateLimitCapacity:   100_000,
					Allowlists: map[user.AllowlistType]map[string]struct{}{
						user.AllowlistTypeOrigins: {"origin_1": {}},
					},
				},
				"user_app_5": {
					ID:        "user_app_5",
					AccountID: "account_2",
					PlanType:  "PLAN_UNLIMITED",
					Allowlists: map[user.AllowlistType]map[string]struct{}{
						user.AllowlistTypeUserAgents: {"user_agent_1": {}},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if testing.Short() {
				t.Skip("skipping driver integration test")
			}

			c := require.New(t)

			driver, cleanup, err := NewPostgresDriver(connectionString)
			c.NoError(err)
			defer cleanup()

			apps, err := driver.GetUserApps(context.Background())
			c.NoError(err)
			c.Equal(test.expected, apps)
		})
	}
}

func Test_convertToUserApps(t *testing.T) {
	tests := []struct {
		name     string
		rows     []SelectUserAppsRow
		expected map[user.UserAppID]user.UserApp
		wantErr  bool
	}{
		{
			name: "should convert rows to portal apps successfully",
			rows: []SelectUserAppsRow{
				{
					ID:                "app1",
					AccountID:         pgtype.Text{String: "acc1", Valid: true},
					SecretKey:         pgtype.Text{String: "secret1", Valid: true},
					SecretKeyRequired: pgtype.Bool{Bool: true, Valid: true},
					Allowlists: json.RawMessage(`{
						"origins": {
							"origin_1": {}
						},
						"user_agents": {
							"user_agent_1": {}
						},
						"services": {
							"service_1": {}
						},
						"contracts": {
							"contract_1": {}
						},
						"methods": {
							"method_1": {}
						}
					}`),
					Plan:                pgtype.Text{String: "plan1", Valid: true},
					RateLimitThroughput: pgtype.Int4{Int32: 30, Valid: true},
					RateLimitCapacity:   pgtype.Int4{Int32: 100_000, Valid: true},
				},
			},
			expected: map[user.UserAppID]user.UserApp{
				"app1": {
					ID:                  "app1",
					AccountID:           "acc1",
					PlanType:            "plan1",
					SecretKey:           "secret1",
					SecretKeyRequired:   true,
					RateLimitThroughput: 30,
					RateLimitCapacity:   100_000,
					Allowlists: map[user.AllowlistType]map[string]struct{}{
						user.AllowlistTypeOrigins:    {"origin_1": {}},
						user.AllowlistTypeUserAgents: {"user_agent_1": {}},
						user.AllowlistTypeServices:   {"service_1": {}},
						user.AllowlistTypeContracts:  {"contract_1": {}},
						user.AllowlistTypeMethods:    {"method_1": {}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "should return error on invalid allowlists JSON",
			rows: []SelectUserAppsRow{
				{
					ID:                  "app1",
					AccountID:           pgtype.Text{String: "acc1", Valid: true},
					SecretKey:           pgtype.Text{String: "secret1", Valid: true},
					SecretKeyRequired:   pgtype.Bool{Bool: true, Valid: true},
					Allowlists:          json.RawMessage(`invalid`),
					Plan:                pgtype.Text{String: "plan1", Valid: true},
					RateLimitThroughput: pgtype.Int4{Int32: 30, Valid: true},
					RateLimitCapacity:   pgtype.Int4{Int32: 100_000, Valid: true},
				},
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			driver := &PostgresDriver{}
			apps, err := driver.convertToUserApps(test.rows)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, apps)
			}
		})
	}
}

/* -------------------- Dockertest Ephemeral DB Container Setup -------------------- */

const (
	containerName      = "db"
	containerRepo      = "postgres"
	containerTag       = "14"
	dbUser             = "postgres"
	password           = "pgpassword"
	db                 = "postgres"
	connStringFormat   = "postgres://%s:%s@%s/%s?sslmode=disable"
	schemaLocation     = "./sqlc/schema.sql"
	seedTestDBLocation = "./testdata/seed-test-db.sql"
	dockerEntrypoint   = ":/docker-entrypoint-initdb.d/init_%s.sql"
	timeOut            = 1200
)

var (
	containerEnvUser     = fmt.Sprintf("POSTGRES_USER=%s", dbUser)
	containerEnvPassword = fmt.Sprintf("POSTGRES_PASSWORD=%s", password)
	containerEnvDB       = fmt.Sprintf("POSTGRES_DB=%s", db)
	schemaDockerPath     = filepath.Join(os.Getenv("PWD"), schemaLocation) + fmt.Sprintf(dockerEntrypoint, "1")
	seedTestDBDockerPath = filepath.Join(os.Getenv("PWD"), seedTestDBLocation) + fmt.Sprintf(dockerEntrypoint, "2")
)

func setupPostgresDocker() (*dockertest.Pool, *dockertest.Resource, string) {
	opts := dockertest.RunOptions{
		Name:       containerName,
		Repository: containerRepo,
		Tag:        containerTag,
		Env:        []string{containerEnvUser, containerEnvPassword, containerEnvDB},
		Mounts:     []string{schemaDockerPath, seedTestDBDockerPath},
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		fmt.Printf("Could not construct pool: %s", err)
		os.Exit(1)
	}
	resource, err := pool.RunWithOptions(&opts, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		fmt.Printf("Could not start resource: %s", err)
		os.Exit(1)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			fmt.Printf("exit signal %d received\n", sig)
			if err := pool.Purge(resource); err != nil {
				fmt.Printf("could not purge resource: %s", err)
			}
		}
	}()

	if err := resource.Expire(timeOut); err != nil {
		fmt.Printf("[ERROR] Failed to set expiration on docker container: %v", err)
		os.Exit(1)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseURL := fmt.Sprintf(connStringFormat, dbUser, password, hostAndPort, db)

	poolRetryChan := make(chan struct{}, 1)
	retryConnectFn := func() error {
		conn, err := pgx.Connect(context.Background(), databaseURL)
		if err != nil {
			return fmt.Errorf("unable to connect to database: %v", err)
		}
		dbConn = conn
		poolRetryChan <- struct{}{}
		return nil
	}
	if err = pool.Retry(retryConnectFn); err != nil {
		fmt.Printf("could not connect to docker: %s", err)
		os.Exit(1)
	}

	<-poolRetryChan

	return pool, resource, databaseURL
}

func cleanupPostgresDocker(_ *testing.M, pool *dockertest.Pool, resource *dockertest.Resource) {
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %s", err)
	}
}
