package cometbft

import (
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// NewQoSInstance builds and returns an instance of the CometBFT QoS service.
func NewQoSInstance(logger polylog.Logger, config CometBFTServiceQoSConfig) *QoS {
	cometBFTChainID := config.GetCometBFTChainID()

	logger = logger.With(
		"qos_instance", "cometbft",
		"cometbft_chain_id", cometBFTChainID,
	)

	// TODO_MVP(@commoddity): add archival check configuration for CometBFT.
	serviceState := &ServiceState{
		logger:  logger,
		chainID: cometBFTChainID,
	}

	cometBFTEndpointStore := &EndpointStore{
		logger:       logger,
		serviceState: serviceState,
	}

	return &QoS{
		logger:        logger,
		ServiceState:  serviceState,
		EndpointStore: cometBFTEndpointStore,
	}
}
