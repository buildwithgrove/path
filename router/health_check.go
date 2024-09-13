package router

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/buildwithgrove/path/relayer/morse"
	"github.com/buildwithgrove/path/relayer/shannon"
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// HealthCheck is an interface that must be implemented
// by components that need to report their health status
type HealthCheck interface {
	// Name returns the name of the component being checked.
	Name() string
	// IsAlive returns true if the component is healthy, otherwise false.
	IsAlive() bool
}

// All components that implement the HealthCheck interface must be registered here.
var _ HealthCheck = &morse.Protocol{}
var _ HealthCheck = &shannon.Protocol{}

// One of `ready` or `not_ready`.
type healthCheckStatus string

const (
	// StatusReady indicates that all PATH components are ready
	statusReady healthCheckStatus = "ready"
	// StatusNotReady indicates that one or more PATH components
	// are still initializing, ie. warming up caches, etc.
	statusNotReady healthCheckStatus = "not_ready"
)

type (
	// The healthChecker struct is used to check the health of all PATH
	// components that require startup time to warm up their caches, etc.
	healthChecker struct {
		components []HealthCheck
		logger     polylog.Logger
	}
	// healthCheckJSON is the JSON structure of the response body
	// returned by the `/healthz` endpoint along with the status code.
	healthCheckJSON struct {
		// Status is either "ready" or "not_ready". "not_ready" indicates
		// that the service is still warming up its caches, etc.
		Status healthCheckStatus `json:"status"`
		// ImageTag is the semver tag of the PATH Docker image, eg. `v0.0.1`
		// Will default to `development` if not set in the image.
		ImageTag string `json:"imageTag"`
		// ReadyStates is a map of component names to their ready status
		ReadyStates map[string]bool `json:"readyStates,omitempty"`
	}
)

// healthCheckHandler returns the health status of PATH as a JSON response.
//
// It will return a 200 OK status code if all components are ready or
// a 503 Service Unavailable status code if any component is not ready.
//
// The image tag is set to the value of the IMAGE_TAG environment variable, which is
// passed to the Docker image as a build argument at build time.
func (c *healthChecker) healthzHandler(w http.ResponseWriter, req *http.Request) {
	readyStates := c.getComponentReadyStates()
	status := getStatus(readyStates)

	responseBytes := c.getHealthCheckResponse(status, readyStates)
	if responseBytes == nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if status == statusReady {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if _, err := w.Write(responseBytes); err != nil {
		c.logger.Error().Msgf("error writing health check response: %s", err.Error())
	}
}

// getHealthCheckResponse returns the health check JSON response body as bytes
func (c *healthChecker) getHealthCheckResponse(status healthCheckStatus, readyStates map[string]bool) []byte {
	imageTag := os.Getenv(imageTagEnvVar)
	if imageTag == "" {
		imageTag = defaultImageTag
	}

	responseBytes, err := json.Marshal(healthCheckJSON{
		Status:      status,
		ReadyStates: readyStates,
		ImageTag:    imageTag,
	})
	if err != nil {
		c.logger.Error().Msgf("error marshalling health check response: %s", err.Error())
		return nil
	}

	return responseBytes
}

// getComponentReadyStates returns a map of component names to their ready status
func (c *healthChecker) getComponentReadyStates() map[string]bool {
	readyStates := make(map[string]bool)
	for _, component := range c.components {
		readyStates[component.Name()] = component.IsAlive()
	}
	return readyStates
}

// getStatus returns false if any component is not ready, otherwise true
func getStatus(readyStates map[string]bool) healthCheckStatus {
	for _, ready := range readyStates {
		if !ready {
			return statusNotReady
		}
	}
	return statusReady
}
