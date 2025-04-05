package solana

import "github.com/buildwithgrove/path/protocol"

// QoSType is the QoS type for the Solana blockchain.
const QoSType = "solana"

// ServiceQoSConfig defines the base interface for service QoS configurations.
// This avoids circular dependency with the config package.
type ServiceQoSConfig interface {
	GetServiceID() protocol.ServiceID
	GetServiceQoSType() string
}

// SolanaServiceQoSConfig is the configuration for the Solana service QoS.
type SolanaServiceQoSConfig interface {
	ServiceQoSConfig // Using locally defined interface to avoid circular dependency
}

// NewSolanaServiceQoSConfig creates a new Solana service configuration.
func NewSolanaServiceQoSConfig(
	serviceID protocol.ServiceID,
) SolanaServiceQoSConfig {
	return solanaServiceQoSConfig{
		serviceID: serviceID,
	}
}

// Ensure implementation satisfies interface
var _ SolanaServiceQoSConfig = (*solanaServiceQoSConfig)(nil)

type solanaServiceQoSConfig struct {
	serviceID protocol.ServiceID
}

// GetServiceID returns the ID of the service.
// Implements the ServiceQoSConfig interface.
func (c solanaServiceQoSConfig) GetServiceID() protocol.ServiceID {
	return c.serviceID
}

// GetServiceQoSType returns the QoS type of the service.
// Implements the ServiceQoSConfig interface.
func (_ solanaServiceQoSConfig) GetServiceQoSType() string {
	return QoSType
}
