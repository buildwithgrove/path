package evm

import "github.com/buildwithgrove/path/protocol"

// QoSTypeEVM is the QoS type for the EVM blockchain.
const QoSType = "evm"

const defaultEVMArchivalThreshold = 128

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

type EVMArchivalCheckConfig struct {
	Enabled            bool
	Threshold          uint64
	ContractAddress    string
	ContractStartBlock uint64
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

func (c ServiceConfig) getArchivalCheckConfig() EVMArchivalCheckConfig {
	if c.ArchivalCheckConfig.Threshold == 0 {
		c.ArchivalCheckConfig.Threshold = defaultEVMArchivalThreshold
	}
	return c.ArchivalCheckConfig
}
