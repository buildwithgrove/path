package evm

import (
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// TODO_IMPROVE: make this configurable per-chain.
const defaultSyncAllowance = 10

// NewQoSInstance builds and returns an instance of the EVM QoS service.
func NewQoSInstance(logger polylog.Logger, evmChainID string) *QoS {
	logger = logger.With(
		"qos_instance", "evm",
		"evm_chain_id", evmChainID,
	)

	serviceState := &ServiceState{
		logger: logger,
		config: serviceStateConfig{
			chainID:       evmChainID,
			syncAllowance: defaultSyncAllowance,
		},
	}

	evmEndpointStore := &EndpointStore{
		logger:       logger,
		serviceState: serviceState,
	}

	return &QoS{
		logger:        logger,
		ServiceState:  serviceState,
		EndpointStore: evmEndpointStore,
	}
}
