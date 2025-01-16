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
		ServiceState: serviceState,
		Logger:       logger,
	}

	return &QoS{
		ServiceState:  serviceState,
		EndpointStore: evmEndpointStore,
		Logger:        logger,
	}
}
