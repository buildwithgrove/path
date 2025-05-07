package judge

// StateParameter stores related values for a single QoS service component
// Example: archival state:
// - contract address
// - count of each reported balance
//
// Supported value types (PR #210):
// - string: e.g., contract address for archival checks
// - integer: e.g., block number from blockchain service
// - consensus: e.g., balance reports as map[string]int{"0x1234": 5}
//
// Examples:
// - BlockNumber: IntValues{"blockNumber": 12345}
// - ArchivalState:
//   - StringValues{"contractAddress": "0xADDR"}
//   - ConsensusValues{"0x12345": 5, "0x5678": 8}
type StateParameter struct {
	// stores string type state value
	strValue *string

	// stores integer type state value
	intValue *int

	// stores consensus type state value
	consensusValues map[string]int
}

func (sp *StateParameter) GetStr() (string, bool) {
	if sp.strValue == nil {
		return "", false
	}

	return *sp.strValue, true
}

func (sp *StateParameter) GetInt() (int, bool) {
	if sp.intValue == nil {
		return 0, false
	}

	return *sp.intValue, true
}

func (sp *StateParameter) GetConsensus() (map[string]int, bool) {
	if sp.consensusValues == nil {
		return nil, false
	}

	consensusCopy := make(map[string]int)
	for k, v := range sp.consensusValues {
		consensusCopy[k] = v
	}

	return consensusCopy, true
}
