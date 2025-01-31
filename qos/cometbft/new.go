package cometbft

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/qos"
)

// NewQoSInstance builds and returns an instance of the CometBFT QoS service.
func NewQoSInstance(logger polylog.Logger) *QoS {
	serviceState := &ServiceState{
		Logger: logger,
	}

	cometbftEndpointStore := &qos.EndpointStore{
		Logger:       logger,
		ServiceState: serviceState,
	}
	// Define the set of quality checks, performed by the Hydrator,
	// that must be satisfied for an endpoint to be considered valid.
	cometbftEndpointStore.RequiredQualityChecks = []gateway.RequestQoSContext{
		getEndpointCheck(cometbftEndpointStore, withHealthCheck),
		getEndpointCheck(cometbftEndpointStore, withBlockHeightCheck),
	}

	return &QoS{
		Logger:        logger,
		ServiceState:  serviceState,
		EndpointStore: cometbftEndpointStore,
	}
}
