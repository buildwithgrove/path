package evm

import (
	"fmt"
	"strconv"
)

// endpoint captures the details required to validate an EVM endpoint.
type endpoint struct {
	ChainID string
	// blockHeight is stored as a string
	// to allow validation of the endpoint's response.
	blockHeight string
	// TODO_FUTURE: support archival endpoints.
}

func (e *endpoint) Apply(observations []observation) {
	for _, observation := range observations {
		if observation.ChainID != "" {
			e.ChainID = observation.ChainID
		}

		if observation.BlockHeight != "" {
			e.blockHeight = observation.BlockHeight
		}
	}
}

func (e endpoint) Validate(expectedChainID string) error {
	if e.ChainID != expectedChainID {
		return fmt.Errorf("invalid chain ID: %s, expected: %s", e.ChainID, expectedChainID)
	}

	_, err := e.GetBlockHeight()
	return err
}

func (e endpoint) GetBlockHeight() (uint64, error) {
	// base 0: use the string's prefix to determine its base.
	height, err := strconv.ParseUint(e.blockHeight, 0, 64)
	if err != nil {
		return 0, fmt.Errorf("getBlockHeight: invalid block height value %q: %v", e.blockHeight, err)
	}

	return height, nil
}
