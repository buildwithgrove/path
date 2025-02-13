package cometbft

import (
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// NewQoSInstance builds and returns an instance of the CometBFT QoS service.
func NewQoSInstance(logger polylog.Logger, chainID string) *QoS {
	logger = logger.With("qos_instance", "cometbft")

	serviceState := &ServiceState{
		logger:  logger,
		chainID: chainID,
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
