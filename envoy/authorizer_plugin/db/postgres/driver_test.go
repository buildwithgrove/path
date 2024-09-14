package postgres

import (
	"context"
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

	"github.com/buildwithgrove/path-authorizer/user"
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

func Test_Integration_GetGatewayEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		expected map[user.EndpointID]user.GatewayEndpoint
	}{
		{
			name: "should retrieve all gateway endpoints correctly",
			expected: map[user.EndpointID]user.GatewayEndpoint{
				"endpoint_1": {
					EndpointID: "endpoint_1",
					Auth: user.Auth{
						APIKey:         "api_key_1",
						APIKeyRequired: true,
					},
					UserAccount: user.UserAccount{
						AccountID: "account_1",
						PlanType:  "PLAN_FREE",
					},
					RateLimiting: user.RateLimiting{
						ThroughputLimit:     30,
						CapacityLimit:       100000,
						CapacityLimitPeriod: "daily",
					},
				},
				"endpoint_2": {
					EndpointID: "endpoint_2",
					Auth: user.Auth{
						APIKey:         "api_key_2",
						APIKeyRequired: true,
					},
					UserAccount: user.UserAccount{
						AccountID: "account_2",
						PlanType:  "PLAN_UNLIMITED",
					},
					RateLimiting: user.RateLimiting{
						ThroughputLimit: 0,
						CapacityLimit:   0,
					},
				},
				"endpoint_3": {
					EndpointID: "endpoint_3",
					Auth: user.Auth{
						APIKey:         "api_key_3",
						APIKeyRequired: true,
					},
					UserAccount: user.UserAccount{
						AccountID: "account_3",
						PlanType:  "PLAN_FREE",
					},
					RateLimiting: user.RateLimiting{
						ThroughputLimit:     30,
						CapacityLimit:       100000,
						CapacityLimitPeriod: "daily",
					},
				},
				"endpoint_4": {
					EndpointID: "endpoint_4",
					Auth: user.Auth{
						APIKey:         "",
						APIKeyRequired: false,
					},
					UserAccount: user.UserAccount{
						AccountID: "account_1",
						PlanType:  "PLAN_FREE",
					},
					RateLimiting: user.RateLimiting{
						ThroughputLimit:     30,
						CapacityLimit:       100000,
						CapacityLimitPeriod: "daily",
					},
				},
				"endpoint_5": {
					EndpointID: "endpoint_5",
					Auth: user.Auth{
						APIKey:         "",
						APIKeyRequired: false,
					},
					UserAccount: user.UserAccount{
						AccountID: "account_2",
						PlanType:  "PLAN_UNLIMITED",
					},
					RateLimiting: user.RateLimiting{
						ThroughputLimit: 0,
						CapacityLimit:   0,
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

			endpoints, err := driver.GetGatewayEndpoints(context.Background())
			c.NoError(err)
			c.Equal(test.expected, endpoints)
		})
	}
}

func Test_convertToGatewayEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		rows     []SelectGatewayEndpointsRow
		expected map[user.EndpointID]user.GatewayEndpoint
		wantErr  bool
	}{
		{
			name: "should convert rows to gateway endpoints successfully",
			rows: []SelectGatewayEndpointsRow{
				{
					ID:                      "endpoint_1",
					AccountID:               pgtype.Text{String: "account_1", Valid: true},
					ApiKey:                  pgtype.Text{String: "api_key_1", Valid: true},
					ApiKeyRequired:          pgtype.Bool{Bool: true, Valid: true},
					Plan:                    pgtype.Text{String: "PLAN_FREE", Valid: true},
					RateLimitThroughput:     pgtype.Int4{Int32: 30, Valid: true},
					RateLimitCapacity:       pgtype.Int4{Int32: 100000, Valid: true},
					RateLimitCapacityPeriod: NullRateLimitCapacityPeriod{RateLimitCapacityPeriod: "daily", Valid: true},
				},
			},
			expected: map[user.EndpointID]user.GatewayEndpoint{
				"endpoint_1": {
					EndpointID: "endpoint_1",
					Auth: user.Auth{
						APIKey:         "api_key_1",
						APIKeyRequired: true,
					},
					UserAccount: user.UserAccount{
						AccountID: "account_1",
						PlanType:  "PLAN_FREE",
					},
					RateLimiting: user.RateLimiting{
						ThroughputLimit:     30,
						CapacityLimit:       100000,
						CapacityLimitPeriod: "daily",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			driver := &postgresDriver{}
			endpoints, err := driver.convertToGatewayEndpoints(test.rows)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, endpoints)
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
	dbName             = "postgres"
	connStringFormat   = "postgres://%s:%s@%s/%s?sslmode=disable"
	schemaLocation     = "./sqlc/schema.sql"
	seedTestDBLocation = "./testdata/seed-test-db.sql"
	dockerEntrypoint   = ":/docker-entrypoint-initdb.d/init_%s.sql"
	timeOut            = 1200
)

var (
	containerEnvUser     = fmt.Sprintf("POSTGRES_USER=%s", dbUser)
	containerEnvPassword = fmt.Sprintf("POSTGRES_PASSWORD=%s", password)
	containerEnvDB       = fmt.Sprintf("POSTGRES_DB=%s", dbName)
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
	databaseURL := fmt.Sprintf(connStringFormat, dbUser, password, hostAndPort, dbName)

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
