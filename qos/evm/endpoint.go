package evm

import "time"

// endpoint captures the details required to validate an EVM endpoint.
// It contains all checks that should be run for the endpoint to validate
// it is providing a valid response to service requests.
// TODO_IMPROVE: Rename to 'endpointValidation'
type endpoint struct {
	hasReturnedEmptyResponse    bool
	hasReturnedInvalidResponse  bool
	invalidResponseLastObserved *time.Time
	checkBlockNumber            endpointCheckBlockNumber
	checkChainID                endpointCheckChainID
	checkArchival               endpointCheckArchival
}
