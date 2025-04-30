package main

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"

	shannonconfig "github.com/buildwithgrove/path/config/shannon"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol/shannon"
)

// getShannonFullNode builds and returns a FullNode implementation for Shannon protocol integration, using the supplied configuration.
func getShannonFullNode(config shannon.FullNodeConfig) (shannon.FullNode, error) {
	// TODO_MVP(@adshmh): rename the variables here once a more accurate name is selected for `LazyFullNode`
	// LazyFullNode skips all caching and queries the onchain data for serving each relay request.
	lazyFullNode, err := shannon.NewLazyFullNode(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Shannon lazy full node: %v", err)
	}

	if config.LazyMode {
		return lazyFullNode, nil
	}

	return shannon.NewCachingFullNode(lazyFullNode), nil
}

// getShannonProtocol returns an instance of the Shannon protocol using the supplied Shannon-specific configuration.
func getShannonProtocol(logger polylog.Logger, config *shannonconfig.ShannonGatewayConfig) (gateway.Protocol, error) {
	logger.Info().Msg("Starting PATH gateway with Shannon protocol")

	fullNode, err := getShannonFullNode(config.FullNodeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon full node instance: %v", err)
	}

	protocol, err := shannon.NewProtocol(logger, fullNode, config.GatewayConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon protocol instance: %v", err)
	}

	return protocol, nil
}
