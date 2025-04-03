package evm

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// archivalConsensusThreshold is the number of endpoints that must agree on the archival balance for the randomly
// selected archival block number before it is considered to be the source of truth for the archival check.
// TODO_TECHDEBT(@commoddity): settle on a final value for this.
const archivalConsensusThreshold = 5

// The archival check verifies that nodes can provide accurate historical blockchain data. Here's how it works:
//   - Identify specific blockchain contracts that have been widely used since early in the chain's history.
//   - The system selects a random historical block from the past (between when the contract was first deployed and what's considered "recent");
//   - PATH queries all endpoints in session for the contract's balance with eth_getBalance at this historical block.
//   - When <archivalConsensusThreshold> endpoints independently report the same balance value, this becomes our "source of truth" for that contract at that block.
//   - Each endpoint is evaluated against this established truth - any endpoint that reports a different balance value is flagged as lacking proper archival data and will be filtered out by QoS.
type archivalState struct {
	logger polylog.Logger
	// archivalCheckConfig contains all configurable values for an EVM archival check.
	archivalCheckConfig EVMArchivalCheckConfig

	// blockNumberHex is a randomly selected block number from which to check the balance of the contract.
	//
	// It is calculated using the `calculateArchivalBlockNumber` method, which selects a block from the range:
	// 		- earliest possible block = <archivalCheckConfig.ContractStartBlock>
	// 		- latest possible block = <perceivedBlockNumber> - <archivalCheckConfig.Threshold>
	//
	// eg. if perceivedBlockNumber = 100, archivalThreshold = 10, contractStartBlock = 10, then the block number will be between 10 and 90.
	//
	// Format will look like: `0x3f8627c`
	blockNumberHex string

	// balance is the balance of the contract at the block number specified in `blockNumberHex`.
	// It is determined by reaching a consensus on the balance among `<archivalConsensusThreshold>` endpoints.
	expectedBalance string

	// balanceConsensus is a map where:
	//   - key: hex balance value for the archival block number
	//   - value: number of endpoints that reported the balance
	//
	// eg. {"0x1ce31607bc8f16a8c53d80": 5, "0x1ce31607bc8f16a8c53d81": 3}
	//
	// When a single hex value with a count of <archivalConsensusThreshold> is reached, the balance
	// is set as the expected archival balance source of truth for archival validation.
	balanceConsensus map[string]int

	// TODO_IMPROVE(@commoddity): set an expiry time for the archival state so that a new random block can be selected on an interval.
}

func (a *archivalState) isEnabled() bool {
	return a.archivalCheckConfig.Enabled
}

// processEndpointArchivalData processes all archival-related data from an endpoint
// and updates the archival state accordingly.
//
// Returns true if the endpoint's archival data was processed successfully,
// false if the endpoint should be skipped (e.g., missing archival balance).
func (a *archivalState) processEndpointArchivalData(endpoint endpoint) bool {
	if !a.isEnabled() {
		return true
	}

	// Attempt to retrieve the archival balance from the endpoint
	balance, err := endpoint.getArchivalBalance()
	if err != nil {
		return false
	}

	// Update the consensus map to determine the balance at the perceived block number
	a.updateConsensusMap(balance)
	return true
}

// updateConsensusMap updates the balance consensus map to determine
// the expected archival balance at the archival block number.
//
// Once <archivalConsensusThreshold> endpoints report the same balance,
// the balance is set as the expected archival balance.
func (a *archivalState) updateConsensusMap(balance string) {
	if a.expectedBalance == "" && balance != "" {
		a.balanceConsensus[balance]++
	}
}

