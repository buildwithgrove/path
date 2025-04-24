//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

/* -------------------- Dockertest Ephemeral PATH Container Setup -------------------- */

// These variables reference the testOptions struct defined in evm_test.go

const (
	imageName            = "path-image"
	containerName        = "path"
	internalPathPort     = "3069"
	buildContextDir      = ".."
	dockerfileName       = "Dockerfile"
	configMountPoint     = ":/app/config/.config.yaml"
	containerEnvImageTag = "IMAGE_TAG=test"
	containerExtraHost   = "host.docker.internal:host-gateway" // allows the container to access the host machine's Docker daemon
	// containerExpirySeconds is the number of seconds after which the started PATH container should be removed by the dockertest library.
	containerExpirySeconds = 240
	// maxPathHealthCheckWaitTimeMillisec is the maximum amount of time a started PATH container has to report its status as healthy.
	// Once this time expires, the associated E2E test is marked as failed and the PATH container is removed.
	maxPathHealthCheckWaitTimeMillisec = 120000
)

// eg. 3069/tcp
var containerPortAndProtocol = internalPathPort + "/tcp"

// setupPathInstance starts an instance of PATH in a Docker container.
//
// Returns:
// - "pathPort": the dynamically selected and exposed port by the ephemeral PATH container
// - "cleanup": a function that must be called to clean up the PATH container
//   - Test functions are responsible for calling this cleanup function
func setupPathInstance(
	t *testing.T,
	configFilePath string,
	dockerOpts dockerOptions,
) (containerPort string, cleanupFn func()) {
	t.Helper()

	// Initialize the ephemeral PATH Docker container
	pool, resource, containerPort, logOutputFile := setupPathDocker(t, configFilePath, dockerOpts)

	cleanupFn = func() {
		// Cleanup the ephemeral PATH Docker container
		cleanupPathDocker(t, pool, resource)
		if logOutputFile != "" {
			fmt.Printf("\n%s===== 👀 LOGS 👀 =====%s\n", BOLD_CYAN, RESET)
			fmt.Printf("\n ✍️ PATH container output logged to %s ✍️ \n\n", logOutputFile)
			fmt.Printf("%s===== 👀 LOGS 👀 =====%s\n\n", BOLD_CYAN, RESET)
		}
	}

	return containerPort, cleanupFn
}

