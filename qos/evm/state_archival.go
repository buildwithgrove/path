package evm

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
)

// archivalConsensusThreshold is the # of endpoints that must agree on the balance:
//   - at the `archivalState.blockNumberHex`
//   - for the contract specified in `EVMArchivalCheckConfig.ContractAddress`
//
// Once a consensus is reached, the balance is set in `expectedBalance`.
const archivalConsensusThreshold = 5

// The archival check verifies that nodes can provide accurate historical blockchain data.
//
// Here's how it works:
//   - Uses a specific contract with frequent balance changes (e.g. USDC) - `EVMArchivalCheckConfig.ContractAddress`
//   - Selects a random historical block from the past for `blockNumberHex`.
//   - >= 5 endpoints agree on the balance at `blockNumberHex`, `expectedBalance` is set.
//   - When filtering valid endpoints their `observedArchivalBalance` is validated against `expectedBalance`.
type archivalState struct {
	logger polylog.Logger

	// archivalCheckConfig contains all configurable values for an EVM archival check.
	archivalCheckConfig eVMArchivalCheckConfig

	// balanceConsensus is a map where:
	//   - key: hex balance value for the archival block number
	//   - value: number of endpoints that reported the balance
	// eg. {"0x1ce31607bc8f16a8c53d80": 5, "0x1ce31607bc8f16a8c53d81": 3}
	//
	// When a single hex balance is agreed on by >= 5 endpoints, it is set as the expected archival balance.
	//
	// Once the expected balance is set, the balanceConsensus map is cleared and no longer used.
	balanceConsensus map[string]int

	// blockNumberHex is the archival block number for which to check the balance of the contract, eg. 0x3f8627c.
	//
	// It is calculated using the `calculateArchivalBlockNumber` method, which selects a block from the range:
	// 		- earliest possible block = <archivalCheckConfig.ContractStartBlock>
	// 		- latest possible block = <perceivedBlockNumber> - <archivalCheckConfig.Threshold>
	//
	// Example: <archivalCheckConfig.ContractStartBlock> = 15, <perceivedBlockNumber> = 100, <archivalCheckConfig.Threshold> = 10
	//     		the calculated block number will be between 15 and 90.
	blockNumberHex string

	// expectedBalance is the agreed upon balance:
	//   - at the block number specified in `blockNumberHex`
	//   - for the contract specified in `EVMArchivalCheckConfig.ContractAddress`
	//
	// It is set once >= 5 endpoints agree on the balance at `blockNumberHex`.
	//
	// Once a consensus is reached, the balance is set in `expectedBalance` and `updateArchivalState` becomes a no-op.
	expectedBalance string

	// TODO_IMPROVE(@commoddity): set an expiry time for the `expectedBalance` so
	// that a new expected balance can be calculated, eg. every hour or two.
}

// updateArchivalState updates the archival state, to determine the `archivalState.expectedBalance`.
// once `archivalState.expectedBalance` is set, this method becomes a no-op.
func (a *archivalState) updateArchivalState(
	perceivedBlockNumber uint64,
	updatedEndpoints map[protocol.EndpointAddr]endpoint,
) {
	// If the expected archival balance is already set, there is no need to update the archival state.
	if a.expectedBalance != "" {
		return
	}

	// If the archival block number is not yet set for the service, calculate it.
	if perceivedBlockNumber != 0 && a.blockNumberHex == "" {
		a.calculateArchivalBlockNumber(perceivedBlockNumber)
	}

	// If the expected archival balance is not yet set for the service, set it.
	if a.blockNumberHex != "" && a.expectedBalance == "" {
		a.updateExpectedBalance(updatedEndpoints)
	}
}

// isEnabled returns true if archival checks are enabled for the service.
// Not all EVM services will require archival checks (for example, if a service is expected to run pruned nodes).
func (a *archivalState) isEnabled() bool {
	return !a.archivalCheckConfig.IsEmpty()
}

// calculateArchivalBlockNumber determines a, archival block number based on the perceived block number.
// See comment on `archivalState.blockNumberHex` in `archivalState` struct for more details on the calculation.
func (a *archivalState) calculateArchivalBlockNumber(perceivedBlockNumber uint64) {
	archivalThreshold := a.archivalCheckConfig.threshold
	minArchivalBlock := a.archivalCheckConfig.contractStartBlock

	var blockNumHex string
	// Case 1: Block number is below or equal to the archival threshold
	if perceivedBlockNumber <= archivalThreshold {
		blockNumHex = blockNumberToHex(1)
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

	a.logger.Info().Msgf("Calculated archival block number: %s", blockNumHex)
	a.blockNumberHex = blockNumHex
}

// blockNumberToHex converts a integer block number to its hexadecimal representation.
// eg. 66609788 -> "0x3f8627c"
func blockNumberToHex(blockNumber uint64) string {
	return fmt.Sprintf("0x%x", blockNumber)
}

// updateExpectedBalance checks for consensus and updates the expected balance in the archival state.
// When >= 5 endpoints agree on the same balance, it is set as the expected archival balance.
func (a *archivalState) updateExpectedBalance(updatedEndpoints map[protocol.EndpointAddr]endpoint) {
	for _, endpoint := range updatedEndpoints {
		// Get the observed balance at the archival block number from the endpoint observation.
		balance, err := endpoint.getArchivalBalance()
		if err != nil {
			a.logger.Info().Err(err).Msg("Skipping endpoint with no observed archival balance")
			continue
		}

		// Update the balance consensus map.
		count := a.balanceConsensus[balance] + 1
		a.balanceConsensus[balance] = count

		// Check for consensus immediately after updating count
		if count >= archivalConsensusThreshold {
			a.expectedBalance = balance
			a.logger.Info().
				Str("archival_block_number", a.blockNumberHex).
				Str("contract_address", a.archivalCheckConfig.contractAddress).
				Str("expected_balance", balance).
				Msg("Updated expected archival balance")

			a.balanceConsensus = make(map[string]int) // Clear map as it's no longer needed.
			return
		}
	}
}
