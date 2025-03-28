package config

import "github.com/buildwithgrove/path/protocol"

// serviceQoSType maps to a gateway.QoSService implementation that builds request QoS context
// and selects endpoints for a given service ID.
type serviceQoSType string

/* ---------- EVM Service Config ---------- */

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

func (c CometBFTServiceConfig) GetArchivalThreshold() int {
	return defaultCometBFTArchivalThreshold
}
