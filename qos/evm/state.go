package evm

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
)

// ServiceState keeps the expected current state of the EVM blockchain based on the endpoints' responses to
// different requests.
type ServiceState struct {
	logger polylog.Logger

	serviceStateLock sync.RWMutex

	// chainID is the expected value of the `Result` field in any endpoint's response to an `eth_chainId` request.
	//
	// See the following link for more details: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_chainid
	//
	// Chain IDs Reference: https://chainlist.org/
	chainID string

	// perceivedBlockNumber is the perceived current block number based on endpoints' responses to `eth_blockNumber` requests.
	// It is calculated as the maximum of block height reported by any of the endpoints.
	//
	// See the following link for more details:
	// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_blocknumber
	perceivedBlockNumber uint64

	// archivalCheckConfig contains all configurable values for an EVM archival check.
	archivalCheckConfig EVMArchivalCheckConfig
	// archivalState contains the current state of the EVM archival check.
	archivalState evmArchivalState
}

// consensusThreshold is the number of endpoints that must agree on the archival balance for the randomly
// selected archival block number before it is considered to be the source of truth for the archival check.
// TODO_IMPROVE(@commoddity): make this value configurable.
const consensusThreshold = 5

// evmArchivalState contains the current state of the EVM archival check for the service.
type evmArchivalState struct {
	// blockNumberHex is a randomly selected block number from which to check the balance of the contract.
	blockNumberHex string
	// balance is the balance of the contract at the block number specified in `blockNumberHex`.
	balance string
	// balanceConsensus is a map of balances and the number of endpoints that reported them.
	balanceConsensus map[string]int
}

// TODO_FUTURE: add an endpoint ranking method which can be used to assign a rank/score to a valid endpoint to guide endpoint selection.
//
// ValidateEndpoint returns an error if the supplied endpoint is not valid based on the perceived state of the EVM blockchain.
func (s *ServiceState) ValidateEndpoint(endpoint endpoint, endpointAddr protocol.EndpointAddr) error {
	s.serviceStateLock.RLock()
	defer s.serviceStateLock.RUnlock()

	if err := endpoint.validateEmptyResponse(); err != nil {
		return err
	}
	if err := endpoint.validateChainID(s.chainID); err != nil {
		return err
	}
	if err := endpoint.validateBlockNumber(s.perceivedBlockNumber); err != nil {
		return err
	}
	if s.shouldPerformArchivalCheck() {
		if err := endpoint.validateArchivalCheck(s.archivalState.balance, endpointAddr); err != nil {
			return err
		}
	}
	return nil
}

// UpdateFromEndpoints updates the service state using estimation(s) derived from the set of updated endpoints.
// This only includes the set of endpoints for which an observation was received.
func (s *ServiceState) UpdateFromEndpoints(updatedEndpoints map[protocol.EndpointAddr]endpoint) error {
	s.serviceStateLock.Lock()
	defer s.serviceStateLock.Unlock()

	// Initialize consensus map if it doesn't exist.
	if s.archivalState.balanceConsensus == nil {
		s.archivalState.balanceConsensus = make(map[string]int)
	}

	for endpointAddr, endpoint := range updatedEndpoints {
		logger := s.logger.With(
			"endpoint_addr", endpointAddr,
			"perceived_block_number", s.perceivedBlockNumber,
		)

		// Validate the endpoint's chain ID; do not update the perceived block number if the chain ID is invalid.
		if err := endpoint.validateChainID(s.chainID); err != nil {
			logger.Info().Err(err).Msg("Skipping endpoint with invalid chain id")
			continue
		}

		// Retrieve the block number from the endpoint.
		blockNumber, err := endpoint.getBlockNumber()
		if err != nil {
			logger.Info().Err(err).Msg("Skipping endpoint with invalid block number")
			continue
		}

		// Update the perceived block number.
		s.perceivedBlockNumber = blockNumber

		// Attempt to retrieve the archival balance from the endpoint.
		balance, err := endpoint.getArchivalBalance()
		if err != nil {
			logger.Info().Err(err).Msg("Skipping endpoint without archival balance")
			continue
		}

		// Only count non-empty balances toward consensus.
		if s.archivalState.balance == "" && balance != "" {
			s.archivalState.balanceConsensus[balance]++
		}
	}

	// Update archival block number and archival balance only if not yet set.
	if s.perceivedBlockNumber != 0 && s.archivalState.blockNumberHex == "" {
		s.assignArchivalBlockNumber()
	}
	if s.archivalState.balance == "" {
		s.updateArchivalBalance(consensusThreshold)
	}

	return nil
}

// assignArchivalBlockNumber returns a random archival block number based on the perceived block number.
func (s *ServiceState) assignArchivalBlockNumber() string {
	archivalThreshold := s.archivalCheckConfig.Threshold
	minArchivalBlock := s.archivalCheckConfig.ContractStartBlock

	var result string
	if s.perceivedBlockNumber <= archivalThreshold {
		result = blockNumberToHex(0)
	} else {
		maxBlockNumber := s.perceivedBlockNumber - archivalThreshold
		if maxBlockNumber < minArchivalBlock {
			result = blockNumberToHex(minArchivalBlock)
		} else {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			rangeSize := maxBlockNumber - minArchivalBlock + 1
			result = blockNumberToHex(minArchivalBlock + (r.Uint64() % rangeSize))
		}
	}

	s.archivalState.blockNumberHex = result
	return result
}

// updateArchivalBalance checks for consensus and updates the archival balance if it hasn't been set yet.
func (s *ServiceState) updateArchivalBalance(consensusThreshold int) {
	for balance, count := range s.archivalState.balanceConsensus {
		if count >= consensusThreshold {
			s.archivalState.balance = balance
			// Reset consensus map after consensus is reached.
			s.archivalState.balanceConsensus = make(map[string]int)
			break
		}
	}
}

// shouldPerformArchivalCheck returns a boolean indicating whether the archival check should be performs.
// This depends on:
//   - The archival check is enabled for the service
//   - The archival block number to check the balance of has been set in the service state.
func (s *ServiceState) shouldPerformArchivalCheck() bool {
	if s.archivalCheckConfig.Enabled && s.getArchivalBlockNumberHex() != "" {
		return true
	}
	return false
}

func (s *ServiceState) getArchivalBlockNumberHex() string {
	return s.archivalState.blockNumberHex
}

func blockNumberToHex(blockNumber uint64) string {
	return fmt.Sprintf("0x%x", blockNumber)
}
