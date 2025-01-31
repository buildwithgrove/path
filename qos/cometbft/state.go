package cometbft

import (
	"fmt"
	"sync"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos"
)

var _ qos.ServiceState = &ServiceState{}

// ServiceState keeps the expected current state of the CometBFT blockchain based on the endpoints' responses to different requests.
type ServiceState struct {
	Logger polylog.Logger

	stateLock sync.RWMutex

	// perceivedBlockNumber is the perceived current block number based on endpoints' responses to `/status` requests.
	// It is calculated as the maximum of block height reported by any of the endpoints.
	//
	// See the following link for more details:
	// https://docs.cometbft.com/v1.0/spec/rpc/#status
	perceivedBlockNumber uint64
}

// TODO_FUTURE: add an endpoint ranking method which can be used to assign a rank/score to a valid endpoint to guide endpoint selection.
//
// ValidateEndpoint returns an error if the supplied endpoint is not valid based on the perceived state of the CometBFT blockchain.
func (s *ServiceState) ValidateEndpoint(qosEndpoint qos.Endpoint) error {
	s.stateLock.RLock()
	defer s.stateLock.RUnlock()

	cometbftEndpoint, ok := qosEndpoint.(endpoint)
	if !ok {
		return fmt.Errorf("endpoint was not of type cometbft.endpoint")
	}

	// CometBFT does not use a chain ID.
	if err := cometbftEndpoint.Validate(""); err != nil {
		return err
	}

	if err := validateEndpointBlockNumber(cometbftEndpoint, s.perceivedBlockNumber); err != nil {
		return err
	}

	return nil
}

// UpdateFromEndpoints updates the service state using estimation(s) derived from the set of updated endpoints.
// This only includes the set of endpoints for which an observation was received.
func (s *ServiceState) UpdateFromEndpoints(updatedEndpoints map[protocol.EndpointAddr]qos.Endpoint) error {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	for endpointAddr, updatedEndpoint := range updatedEndpoints {
		logger := s.Logger.With(
			"endpoint_addr", endpointAddr,
			"perceived_block_number", s.perceivedBlockNumber,
		)

		cometbftEndpoint, ok := updatedEndpoint.(endpoint)
		if !ok {
			logger.Info().Msg("endpoint was not of type cometbft.endpoint. Skipping.")
			continue
		}

		// Do NOT use the endpoint for updating the perceived state of the CometBFT blockchain if the endpoint is not considered valid.
		// e.g. an endpoint with an invalid response to `/status` will not be used to update the perceived block number.
		if err := cometbftEndpoint.Validate(""); err != nil {
			logger.Info().Err(err).Msg("Skipping endpoint with invalid chain id")
			continue
		}

		// TODO_TECHDEBT: use a more resilient method for updating block height.
		// e.g. one endpoint returning a very large number as block height should
		// not result in all other endpoints being marked as invalid.
		blockNumber, err := cometbftEndpoint.getBlockNumber()
		if err != nil {
			logger.Info().Err(err).Msg("Skipping endpoint with invalid block number")
			continue
		}

		s.perceivedBlockNumber = blockNumber

		logger.With("endpoint_block_number", blockNumber).Info().Msg("Updating latest block height")
	}

	return nil
}

// validateEndpointBlockNumber validates the supplied endpoint against the supplied perceived block number for the CometBFT blockchain.
func validateEndpointBlockNumber(cometbftEndpoint endpoint, perceivedBlockNumber uint64) error {
	blockNumber, err := cometbftEndpoint.getBlockNumber()
	if err != nil {
		return err
	}

	if blockNumber < perceivedBlockNumber {
		return fmt.Errorf("endpoint has block height %d, perceived block height is %d", blockNumber, perceivedBlockNumber)
	}

	return nil
}
