package evm

import (
	"github.com/buildwithgrove/path/observation/qos"
)

// response defines methods needed for response metrics collection.
// Abstracts proto-specific details from metrics logic.
// DEV_NOTE: You MUST update this when adding new metric requirements.
type response interface {
	GetValid() bool
	GetInvalidReason() qos.EVMResponseInvalidReason
}

// getResponseFromObservation extracts the response data from an endpoint observation.
// Returns nil if no response is present.
// DEV_NOTE: You MUST update this when adding new metric requirements.
func extractEndpointResponseFromObservation(observation *qos.EVMEndpointObservation) response {
	if observation == nil {
		return nil
	}

	// handle chain_id response
	if response := observation.GetChainIdResponse(); response != nil {
		return response
	}

	// handle block_number response
	if response := observation.GetBlockNumberResponse(); response != nil {
		return response
	}

	// handle unrecognized response
	if response := observation.GetUnrecognizedResponse(); response != nil {
		return response
	}

	// handle empty response
	if response := observation.GetEmptyResponse(); response != nil {
		return response
	}

	return nil
}
