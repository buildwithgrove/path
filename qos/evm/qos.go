package evm

import (
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// NewQoSInstance builds and returns an instance of the EVM QoS service.
func NewQoSInstance(logger polylog.Logger, config EVMServiceQoSConfig) *QoS {
	evmChainID := config.getEVMChainID()

	logger = logger.With(
		"qos_instance", "evm",
		"evm_chain_id", evmChainID,
	)

	serviceState := &serviceState{
		logger:        logger,
		serviceConfig: config,
	}

	// TODO_CONSIDERATION(@olshansk): Archival checks are currently optional to enable iteration and optionality.
	// In the future, evaluate whether it should be mandatory for all EVM services.
	if config.archivalCheckEnabled() {
		serviceState.archivalState = archivalState{
			logger:              logger.With("state", "archival"),
			archivalCheckConfig: config.getEVMArchivalCheckConfig(),
			// Initialize the balance consensus map.
			// It keeps track and maps a balance (at the configured address and contract)
			// to the number of occurrences seen across all endpoints.
			balanceConsensus: make(map[string]int),
		}
	}

	evmEndpointStore := &endpointStore{
		logger:       logger,
		serviceState: serviceState,
	}

	evmRequestValidator := &evmRequestValidator{
		logger:        logger,
		chainID:       evmChainID,
		endpointStore: evmEndpointStore,
	}

	return &QoS{
		logger:              logger,
		serviceState:        serviceState,
		endpointStore:       evmEndpointStore,
		evmRequestValidator: evmRequestValidator,
	}
}
