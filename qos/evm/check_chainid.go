package evm

import (
	"fmt"
	"time"
)

const (
	endpointCheckNameChainID endpointCheckName = "chain_id"
	chainIDCheckInterval                       = 60 * time.Minute
)

var (
	errNoChainIDObs      = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodChainID)
	errInvalidChainIDObs = fmt.Errorf("endpoint returned an invalid response to a %q request", methodChainID)
)

type endpointCheckChainID struct {
	// chainIDResponse stores the result of processing the endpoint's response to an `eth_chainId` request.
	// It is nil if there has NOT been an observation of the endpoint's response to an `eth_chainId` request.
	chainID   *string
	expiresAt time.Time
}

func (e *endpointCheckChainID) CheckName() string {
	return string(endpointCheckNameChainID)
}

func (e *endpointCheckChainID) IsValid(serviceState *ServiceState) error {
	if e.chainID == nil {
		return errNoChainIDObs
	}
	if serviceState.chainID != *e.chainID {
		return errInvalidChainIDObs
	}
	return nil
}

func (e *endpointCheckChainID) ExpiresAt() time.Time {
	return e.expiresAt
}
