const (
	stateParameterNameBlockNumber = "blockNumber"
	endpointResultNameBlockNumber = stateParameterNameBlockNumber
)

func updateArchivalState(ctx ServiceStateUpdateContext) *StateParamUpdates {
	// TODO_IMPROVE(@commoddity): apply an expiry time to the `expectedBalance`.
	// 
	expectedBalance, found := ctx.GetStrParam(paramArchivalBalance)
	// expectedBalance already set: skip the rest of the processing
	if found {
		return nil
	}

	perceivedBlockNumber, found := ctx.GetIntParam(paramBlockNumber)
	// no perceived block number set yet: skip archival state update.
	if !found {
		return nil
	}

	// Set the archival block if not set already.
	archivalBlock, found := ctx.GetStrParam(paramArchivalBlock)
	if !found {
		archivalBlock = calculateArchivalBlock(config, perceivedBlockNumber)
		ctx.SetStrParam(paramArchivalBlock, archivalBlock)
	}

	// fetch the latest archival balance consensus map.
	balanceConsensus := ctx.GetConsensusParam(paramArchivalBalanceConsensus)

	// Update the archival balance consensus map from the endpoint results.
	for _, updatedEndpoint := range ctx.GetUpdatedEndpoints() {
		// skip out of sync endpoints
		endpointBlockNumber, found := updatedEndpoint.GetIntResult(resultBlockNumber)
		if !found || endpointBlockNumber < perceivedBlockNumber {
			continue
		}

		// skip endpoints with no archival balance result.
		endpointArchivalBalance, found := updatedEndpoint.GetStrResult(resultArchivalBalance)
		if !found {
			continue
		}

		// update the balance consensus with the endpoint's result.
		balanceConsensus[endpointArchivalBalance]++
	}

	// Set the archival balance concensus parameter on the context.
	ctx.SetConsensusParam(paramArchivalBalanceConcensus, balanceConsensus)

	// Build and return the set of updated state parameters.
	return ctx.BuildStateParameterUpdateSet()
}
