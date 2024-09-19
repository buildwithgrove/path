//go:build e2e

package e2e

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

/* -------------------- Dockertest Ephemeral PATH Container Setup -------------------- */

const (
	containerName        = "path"
	internalPathPort     = "3000"
	dockerfilePath       = "../Dockerfile"
	configFilePath       = "./.config.test.yaml"
	configMountPoint     = ":/app/.config.yaml"
	containerEnvImageTag = "IMAGE_TAG=test"
	containerExtraHost   = "host.docker.internal:host-gateway" // allows the container to access the host machine's Docker daemon
	timeoutSeconds       = 120
)

var (
	// eg. {file_path}/path/e2e/.config.test.yaml:/app/.config.yaml
	containerConfigMount = filepath.Join(os.Getenv("PWD"), configFilePath) + configMountPoint
	// eg. 3000/tcp
	containerPortAndProtocol = internalPathPort + "/tcp"
)

// setupPathDocker sets up and starts a Docker container for the PATH service using dockertest.
//
// Key steps:
//
// - Builds the container from a specified Dockerfile.
//
// - Mounts necessary configuration files.
//
// - Sets environment variables for the container.
//
// - Exposes required ports and sets extra hosts.
//
// - Sets up a signal handler to clean up the container on termination signals.
//
// - Performs a health check to ensure the container is ready for requests.
//
// - Returns the dockertest pool, resource, and the container port.
func setupPathDocker() (*dockertest.Pool, *dockertest.Resource, string) {
	opts := &dockertest.RunOptions{
		Name:         containerName,
		Mounts:       []string{containerConfigMount},
		Env:          []string{containerEnvImageTag},
		ExposedPorts: []string{containerPortAndProtocol},
		ExtraHosts:   []string{containerExtraHost},
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		fmt.Printf("Could not construct pool: %s", err)
		os.Exit(1)
	}
	resource, err := pool.BuildAndRunWithOptions(dockerfilePath, opts, func(config *docker.HostConfig) {
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

	if err := resource.Expire(timeoutSeconds); err != nil {
		fmt.Printf("[ERROR] Failed to set expiration on docker container: %v", err)
		os.Exit(1)
	}

	// performs a health check on the PATH container to ensure it is ready for requests
	healthCheckURL := fmt.Sprintf("http://%s/healthz", resource.GetHostPort(containerPortAndProtocol))

	poolRetryChan := make(chan struct{}, 1)
	retryConnectFn := func() error {
		resp, err := http.Get(healthCheckURL)
		if err != nil {
			return fmt.Errorf("unable to connect to health check endpoint: %v", err)
		}
		defer resp.Body.Close()

		// the health check endpoint returns a 200 OK status if the service is ready
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("health check endpoint returned non-200 status: %d", resp.StatusCode)
		}

		// notify the pool that the health check was successful
		poolRetryChan <- struct{}{}
		return nil
	}
	if err = pool.Retry(retryConnectFn); err != nil {
		fmt.Printf("could not connect to docker: %s", err)
		os.Exit(1)
	}

	<-poolRetryChan

	return pool, resource, resource.GetPort(containerPortAndProtocol)
}

// cleanupPathDocker purges the Docker container and resource from the provided dockertest pool and resource.
func cleanupPathDocker(_ *testing.M, pool *dockertest.Pool, resource *dockertest.Resource) {
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %s", err)
	}
}
