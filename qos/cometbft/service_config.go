package cometbft

import "github.com/buildwithgrove/path/protocol"

// QoSType is the QoS type for the CometBFT blockchain.
const QoSType = "cometbft"

// 128 is the default archival threshold for the CometBFT blockchain.
// This is an opinionated value that aligns with industry standard
// practices for defining what constitutes an archival block.
const defaultCometBFTArchivalThreshold = 128

type ServiceConfig struct {
	ServiceID       protocol.ServiceID
	CometBFTChainID string
}

func (c ServiceConfig) GetServiceID() protocol.ServiceID {
	return c.ServiceID
}

func (c ServiceConfig) GetServiceQoSType() string {
	return QoSType
}

func (c ServiceConfig) GetCometBFTChainID() string {
	return c.CometBFTChainID
}

func (c ServiceConfig) GetArchivalThreshold() int {
	return defaultCometBFTArchivalThreshold
}
