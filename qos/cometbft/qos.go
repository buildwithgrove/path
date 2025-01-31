package cometbft

import (
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// NewQoSInstance builds and returns an instance of the CometBFT QoS service.
func NewQoSInstance(logger polylog.Logger) *QoS {
	logger = logger.With("qos_instance", "cometbft")

	serviceState := &ServiceState{
		logger: logger,
	}

	cometBFTEndpointStore := &EndpointStore{
		logger:       logger,
		ServiceState: serviceState,
	}

	return &QoS{
		logger:        logger,
		ServiceState:  serviceState,
		EndpointStore: cometBFTEndpointStore,
	}
}
