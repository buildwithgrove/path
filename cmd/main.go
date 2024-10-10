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
	"github.com/buildwithgrove/path/message"
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

	protocol, err := getProtocol(config, logger)
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
	hydrator, err := setupEndpointHydrator(protocol, relayer, qosPublisher, hydratorQoSGenerators, logger)
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

func getProtocol(config config.GatewayConfig, logger polylog.Logger) (relayer.Protocol, error) {
	if shannonConfig := config.GetShannonConfig(); shannonConfig != nil {
		return getShannonProtocol(shannonConfig, logger)
	}

	if morseConfig := config.GetMorseConfig(); morseConfig != nil {
		return getMorseProtocol(morseConfig, logger)
	}

	return nil, fmt.Errorf("no protocol config set")
}

func getShannonProtocol(config *shannonConfig.ShannonGatewayConfig, logger polylog.Logger) (relayer.Protocol, error) {
	logger.Info().Msg("Starting PATH gateway with Shannon protocol")

	fullNode, err := shannon.NewFullNode(config.FullNodeConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create shannon full node: %v", err)
	}

	protocol, err := shannon.NewProtocol(context.Background(), fullNode)
	if err != nil {
		return nil, fmt.Errorf("failed to create shannon protocol: %v", err)
	}

	return protocol, nil
}

func getMorseProtocol(config *morseConfig.MorseGatewayConfig, logger polylog.Logger) (relayer.Protocol, error) {
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

// TODO_UPNEXT(@adshmh): Remove this after implementing the QoS Publisher
type noopQoSPublisher struct{}

func (noopQoSPublisher) Publish(message.ObservationSet) error {
	return nil
}
