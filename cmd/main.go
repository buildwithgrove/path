package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"

	"github.com/buildwithgrove/path/config"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/health"
	"github.com/buildwithgrove/path/request"
	"github.com/buildwithgrove/path/router"
)

// defaultConfigPath will be appended to the location of
// the executable to get the full path to the config file.
const defaultConfigPath = "config/.config.yaml"

func main() {
	configPath, err := getConfigPath()
	if err != nil {
		log.Fatalf("failed to get config path: %v", err)
	}

	config, err := config.LoadGatewayConfigFromYAML(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("Initializing PATH logger with level: %s", config.Logger.Level)

	loggerOpts := []polylog.LoggerOption{
		polyzero.WithLevel(polyzero.ParseLevel(config.Logger.Level)),
	}

	logger := polyzero.NewLogger(loggerOpts...)

	logger.Info().Msgf("Starting PATH using config file: %s", configPath)

	protocol, err := getProtocol(config, logger)
	if err != nil {
		log.Fatalf("failed to create protocol: %v", err)
	}

	qosInstances, err := getServiceQoSInstances(config, logger)
	if err != nil {
		log.Fatalf("failed to setup QoS instances: %v", err)
	}

	// TODO_IMPROVE: consider using a separate protocol instance for the hydrator,
	// to enable configuring separate worker pools for the user requests
	// and the endpoint hydrator requests.
	hydrator, err := setupEndpointHydrator(config.HydratorConfig, protocol, qosInstances, logger)
	if err != nil {
		log.Fatalf("failed to setup endpoint hydrator: %v", err)
	}

	// setup the request parser which maps requests to the correst QoS instance.
	requestParser := &request.Parser{
		QoSServices: qosInstances,
		Logger:      logger,
	}

	// NOTE: the gateway uses the requestParser to get the correct QoS instance for any incoming request.
	gateway := &gateway.Gateway{
		HTTPRequestParser: requestParser,
		Protocol:          protocol,
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

// getConfigPath returns the full path to the config file
// based on the current working directory of the executable.
//
// For example if the binary is in:
// - `/app` the full config path will be `/app/config/.config.yaml`
// - `./bin` the full config path will be `./bin/config/.config.yaml`
func getConfigPath() (string, error) {
	var configPath string

	// The config path can be overridden using the `-config` flag.
	flag.StringVar(&configPath, "config", "", "override the default config path")
	flag.Parse()
	if configPath != "" {
		return configPath, nil
	}

	// Otherwise, use the default config path based on the executable path
	exeDir, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %v", err)
	}

	configPath = filepath.Join(filepath.Dir(exeDir), defaultConfigPath)

	return configPath, nil
}

// getProtocol returns the protocol instance based on the config YAML.
//
// - If `shannon_config` is set it returns a Shannon protocol instance.
// - If `morse_config` is set it returns a Morse protocol instance.
// - If neither is set, it returns an error.
func getProtocol(config config.GatewayConfig, logger polylog.Logger) (gateway.Protocol, error) {
	if shannonConfig := config.GetShannonConfig(); shannonConfig != nil {
		return getShannonProtocol(shannonConfig, logger)
	}

	if morseConfig := config.GetMorseConfig(); morseConfig != nil {
		return getMorseProtocol(morseConfig, logger)
	}

	return nil, fmt.Errorf("no protocol config set")
}
