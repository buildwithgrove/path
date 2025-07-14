package cosmos

import (
	"fmt"
	"net/http"
	"time"
)

// Get CometBFT status including node info, pubkey, latest block hash, app hash, block height and time.
// Reference: https://docs.cometbft.com/v1.0/spec/rpc/#status
const apiPathStatus = "/status"

// TODO_IMPROVE(@commoddity): determine an appropriate interval for checking the status.
const checkStatusInterval = 10 * time.Second

var (
	errNoStatusObs       = fmt.Errorf("endpoint has not had an observation of its response to a %q request", apiPathStatus)
	errInvalidStatusObs  = fmt.Errorf("endpoint returned an invalid response to a %q request", apiPathStatus)
	errInvalidChainIDObs = fmt.Errorf("endpoint returned an invalid chain ID in its response to a %q request", apiPathStatus)
	errCatchingUpObs     = fmt.Errorf("endpoint is catching up to the network in its response to a %q request", apiPathStatus)
)

// endpointCheckStatus is a check that ensures the endpoint's status information is valid.
// It is used to verify the endpoint is on the correct chain and not catching up.
type endpointCheckStatus struct {
	// chainID stores the chain ID from the endpoint's response to a `/status` request.
	// It is nil if there has NOT been an observation of the endpoint's response to a `/status` request.
	chainID *string

	// catchingUp stores whether the endpoint is catching up from the endpoint's response to a `/status` request.
	// It is nil if there has NOT been an observation of the endpoint's response to a `/status` request.
	catchingUp *bool

	// latestBlockHeight stores the latest block height from the endpoint's response to a `/status` request.
	// It is nil if there has NOT been an observation of the endpoint's response to a `/status` request.
	latestBlockHeight *uint64

	expiresAt time.Time
}

// GetRequest returns an HTTP request to check the status.
// e.g. GET /status
func (e *endpointCheckStatus) GetRequest() *http.Request {
	req, _ := http.NewRequest(http.MethodGet, apiPathStatus, nil)
	return req
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
