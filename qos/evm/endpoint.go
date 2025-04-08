package evm

import (
	"errors"
)

var errHasReturnedEmptyResponse = errors.New("endpoint is invalid: history of empty responses")

// Each endpoint check should use its own ID to avoid potential conflicts.
// ID of JSON-RPC requests for any new checks should be added to the list below.
type endpointCheckID int

// endpoint captures the details required to validate an EVM endpoint.
// It contains all checks that should be run for the endpoint to validate
// it is providing a valid response to service requests.
type endpoint struct {
	hasReturnedEmptyResponse bool
	checkBlockNumber         endpointCheckBlockNumber
	checkChainID             endpointCheckChainID
	checkArchival            endpointCheckArchival
}

// newEndpoint initializes a new endpoint with the checks that should be run for the endpoint.
func newEndpoint() endpoint {
	return endpoint{
		checkBlockNumber: endpointCheckBlockNumber{},
		checkChainID:     endpointCheckChainID{},
		checkArchival:    endpointCheckArchival{},
	}
}
