package main

import (
	"context"
	"fmt"
	"log"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"

	"github.com/buildwithgrove/path/config"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/relayer"
	"github.com/buildwithgrove/path/relayer/morse"
	"github.com/buildwithgrove/path/relayer/shannon"
	"github.com/buildwithgrove/path/request"
	"github.com/buildwithgrove/path/router"
)

const configPath = ".config.yaml"

func getProtocol(config config.GatewayConfig, logger polylog.Logger) (relayer.Protocol, error) {

	// Config YAML validation enforces that exactly one protocol config is set,
	// so first check if the protocol config is set for Shannon.
	if shannonConfig := config.GetShannonConfig(); shannonConfig != nil {
		logger.Info().Msg("Starting PATH gateway with Shannon protocol")

		fullNode, err := shannon.NewFullNode(shannonConfig.FullNodeConfig, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create shannon full node: %v", err)
		}

		protocol, err := shannon.NewProtocol(context.Background(), fullNode)
		if err != nil {
			return nil, fmt.Errorf("failed to create shannon protocol: %v", err)
		}

		return protocol, nil
	}

	// If the protocol config is not set for Shannon, then it must be set for Morse.
	if morseConfig := config.GetMorseConfig(); morseConfig != nil {
		logger.Info().Msg("Starting PATH gateway with Morse protocol")

		fullNode, err := morse.NewFullNode(morseConfig.FullNodeConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create morse full node: %v", err)
		}

		protocol, err := morse.NewProtocol(context.Background(), fullNode, morseConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create morse protocol: %v", err)
		}

		return protocol, nil
	}

	// this should never happen but guard against it
	return nil, fmt.Errorf("no protocol config set")
}

func main() {
	logger := polyzero.NewLogger()

	config, err := config.LoadGatewayConfigFromYAML(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	protocol, err := getProtocol(config, logger)
	if err != nil {
		log.Fatalf("failed to create protocol: %v", err)
	}

	requestParser, err := request.NewParser(config, logger)
	if err != nil {
		log.Fatalf("failed to create request parser: %v", err)
	}

	relayer := &relayer.Relayer{Protocol: protocol}

	gateway := &gateway.Gateway{
		HTTPRequestParser: requestParser,
		Relayer:           relayer,
	}

	// Until all components are ready, the `/healthz` endpoint will return a 503 Service
	// Unavailable status; once all components are ready, it will return a 200 OK status.
	// health check components must implement the router.HealthCheckComponent
	// interface to be able to signal they are ready to service requests.
	healthCheckComponents := []router.HealthCheckComponent{
		protocol,
	}

	apiRouter := router.NewRouter(gateway, healthCheckComponents, config.GetRouterConfig(), logger)
	if err != nil {
		log.Fatalf("failed to create API router: %v", err)
	}

	if err := apiRouter.Start(); err != nil {
		log.Fatalf("failed to start API router: %v", err)
	}
}
