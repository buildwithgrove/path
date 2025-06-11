package main

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"

	shannonconfig "github.com/buildwithgrove/path/config/shannon"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/protocol/shannon"
)

// getShannonFullNode builds and returns a FullNode implementation for Shannon protocol integration, using the supplied configuration.
// It also returns the owned apps if the gateway mode is Centralized.
func getShannonFullNode(logger polylog.Logger, config *shannonconfig.ShannonGatewayConfig) (shannon.FullNode, map[protocol.ServiceID][]string, error) {
	fullNodeConfig := config.FullNodeConfig

	// TODO_MVP(@adshmh): rename the variables here once a more accurate name is selected for `LazyFullNode`
	// LazyFullNode skips all caching and queries the onchain data for serving each relay request.
	lazyFullNode, err := shannon.NewLazyFullNode(fullNodeConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Shannon lazy full node: %v", err)
	}

	if fullNodeConfig.LazyMode {
		return lazyFullNode, nil, nil
	}

	// Initialize the owned apps for the gateway mode. This value will be nil if gateway mode is not Centralized.
	var ownedApps map[protocol.ServiceID][]string
	if config.GatewayConfig.GatewayMode == protocol.GatewayModeCentralized {
		ownedApps, err = shannon.GetCentralizedModeOwnedApps(logger, config.GatewayConfig.OwnedAppsPrivateKeysHex, lazyFullNode)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get app addresses from config: %v", err)
		}
	}

	fullNode, err := shannon.NewCachingFullNode(logger, lazyFullNode, fullNodeConfig.CacheConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create a Shannon caching full node instance: %v", err)
	}

	return fullNode, ownedApps, nil
}

// getShannonProtocol returns an instance of the Shannon protocol using the supplied Shannon-specific configuration.
func getShannonProtocol(logger polylog.Logger, config *shannonconfig.ShannonGatewayConfig) (gateway.Protocol, error) {
	logger.Info().Msg("Starting PATH gateway with Shannon protocol")

	fullNode, ownedApps, err := getShannonFullNode(logger, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon full node instance: %v", err)
	}

	protocol, err := shannon.NewProtocol(logger, fullNode, config.GatewayConfig, ownedApps)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon protocol instance: %v", err)
	}

	return protocol, nil
}
