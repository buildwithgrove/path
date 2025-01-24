package cometbft

import (
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// NewQoSInstance builds and returns an instance of the CometBFT QoS service.
func NewQoSInstance(logger polylog.Logger) *QoS {
	serviceState := &ServiceState{
		Logger: logger,
	}

	cometbftEndpointStore := &EndpointStore{
		Logger:       logger,
		ServiceState: serviceState,
	}

	return &QoS{
		Logger:        logger,
		ServiceState:  serviceState,
		EndpointStore: cometbftEndpointStore,
	}
}
