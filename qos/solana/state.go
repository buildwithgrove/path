package solana

import (
	"fmt"
	"sync"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos"
)

// ServiceState keeps the expected current state of the solana blockchain based on the endpoints' responses to
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
func (s *ServiceState) ValidateEndpoint(qosEndpoint qos.Endpoint) error {
	s.stateLock.RLock()
	defer s.stateLock.RUnlock()

	solanaEndpoint, ok := qosEndpoint.(endpoint)
	if !ok {
		return fmt.Errorf("endpoint was not of type solana.endpoint")
	}

	if err := solanaEndpoint.Validate(""); err != nil {
		return err
	}

	if solanaEndpoint.SolanaGetEpochInfoResponse.Epoch < s.perceivedEpoch {
		return fmt.Errorf("endpoint has epoch %d, perceived current epoch is %d", solanaEndpoint.SolanaGetEpochInfoResponse.Epoch, s.perceivedEpoch)
	}

	if solanaEndpoint.SolanaGetEpochInfoResponse.BlockHeight < s.perceivedBlockHeight {
		return fmt.Errorf("endpoint has block height %d, perceived block height is %d", solanaEndpoint.SolanaGetEpochInfoResponse.BlockHeight, s.perceivedBlockHeight)
	}

	return nil
}

// UpdateFromObservations updates the service state using estimation(s) deriven from the set of updated endpoints, i.e. the set of endpoints for which
// an observation was received.
func (s *ServiceState) UpdateFromEndpoints(updatedEndpoints map[protocol.EndpointAddr]qos.Endpoint) error {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	for endpointAddr, updatedEndpoint := range updatedEndpoints {
		solanaEndpoint, ok := updatedEndpoint.(endpoint)
		if !ok {
			continue
		}

		if err := solanaEndpoint.Validate(""); err != nil {
			continue
		}

		// The endpoint's Epoch should be at-least equal to the perceived epoch before being used to update the perceived state of Solana blockchain.
		if solanaEndpoint.SolanaGetEpochInfoResponse.Epoch < s.perceivedEpoch {
			continue
		}

		// The endpoint's BlockHeight should be greater than the perceived block height before being used to update the perceived state of Solana blockchain.
		if solanaEndpoint.SolanaGetEpochInfoResponse.BlockHeight <= s.perceivedBlockHeight {
			continue
		}

		// TODO_IMPROVE: use a more resilient method for updating block height.
		// e.g. one endpoint returning a very large number as block height should
		// not result in all other endpoints being marked as invalid.
		s.perceivedEpoch = solanaEndpoint.SolanaGetEpochInfoResponse.Epoch
		s.perceivedBlockHeight = solanaEndpoint.SolanaGetEpochInfoResponse.BlockHeight

		s.Logger.With(
			"endpoint", endpointAddr,
			"block height", s.perceivedBlockHeight,
			"epoch", s.perceivedEpoch,
		).Info().Msg("Updating latest block height")
	}

	return nil
}
