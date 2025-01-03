package main

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/config"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/evm"
)

// TODO_UPNEXT(@adshmh): enable Solana QoS instance through the following steps:
// 1. Add Solana alias + config to the configuration
// 2. Build a Solana QoS instance using any required configuration options.
// 3. Pass the Solana QoS instance to the endpoint hydrator, if enabled.
// 4. Pass the Solana QoS instance to the gateway.

// getServiceQoSInstances returns all QoS instances to be used by the Gateway and the EndpointHydrator.
func getServiceQoSInstances(
	gatewayConfig config.GatewayConfig,
	logger polylog.Logger,
) (
	map[protocol.ServiceID]gateway.QoSService,
	error,
) {
	// TODO_TECHDEBT(@adshmh): refactor this function to remove the
	// need to manually add entries for every new QoS implementation.
	qosServices := make(map[protocol.ServiceID]gateway.QoSService)

	// Initialize QoS services for all service IDs with a corresponding QoS
	// implementation, as defined in the `config/service_qos.go` file.
	for serviceID, serviceQoSType := range config.ServiceQoSTypes {
		switch serviceQoSType {

		case config.ServiceIDEVM:
			evmQoS := evm.BuildEVMQoSInstance(logger)
			qosServices[serviceID] = evmQoS

		// TODO_FUTURE(@adshmh): The logic here is complex enough to justify using a builder/factory function pattern.
		// At-least having something like func buildSolanaQoSInstance(...) in a solana.go file either here or under
		// config package will make the initialization/configuration code easier to read and maintain.
		case config.ServiceIDSolana:
			// TODO_TECHDEBT: add solana qos service here

		case config.ServiceIDPOKT:
			// TODO_TECHDEBT: add pokt qos service here

		default: // this should never happen
			return nil, fmt.Errorf("error building QoS instances: service ID %q not supported by PATH", serviceID)
		}
	}

	return qosServices, nil
}
