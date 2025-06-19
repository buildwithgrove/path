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

	onchainDataFetcher, err := getOnchainDataFetcher(logger, config.GatewayClientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Shannon full node: %v", err)
	}

	gatewayClient, err := client.NewGatewayClient(
		logger,
		onchainDataFetcher,
		config.GatewayConfig.GatewayAddress,
		config.GatewayConfig.GatewayPrivateKeyHex,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create a gateway client: %v", err)
	}

	protocolGatewayClient, err := getGatewayModeClient(logger, gatewayClient, config.GatewayConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create a gateway mode client: %v", err)
	}

	protocol, err := shannon.NewProtocol(logger, protocolGatewayClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create a Shannon protocol instance: %v", err)
	}

	return protocol, nil
}

// getOnchainDataFetcher builds and returns an OnchainDataFetcher instance using the supplied configuration.
// It is passed to the GatewayClient to fetch and optionally cache onchain data for the Shannon protocol.
//
// May return one of the following:
//   - client.GatewayClientCache: if caching is enabled
//   - client.GRPCClient: if caching is disabled
func getOnchainDataFetcher(logger polylog.Logger, config client.GatewayClientConfig) (client.OnchainDataFetcher, error) {
	// Instantiate a gRPC client for fetching onchain data from the Shannon full node.
	grpcClient, err := client.NewGRPCClient(logger, config.GRPCConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create a gRPC client: %v", err)
	}

	// If caching is disabled, return the gRPC client directly.
	if config.CacheConfig.UseCache != nil && !*config.CacheConfig.UseCache {
		return grpcClient, nil
	}

	// If caching is enabled, return a GatewayClientCache instance that wraps the gRPC client.
	return client.NewGatewayClientCache(logger, grpcClient, config.CacheConfig)
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
