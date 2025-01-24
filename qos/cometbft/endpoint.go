package cometbft

import (
	"fmt"
	"strconv"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
)

var (
	// The errors below list all the possible validation errors on an endpoint.
	errNoHealthObs           = fmt.Errorf("endpoint has not had an observation of its response to a health check request")
	errInvalidHealthObs      = fmt.Errorf("endpoint returned an invalid response to a health check request")
	errNoBlockNumberObs      = fmt.Errorf("endpoint has not had an observation of its response to a block height request")
	errInvalidBlockNumberObs = fmt.Errorf("endpoint returned an invalid response to a block height request")
)

// endpoint captures the details required to validate an CometBFT endpoint.
type endpoint struct {
	// healthResponse stores the result of processing the endpoint's response to a `/health` request.
	// It is nil if there has NOT been an observation of the endpoint's response to a `/health` request.
	healthResponse *bool

	// parsedBlockNumberResponse stores the result of processing the endpoint's response to an block height request.
	// It is nil if there has NOT been an observation of the endpoint's response to an block height request.
	parsedBlockNumberResponse *uint64
}

// Validate returns an error if the endpoint is invalid.
// e.g. an endpoint without an observation of its response to an `eth_chainId` request is not considered valid.
func (e endpoint) Validate(_ string) error {
	switch {
	case e.healthResponse == nil:
		return errNoHealthObs
	case !*e.healthResponse:
		return fmt.Errorf("invalid response: endpoint is not healthy: %w", errInvalidHealthObs)
	case e.parsedBlockNumberResponse == nil:
		return errNoBlockNumberObs
	case *e.parsedBlockNumberResponse == 0:
		return errInvalidBlockNumberObs
	default:
		return nil
	}
}

// ApplyObservation updates the data stored regarding the endpoint using the supplied observation.
// It returns true if the observation was not unrecognized, i.e. mutated the endpoint.
func (e *endpoint) ApplyObservation(obs *qosobservations.CometBFTEndpointObservation) bool {
	if healthResponse := obs.GetHealthResponse(); healthResponse != nil {
		observedHealth := healthResponse.GetHealthStatusResponse()
		e.healthResponse = &observedHealth
		return true
	}

	if blockNumberResponse := obs.GetLatestBlockHeightResponse(); blockNumberResponse != nil {
		// base 0: use the string's prefix to determine its base.
		parsedBlockNumber, err := strconv.ParseUint(blockNumberResponse.GetLatestBlockHeightResponse(), 0, 64)
		// The endpoint returned an invalid response to a block height request.
		// Explicitly set the parsedBlockNumberResponse to a zero value as the ParseUInt does not guarantee returning a 0 on all error cases.
		if err != nil {
			zero := uint64(0)
			e.parsedBlockNumberResponse = &zero
			return true
		}

		e.parsedBlockNumberResponse = &parsedBlockNumber
		return true
	}

	return false
}

// GetBlockNumber returns the parsed block number value for the endpoint.
func (e endpoint) GetBlockNumber() (uint64, error) {
	if e.parsedBlockNumberResponse == nil {
		return 0, errNoBlockNumberObs
	}

	return *e.parsedBlockNumberResponse, nil
}