// updateArchivalState updates the archival state based on the perceived block number.
// This handles calculating the archival block number if needed and updating the
// consensus-based archival balance.
func (a *archivalState) updateArchivalState(perceivedBlockNumber uint64) {
	// Skip further processing if archival checks are not enabled
	if !a.isEnabled() {
		return
	}

	// If the archival block number is not yet set for the service, calculate it.
	// This requires that the perceived block number is set to determine the latest possible block number.
	if perceivedBlockNumber != 0 && a.blockNumberHex == "" {
		a.logger.Info().Msg("Calculating archival block number")
		a.calculateArchivalBlockNumber(perceivedBlockNumber)
	}

	// If the expected archival balance is not yet set for the service, set it.
	// This utilizes the consensus map to determine a source of truth for the archival balance.
	// If <archivalConsensusThreshold> endpoints report the same balance, it is considered the source of truth.
	if a.expectedBalance == "" {
		a.updateArchivalBalance(archivalConsensusThreshold)
	}
}

// calculateArchivalBlockNumber determines a random archival block number based on the perceived block number.
// The function applies the following logic:
//   - If perceived block is below threshold, returns block 0
//   - Otherwise, calculates a random block between minArchivalBlock and (perceivedBlockNumber - threshold)
//   - Ensures the returned block number is never below the contract start block
func (a *archivalState) calculateArchivalBlockNumber(perceivedBlockNumber uint64) string {
	archivalThreshold := a.archivalCheckConfig.Threshold
	minArchivalBlock := a.archivalCheckConfig.ContractStartBlock

	var blockNumHex string
	// Case 1: Block number is below or equal to the archival threshold
	if perceivedBlockNumber <= archivalThreshold {
		blockNumHex = blockNumberToHex(0)
	} else {
		// Case 2: Block number is above the archival threshold
		maxBlockNumber := perceivedBlockNumber - archivalThreshold

		// Ensure we don't go below the minimum archival block
		if maxBlockNumber < minArchivalBlock {
			blockNumHex = blockNumberToHex(minArchivalBlock)
		} else {
			// Generate a random block number within valid range
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			rangeSize := maxBlockNumber - minArchivalBlock + 1
			blockNumHex = blockNumberToHex(minArchivalBlock + (r.Uint64() % rangeSize))
		}
	}

	// Store the calculated block number in the archival state
	a.blockNumberHex = blockNumHex
	return blockNumHex
}

// updateArchivalBalance checks for consensus and updates the archival balance if it hasn't been set yet.
// For example, if more than 5 endpoints report the same balance, the archival balance is updated to the consensus balance.
func (a *archivalState) updateArchivalBalance(consensusThreshold int) {
	for balance, count := range a.balanceConsensus {
		// If we've reached the threshold, update the expected balance
		if count >= consensusThreshold {
			a.logger.Info().Msgf("Updating expected archival balance for block number %s to %s", a.blockNumberHex, balance)
			a.expectedBalance = balance
			// Reset consensus map after consensus is reached.
			a.balanceConsensus = make(map[string]int)
			break
		}
	}
}

// blockNumberToHex converts a integer block number to its hexadecimal representation.
func blockNumberToHex(blockNumber uint64) string {
	return fmt.Sprintf("0x%x", blockNumber)
}

// getArchivalCheckRequest returns a JSONRPC request to check the balance of:
//   - the contract specified in `a.archivalCheckConfig.ContractAddress`
//   - at the block number specified in `a.blockNumberHex`
//
// It returns false if the archival check is not enabled for the service or if the block number has not been set.
func (a *archivalState) getArchivalCheckRequest() (jsonrpc.Request, bool) {
	if a.archivalCheckConfig.Enabled && a.blockNumberHex != "" {
		archivalCheckReq := jsonrpc.NewRequest(
			idArchivalBlockCheck,
			methodGetBalance,
			// Pass params in this order, eg. "params":["0x28C6c06298d514Db089934071355E5743bf21d60", "0xe71e1d"]
			// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_getbalance
			a.archivalCheckConfig.ContractAddress,
			a.blockNumberHex,
		)

		return archivalCheckReq, true
	}

	return jsonrpc.Request{}, false
}
