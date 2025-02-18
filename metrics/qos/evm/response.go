package evm

import (
	"github.com/buildwithgrove/path/observation/qos"
)

// response defines methods needed for response metrics collection.
// Abstracts proto-specific details from metrics logic.
// Note: Update when adding new metric requirements.
type response interface {
	GetValid() bool
	GetInvalidReason() qos.EVMResponseInvalidReason
}

// getResponseFromObservation extracts the response data from an endpoint observation.
// Returns nil if no response is present.
// Note: Update when adding new response types.
func extractEndpointResponseFromObservation(observation *qos.EVMEndpointObservation) response {
	if observation == nil {
		return nil
	}

	if response := observation.GetChainIdResponse(); response != nil {
		return response
	}

	if response := observation.GetBlockNumberResponse(); response != nil {
		return response
	}

	if response := observation.GetUnrecognizedResponse(); response != nil {
		return response
	}

	if response := observation.GetEmptyResponse(); response != nil {
		return response
	}

	return nil
}
