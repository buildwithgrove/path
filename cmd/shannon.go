package main

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"

	shannonconfig "github.com/buildwithgrove/path/config/shannon"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol/shannon"
)

// getShannonFullNode builds and returns a FullNode implementation for Shannon protocol integration, using the supplied configuration.
func getShannonFullNode(config shannon.FullNodeConfig, logger polylog.Logger) (shannon.FullNode, error) {
	// LazyFullNode skips all caching and queries the onchain data for serving each relay request.
	lazyFullNode, err := shannon.NewLazyFullNode(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Shannon lazy full node: %v", err)
	}

	if config.LazyMode {
		return lazyFullNode, nil
	}

	// Use a Caching FullNode implementation if LazyMode flag is not set.
	cachingFullNode, err := shannon.NewCachingFullNode(lazyFullNode, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Shannon caching full node: %v", err)
	}

	return cachingFullNode, nil
}

// getShannonProtocol returns an instance of the Shannon protocol using the supplied Shannon-specific configuration.
func getShannonProtocol(config *shannonconfig.ShannonGatewayConfig, logger polylog.Logger) (gateway.Protocol, error) {
	logger.Info().Msg("Starting PATH gateway with Shannon protocol")

	fullNode, err := getShannonFullNode(config.FullNodeConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon full node instance: %v", err)
	}

	protocol, err := shannon.NewProtocol(logger, fullNode, config.GatewayConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon protocol instance: %v", err)
	}

	return protocol, nil
}
