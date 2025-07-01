package main

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"

	shannonconfig "github.com/buildwithgrove/path/config/shannon"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol/shannon"
)

// getShannonFullNode builds and returns a Shannon FullNode configuration.
// If the configuration provided is "lazy", it short circuits to a lazy node and bypasses caching.
func getShannonFullNode(logger polylog.Logger, config *shannonconfig.ShannonGatewayConfig) (shannon.FullNode, error) {
	fullNodeConfig := config.FullNodeConfig

	// TODO_MVP(@adshmh): rename the variables here once a more accurate name is selected for `LazyFullNode`
	// LazyFullNode skips all caching and queries the onchain data for serving each relay request.
	lazyFullNode, err := shannon.NewLazyFullNode(fullNodeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Shannon lazy full node: %w", err)
	}

	// Bypass caching if the configuration is "lazy".
	if fullNodeConfig.CacheConfig.UseCache != nil && !*fullNodeConfig.CacheConfig.UseCache {
		return lazyFullNode, nil
	}

	fullNode, err := shannon.NewCachingFullNode(logger, lazyFullNode)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon caching full node instance: %w", err)
	}

	return fullNode, nil
}

// getShannonProtocol returns an instance of the Shannon protocol using the supplied Shannon-specific configuration.
func getShannonProtocol(logger polylog.Logger, config *shannonconfig.ShannonGatewayConfig) (gateway.Protocol, error) {
	logger.Info().Msg("Starting PATH gateway with Shannon protocol")

	fullNode, err := getShannonFullNode(logger, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon full node instance: %w", err)
	}

	protocol, err := shannon.NewProtocol(logger, config.GatewayConfig, fullNode)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon protocol instance: %w", err)
	}

	return protocol, nil
}
