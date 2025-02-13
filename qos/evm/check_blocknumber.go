package evm

import (
	"fmt"
	"time"
)

const (
	endpointCheckNameBlockHeight endpointCheckName = "block_height"
	blockHeightCheckInterval                       = 60 * time.Second
)

var (
	errNoBlockNumberObs      = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodBlockNumber)
	errInvalidBlockNumberObs = fmt.Errorf("endpoint returned an invalid response to a %q request", methodBlockNumber)
)

type endpointCheckBlockNumber struct {
	// parsedBlockNumberResponse stores the result of processing the endpoint's response to an `eth_blockNumber` request.
	// It is nil if there has NOT been an observation of the endpoint's response to an `eth_blockNumber` request.
	blockHeight *uint64
	expiresAt   time.Time
}

func (e *endpointCheckBlockNumber) CheckName() string {
	return string(endpointCheckNameBlockHeight)
}

func (e *endpointCheckBlockNumber) IsValid(serviceState *ServiceState) error {
	if e.blockHeight == nil {
		return errNoBlockNumberObs
	}
	if serviceState.perceivedBlockNumber <= *e.blockHeight {
		return errInvalidBlockNumberObs
	}
	return nil
}

func (e *endpointCheckBlockNumber) ExpiresAt() time.Time {
	return e.expiresAt
}
