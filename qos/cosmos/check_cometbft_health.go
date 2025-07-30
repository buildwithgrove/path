package cosmos

import (
	"fmt"
	"time"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

/* -------------------- CometBFT Health Check -------------------- */

// ID for the CometBFT health check.
// This number may be any arbitrary ID and is selected
// to maintain a convention in the QoS packages of
// consistent ID for a given check type.
//
// CometBFT checks begin with 2.
const idCometBFTHealthCheck = 2001

// methodHealth is the CometBFT JSON-RPC method for getting the node health.
// Reference: https://docs.cometbft.com/v1.0/spec/rpc/#health
const methodCometBFTHealth = jsonrpc.Method("health")

// TODO_IMPROVE(@commoddity): determine an appropriate interval for checking the health.
const checkHealthInterval = 30 * time.Second

var (
	errNoCometBFTHealthObs      = fmt.Errorf("endpoint has not had an observation of its response to a CometBFT '%q' request", methodCometBFTHealth)
	errInvalidCometBFTHealthObs = fmt.Errorf("endpoint returned an invalid response to a CometBFT '%q' request", methodCometBFTHealth)
)

// endpointCheckHealth is a check that ensures the endpoint is healthy.
// It is used to verify the endpoint is healthy and responding to requests.
//
// Note that this check has an expiry as health checks should be performed periodically
// to ensure the endpoint remains responsive.
type endpointCheckCometBFTHealth struct {
	// healthy stores the health status from the endpoint's response to a `health` request.
	// It is nil if there has NOT been an observation of the endpoint's response to a `health` request.
	healthy *bool

	// expiresAt stores the time at which the last check expires.
	expiresAt time.Time
}

// getRequest returns a JSONRPC request to check if the endpoint is healthy.
// eg. '{"jsonrpc":"2.0","id":1002,"method":"health"}'
//
// It is called in `request_validator_checks.go` to generate the endpoint checks.
func (e *endpointCheckCometBFTHealth) getRequest() jsonrpc.Request {
	return jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(idCometBFTHealthCheck),
		Method:  jsonrpc.Method(methodCometBFTHealth),
	}
}

// GetHealthy returns the parsed health status for the endpoint.
func (e *endpointCheckCometBFTHealth) GetHealthy() (bool, error) {
	if e.healthy == nil {
		return false, errNoCometBFTHealthObs
	}
	return *e.healthy, nil
}

// IsExpired returns true if the check has expired and needs to be refreshed.
func (e *endpointCheckCometBFTHealth) IsExpired() bool {
	return time.Now().After(e.expiresAt)
}
