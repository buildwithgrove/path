package main

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"

	morseconfig "github.com/buildwithgrove/path/config/morse"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol/morse"
)

// getMorseProtocol returns an instance of the Morse protocol using the supplied Morse-specific configuration.
func getMorseProtocol(
	logger polylog.Logger,
	config *morseconfig.MorseGatewayConfig,
) (gateway.Protocol, error) {
	logger.Info().Msg("Starting PATH gateway with Morse protocol")

	fullNode, err := morse.NewFullNode(config.FullNodeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create morse full node: %w", err)
	}

	protocol := morse.NewProtocol(logger, fullNode, config)
	return protocol, nil
}
