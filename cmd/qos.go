package main

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/config"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/qos/cometbft"
	"github.com/buildwithgrove/path/qos/evm"
	"github.com/buildwithgrove/path/qos/solana"
)

// getServiceQoSInstances returns all QoS instances to be used by the Gateway and the EndpointHydrator.
func getServiceQoSInstances(
	logger polylog.Logger,
	gatewayConfig config.GatewayConfig,
	protocolInstance gateway.Protocol,
) (map[sdk.ServiceID]gateway.QoSService, error) {
	// TODO_TECHDEBT(@adshmh): refactor this function to remove the
	// need to manually add entries for every new QoS implementation.
	qosServices := make(map[sdk.ServiceID]gateway.QoSService)

	logger = logger.With("module", "qos")

	// Wait for the protocol to become healthy BEFORE configuring and starting the hydrator.
	// - Ensures the protocol instance's configured service IDs are available before hydrator startup.
	err := waitForProtocolHealth(logger, protocolInstance, defaultProtocolHealthTimeout)
	if err != nil {
		return nil, err
	}

	// Get configured service IDs from the protocol instance.
	//   - Used to run hydrator checks on all configured service IDs (except those manually disabled by the user).
	configuredServiceIDs := protocolInstance.GetConfiguredServiceIDs()

	// Remove any service IDs that are manually disabled by the user.
	for _, disabledServiceID := range gatewayConfig.HydratorConfig.QoSDisabledServiceIDs {
		// Throw error if any manually disabled service IDs are not found in the protocol's configured service IDs.
		if _, found := configuredServiceIDs[disabledServiceID]; !found {
			return nil, fmt.Errorf("invalid configuration: QoS manually disabled for service ID: %s, but not found in protocol's configured service IDs", disabledServiceID)
		}
		logger.Info().Msgf("QoS manually disabled for service ID: %s", disabledServiceID)
		delete(configuredServiceIDs, disabledServiceID)
	}

	// Get the service configs for the current protocol
	serviceConfigs := config.ServiceConfigs.GetServiceConfigs(gatewayConfig)

	// Initialize QoS services for all service IDs with a corresponding QoS
	// implementation, as defined in the `config/service_qos.go` file.
	for _, serviceConfig := range serviceConfigs {
		// Skip service IDs that are not configured for the PATH instance.
		if _, found := configuredServiceIDs[serviceConfig.GetServiceID()]; !found {
			continue
		}

		serviceID := serviceConfig.GetServiceID()

		switch serviceConfig.GetServiceQoSType() {
		case evm.QoSType:
			evmServiceQoSConfig, ok := serviceConfig.(evm.EVMServiceQoSConfig)
			if !ok {
				return nil, fmt.Errorf("SHOULD NEVER HAPPEN: error building QoS instances: service ID %q is not an EVM service", serviceID)
			}

			evmQoS := evm.NewQoSInstance(logger, evmServiceQoSConfig)
			qosServices[serviceID] = evmQoS

			logger.With("service_id", serviceID).Debug().Msg("Added EVM QoS instance for the service ID.")

		case cometbft.QoSType:
			cometBFTServiceQoSConfig, ok := serviceConfig.(cometbft.CometBFTServiceQoSConfig)
			if !ok {
				return nil, fmt.Errorf("SHOULD NEVER HAPPEN: error building QoS instances: service ID %q is not a CometBFT service", serviceID)
			}

			cometBFTQoS := cometbft.NewQoSInstance(logger, cometBFTServiceQoSConfig)
			qosServices[serviceID] = cometBFTQoS

		case solana.QoSType:
			solanaServiceQoSConfig, ok := serviceConfig.(solana.SolanaServiceQoSConfig)
			if !ok {
				return nil, fmt.Errorf("SHOULD NEVER HAPPEN: error building QoS instances: service ID %q is not a Solana service", serviceID)
			}

			solanaQoS := solana.NewQoSInstance(logger, solanaServiceQoSConfig)
			qosServices[serviceID] = solanaQoS

			logger.With("service_id", serviceID).Debug().Msg("Added Solana QoS instance for the service ID.")
		default:
			return nil, fmt.Errorf("SHOULD NEVER HAPPEN: error building QoS instances: service ID %q not supported by PATH", serviceID)
		}
	}

	return qosServices, nil
}
