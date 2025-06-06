package main

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"

	shannonconfig "github.com/buildwithgrove/path/config/shannon"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol/shannon"
)

// getShannonFullNode builds and returns a FullNode implementation for Shannon protocol integration, using the supplied configuration.
func getShannonFullNode(logger polylog.Logger, config *shannonconfig.ShannonGatewayConfig) (shannon.FullNode, []shannon.OwnedApp, error) {
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

	ownedApps, err := shannon.GetCentralizedModeOwnedApps(logger, config.GatewayConfig.OwnedAppsPrivateKeysHex, lazyFullNode)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get app addresses from config: %v", err)
	}

	return shannon.NewCachingFullNode(logger, lazyFullNode, ownedApps, fullNodeConfig.CacheConfig), ownedApps, nil
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
