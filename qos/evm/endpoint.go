package evm

import (
	"fmt"
	"strconv"
	"time"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
)

// The errors below list all the possible validation errors on an endpoint.
var (
	errNoChainIDObs          = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodChainID)
	errInvalidChainIDObs     = fmt.Errorf("endpoint returned an invalid response to a %q request", methodChainID)
	errNoBlockNumberObs      = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodBlockNumber)
	errInvalidBlockNumberObs = fmt.Errorf("endpoint returned an invalid response to a %q request", methodBlockNumber)
)

type endpointCheckName string

type endpointCheck interface {
	IsValid(serviceState *ServiceState) error
	ExpiresAt() time.Time
}

const (
	endpointCheckNameChainID endpointCheckName = "chain_id"
	chainIDCheckInterval                       = 60 * time.Minute
)

type endpointCheckChainID struct {
	// chainIDResponse stores the result of processing the endpoint's response to an `eth_chainId` request.
	// It is nil if there has NOT been an observation of the endpoint's response to an `eth_chainId` request.
	chainID   *string
	expiresAt time.Time
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

const (
	endpointCheckNameBlockHeight endpointCheckName = "block_height"
	blockHeightCheckInterval                       = 60 * time.Second
)

type endpointCheckBlockHeight struct {
	// parsedBlockNumberResponse stores the result of processing the endpoint's response to an `eth_blockNumber` request.
	// It is nil if there has NOT been an observation of the endpoint's response to an `eth_blockNumber` request.
	blockHeight *uint64
	expiresAt   time.Time
}

func (e *endpointCheckBlockHeight) IsValid(serviceState *ServiceState) error {
	if e.blockHeight == nil {
		return errNoBlockNumberObs
	}
	if serviceState.perceivedBlockNumber <= *e.blockHeight {
		return errInvalidBlockNumberObs
	}
	return nil
}

func (e *endpointCheckBlockHeight) ExpiresAt() time.Time {
	return e.expiresAt
}

// endpoint captures the details required to validate an EVM endpoint.
type endpoint struct {
	checks map[endpointCheckName]endpointCheck
}

func newEndpoint() endpoint {
	return endpoint{
		checks: make(map[endpointCheckName]endpointCheck),
	}
}

// Validate returns an error if the endpoint is invalid.
// e.g. an endpoint without an observation of its response to an `eth_chainId` request is not considered valid.
func (e endpoint) Validate(serviceState *ServiceState) error {
	for _, check := range e.checks {
		if err := check.IsValid(serviceState); err != nil {
			return err
		}
	}
	return nil
}

// ApplyObservation updates the data stored regarding the endpoint using the supplied observation.
// It Returns true if the observation was not unrecognized, i.e. mutated the endpoint.
// TODO_TECHDEBT(@adshmh): add a method to distinguish the following two scenarios:
//   - an endpoint that returned in invalid response.
//   - an endpoint with no/incomplete observations.
func (e *endpoint) ApplyObservation(obs *qosobservations.EVMEndpointObservation) bool {
	if chainIDResponse := obs.GetChainIdResponse(); chainIDResponse != nil {
		observedChainID := chainIDResponse.GetChainIdResponse()
		e.checks[endpointCheckNameChainID] = &endpointCheckChainID{
			chainID:   &observedChainID,
			expiresAt: time.Now().Add(chainIDCheckInterval),
		}
		return true
	}

	if blockNumberResponse := obs.GetBlockNumberResponse(); blockNumberResponse != nil {
		// base 0: use the string's prefix to determine its base.
		parsedBlockNumber, err := strconv.ParseUint(blockNumberResponse.GetBlockNumberResponse(), 0, 64)
		// The endpoint returned an invalid response to an `eth_blockNumber` request.
		// Explicitly set the parsedBlockNumberResponse to a zero value as the ParseUInt does not guarantee returning a 0 on all error cases.
		if err != nil {
			zero := uint64(0)
			e.checks[endpointCheckNameBlockHeight] = &endpointCheckBlockHeight{
				blockHeight: &zero,
				expiresAt:   time.Now().Add(blockHeightCheckInterval),
			}
			return true
		}

		e.checks[endpointCheckNameBlockHeight] = &endpointCheckBlockHeight{
			blockHeight: &parsedBlockNumber,
			expiresAt:   time.Now().Add(blockHeightCheckInterval),
		}
		return true
	}

	return false
}

// GetBlockNumber returns the parsed block number value for the endpoint.
func (e endpoint) GetBlockNumber() (uint64, error) {
	observedBlockHeight, ok := e.checks[endpointCheckNameBlockHeight].(*endpointCheckBlockHeight)
	if !ok {
		return 0, errNoBlockNumberObs
	}
	if observedBlockHeight.blockHeight == nil {
		return 0, errNoBlockNumberObs
	}

	return *observedBlockHeight.blockHeight, nil
}
