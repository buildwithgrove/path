package evm

import (
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// NewQoSInstance builds and returns an instance of the EVM QoS service.
func NewQoSInstance(logger polylog.Logger, evmChainID string) *QoS {
	logger = logger.With(
		"qos_instance", "evm",
		"evm_chain_id", evmChainID,
	)

	serviceState := &ServiceState{
		logger:  logger,
		chainID: evmChainID,
	}

	evmEndpointStore := &EndpointStore{
		logger:       logger,
		serviceState: serviceState,
	}

	evmRequestValidator := &evmRequestValidator{
		logger:        logger,
		chainID:       evmChainID,
		endpointStore: evmEndpointStore,
	}

	return &QoS{
		logger:           logger,
		ServiceState:     serviceState,
		EndpointStore:    evmEndpointStore,
		evmRequestValidator: evmRequestValidator,
	}
}
