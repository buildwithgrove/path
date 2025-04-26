package evm

// endpoint captures the details required to validate an EVM endpoint.
// It contains all checks that should be run for the endpoint to validate
// it is providing a valid response to service requests.
type endpoint struct {
	hasReturnedEmptyResponse bool
	checkBlockNumber         endpointCheckBlockNumber
	checkChainID             endpointCheckChainID
	checkArchival            endpointCheckArchival
}
