package evm

// endpoint captures the details required to validate an EVM endpoint.
type endpoint struct {
	ChainID     string
	BlockHeight uint64
	// TODO_FUTURE: support archival endpoints.
}

func (e *endpoint) ApplyObservations(observations []observation) {
	for _, observation := range observations {
		if observation.ChainID != "" {
			e.ChainID = observation.ChainID
		}

		if observation.BlockHeight > 0 {
			e.BlockHeight = observation.BlockHeight
		}
	}

}

func (e endpoint) Validate(expectedChainID string) error {
	if e.ChainID != chainID {
		return fmt.Errorf("invalid chain ID: %s, expected: %s", e.ChainID, expectedChainID)
	}

	return nil
}
