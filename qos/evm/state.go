package solana

import (
	"fmt"
	"sync"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
)

// ServiceState keeps the expected current state of the Solana blockchain based on the endpoints' responses to
// different requests.
type ServiceState struct {
	Logger polylog.Logger

	stateLock sync.RWMutex
	// estimatedEpoch is the estimated current epoch based on endpoints' responses to `getEpochInfo` requests.
	// See the following link for more details:
	// https://solana.com/docs/rpc/http/getepochinfo
	estimatedEpoch uint64
	// estimatedBlockHeight is the estimated blockheight based on endpoints' responses to `getEpochInfo` requests.
	estimatedBlockHeight uint64
}

// TODO_FUTURE: add an endpoint ranking method which can be used to assign a rank/score to a valid endpoint to guide endpoint selection.
//
// ValidateEndpoint returns an error if the supplied endpoint is not valid based on the estimated state of Solana blockchain.
func (s *ServiceState) ValidateEndpoint(endpoint endpoint) error {
	s.stateLock.RLock()
	defer s.stateLock.RUnlock()

	if err := endpoint.ValidateBasic(); err != nil {
		return err
	}

	if endpoint.GetEpochInfoResult.Epoch < s.estimatedEpoch {
		return fmt.Errorf("endpoint has epoch %d, estimated current epoch is %d", endpoint.GetEpochInfoResult.Epoch, s.estimatedEpoch)
	}

	if endpoint.GetEpochInfoResult.BlockHeight < s.estimatedBlockHeight {
		return fmt.Errorf("endpoint has block height %d, estimated block height is %d", endpoint.GetEpochInfoResult.BlockHeight, s.estimatedBlockHeight)
	}

	return nil
}

// UpdateFromObservations updates the service state using estimation(s) deriven from the set of updated endpoints, i.e. the set of endpoints for which
// an observation was received.
func (s *ServiceState) UpdateFromEndpoints(updatedEndpoints map[protocol.EndpointAddr]endpoint) error {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	for endpointAddr, endpoint := range updatedEndpoints {
		if err := endpoint.ValidateBasic(); err != nil {
			continue
		}

		// The endpoint's Epoch should be at-least equal to the estimated epoch before being used to update the estimated state of Solana blockchain.
		if endpoint.GetEpochInfoResult.Epoch < s.estimatedEpoch {
			continue
		}

		// The endpoint's BlockHeight should be greater than the estimated block height before being used to update the estimated state of Solana blockchain.
		if endpoint.GetEpochInfoResult.BlockHeight <= s.estimatedBlockHeight {
			continue
		}

		// TODO_TECHDEBT: use a more resilient method for updating block height.
		// e.g. one endpoint returning a very large number as block height should
		// not result in all other endpoints being marked as invalid.
		s.estimatedEpoch = endpoint.GetEpochInfoResult.Epoch
		s.estimatedBlockHeight = endpoint.GetEpochInfoResult.BlockHeight

		s.Logger.With(
			"endpoint", endpointAddr,
			"block height", s.estimatedBlockHeight,
			"epoch", s.estimatedEpoch,
		).Info().Msg("Updating latest block height")
	}

	return nil
}
