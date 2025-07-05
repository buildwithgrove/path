package cometbft

import (
	"fmt"
	"strconv"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
)

// The errors below list all the possible QoS validation errors of an endpoint.
var (
	// Health request validation errors.
	errHealthReqNoObs      = fmt.Errorf("endpoint has not had an observation of its response to a health check request")
	errHealthReqInvalidObs = fmt.Errorf("endpoint not healthy and returned an invalid response to a health check request")

	// Status request validation errors.
	errStatusReqNoChainIDObs      = fmt.Errorf("endpoint has not had an observation of its chain ID response to a status request")
	errStatusReqInvalidChainIDObs = fmt.Errorf("endpoint did not return a valid chain ID in its response to a status request")
	errStatusReqInvalidSyncedObs  = fmt.Errorf("endpoint has returned a response indicating it is catching up")

	errStatusReqNoBlockNumberObs      = fmt.Errorf("endpoint has not had an observation of its block height response to a status request")
	errStatusReqInvalidBlockNumberObs = fmt.Errorf("endpoint returned an invalid block height in its response to a status request")
)

// endpoint stores validation details for a CometBFT endpoint.
type endpoint struct {
	// healthResponse indicates if the endpoint passed health check via `/health` request.
	// nil if no health check has been performed yet.
	healthResponse *bool

	// chainIDResponse stores the chain ID of the endpoint.
	// Based off the response of NodeInfo.Network in the `/status` response.
	chainIDResponse string

	// catchingUpResponse stores if the endpoint is not catching up.
	// Based off the response of SyncInfo.CatchingUp in the `/status` response.
	catchingUpResponse bool

	// latestBlockHeightResponse stores latest block height reported by the endpoint.
	// nil if no block height request has been made yet.
	// Based off the response of SyncInfo.LatestBlockHeight in the `/status` response.
	latestBlockHeightResponse *uint64
}

// validate checks if endpoint has the required observations to be considered valid.
// Returns error if the necessary responses are either lacking or invalid.
func (e endpoint) validate(chainID string) error {
	switch {

	// No health check has been performed yet.
	case e.healthResponse == nil:
		return errHealthReqNoObs

	// Invalid health check response.
	case !*e.healthResponse:
		return errHealthReqInvalidObs

	// No chain ID response.
	case e.chainIDResponse == "":
		return errStatusReqNoChainIDObs

	// Invalid chain ID response.
	case e.chainIDResponse != chainID:
		return errStatusReqInvalidChainIDObs

	// Invalid catching up response.
	case e.catchingUpResponse:
		return errStatusReqInvalidSyncedObs

	// No block height request has been made yet.
	case e.latestBlockHeightResponse == nil:
		return errStatusReqNoBlockNumberObs

	// Invalid block height response.
	case *e.latestBlockHeightResponse == 0:
		return errStatusReqInvalidBlockNumberObs

	default:
		return nil
	}
}

// applyObservation updates the endpoint data using the provided observation.
// Returns true if the observation was recognized.
// IMPORTANT: This function mutates the endpoint.
func (e *endpoint) applyObservation(obs *qosobservations.CometBFTEndpointObservation) bool {
	// Health check observation made - update healthResponse.
	if healthResponse := obs.GetHealthResponse(); healthResponse != nil {
		observedHealth := healthResponse.GetHealthStatusResponse()
		e.healthResponse = &observedHealth
		return true
	}

	// Block height observation made - update parsedBlockNumberResponse.
	if blockNumberResponse := obs.GetStatusResponse(); blockNumberResponse != nil {
		e.chainIDResponse = blockNumberResponse.GetChainIdResponse()
		e.catchingUpResponse = blockNumberResponse.GetCatchingUpResponse()

		// base0 uses the string's prefix to determine its base.
		parsedBlockNumber, err := strconv.ParseUint(blockNumberResponse.GetLatestBlockHeightResponse(), 0, 64)

		// The endpoint returned an invalid response to a block height request.
		// Explicitly set the parsedBlockNumberResponse to zero.
		// This ensures consistent behavior since ParseUInt may return non-zero values on errors.
		if err != nil {
			zero := uint64(0)
			e.latestBlockHeightResponse = &zero
			return true
		}

		e.latestBlockHeightResponse = &parsedBlockNumber
		return true
	}

	// No observation made or recognized.
	return false
}

// getBlockNumber returns the parsed block number value for the endpoint if available.
func (e endpoint) getBlockNumber() (uint64, error) {
	// No block height request has been made yet.
	if e.latestBlockHeightResponse == nil {
		return 0, errStatusReqNoBlockNumberObs
	}

	// Return the parsed block number value.
	return *e.latestBlockHeightResponse, nil
}
