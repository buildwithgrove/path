package cosmos

import (
	"fmt"
	"time"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_IMPROVE: Create an interface that partially or completely captures the
// interface exposed by endpointCheckCometBFTStatus.

/* -------------------- CometBFT Status Check -------------------- */

// TODO_DOCUMENT(@commoddity): Document this loose convention.
// CometBFT ID checks begin with 2 for JSON-RPC requests.
//
// This is an arbitrary ID selected by the engineering team at Grove.
// It is used for compatibility with the JSON-RPC spec.
// It is a loose convention in the QoS package.

// ID for the CometBFT /status check.
const idStatusCheck = 2002

// methodStatus is the CometBFT JSON-RPC method for getting the node status.
// Reference: https://docs.cometbft.com/v1.0/spec/rpc/#status
const methodStatusCheck = jsonrpc.Method("status")

// TODO_IMPROVE(@commoddity): determine an appropriate interval for checking the status and/or make it configurable.
const checkStatusInterval = 30 * time.Second

var (
	errNoStatusObs       = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodStatusCheck)
	errInvalidStatusObs  = fmt.Errorf("endpoint returned an invalid response to a %q request", methodStatusCheck)
	errInvalidChainIDObs = fmt.Errorf("endpoint returned an invalid chain ID in its response to a %q request", methodStatusCheck)
	errCatchingUpObs     = fmt.Errorf("endpoint is catching up to the network in its response to a %q request", methodStatusCheck)
)

// endpointCheckCometBFTStatus is a check that ensures the endpoint's status information is valid.
// It is used to verify the endpoint is on the correct chain and not catching up.
//
// DEV_NOTE: The CometBFT status check returns a number of fields that we do not currently use but may wish to include as part of the status check in the future.
// To see the full list of fields, see the CometBFT docs reference:
//
//	https://docs.cometbft.com/v1.0/spec/rpc/#status
type endpointCheckCometBFTStatus struct {
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
// eg. '{"jsonrpc":"2.0","id":2002,"method":"status"}'
//
// It is called in `request_validator_checks.go` to generate the endpoint checks.
func (e *endpointCheckCometBFTStatus) getRequest() jsonrpc.Request {
	return jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(idStatusCheck),
		Method:  jsonrpc.Method(methodStatusCheck),
	}
}

// GetChainID returns the parsed chain ID value for the endpoint.
func (e *endpointCheckCometBFTStatus) GetChainID() (string, error) {
	if e.chainID == nil {
		return "", errNoStatusObs
	}
	if *e.chainID == "" {
		return "", errInvalidChainIDObs
	}
	return *e.chainID, nil
}

// GetCatchingUp returns whether the endpoint is catching up.
func (e *endpointCheckCometBFTStatus) GetCatchingUp() (bool, error) {
	if e.catchingUp == nil {
		return false, errNoStatusObs
	}
	return *e.catchingUp, nil
}

// GetLatestBlockHeight returns the parsed latest block height value for the endpoint.
func (e *endpointCheckCometBFTStatus) GetLatestBlockHeight() (uint64, error) {
	if e.latestBlockHeight == nil {
		return 0, errNoStatusObs
	}
	if *e.latestBlockHeight == 0 {
		return 0, errInvalidStatusObs
	}
	return *e.latestBlockHeight, nil
}

// IsExpired returns true if the check has expired and needs to be refreshed.
func (e *endpointCheckCometBFTStatus) IsExpired() bool {
	return time.Now().After(e.expiresAt)
}
