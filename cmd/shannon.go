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

	gatewayClientCache, err := getGatewayClientCache(logger, config.FullNodeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Shannon full node: %v", err)
	}

	gatewayClient, err := client.NewGatewayClient(
		logger,
		gatewayClientCache,
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

// getGatewayClientCache builds and returns a GatewayClientCache instance using the supplied configuration.
// It is passed to the GatewayClient to fetch and cache onchain data for the Shannon protocol.
func getGatewayClientCache(logger polylog.Logger, config client.FullNodeConfig) (*client.GatewayClientCache, error) {
	return client.NewGatewayClientCache(logger, config.RpcURL, config)
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
