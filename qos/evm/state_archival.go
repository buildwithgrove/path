package evm

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// archivalConsensusThreshold is the number of endpoints that must agree on the archival balance for the randomly
// selected archival block number before it is considered to be the source of truth for the archival check.
// TODO_TECHDEBT(@commoddity): settle of a final value for this.
const archivalConsensusThreshold = 5

type archivalState struct {
	// archivalCheckConfig contains all configurable values for an EVM archival check.
	archivalCheckConfig EVMArchivalCheckConfig

	// blockNumberHex is a randomly selected block number from which to check the balance of the contract.
	// It is calculated using the `calculateArchivalBlockNumber` method, which selects a block from the range:
	// earliest possible block = <archivalCheckConfig.ContractStartBlock>
	// latest possible block = <perceivedBlockNumber> - <archivalCheckConfig.Threshold>
	blockNumberHex string

	// balance is the balance of the contract at the block number specified in `blockNumberHex`.
	// It is determined by reaching a consensus on the balance among `<archivalConsensusThreshold>` endpoints.
	balance string

	// balanceConsensus is a map where:
	//   - key: hex balance value for the archival block number
	//   - value: number of endpoints that reported the balance
	//
	// eg. {"0x1ce31607bc8f16a8c53d80": 5, "0x1ce31607bc8f16a8c53d81": 3}
	//
	// When a single hex value with a count of <archivalConsensusThreshold> is reached, the balance
	// is set as the expected archival balance source of truth for archival validation.
	balanceConsensus map[string]int
}

// initializeConsensusMap initializes the balance consensus map if it doesn't exist.
func (a *archivalState) initializeConsensusMap() {
	if a.balanceConsensus == nil {
		a.balanceConsensus = make(map[string]int)
	}
}

// updateConsensusMap updates the balance consensus map to determine
// the expected archival balance at the archival block number.
//
// Once <archivalConsensusThreshold> endpoints report the same balance,
// the balance is set as the expected archival balance.
func (a *archivalState) updateConsensusMap(balance string) {
	if a.getBalance() == "" && balance != "" {
		a.balanceConsensus[balance]++
	}
}

// getBalance returns the current archival balance.
func (a *archivalState) getBalance() string {
	return a.balance
}

// getBlockNumberHex returns the current archival block number in hexadecimal format.
func (a *archivalState) getBlockNumberHex() string {
	return a.blockNumberHex
}

// getArchivalCheckRequest returns a JSONRPC request to check the balance of the contract at
// the block number specified in `blockNumberHex`.
//
// It returns false if the archival check is not enabled for the service or if the block number has not been set.
func (a *archivalState) getArchivalCheckRequest() (jsonrpc.Request, bool) {
	if a.archivalCheckConfig.Enabled && a.getBlockNumberHex() != "" {
		archivalCheckReq := jsonrpc.NewRequest(
			idArchivalBlockCheck,
			methodGetBalance,
			// Pass params in this order, eg. "params":["0x28C6c06298d514Db089934071355E5743bf21d60", "0xe71e1d"]
			// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_getbalance
			a.archivalCheckConfig.ContractAddress,
			a.getBlockNumberHex(),
		)

		return archivalCheckReq, true
	}

	return jsonrpc.Request{}, false
}

// calculateArchivalBlockNumber returns a random archival block number based on the perceived block number.
// The function applies the following logic:
// - If perceived block is below threshold, returns block 0
// - Otherwise, calculates a random block between min archival block and (perceived block - threshold)
// - Ensures the returned block number is never below the contract start block
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

	// Store the calculated block number in the service state
	a.blockNumberHex = blockNumHex
	return blockNumHex
}

// updateArchivalBalance checks for consensus and updates the archival balance if it hasn't been set yet.
// For example, if more than 5 endpoints report the same balance, the archival balance is updated to the consensus balance.
func (a *archivalState) updateArchivalBalance(consensusThreshold int) {
	for balance, count := range a.balanceConsensus {
		if count >= consensusThreshold {
			a.balance = balance
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
