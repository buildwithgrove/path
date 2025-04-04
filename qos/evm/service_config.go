package evm

import "github.com/buildwithgrove/path/protocol"

// QoSTypeEVM is the QoS type for the EVM blockchain.
const QoSType = "evm"

// DefaultEVMArchivalThreshold is the default archival threshold for EVM-based chains.
// This is an opinionated value that aligns with industry standard
// practices for defining what constitutes an archival block.
const DefaultEVMArchivalThreshold = 128

// defaultEVMBlockNumberSyncAllowance is the default sync allowance for EVM-based chains.
// This number indicates how many blocks behind the perceived
// block number the endpoint may be and still be considered valid.
const defaultEVMBlockNumberSyncAllowance = 5

type ServiceConfig struct {
	// serviceID returns the ID of the service.
	ServiceID protocol.ServiceID
	// EVMChainID is the expected value of the `Result` field in any endpoint's response to an `eth_chainId` request.
	// See the following link for more details: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_chainid
	//
	// Chain IDs Reference: https://chainlist.org/
	EVMChainID string
	// SyncAllowance is the number of blocks that the perceived block number is allowed to be behind the actual block number.
	SyncAllowance uint64
	// ArchivalCheckConfig is the configuration for the archival check.
	ArchivalCheckConfig EVMArchivalCheckConfig
}

// EVMArchivalCheckConfig is the configuration for the archival check.
//
// The basic methodology is:
//  1. Select a `ContractAddress` for the chain with a frequent transaction volume and large balance.
//  2. Determine its starting block height (`ContractStartBlock`).
//  3. Set a `Threshold` for how many blocks below the current block number are considered "archival" data.
//
// With all of this data, the QoS implementation can select a random block number to check using `eth_getBalance`.
type EVMArchivalCheckConfig struct {
	Enabled            bool   // Whether to require an archival check for the service.
	Threshold          uint64 // The number of blocks below the current block number to be considered "archival" data
	ContractAddress    string // The address of the contract to check for the archival balance.
	ContractStartBlock uint64 // The start block of the contract address (ie. when it first had a balance)
}

func (c ServiceConfig) GetServiceID() protocol.ServiceID {
	return c.ServiceID
}

func (c ServiceConfig) GetServiceQoSType() string {
	return QoSType
}

func (c ServiceConfig) getEVMChainID() string {
	return c.EVMChainID
}

func (c ServiceConfig) getSyncAllowance() uint64 {
	if c.SyncAllowance == 0 {
		c.SyncAllowance = defaultEVMBlockNumberSyncAllowance
	}
	return c.SyncAllowance
}

func (c ServiceConfig) getEVMArchivalCheckConfig() (EVMArchivalCheckConfig, bool) {
	if !c.ArchivalCheckConfig.Enabled {
		return EVMArchivalCheckConfig{}, false
	}

	if c.ArchivalCheckConfig.Threshold == 0 {
		c.ArchivalCheckConfig.Threshold = DefaultEVMArchivalThreshold
	}
	return c.ArchivalCheckConfig, true
}
