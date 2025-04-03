package evm

import (
	framework "github.com/buildwithgrove/path/qos/framework/jsonrpc"
)

type evmStateUpdater struct {
	logger polylog.Logger
	config Config
}

// implements the framework.StateUpdater interface
func (esu evmStateUpdater) UpdateServiceState(ctx *StateUpdateContext) *framework.StateUpdate {

	var maxObservedBlockNumber int

	// Loop over all updated endpoints.
	for _, endpoint := range ctx.GetUpdatedEndpoints() {

		// validate endpoint's Chain ID attribute.
		endpointChainID, err := endpoint.GetStringAttribute(attrETHChainID)
		// Do not use endpoints with invalid/missing chain ID for service state update.
		if err != nil {
			logger.Debug().Err(err).Msg("Skipping endpoint with missing/invalid chain id")
			continue
		}

		// TODO_TECHDEBT(@adshmh): use a more resilient method for updating block height.
		// E.g. one endpoint returning a very large number as block height should
		// not result in all other endpoints being marked as invalid.
		//
		// validate endpoint's Block Number attribute.
		endpointBlockNumber, err := endpoint.GetIntAttribute(attrETHBlockNumber)
		if err != nil {
			logger.Debug().Err(err).Msg("Skipping endpoint with missing/invalid block number")
			continue
		}

		if endpointBlockNumber > maxObservedBlockNumber {
			maxObservedBlockNumber = endpointBlockNumber
		}
	}

	// Fetch the latest block number from the service state.
	perceivedBlockNumber := ctx.GetStateIntAttribute(attrETHBlockNumber)

	// Skip state update if block number has not increased.
	if perceivedBlockNumber >= maxObservedBlockNumber {
		return nil
	}

	// Update the service state to maximum observed block number.
	s.MarkStateIntAttributeForUpdate(attrETHBlockNumber, maxObservedBlockNumber)
	return s.BuildStateUpdateData()
}
