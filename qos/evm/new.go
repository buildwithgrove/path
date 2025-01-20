package evm

import (
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// NewQoSInstance builds and returns an instance of the EVM QoS service.
func NewQoSInstance(logger polylog.Logger) *QoS {
	serviceState := &ServiceState{
		Logger: logger,

		// TODO_MVP(@adshmh): Read the chain ID from the configuration.
		ChainID: "0x1",
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
