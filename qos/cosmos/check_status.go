package cosmos

import (
	"fmt"
	"time"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

/* -------------------- CometBFT Status Check -------------------- */

const idStatusCheck = 1003

// methodStatus is the CometBFT JSON-RPC method for getting the node status.
// Reference: https://docs.cometbft.com/v1.0/spec/rpc/#status
const methodStatusCheck = jsonrpc.Method("status")

// TODO_IMPROVE(@commoddity): determine an appropriate interval for checking the status and/or make it configurable.
const checkStatusInterval = 10 * time.Second

var (
	errNoStatusObs       = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodStatusCheck)
	errInvalidStatusObs  = fmt.Errorf("endpoint returned an invalid response to a %q request", methodStatusCheck)
	errInvalidChainIDObs = fmt.Errorf("endpoint returned an invalid chain ID in its response to a %q request", methodStatusCheck)
	errCatchingUpObs     = fmt.Errorf("endpoint is catching up to the network in its response to a %q request", methodStatusCheck)
)

// endpointCheckStatus is a check that ensures the endpoint's status information is valid.
// It is used to verify the endpoint is on the correct chain and not catching up.
type endpointCheckStatus struct {
	// chainID stores the chain ID from the endpoint's response to a `status` request.
	// It is nil if there has NOT been an observation of the endpoint's response to a `status` request.
	chainID *string

	// catchingUp stores whether the endpoint is catching up from the endpoint's response to a `status` request.
	// It is nil if there has NOT been an observation of the endpoint's response to a `status` request.
	catchingUp *bool

	// latestBlockHeight stores the latest block height from the endpoint's response to a `status` request.
	// It is nil if there has NOT been an observation of the endpoint's response to a `status` request.
	latestBlockHeight *uint64

	// expiresAt stores the time at which the last check expires.
	expiresAt time.Time
}

// getRequest returns a JSONRPC request to check the status.
// eg. '{"jsonrpc":"2.0","id":1003,"method":"status"}'
func (e *endpointCheckStatus) getRequest() jsonrpc.Request {
	return jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(idStatusCheck),
		Method:  jsonrpc.Method(methodStatusCheck),
	}
}

// GetChainID returns the parsed chain ID value for the endpoint.
func (e *endpointCheckStatus) GetChainID() (string, error) {
	if e.chainID == nil {
		return "", errNoStatusObs
	}
	if *e.chainID == "" {
		return "", errInvalidChainIDObs
	}
	return *e.chainID, nil
}

// GetCatchingUp returns whether the endpoint is catching up.
func (e *endpointCheckStatus) GetCatchingUp() (bool, error) {
	if e.catchingUp == nil {
		return false, errNoStatusObs
	}
	return *e.catchingUp, nil
}

// GetLatestBlockHeight returns the parsed latest block height value for the endpoint.
func (e *endpointCheckStatus) GetLatestBlockHeight() (uint64, error) {
	if e.latestBlockHeight == nil {
		return 0, errNoStatusObs
	}
	if *e.latestBlockHeight == 0 {
		return 0, errInvalidStatusObs
	}
	return *e.latestBlockHeight, nil
}

// IsExpired returns true if the check has expired and needs to be refreshed.
func (e *endpointCheckStatus) IsExpired() bool {
	return time.Now().After(e.expiresAt)
}
