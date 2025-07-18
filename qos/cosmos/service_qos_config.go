package cosmos

import (
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"

	"github.com/buildwithgrove/path/protocol"
)

// QoSType is the QoS type for the CosmosSDK blockchain.
const QoSType = "cosmossdk"

// defaultCosmosSDKBlockNumberSyncAllowance is the default sync allowance for CosmosSDK-based chains.
// This number indicates how many blocks behind the perceived
// block number the endpoint may be and still be considered valid.
const defaultCosmosSDKBlockNumberSyncAllowance = 5

// ServiceQoSConfig defines the base interface for service QoS configurations.
// This avoids circular dependency with the config package.
type ServiceQoSConfig interface {
	GetServiceID() protocol.ServiceID
	GetServiceQoSType() string
}

// CosmosSDKServiceQoSConfig is the configuration for the CosmosSDK service QoS.
type CosmosSDKServiceQoSConfig interface {
	ServiceQoSConfig // Using locally defined interface to avoid circular dependency
	getCosmosSDKChainID() string
	getSyncAllowance() uint64
	getRPCTypes() map[sharedtypes.RPCType]struct{}
}

// NewCosmosSDKServiceQoSConfig creates a new CosmosSDK service configuration.
func NewCosmosSDKServiceQoSConfig(
	serviceID protocol.ServiceID,
	cosmosSDKChainID string,
	rpcTypes map[sharedtypes.RPCType]struct{},
) CosmosSDKServiceQoSConfig {
	return cosmosSDKServiceQoSConfig{
		serviceID:        serviceID,
		cosmosSDKChainID: cosmosSDKChainID,
		rpcTypes:         rpcTypes,
	}
}

// Ensure implementation satisfies interface
var _ CosmosSDKServiceQoSConfig = (*cosmosSDKServiceQoSConfig)(nil)

type cosmosSDKServiceQoSConfig struct {
	serviceID        protocol.ServiceID
	cosmosSDKChainID string
	syncAllowance    uint64
	rpcTypes         map[sharedtypes.RPCType]struct{}
}

// GetServiceID returns the ID of the service.
// Implements the ServiceQoSConfig interface.
func (c cosmosSDKServiceQoSConfig) GetServiceID() protocol.ServiceID {
	return c.serviceID
}

// GetServiceQoSType returns the QoS type of the service.
// Implements the ServiceQoSConfig interface.
func (cosmosSDKServiceQoSConfig) GetServiceQoSType() string {
	return QoSType
}

// getCosmosSDKChainID returns the chain ID.
// Implements the CosmosSDKServiceQoSConfig interface.
func (c cosmosSDKServiceQoSConfig) getCosmosSDKChainID() string {
	return c.cosmosSDKChainID
}

// getSyncAllowance returns the amount of blocks behind the perceived
// block number the endpoint may be and still be considered valid.
func (c cosmosSDKServiceQoSConfig) getSyncAllowance() uint64 {
	if c.syncAllowance == 0 {
		c.syncAllowance = defaultCosmosSDKBlockNumberSyncAllowance
	}
	return c.syncAllowance
}

// getRPCTypes returns the RPC types supported by the service.
// For example, XRPLEVM supports the following RPC types:
//   - JSON_RPC
//   - REST
//   - COMET_BFT
//   - WEBSOCKET (does not currently have a QoS quality check system in PATH)
//
// This is used to determine the appropriate QoS endpoint checks to run
func (c cosmosSDKServiceQoSConfig) getRPCTypes() map[sharedtypes.RPCType]struct{} {
	return c.rpcTypes
}
