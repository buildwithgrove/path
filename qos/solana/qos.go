package solana

import (
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// NewQoSInstance builds and returns an instance of the Solana QoS service.
func NewQoSInstance(logger polylog.Logger, serviceConfig ServiceConfig) *QoS {
	logger = logger.With("qos_instance", "solana")

	logger.Warn().Msgf("TODO_MVP(@commoddity): Use the Solana ServiceConfig for QoS")

	serviceState := &ServiceState{
		logger: logger,
	}

	solanaEndpointStore := &EndpointStore{
		logger:       logger,
		serviceState: serviceState,
	}

	return &QoS{
		logger:        logger,
		ServiceState:  serviceState,
		EndpointStore: solanaEndpointStore,
	}
}
