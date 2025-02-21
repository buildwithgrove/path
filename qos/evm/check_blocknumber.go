package evm

import (
	"fmt"
	"time"
)

const (
	endpointCheckNameBlockHeight endpointCheckName = "block_height"
	// TODO_IMPROVE: determine an appropriate interval for checking the block height.
	blockHeightCheckInterval = 60 * time.Second
)

var (
	errNoBlockNumberObs      = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodBlockNumber)
	errInvalidBlockNumberObs = fmt.Errorf("endpoint returned an invalid response to a %q request", methodBlockNumber)
)

// endpointCheckBlockNumber is a check that ensures the endpoint's block height is greater than the perceived block height.
// It is used to ensure that the endpoint is not behind the chain.
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
	// If the endpoint's block height is less than the perceived block height minus the sync allowance,
	// then the endpoint is behind the chain and should be filtered out.
	minAllowedBlockHeight := serviceState.perceivedBlockNumber - serviceState.config.syncAllowance

	if *e.blockHeight < minAllowedBlockHeight {
		return errInvalidBlockNumberObs
	}

	return nil
}

func (e *endpointCheckBlockNumber) ExpiresAt() time.Time {
	return e.expiresAt
}
