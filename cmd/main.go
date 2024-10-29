package main

import (
	"context"
	"fmt"
	"log"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"

	"github.com/buildwithgrove/path/config"
	morseConfig "github.com/buildwithgrove/path/config/morse"
	shannonConfig "github.com/buildwithgrove/path/config/shannon"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/health"
	"github.com/buildwithgrove/path/relayer"
	"github.com/buildwithgrove/path/relayer/morse"
	"github.com/buildwithgrove/path/relayer/shannon"
	"github.com/buildwithgrove/path/request"
	"github.com/buildwithgrove/path/router"
)

const configPath = ".config.yaml"

func main() {
	logger := polyzero.NewLogger()

	config, err := config.LoadGatewayConfigFromYAML(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	protocol, endpointLister, err := getProtocol(config, logger)
	if err != nil {
		log.Fatalf("failed to create protocol: %v", err)
	}

	relayer := &relayer.Relayer{Protocol: protocol}

	qosPublisher, err := getQoSPublisher(config.MessagingConfig)
	if err != nil {
		log.Fatalf("failed to setup the QoS publisher: %v", err)
	}

	gatewayQoSInstances, hydratorQoSGenerators, err := getServiceQoSInstances(config, logger)
	if err != nil {
		log.Fatalf("failed to setup QoS instances: %v", err)
	}

	// TODO_IMPROVE: consider using a separate relayer for the hydrator,
	// to enable configuring separate worker pools for the user requests
	// and the endpoint hydrator requests.
	hydrator, err := setupEndpointHydrator(endpointLister, relayer, qosPublisher, hydratorQoSGenerators, logger)
	if err != nil {
		log.Fatalf("failed to setup endpoint hydrator: %v", err)
	}

	requestParser, err := request.NewParser(config, gatewayQoSInstances, logger)
	if err != nil {
		log.Fatalf("failed to create request parser: %v", err)
	}

	gateway := &gateway.Gateway{
		HTTPRequestParser: requestParser,
		Relayer:           relayer,
		QoSPublisher:      qosPublisher,
		Logger:            logger,
	}

	// Until all components are ready, the `/healthz` endpoint will return a 503 Service
	// Unavailable status; once all components are ready, it will return a 200 OK status.
	// health check components must implement the health.Check interface
	// to be able to signal they are ready to service requests.
	components := []health.Check{protocol}
	if hydrator != nil {
		components = append(components, hydrator)
	}

	healthChecker := &health.Checker{
		Components: components,
		Logger:     logger,
	}

	apiRouter := router.NewRouter(gateway, healthChecker, config.GetRouterConfig(), logger)
	if err != nil {
		log.Fatalf("failed to create API router: %v", err)
	}

	if err := apiRouter.Start(); err != nil {
		log.Fatalf("failed to start API router: %v", err)
	}
}

/* -------------------- Gateway Init Helpers -------------------- */

func getProtocol(config config.GatewayConfig, logger polylog.Logger) (relayer.Protocol, gateway.EndpointLister, error) {
	if shannonConfig := config.GetShannonConfig(); shannonConfig != nil {
		return getShannonProtocol(shannonConfig, logger)
	}

	if morseConfig := config.GetMorseConfig(); morseConfig != nil {
		return getMorseProtocol(morseConfig, logger)
	}

	return nil, nil, fmt.Errorf("no protocol config set")
}

func getShannonProtocol(config *shannonConfig.ShannonGatewayConfig, logger polylog.Logger) (relayer.Protocol, gateway.EndpointLister, error) {
	logger.Info().Msg("Starting PATH gateway with Shannon protocol")

	// LazyFullNode skips all caching and queries the onchain data for serving each relay request.
	lazyFullNode, err := shannon.NewLazyFullNode(config.FullNodeConfig, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Shannon lazy full node: %v", err)
	}

	if config.FullNodeConfig.LazyMode {
		protocol := &shannon.Protocol{lazyFullNode, logger}
		// return the same protocol instance as two different interfaces for consumption by the relayer and the endpoint hydrator components.
		return protocol, protocol, nil
	}

	// Use a Caching FullNode implementation if LazyMode flag is not set.
	cachingFullNode, err := shannon.NewCachingFullNode(lazyFullNode, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Shannon caching full node: %v", err)
	}

	protocol := &shannon.Protocol{cachingFullNode, logger}

	// return the same protocol instance as two different interfaces for consumption by the relayer and the endpoint hydrator components.
	return protocol, protocol, nil
}

func getMorseProtocol(config *morseConfig.MorseGatewayConfig, logger polylog.Logger) (relayer.Protocol, gateway.EndpointLister, error) {
	logger.Info().Msg("Starting PATH gateway with Morse protocol")

	fullNode, err := morse.NewFullNode(config.FullNodeConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create morse full node: %v", err)
	}

	protocol, err := morse.NewProtocol(context.Background(), fullNode, config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create morse protocol: %v", err)
	}

	// return the same protocol instance as two different interfaces for consumption by the relayer and the endpoint hydrator components.
	return protocol, protocol, nil
}
