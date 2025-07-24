package cosmos

import (
	"fmt"
	"time"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

/* -------------------- CometBFT Health Check -------------------- */

const idHealthCheck = 1002

// methodHealth is the CometBFT JSON-RPC method for getting the node health.
// Reference: https://docs.cometbft.com/v1.0/spec/rpc/#health
const methodHealth = jsonrpc.Method("health")

// TODO_IMPROVE(@commoddity): determine an appropriate interval for checking the health.
const checkHealthInterval = 30 * time.Second

var (
	errNoHealthObs      = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodHealth)
	errInvalidHealthObs = fmt.Errorf("endpoint returned an invalid response to a %q request", methodHealth)
)

// endpointCheckHealth is a check that ensures the endpoint's health status is valid.
// It is used to verify the endpoint is healthy and responding to requests.
//
// Note that this check has an expiry as health checks should be performed periodically
// to ensure the endpoint remains responsive.
type endpointCheckCometBFTHealth struct {
	// healthy stores the health status from the endpoint's response to a `status` request.
	// It is nil if there has NOT been an observation of the endpoint's response to a `status` request.
	healthy *bool

	// expiresAt stores the time at which the last check expires.
	expiresAt time.Time
}

// getRequest returns a JSONRPC request to check the health/status.
// eg. '{"jsonrpc":"2.0","id":1002,"method":"health"}'
func (e *endpointCheckCometBFTHealth) getRequest() jsonrpc.Request {
	return jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(idHealthCheck),
		Method:  jsonrpc.Method(methodHealth),
	}
}

// GetHealthy returns the parsed health status for the endpoint.
func (e *endpointCheckCometBFTHealth) GetHealthy() (bool, error) {
	if e.healthy == nil {
		return false, errNoHealthObs
	}
	return *e.healthy, nil
}

// IsExpired returns true if the check has expired and needs to be refreshed.
func (e *endpointCheckCometBFTHealth) IsExpired() bool {
	return time.Now().After(e.expiresAt)
}
