package evm

import "github.com/buildwithgrove/path/protocol"

// QoSType is the QoS type for the EVM blockchain.
const QoSType = "evm"

// 128 is the default archival threshold for EVM-based chains.
// This is an opinionated value that aligns with industry standard
// practices for defining what constitutes an archival block.
const defaultEVMArchivalThreshold = 128

// ServiceQoSConfig defines the base interface for service QoS configurations.
// This avoids circular dependency with the config package.
type ServiceQoSConfig interface {
	GetServiceID() protocol.ServiceID
	GetServiceQoSType() string
}

// EVMServiceQoSConfig is the configuration for the EVM service QoS.
type EVMServiceQoSConfig interface {
	ServiceQoSConfig // Using locally defined interface to avoid circular dependency
	GetEVMChainID() string

	GetEVMArchivalCheckConfig() evmArchivalCheckConfig
	ArchivalCheckEnabled() bool
}

// evmArchivalCheckConfig is the configuration for the archival check.
//
// The basic methodology is:
//  1. Select a `ContractAddress` for the chain with a frequent transaction volume and large balance.
//  2. Determine its starting block height (`ContractStartBlock`).
//  3. Set a `Threshold` for how many blocks below the current block number are considered "archival" data.
//
// With all of this data, the QoS implementation can select a random block number to check using `eth_getBalance`.
type evmArchivalCheckConfig struct {
	threshold          uint64 // The number of blocks below the current block number to be considered "archival" data
	contractAddress    string // The address of the contract to check for the archival balance.
	contractStartBlock uint64 // The start block of the contract address (ie. when it first had a balance)
}

func (c evmArchivalCheckConfig) IsEmpty() bool {
	return c.contractAddress == "" || c.contractStartBlock == 0 || c.threshold == 0
}

// NewEVMServiceQoSConfig creates a new EVM service configuration with the specified archival check settings.
func NewEVMServiceQoSConfig(
	serviceID protocol.ServiceID,
	evmChainID string,
	archivalCheckConfig *evmArchivalCheckConfig,
) EVMServiceQoSConfig {
	return evmServiceQoSConfig{
		serviceID:           serviceID,
		evmChainID:          evmChainID,
		archivalCheckConfig: archivalCheckConfig,
	}
}

func NewEVMArchivalCheckConfig(
	contractAddress string,
	contractStartBlock uint64,
) *evmArchivalCheckConfig {
	return &evmArchivalCheckConfig{
		threshold:          defaultEVMArchivalThreshold,
		contractAddress:    contractAddress,
		contractStartBlock: contractStartBlock,
	}
}

// Ensure implementation satisfies interface
var _ EVMServiceQoSConfig = (*evmServiceQoSConfig)(nil)

type evmServiceQoSConfig struct {
	serviceID           protocol.ServiceID
	evmChainID          string
	archivalCheckConfig *evmArchivalCheckConfig
}

// GetServiceID returns the ID of the service.
// Implements the ServiceQoSConfig interface.
func (c evmServiceQoSConfig) GetServiceID() protocol.ServiceID {
	return c.serviceID
}

// GetServiceQoSType returns the QoS type of the service.
// Implements the ServiceQoSConfig interface.
func (_ evmServiceQoSConfig) GetServiceQoSType() string {
	return QoSType
}

// GetEVMChainID returns the chain ID.
// Implements the EVMServiceQoSConfig interface.
func (c evmServiceQoSConfig) GetEVMChainID() string {
	return c.evmChainID
}

// ArchivalCheckEnabled returns true if the archival check is enabled.
// If the archival check is not enabled for the service, this will always return false.
func (c evmServiceQoSConfig) ArchivalCheckEnabled() bool {
	return c.archivalCheckConfig != nil
}

// GetEVMArchivalCheckConfig returns the archival check configuration.
// Implements the EVMServiceQoSConfig interface.
func (c evmServiceQoSConfig) GetEVMArchivalCheckConfig() evmArchivalCheckConfig {
	return *c.archivalCheckConfig
}
