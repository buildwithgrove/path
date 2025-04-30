package evm

import (
	"github.com/buildwithgrove/path/protocol"
	framework "github.com/buildwithgrove/path/qos/framework/jsonrpc"
)

// The errors below list all the possible validation errors on an endpoint.
var (
	errNoChainIDObs             = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodChainID)
	errInvalidChainIDObs        = fmt.Errorf("endpoint returned an invalid response to a %q request", methodChainID)
	errNoBlockNumberObs         = fmt.Errorf("endpoint has not had an observation of its response to a %q request", methodBlockNumber)
	errInvalidBlockNumberObs    = fmt.Errorf("endpoint returned an invalid response to a %q request", methodBlockNumber)
	errHasReturnedEmptyResponse = errors.New("endpoint is invalid: history of empty responses")
)

var _ framework.EndpointSelector = evmEndpointSelector

// evmEndpointSelector selects an endpoint from the set of available ones.
// It uses the configuration and service state to filter out misconfigured/out-of-sync/invalid endpoints.
func evmEndpointSelector(
	ctx *EndpointSelectionContext,
	config EVMConfig,
) (protocol.EndpointAddr, error) {
	// Fetch latest block number from the service state.
	perceivedBlockNumber, _ := ctx.GetIntParam(methodETHBlockNumber)

	// The perceived block number not set yet: return a random endpoint
	if perceivedBlockNumber <= 0 {
		return ctx.SelectRandomQualifiedEndpoint()
	}

	return ctx.SelectRandomQualifiedEndpoint(
		func(endpoint *Endpoint) error {
			endpointChainID, _ := endpoint.GetStrResult(methodETHChainID)
			// Endpoint's chain ID does not match the expected value.
			// Disqualify the endpoint.
			if endpointChainID != config.GetChainID() {
				return ctx.DisqualifyEndpoint(endpoint, fmt.Sprintf("invalid chain ID %s, expected: %s", endpointChainID, config.GetChainID()))
			}

			endpointBlockNumber, _ := endpoint.GetIntResult(methodETHBlockNumber)
			// TODO_IN_THIS_PR: add slack from the configuration.
			// endpoint is out-of-sync: Disqualify.
			if endpointBlockNumber < perceivedBlockNumber {
				return ctx.DisqualifyEndpoint(endpoint, fmt.Errorf("out of sync: %d block number, perceived: %d", endpointBlockNumber, perceivedBlockNumber))
			}

			// TODO_IN_THIS_PR: validate archival state.
			return nil
		}
	)
}
