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
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

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
		_, err := pgx.Connect(context.Background(), databaseURL)
		if err != nil {
			return fmt.Errorf("unable to connect to database: %v", err)
		}
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
