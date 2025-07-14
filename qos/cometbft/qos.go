package cometbft

import (
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// NewQoSInstance builds and returns an instance of the CometBFT QoS service.
func NewQoSInstance(logger polylog.Logger, config CometBFTServiceQoSConfig) *QoS {
	cometBFTChainID := config.GetCometBFTChainID()
	serviceID := config.GetServiceID()

	logger = logger.With(
		"qos_instance", "cometbft",
		"service_id", serviceID,
		"cometbft_chain_id", cometBFTChainID,
	)

	// TODO_MVP(@commoddity): add archival check configuration for CometBFT.
	serviceState := &ServiceState{
		logger:    logger,
		chainID:   cometBFTChainID,
		serviceID: serviceID,
	}

	cometBFTEndpointStore := &EndpointStore{
		logger:       logger,
		serviceState: serviceState,
	}

	// TODO_TECHDEBT(@adshmh): Add a Request Validator for CometBFT services.
	// Use evm or solana package as template.
	return &QoS{
		logger:        logger,
		ServiceState:  serviceState,
		EndpointStore: cometBFTEndpointStore,
	}
}
