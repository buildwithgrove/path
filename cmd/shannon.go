package main

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/shannon-sdk/client"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol/shannon"
)

// getShannonProtocol returns an instance of the Shannon protocol using the supplied Shannon-specific configuration.
func getShannonProtocol(logger polylog.Logger, config *shannon.ShannonGatewayConfig) (gateway.Protocol, error) {
	logger.Info().Msg("Starting PATH gateway with Shannon protocol")

	fullNode, err := getFullNode(logger, config.FullNodeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Shannon full node: %v", err)
	}

	gatewayClient, err := client.NewGatewayClient(
		logger,
		fullNode,
		config.GatewayConfig.GatewayAddress,
		config.GatewayConfig.GatewayPrivateKeyHex,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon protocol gateway client: %v", err)
	}

	protocolGatewayClient, err := getGatewayModeClient(logger, gatewayClient, config.GatewayConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon protocol gateway client: %v", err)
	}

	protocol, err := shannon.NewProtocol(logger, protocolGatewayClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon protocol instance: %v", err)
	}

	return protocol, nil
}

// getFullNode builds and returns a FullNode implementation for Shannon protocol integration.
//
// It may return a `fullNode` or a `fullNodeWithCache` depending on the caching configuration.
func getFullNode(logger polylog.Logger, config client.FullNodeConfig) (client.ShannonFullNode, error) {
	// With or without caching, we use the full node to fetch the onchain data.
	fullNode, err := client.NewFullNode(logger, config.RpcURL, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Shannon full node: %v", err)
	}

	// If caching is disabled, return the full node directly.
	if !config.CacheConfig.CachingEnabled {
		return fullNode, nil
	}

	// If caching is enabled, return the full node with cache.
	return client.NewFullNodeWithCache(logger, fullNode, config.CacheConfig)
}

// getGatewayModeClient gets the configured gateway client for the PATH instance.
func getGatewayModeClient(
	logger polylog.Logger,
	gatewayClient *client.GatewayClient,
	gatewayConfig shannon.GatewayConfig,
) (shannon.GatewayClient, error) {
	switch gatewayConfig.GatewayMode {

	case shannon.GatewayModeCentralized:
		logger.Info().Msg("getGatewayClient: PATH configured for centralized gateway mode")
		return shannon.NewCentralizedGatewayClient(
			logger,
			gatewayClient,
			gatewayConfig.OwnedAppsPrivateKeysHex,
		)

	case shannon.GatewayModeDelegated:
		logger.Info().Msg("getGatewayClient: PATH configured for delegated gateway mode")
		return shannon.NewDelegatedGatewayClient(
			logger,
			gatewayClient,
		)

		// TODO_IMPROVE(@commoddity, @adshmh): add new gateway client for permissionless mode once implemented in the SDK.

	default:
		return nil, fmt.Errorf("unsupported gateway mode: %s", gatewayConfig.GatewayMode)
	}
}