// setupPathDocker sets up and starts a Docker container for the PATH service using dockertest.
//
// Key steps:
// - Builds the container from a specified Dockerfile.
// - Mounts necessary configuration files.
// - Sets environment variables for the container.
// - Exposes required ports and sets extra hosts.
// - Sets up a signal handler to clean up the container on termination signals.
// - Performs a health check to ensure the container is ready for requests.
// - Returns the dockertest pool, resource, and the container port.
func setupPathDocker(
	t *testing.T,
	configFilePath string,
	dockerOpts dockerOptions,
) (*dockertest.Pool, *dockertest.Resource, string, string) {
	t.Helper()

	// Get docker options from the global test options
	logContainer := dockerOpts.logOutput
	forceRebuild := dockerOpts.forceRebuild

	// eg. {file_path}/path/e2e/.shannon.config.yaml
	configFilePath = filepath.Join(os.Getenv("PWD"), configFilePath)

	// Check if config file exists and exit if it does not
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		t.Fatalf("config file does not exist: %s", configFilePath)
	}

	// eg. {file_path}/path/e2e/.shannon.config.yaml:/app/config/.config.yaml
	containerConfigMount := configFilePath + configMountPoint

	// Initialize the dockertest pool
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not construct pool: %s", err)
	}
	pool.MaxWait = time.Duration(maxPathHealthCheckWaitTimeMillisec) * time.Millisecond

	// Check if the image already exists and we're not forcing a rebuild
	imageExists := false
	if !forceRebuild {
		if _, err := pool.Client.InspectImage(imageName); err == nil {
			imageExists = true
			fmt.Println("\n🐳 Using existing Docker image, skipping build...")
			fmt.Println("  💡 TIP: Set DOCKER_FORCE_REBUILD=true to rebuild the image if needed 💡")
		}
	} else {
		fmt.Println("🔄 Force rebuild requested, will build Docker image...")
	}

	// Only build the image if it doesn't exist or force rebuild is set
	if !imageExists || forceRebuild {
		fmt.Println("🏗️  Building Docker image...")

		// Build the image and log build output
		buildOptions := docker.BuildImageOptions{
			Name:           imageName,
			ContextDir:     buildContextDir,
			Dockerfile:     dockerfileName,
			OutputStream:   os.Stdout,
			SuppressOutput: false,
			NoCache:        forceRebuild, // If force rebuilding, also disable cache
		}
		if err := pool.Client.BuildImage(buildOptions); err != nil {
			t.Fatalf("could not build path image: %s", err)
		}
		fmt.Println("🐳 Docker image built successfully!")
	}

	fmt.Println("\n🌿 Starting PATH test container...")

	// Run the built image
	runOpts := &dockertest.RunOptions{
		Name:         containerName,
		Repository:   imageName,
		Mounts:       []string{containerConfigMount},
		Env:          []string{containerEnvImageTag},
		ExposedPorts: []string{containerPortAndProtocol},
		ExtraHosts:   []string{containerExtraHost},
	}
	resource, err := pool.RunWithOptions(runOpts, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}

	// Optionally log the PATH container output
	var logOutputFile string
	if logContainer {
		// Determine log output file
		logOutputFile = os.Getenv("DOCKER_LOG_OUTPUT_FILE")
		if logOutputFile == "" {
			logOutputFile = fmt.Sprintf("/tmp/path_log_e2e_test_%d.txt", time.Now().Unix())
		}
		fmt.Printf("\n ✍️ PATH container output will be logged to %s ✍️ \n\n", logOutputFile)

		// Print container logs in a goroutine to prevent blocking
		go func() {
			f, err := os.Create(logOutputFile)
			if err != nil {
				fmt.Printf("could not create log file %s: %v\n", logOutputFile, err)
				return
			}
			defer f.Close()

			if err := pool.Client.Logs(docker.LogsOptions{
				Container:    resource.Container.ID,
				OutputStream: f,
				ErrorStream:  f,
				Stdout:       true,
				Stderr:       true,
				Follow:       true,
			}); err != nil {
				fmt.Printf("could not fetch logs for PATH container: %s", err)
			}
		}()
	}

	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Set up signal handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Create a channel to wait for cleanup completion
	cleanupDone := make(chan struct{})

	// Handle signals
	go func() {
		select {
		case <-signalChan:
			fmt.Println("\nReceived Ctrl+C, cleaning up containers...")
			// Cancel the context
			cancel()

			// Perform cleanup
			if err := pool.Purge(resource); err != nil {
				log.Printf("Could not purge resource: %s", err)
			}

			// Signal that cleanup is done
			close(cleanupDone)
		case <-ctx.Done():
			// Context was canceled elsewhere
		}
	}()

	if err := resource.Expire(containerExpirySeconds); err != nil {
		t.Fatalf("[ERROR] Failed to set expiration on docker container: %v", err)
	}

	t.Logf("  ✅ PATH test container started successfully!")

	// performs a health check on the PATH container to ensure it is ready for requests
	healthCheckURL := fmt.Sprintf("http://%s/healthz", resource.GetHostPort(containerPortAndProtocol))

	t.Logf("🏥 Performing health check on PATH test container at %s...", healthCheckURL)

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
		t.Fatalf("could not connect to docker: %s", err)
	}

	t.Logf("  ✅ PATH test container is healthy and ready for tests!")

	<-poolRetryChan

	return pool, resource, resource.GetPort(containerPortAndProtocol), logOutputFile
}

// cleanupPathDocker purges the Docker container and resource from the provided dockertest pool and resource.
func cleanupPathDocker(t *testing.T, pool *dockertest.Pool, resource *dockertest.Resource) {
	t.Helper()

	if err := pool.Purge(resource); err != nil {
		t.Fatalf("could not purge resource: %s", err)
	}
}
