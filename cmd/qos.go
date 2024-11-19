package main

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/config"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/qos/evm"
	"github.com/buildwithgrove/path/relayer"
)

// TODO_UPNEXT(@adshmh): enable Solana QoS instance through the following steps:
// 1. Add Solana alias + config to the configuration
// 2. Build a Solana QoS instance using any required configuration options.
// 3. Pass the Solana QoS instance to the endpoint hydrator, if enabled.
// 4. Pass the Solana QoS instance to the gateway.
//
// getServiceQoSInstances returns all QoS instances
// to be used by the Gateway and EndpointHydrator, respectively.
// This is done to ensure the same QoS instance is used in both
// Gateway and EndpointHydrator, if the Service QoS implements
// the gateway packages' QoSEndpointCheckGenerator interface.
func getServiceQoSInstances(
	gatewayConfig config.GatewayConfig,
	logger polylog.Logger,
) (
	map[relayer.ServiceID]gateway.QoSService,
	map[relayer.ServiceID]gateway.QoSEndpointCheckGenerator,
	error,
) {
	// Build a map of services configured for the hydrator
	// and the gateway to allow easy lookup.
	hydratorServiceIDsIdx := buildServiceIDsIdx(gatewayConfig.HydratorConfig.ServiceIDs)
	gatewayServiceIDsIdx := buildServiceIDsIdx(gatewayConfig.GetEnabledServiceIDs())

	// TODO_TECHDEBT: refactor this function to remove
	// the need for manually adding entries for every new QoS implmenetation.
	gatewayQoSService := make(map[relayer.ServiceID]gateway.QoSService)
	hydratorQoSGenerators := make(map[relayer.ServiceID]gateway.QoSEndpointCheckGenerator)

	// TODO_FUTURE: support serviceQoS-specific configuration.
	allServiceIDs := append(gatewayConfig.GetEnabledServiceIDs(), gatewayConfig.HydratorConfig.ServiceIDs...)
	for _, serviceID := range allServiceIDs {
		switch serviceID {
		case config.ServiceIDEVM:
			evmEndpointStore := &evm.EndpointStore{
				Config: evm.EndpointStoreConfig{
					// TODO_MVP(@adshmh): Read the chain ID from the configuration.
					ChainID: "0x1",
				},
				Logger: logger,
			}
			if _, ok := gatewayServiceIDsIdx[serviceID]; ok {
				gatewayQoSService[serviceID] = evm.NewServiceQoS(evmEndpointStore, logger)
			}
			if _, ok := hydratorServiceIDsIdx[serviceID]; ok {
				hydratorQoSGenerators[serviceID] = evmEndpointStore
			}
		case config.ServiceIDSolana:
			// TODO_TECHDEBT: add solana qos service here
		case config.ServiceIDPOKT:
			// TODO_TECHDEBT: add pokt qos service here
		case config.ServiceIDE2E:
			evmEndpointStore := &evm.EndpointStore{
				Config: evm.EndpointStoreConfig{
					// TODO_MVP(@adshmh): Read the chain ID from the configuration.
					ChainID: "0x1",
				},
				Logger: logger,
			}
			if _, ok := gatewayServiceIDsIdx[serviceID]; ok {
				gatewayQoSService[serviceID] = evm.NewServiceQoS(evmEndpointStore, logger)
			}
		default:
			return nil, nil, fmt.Errorf("error building QoS instances: service ID %q not recognized", serviceID)
		}
	}

	return gatewayQoSService, hydratorQoSGenerators, nil
}

// buildServiceIDsIdx builds a map of the provided service IDs to allow one-line lookups.
func buildServiceIDsIdx(ids []relayer.ServiceID) map[relayer.ServiceID]struct{} {
	idx := make(map[relayer.ServiceID]struct{})
	for _, id := range ids {
		idx[id] = struct{}{}
	}

	return idx
}
