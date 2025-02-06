package evm

import (
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// NewQoSInstance builds and returns an instance of the EVM QoS service.
func NewQoSInstance(logger polylog.Logger, evmChainID string) *QoS {
	logger = logger.With("qos_instance", "evm").With("evm_chain_id", evmChainID)

	serviceState := &ServiceState{
		chainID: evmChainID,
		logger:  logger,
	}

	evmEndpointStore := &EndpointStore{
		logger: logger,

		ServiceState: serviceState,
	}

	return &QoS{
		logger: logger,

		ServiceState:  serviceState,
		EndpointStore: evmEndpointStore,
	}
}
