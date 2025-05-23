package main

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/config"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/cometbft"
	"github.com/buildwithgrove/path/qos/evm"
	"github.com/buildwithgrove/path/qos/solana"
)

// getServiceQoSInstances returns all QoS instances to be used by the Gateway and the EndpointHydrator.
func getServiceQoSInstances(logger polylog.Logger, gatewayConfig config.GatewayConfig) (map[protocol.ServiceID]gateway.QoSService, error) {
	// TODO_TECHDEBT(@adshmh): refactor this function to remove the
	// need to manually add entries for every new QoS implementation.
	qosServices := make(map[protocol.ServiceID]gateway.QoSService)

	logger = logger.With("module", "qos")

	// Get the service configs for the current protocol
	serviceConfigs := config.ServiceConfigs.GetServiceConfigs(gatewayConfig)

	// Initialize QoS services for all service IDs with a corresponding QoS
	// implementation, as defined in the `config/service_qos.go` file.
	for _, serviceConfig := range serviceConfigs {
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
