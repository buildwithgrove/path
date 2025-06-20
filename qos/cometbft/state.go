package cometbft

import (
	"fmt"
	"sync"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
)

// ServiceState keeps the expected current state of the CometBFT blockchain
// based on the endpoints' responses to different requests.
type ServiceState struct {
	logger polylog.Logger

	serviceStateLock sync.RWMutex

	// chainID is the chain ID of the CometBFT blockchain.
	// Corresponds with with the `network` field returned by the `/status` endpoint.
	chainID string

	// perceivedBlockNumber is the perceived current block number based on endpoints' responses to `/status` requests.
	// It is calculated as the maximum of block height reported by any of the endpoints.
	//
	// See the following link for more details:
	// https://docs.cometbft.com/v1.0/spec/rpc/#status
	perceivedBlockNumber uint64
}

// ValidateEndpoint returns an error if the supplied endpoint is not valid based on the perceived state of the CometBFT blockchain.
func (s *ServiceState) ValidateEndpoint(endpoint endpoint) error {
	s.serviceStateLock.RLock()
	defer s.serviceStateLock.RUnlock()

	// Basic validation of the endpoint based on prior observations.
	if err := endpoint.Validate(s.chainID); err != nil {
		return err
	}

	// Validate the endpoint based on the perceived block number.
	if err := validateEndpointBlockNumber(endpoint, s.perceivedBlockNumber); err != nil {
		return err
	}

	return nil
}

// UpdateFromObservations updates the service state using estimation(s) derived from the set of updated endpoints.
// NOTE: This only includes the set of endpoints for which an observation was received.
func (s *ServiceState) UpdateFromEndpoints(updatedEndpoints map[protocol.EndpointAddr]endpoint) error {
	s.serviceStateLock.Lock()
	defer s.serviceStateLock.Unlock()

	for endpointAddr, endpoint := range updatedEndpoints {
		logger := s.logger.With(
			"endpoint_addr", endpointAddr,
			"perceived_block_number", s.perceivedBlockNumber,
			"service_id", s.chainID,
		)

		// DO NOT use the endpoint for updating the perceived state of the CometBFT blockchain if the endpoint is not considered valid.
		// E.g. an endpoint with an invalid response to `/status` will not be used to update the perceived block number.
		if err := endpoint.Validate(s.chainID); err != nil {
			logger.Error().Err(err).Msgf("❌ Skipping endpoint with invalid chain id '%s' for endpoint '%s'", s.chainID, endpointAddr)
			continue
		}

		// TODO_TECHDEBT: use a more resilient method for updating block height.
		// E.g. one endpoint returning a very large number as block height should
		// not result in all other endpoints being marked as invalid.
		blockNumber, err := endpoint.GetBlockNumber()
		if err != nil {
			logger.Error().Err(err).Msgf("❌ Skipping endpoint with invalid block number '%d' for endpoint '%s'", blockNumber, endpointAddr)
			continue
		}

		s.perceivedBlockNumber = blockNumber

		logger.With("endpoint_block_number", blockNumber).Info().Msgf("✅ Updating latest block height for endpoint '%s'", endpointAddr)
	}

	return nil
}

// validateEndpointBlockNumber validates the supplied endpoint against the supplied
// perceived block number for the CometBFT blockchain.
func validateEndpointBlockNumber(endpoint endpoint, perceivedBlockNumber uint64) error {
	blockNumber, err := endpoint.GetBlockNumber()
	if err != nil {
		return err
	}

	// The endpoint is behind (out of sync) with the network (i.e. the perceived block number).
	if blockNumber < perceivedBlockNumber {
		return fmt.Errorf("endpoint has block height %d, which is SMALLER THAN the perceived block height is %d", blockNumber, perceivedBlockNumber)
	}

	return nil
}
