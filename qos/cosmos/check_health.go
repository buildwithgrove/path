package cosmos

import (
	"fmt"
	"net/http"
	"time"
)

// Get node health. Returns empty result (200 OK) on success, no response - in case of an error.
// Reference: https://docs.cometbft.com/v1.0/spec/rpc/#health
const apiPathHealthCheck = "/health"

// TODO_IMPROVE(@commoddity): determine an appropriate interval for checking the health.
const checkHealthInterval = 30 * time.Second

var (
	errNoHealthObs      = fmt.Errorf("endpoint has not had an observation of its response to a %q request", apiPathHealthCheck)
	errInvalidHealthObs = fmt.Errorf("endpoint returned an invalid response to a %q request", apiPathHealthCheck)
)

// endpointCheckHealth is a check that ensures the endpoint's health status is valid.
// It is used to verify the endpoint is healthy and responding to requests.
//
// Note that this check has an expiry as health checks should be performed periodically
// to ensure the endpoint remains responsive.
type endpointCheckHealth struct {
	// healthy stores the health status from the endpoint's response to a `/health` request.
	// It is nil if there has NOT been an observation of the endpoint's response to a `/health` request.
	healthy *bool

	// expiresAt stores the time at which the last check expires.
	expiresAt time.Time
}

// GetRequest returns an HTTP request to check the health.
// e.g. GET /health
func (e *endpointCheckHealth) GetRequest() *http.Request {
	req, _ := http.NewRequest(http.MethodGet, apiPathHealthCheck, nil)
	return req
}

// GetHealthy returns the parsed health status for the endpoint.
func (e *endpointCheckHealth) GetHealthy() (bool, error) {
	if e.healthy == nil {
		return false, errNoHealthObs
	}
	return *e.healthy, nil
}

// IsExpired returns true if the check has expired and needs to be refreshed.
func (e *endpointCheckHealth) IsExpired() bool {
	return time.Now().After(e.expiresAt)
}
