package main

import (
	"context"
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"

	morseconfig "github.com/buildwithgrove/path/config/morse"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol/morse"
)

// getMorseProtocol returns an instance of the Morse protocol using the supplied Morse-specific configuration.
func getMorseProtocol(
	config *morseconfig.MorseGatewayConfig,
	logger polylog.Logger,
) (gateway.Protocol, error) {
	logger.Info().Msg("Starting PATH gateway with Morse protocol")

	fullNode, err := morse.NewFullNode(config.FullNodeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create morse full node: %v", err)
	}

	protocol, err := morse.NewProtocol(context.Background(), fullNode, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create morse protocol: %v", err)
	}

	return protocol, nil
}
