package evm

import (
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// NewQoSInstance builds and returns an instance of the EVM QoS service.
func NewQoSInstance(logger polylog.Logger, evmChainID string) *QoS {
	serviceState := &ServiceState{
		ChainID: evmChainID,
		Logger:  logger,
	}

	evmEndpointStore := &EndpointStore{
		Logger: logger,

		ServiceState: serviceState,
	}

	return &QoS{
		Logger: logger,

		ServiceState:  serviceState,
		EndpointStore: evmEndpointStore,
	}
}
