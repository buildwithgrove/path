package solana

import (
	"fmt"
	"sync"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
)

// SerciceState keeps the expected current state of the solana blockchain based on the endpoints' responses to
// different requests.
type ServiceState struct {
	Logger polylog.Logger

	stateLock sync.RWMutex
	// perceivedEpoch is the perceived current epoch based on endpoints' responses to `getEpochInfo` requests.
	// See the following link for more details:
	// https://solana.com/docs/rpc/http/getepochinfo
	perceivedEpoch uint64
	// perceivedBlockHeight is the perceived blockheight based on endpoints' responses to `getEpochInfo` requests.
	perceivedBlockHeight uint64
}

// TODO_FUTURE: add an endpoint ranking method which can be used to assign a rank/score to a valid endpoint to guide endpoint selection.
//
// ValidateEndpoint returns an error if the supplied endpoint is not valid based on the perceived state of Solana blockchain.
func (s *ServiceState) ValidateEndpoint(endpoint endpoint) error {
	s.stateLock.RLock()
	defer s.stateLock.RUnlock()

	if err := endpoint.ValidateBasic(); err != nil {
		return err
	}

	if endpoint.SolanaGetEpochInfoResponse.Epoch < s.perceivedEpoch {
		return fmt.Errorf("endpoint has epoch %d, perceived current epoch is %d", endpoint.SolanaGetEpochInfoResponse.Epoch, s.perceivedEpoch)
	}

	if endpoint.SolanaGetEpochInfoResponse.BlockHeight < s.perceivedBlockHeight {
		return fmt.Errorf("endpoint has block height %d, perceived block height is %d", endpoint.SolanaGetEpochInfoResponse.BlockHeight, s.perceivedBlockHeight)
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

		// The endpoint's Epoch should be at-least equal to the perceived epoch before being used to update the perceived state of Solana blockchain.
		if endpoint.SolanaGetEpochInfoResponse.Epoch < s.perceivedEpoch {
			continue
		}

		// The endpoint's BlockHeight should be greater than the perceived block height before being used to update the perceived state of Solana blockchain.
		if endpoint.SolanaGetEpochInfoResponse.BlockHeight <= s.perceivedBlockHeight {
			continue
		}

		// TODO_IMPROVE: use a more resilient method for updating block height.
		// e.g. one endpoint returning a very large number as block height should
		// not result in all other endpoints being marked as invalid.
		s.perceivedEpoch = endpoint.SolanaGetEpochInfoResponse.Epoch
		s.perceivedBlockHeight = endpoint.SolanaGetEpochInfoResponse.BlockHeight

		s.Logger.With(
			"endpoint", endpointAddr,
			"block height", s.perceivedBlockHeight,
			"epoch", s.perceivedEpoch,
		).Info().Msg("Updating latest block height")
	}

	return nil
}
