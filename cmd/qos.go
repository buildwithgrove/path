package main

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/config"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/evm"
	"github.com/buildwithgrove/path/qos/noop"
)

// TODO_UPNEXT(@adshmh): enable Solana QoS instance through the following steps:
// 1. Add Solana alias + config to the configuration
// 2. Build a Solana QoS instance using any required configuration options.
// 3. Pass the Solana QoS instance to the endpoint hydrator, if enabled.
// 4. Pass the Solana QoS instance to the gateway.

// getServiceQoSInstances returns all QoS instances
// to be used by the Gateway and EndpointHydrator, respectively.
// This is done to ensure the same QoS instance is used in both
// Gateway and EndpointHydrator, if the Service QoS implements
// the gateway packages' QoSEndpointCheckGenerator interface.
func getServiceQoSInstances(
	gatewayConfig config.GatewayConfig,
	logger polylog.Logger,
) (
	map[protocol.ServiceID]gateway.QoSService,
	map[protocol.ServiceID]gateway.QoSEndpointCheckGenerator,
	error,
) {
	// Build a map of services configured for the hydrator to allow easy lookup.
	hydratorServiceIDsIdx := buildServiceIDsIdx(gatewayConfig.HydratorConfig.ServiceIDs)

	// TODO_TECHDEBT(@adshmh): refactor this function to remove the
	// need to manually add entries for every new QoS implementation.
	gatewayQoSService := make(map[protocol.ServiceID]gateway.QoSService)
	hydratorQoSGenerators := make(map[protocol.ServiceID]gateway.QoSEndpointCheckGenerator)

	// Initialize NoOp QoS service to handle services without a corresponding QoS implementation.
	gatewayQoSService[protocol.ServiceIDNoOp] = noop.NoOpQoS{}

	// Initialize QoS services for all service IDs with a corresponding QoS
	// implementation, as defined in the `config/service_qos.go` file.
	for serviceID, serviceQoSType := range config.ServiceQoSTypes {
		switch serviceQoSType {

		case config.ServiceIDEVM:
			evmEndpointStore := &evm.EndpointStore{
				Config: evm.EndpointStoreConfig{
					// TODO_MVP(@adshmh): Read the chain ID from the configuration.
					ChainID: "0x1",
				},
				Logger: logger,
			}
			gatewayQoSService[serviceID] = evm.NewServiceQoS(evmEndpointStore, logger)

			// If the service is configured for the hydrator in the configuration YAML file,
			// add its authoritative service ID to the hydrator's QoS generators map.
			if _, ok := hydratorServiceIDsIdx[serviceID]; ok {
				hydratorQoSGenerators[serviceID] = evmEndpointStore
			}

		// TODO_FUTURE(@adshmh): The logic here is complex enough to justify using a builder/factory function pattern.
		// At-least having something like func buildSolanaQoSInstance(...) in a solana.go file either here or under
		// config package will make the initialization/configuration code easier to read and maintain.
		case config.ServiceIDSolana:
			// TODO_TECHDEBT: add solana qos service here

		case config.ServiceIDPOKT:
			// TODO_TECHDEBT: add pokt qos service here

		default: // this should never happen
			return nil, nil, fmt.Errorf("error building QoS instances: service ID %q not supported by PATH", serviceID)
		}
	}

	return gatewayQoSService, hydratorQoSGenerators, nil
}

// buildServiceIDsIdx builds a map of the provided service IDs to allow one-line lookups.
func buildServiceIDsIdx(ids []protocol.ServiceID) map[protocol.ServiceID]struct{} {
	idx := make(map[protocol.ServiceID]struct{})
	for _, id := range ids {
		idx[id] = struct{}{}
	}

	return idx
}
