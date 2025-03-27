package config

import "github.com/buildwithgrove/path/protocol"

type ServiceConfig interface {
	GetServiceID() protocol.ServiceID
	GetServiceQoSType() serviceQoSType
}

// serviceQoSType maps to a gateway.QoSService implementation that builds request QoS context
// and selects endpoints for a given service ID.
type serviceQoSType string

/* ---------- EVM Service Config ---------- */

// ServiceIDEVM is the service ID for the EVM blockchain.
const ServiceIDEVM serviceQoSType = "evm"

const defaultEVMChainID = "0x1" // ETH Mainnet (1)

// TODO_TECHDEBT(@commoddity): this should be configurable.
const defaultEVMArchivalThreshold = 128

type EVMServiceConfig struct {
	// serviceID returns the ID of the service.
	serviceID protocol.ServiceID
	// evmChainID returns the chain ID.
	evmChainID string
	// archivalThreshold is the numer of blocks behind the current block to be considered an archival node.
	archivalThreshold uint64
}

func (c EVMServiceConfig) GetServiceID() protocol.ServiceID {
	return c.serviceID
}

func (c EVMServiceConfig) GetServiceQoSType() serviceQoSType {
	return ServiceIDEVM
}

func (c EVMServiceConfig) GetServiceChainID() string {
	return c.evmChainID
}

func (c EVMServiceConfig) GetArchivalThreshold() uint64 {
	if c.archivalThreshold == 0 {
		return defaultEVMArchivalThreshold
	}
	return c.archivalThreshold
}

/* ---------- Solana Service Config ---------- */

// ServiceIDSolana is the service ID for the Solana blockchain.
const ServiceIDSolana serviceQoSType = "solana"

type SolanaServiceConfig struct {
	serviceID protocol.ServiceID
}

func (c SolanaServiceConfig) GetServiceID() protocol.ServiceID {
	return c.serviceID
}

func (c SolanaServiceConfig) GetServiceQoSType() serviceQoSType {
	return ServiceIDSolana
}

/* ---------- CometBFT Service Config ---------- */

// ServiceIDCometBFT is the service ID for the CometBFT blockchain.
const ServiceIDCometBFT serviceQoSType = "cometbft"

const defaultCometBFTChainID = "cosmoshub-4"

// TODO_TECHDEBT(@commoddity): this should be configurable.
const defaultCometBFTArchivalThreshold = 128

type CometBFTServiceConfig struct {
	serviceID       protocol.ServiceID
	cometBFTChainID string
}

func (c CometBFTServiceConfig) GetServiceID() protocol.ServiceID {
	return c.serviceID
}

func (c CometBFTServiceConfig) GetServiceQoSType() serviceQoSType {
	return ServiceIDCometBFT
}

func (c CometBFTServiceConfig) GetServiceChainID() string {
	return c.cometBFTChainID
}

func (c CometBFTServiceConfig) GetArchivalThreshold() uint64 {
	return defaultCometBFTArchivalThreshold
}
