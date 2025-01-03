package evm

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/qos/evm"
)

// BuildEVMQoSInstance builds and returns an instance of the EVM QoS service.
func BuildEVMQoSInstance(logger polylog.Logger) *QoS {
	serviceState := &ServiceState{
		// TODO_MVP(@adshmh): Read the chain ID from the configuration.
		ChainID: "0x1",
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
