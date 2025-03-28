package evm

import "github.com/buildwithgrove/path/protocol"

// QoSTypeEVM is the QoS type for the EVM blockchain.
const QoSTypeEVM = "evm"

// TODO_TECHDEBT(@commoddity): this should be configurable.
const defaultEVMArchivalThreshold = 128

type ServiceConfig struct {
	// serviceID returns the ID of the service.
	ServiceID protocol.ServiceID
	// evmChainID returns the chain ID.
	EVMChainID string
	// ArchivalCheckConfig is the configuration for the archival check.
	ArchivalCheckConfig EVMArchivalCheckConfig
}

type EVMArchivalCheckConfig struct {
	Enabled            bool
	Threshold          uint64
	ContractAddress    string
	ContractStartBlock uint64

	archivalBlockNumber    string
	archivalBalance        string
	parsedBalanceConsensus map[string]int
}

func (c ServiceConfig) GetServiceID() protocol.ServiceID {
	return c.ServiceID
}

func (c ServiceConfig) GetServiceQoSType() string {
	return QoSTypeEVM
}

func (c ServiceConfig) GetEVMChainID() string {
	return c.EVMChainID
}

func (c ServiceConfig) GetEVMArchivalCheckConfig() EVMArchivalCheckConfig {
	if c.ArchivalCheckConfig.Threshold == 0 {
		c.ArchivalCheckConfig.Threshold = defaultEVMArchivalThreshold
	}
	return c.ArchivalCheckConfig
}
