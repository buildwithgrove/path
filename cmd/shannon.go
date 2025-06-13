package main

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/shannon-sdk/fullnode"

	shannonconfig "github.com/buildwithgrove/path/config/shannon"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol/shannon"
)

// TODO_NEXT(@commoddity): Replace getShannonFullNode with getShannonGatewayClient method
// once the GatewayClient is implemented in the SDK.
//
// getShannonFullNode builds and returns a FullNode implementation for Shannon protocol integration, using the supplied configuration.
// It also returns the owned apps if the gateway mode is Centralized.
func getShannonFullNode(logger polylog.Logger, config *shannonconfig.ShannonGatewayConfig) (shannon.FullNode, error) {
	fullNodeConfig := config.FullNodeConfig

	// TODO_MVP(@adshmh): rename the variables here once a more accurate name is selected for `LazyFullNode`
	// LazyFullNode skips all caching and queries the onchain data for serving each relay request.
	lazyFullNode, err := fullnode.NewLazyFullNode(fullNodeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Shannon lazy full node: %v", err)
	}

	// Bypass caching if the configuration is "lazy".
	if fullNodeConfig.LazyMode {
		return lazyFullNode, nil
	}

	fullNode, err := fullnode.NewCachingFullNode(logger, lazyFullNode, fullNodeConfig.CacheConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon caching full node instance: %v", err)
	}

	return fullNode, nil
}

// getShannonProtocol returns an instance of the Shannon protocol using the supplied Shannon-specific configuration.
func getShannonProtocol(logger polylog.Logger, config *shannonconfig.ShannonGatewayConfig) (gateway.Protocol, error) {
	logger.Info().Msg("Starting PATH gateway with Shannon protocol")

	fullNode, err := getShannonFullNode(logger, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon full node instance: %v", err)
	}

	protocol, err := shannon.NewProtocol(logger, config.GatewayConfig, fullNode)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon protocol instance: %v", err)
	}

	return protocol, nil
}
