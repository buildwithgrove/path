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
func evmEndpointSelector(ctx *EndpointSelectionContext) (protocol.EndpointAddr, error) {
	// TODO_IN_THIS_PR: the framework should log an entry if the state attribute is not set.
	// Fetch latest block number from the service state.
	perceivedBlockNumber := ctx.GetState().GetIntAttribute(attrETHBlockNumber)

	// TODO_FUTURE(@adshmh): use service-specific metrics to add an endpoint ranking method.
	// e.g. use latency to break the tie between valid endpoints.
	for _, endpoint := range ctx.GetAvailableEndpoints() {
		endpointChainID, err := endpoint.GetStringAttribute(attrETHChainID)
		// ChainID attribute not set: Disqualify the endpoint.
		if err != nil {
			ctx.DisqualifyEndpoint(endpoint, err)
			continue
		}

		// TODO_MVP(@adshmh): pass the EVM config to endpoint selector.
		// Invalid ChainID returned by the endpoint: Disqualify.
		if endpointChainID != config.GetChainID() {
			ctx.DisqualifyEndpoint(endpoint, fmt.Errorf("invalid chain ID %s, expected: %s", endpointChainID, config.GetChainID()))
			continue
		}

		endpointBlockNumber, err := endpoint.GetIntAttribute(attrETHBlockNumber)
		// BlockNumber attribute not set: Disqualify the endpoint.
		if err != nil {
			ctx.DisqualifyEndpoint(endpoint, err)
			continue
		}

		// endpoint will only be disqualified if the State has reported a block number.
		if perceivedBlockNumber <= 0 {
			continue
		}

		// endpoint is out-of-sync: Disqualify.
		if endpointBlockNumber < perceivedBlockNumber {
			ctx.DisqualifyEndpoint(endpoint, fmt.Errorf("endpoint out of sync got %d block number, expected: %d", endpointBlockNumber, perceivedBlockNumber))
			continue
		}

		// TODO_IN_THIS_PR: validate archival state.
	}

	// All invalid endpoints have been marked as disqualified.
	// Return a randomly selected Qualified endpoint.
	return ctx.SelectRandomQualifiedEndpoint()
}
