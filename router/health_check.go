package router

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/pokt-network/poktroll/pkg/polylog"
)

type healthCheckStatus string

const (
	statusReady    healthCheckStatus = "ok"
	statusNotReady healthCheckStatus = "initializing"
)

type (
	HealthCheckComponent interface {
		Name() string
		IsReady() bool
	}
	healthCheckJSON struct {
		Status      healthCheckStatus `json:"status"`
		ImageTag    string            `json:"imageTag"`
		ReadyStates map[string]bool   `json:"readyStates,omitempty"`
	}
)
type healthCheck struct {
	components []HealthCheckComponent
	logger     polylog.Logger
}

func (c *healthCheck) healthCheckHandler(w http.ResponseWriter, req *http.Request) {
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

func (c *healthCheck) getHealthCheckResponse(status healthCheckStatus, readyStates map[string]bool) []byte {
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

func (c *healthCheck) getComponentReadyStates() map[string]bool {
	readyStates := make(map[string]bool)
	for _, component := range c.components {
		readyStates[component.Name()] = component.IsReady()
	}
	return readyStates
}

func getStatus(readyStates map[string]bool) healthCheckStatus {
	for _, ready := range readyStates {
		if !ready {
			return statusNotReady
		}
	}
	return statusReady
}