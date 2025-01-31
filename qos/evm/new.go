package evm

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/qos"
)

// NewQoSInstance builds and returns an instance of the EVM QoS service.
func NewQoSInstance(logger polylog.Logger, evmChainID string) *QoS {
	serviceState := &ServiceState{
		ChainID: evmChainID,
		Logger:  logger,
	}

	evmEndpointStore := &qos.EndpointStore{
		Logger:       logger,
		ServiceState: serviceState,
	}
	// Define the set of quality checks, performed by the Hydrator,
	// that must be satisfied for an endpoint to be considered valid.
	evmEndpointStore.RequiredQualityChecks = []gateway.RequestQoSContext{
		getEndpointCheck(evmEndpointStore, withChainIDCheck),
		getEndpointCheck(evmEndpointStore, withBlockHeightCheck),
	}

	return &QoS{
		Logger: logger,

		ServiceState:  serviceState,
		EndpointStore: evmEndpointStore,
	}
}
