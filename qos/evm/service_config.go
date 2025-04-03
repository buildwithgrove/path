package evm

import "github.com/buildwithgrove/path/protocol"

// QoSTypeEVM is the QoS type for the EVM blockchain.
const QoSType = "evm"

// 128 is the default archival threshold for EVM-based chains.
// This is an opinionated value that aligns with industry standard
// practices for defining what constitutes an archival block.
const defaultEVMArchivalThreshold = 128

type ServiceConfig struct {
	// serviceID returns the ID of the service.
	ServiceID protocol.ServiceID
	// evmChainID returns the chain ID.
	EVMChainID string
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

func (c ServiceConfig) GetEVMChainID() string {
	return c.EVMChainID
}

func (c ServiceConfig) GetEVMArchivalCheckConfig() (EVMArchivalCheckConfig, bool) {
	if !c.ArchivalCheckConfig.Enabled {
		return EVMArchivalCheckConfig{}, false
	}

	if c.ArchivalCheckConfig.Threshold == 0 {
		c.ArchivalCheckConfig.Threshold = defaultEVMArchivalThreshold
	}
	return c.ArchivalCheckConfig, true
}
