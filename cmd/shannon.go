package main

import (
	"fmt"

	shannonconfig "github.com/buildwithgrove/path/config/shannon"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/protocol/shannon"
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// getShannonFullNode builds and returns a FullNode implementation for Shannon protocol integration, using the supplied configuration.
func getShannonFullNode(config *shannonconfig.ShannonGatewayConfig, logger polylog.Logger) (shannon.FullNode, error) {
	var fullNode shannon.FullNode

	// LazyFullNode skips all caching and queries the onchain data for serving each relay request.
	// It is utilized in the CachingFullNode, so must be initialized in all cases.
	lazyFullNode, err := shannon.NewLazyFullNode(config.FullNodeConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Shannon lazy full node: %v", err)
	}

	// Only use a CachingFullNode if the gateway mode is centralized and LazyMode is disabled.
	if config.GatewayConfig.GatewayMode == protocol.GatewayModeCentralized && !config.FullNodeConfig.LazyMode {
		cachingFullNode, err := shannon.NewCachingFullNode(lazyFullNode, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create Shannon caching full node: %v", err)
		}
		fullNode = cachingFullNode
	} else {
		// All other configurations use the lazy full node with no caching.
		fullNode = lazyFullNode
	}

	return fullNode, nil
}

// getShannonProtocol returns an instance of the Shannon protocol using the supplied Shannon-specific configuration.
func getShannonProtocol(config *shannonconfig.ShannonGatewayConfig, logger polylog.Logger) (gateway.Protocol, error) {
	logger.Info().Msg("Starting PATH gateway with Shannon protocol")

	fullNode, err := getShannonFullNode(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon full node instance: %v", err)
	}

	protocol, err := shannon.NewProtocol(fullNode, logger, config.GatewayConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon protocol instance: %v", err)
	}

	return protocol, nil
}
