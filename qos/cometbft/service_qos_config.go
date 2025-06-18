package cometbft

import sdk "github.com/pokt-network/shannon-sdk"

// QoSType is the QoS type for the CometBFT blockchain.
const QoSType = "cometbft"

// 128 is the default archival threshold for the CometBFT blockchain.
// This is an opinionated value that aligns with industry standard
// practices for defining what constitutes an archival block.
// TODO_IMPROVE(@commoddity): Implement proper archival check configuration for CometBFT.
const defaultCometBFTArchivalThreshold = 128

// ServiceQoSConfig defines the base interface for service QoS configurations.
// This avoids circular dependency with the config package.
type ServiceQoSConfig interface {
	GetServiceID() sdk.ServiceID
	GetServiceQoSType() string
}

type CometBFTServiceQoSConfig interface {
	ServiceQoSConfig
	GetCometBFTChainID() string
	GetArchivalThreshold() int
}

// NewCometBFTServiceQoSConfig creates a new EVM service configuration with the specified archival check settings.
func NewCometBFTServiceQoSConfig(
	serviceID sdk.ServiceID,
	cometBFTChainID string,
) CometBFTServiceQoSConfig {
	return cometBFTServiceQoSConfig{
		serviceID:       serviceID,
		cometBFTChainID: cometBFTChainID,
	}
}

var _ CometBFTServiceQoSConfig = (*cometBFTServiceQoSConfig)(nil)

type cometBFTServiceQoSConfig struct {
	serviceID       sdk.ServiceID
	cometBFTChainID string
}

// GetServiceID returns the ID of the service.
// Implements the config.ServiceQoSConfig interface.
func (c cometBFTServiceQoSConfig) GetServiceID() sdk.ServiceID {
	return c.serviceID
}

// GetServiceQoSType returns the QoS type of the service.
// Implements the config.ServiceQoSConfig interface.
func (_ cometBFTServiceQoSConfig) GetServiceQoSType() string {
	return QoSType
}

// GetCometBFTChainID returns the CometBFT chain ID.
// Implements the config.CometBFTServiceQoSConfig interface.
func (c cometBFTServiceQoSConfig) GetCometBFTChainID() string {
	return c.cometBFTChainID
}

// GetArchivalThreshold returns the archival threshold for the CometBFT service.
// Implements the config.CometBFTServiceQoSConfig interface.
func (_ cometBFTServiceQoSConfig) GetArchivalThreshold() int {
	return defaultCometBFTArchivalThreshold
}
