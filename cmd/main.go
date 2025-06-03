package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"

	configpkg "github.com/buildwithgrove/path/config"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/health"
	"github.com/buildwithgrove/path/request"
	"github.com/buildwithgrove/path/router"
)

// defaultConfigPath will be appended to the location of
// the executable to get the full path to the config file.
const defaultConfigPath = "config/.config.yaml"

func main() {
	configPath, err := getConfigPath(defaultConfigPath)
	if err != nil {
		log.Fatalf("failed to get config path: %v", err)
	}

	config, err := configpkg.LoadGatewayConfigFromYAML(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("Initializing PATH logger with level: %s", config.Logger.Level)

	loggerOpts := []polylog.LoggerOption{
		polyzero.WithLevel(polyzero.ParseLevel(config.Logger.Level)),
	}

	logger := polyzero.NewLogger(loggerOpts...)

	logger.Info().Msgf("Starting PATH using config file: %s", configPath)

	protocol, err := getShannonProtocol(logger, config.GetGatewayConfig())
	if err != nil {
		log.Fatalf("failed to create protocol: %v", err)
	}

	qosInstances, err := getServiceQoSInstances(logger, config)
	if err != nil {
		log.Fatalf("failed to setup QoS instances: %v", err)
	}

	// setup metrics reporter, to be used by Gateway and Hydrator.
	metricsReporter, err := setupMetricsServer(logger, prometheusMetricsServerAddr)
	if err != nil {
		log.Fatalf("failed to start metrics server: %v", err)
	}

	setupPprofServer(context.TODO(), logger, pprofAddr)

	// setup data reporter, to be used by Gateway and Hydrator.
	dataReporter, err := setupHTTPDataReporter(logger, config.DataReporterConfig)
	if err != nil {
		log.Fatalf("failed to start the configured HTTP data reporter: %v", err)
	}

	// TODO_IMPROVE: consider using a separate protocol instance for the hydrator,
	// to enable configuring separate worker pools for the user requests
	// and the endpoint hydrator requests.
	hydrator, err := setupEndpointHydrator(
		logger,
		protocol,
		qosInstances,
		metricsReporter,
		dataReporter,
		config.HydratorConfig,
	)
	if err != nil {
		log.Fatalf("failed to setup endpoint hydrator: %v", err)
	}

	// setup the request parser which maps requests to the correct QoS instance.
	requestParser := &request.Parser{
		Logger: logger,

		QoSServices: qosInstances,
	}

	// NOTE: the gateway uses the requestParser to get the correct QoS instance for any incoming request.
	gateway := &gateway.Gateway{
		Logger: logger,

		HTTPRequestParser: requestParser,
		Protocol:          protocol,
		MetricsReporter:   metricsReporter,
		DataReporter:      dataReporter,
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
		Logger:            logger,
		Components:        components,
		ServiceIDReporter: protocol,
	}

	apiRouter := router.NewRouter(logger, gateway, healthChecker, config.GetRouterConfig())
	if err != nil {
		log.Fatalf("failed to create API router: %v", err)
	}

	// Log out some basic info about the running PATH instance.
	configuredServiceIDs := make([]string, 0, len(protocol.ConfiguredServiceIDs()))
	for serviceID := range protocol.ConfiguredServiceIDs() {
		configuredServiceIDs = append(configuredServiceIDs, string(serviceID))
	}
	// log.Printf is used here to ensure this info is printed to the console regardless of the log level.
	log.Printf("🌿 PATH gateway started.\n  Port: %d\n  Protocol: %s\n  Configured Service IDs: %s",
		config.GetRouterConfig().Port, protocol.Name(), strings.Join(configuredServiceIDs, ", "))

	// Start the API router.
	// This will block until the router is stopped.
	if err := apiRouter.Start(); err != nil {
		log.Fatalf("failed to start API router: %v", err)
	}
}

/* -------------------- Gateway Init Helpers -------------------- */

// getConfigPath returns the full path to the config file relative to the executable.
//
// Priority for determining config path:
// - If `-config` flag is set, use its value
// - Otherwise, use defaultConfigPath relative to executable directory
//
// Examples:
// - Executable in `/app` → config at `/app/config/.config.yaml`
// - Executable in `./bin` → config at `./bin/config/.config.yaml`
// - Executable in `./local/path` → config at `./local/path/.config.yaml`
func getConfigPath(defaultConfigPath string) (string, error) {
	var configPath string

	// Check for -config flag override
	flag.StringVar(&configPath, "config", "", "override the default config path")
	flag.Parse()
	if configPath != "" {
		return configPath, nil
	}

	// Get executable directory for default path
	exeDir, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %v", err)
	}

	configPath = filepath.Join(filepath.Dir(exeDir), defaultConfigPath)

	return configPath, nil
}
