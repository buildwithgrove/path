package main

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"
	sdk "github.com/pokt-network/shannon-sdk"
	"github.com/pokt-network/shannon-sdk/fullnode"

	shannonconfig "github.com/buildwithgrove/path/config/shannon"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol/shannon"
)

// getShannonFullNode builds and returns a FullNode implementation for Shannon protocol integration.
func getShannonFullNode(logger polylog.Logger, config *shannonconfig.ShannonGatewayConfig) (sdk.FullNode, error) {
	fullNodeConfig := config.FullNodeConfig

	// In both lazy and caching modes, we use the full node to fetch the onchain data.
	fullNode, err := fullnode.NewFullNode(fullNodeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Shannon full node: %v", err)
	}

	// If PATH is in lazy mode, return the full node directly.
	if fullNodeConfig.LazyMode {
		return fullNode, nil
	}

	// If PATH is in caching mode, return the full node with cache.
	return fullnode.NewFullNodeWithCache(logger, fullNode, fullNodeConfig.CacheConfig)
}

// getShannonProtocol returns an instance of the Shannon protocol using the supplied Shannon-specific configuration.
func getShannonProtocol(logger polylog.Logger, config *shannonconfig.ShannonGatewayConfig) (gateway.Protocol, error) {
	logger.Info().Msg("Starting PATH gateway with Shannon protocol")

	fullNode, err := getShannonFullNode(logger, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon full node instance: %v", err)
	}

	protocol, err := shannon.NewProtocol(logger, fullNode, config.GatewayConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon protocol instance: %v", err)
	}

	return protocol, nil
}
