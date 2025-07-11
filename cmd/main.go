package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"

	configpkg "github.com/buildwithgrove/path/config"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/health"
	"github.com/buildwithgrove/path/metrics"
	"github.com/buildwithgrove/path/metrics/devtools"
	protocolPkg "github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/request"
	"github.com/buildwithgrove/path/router"
)

// Version information injected at build time via ldflags
var (
	Version   string
	Commit    string
	BuildDate string
)

// defaultConfigPath will be appended to the location of
// the executable to get the full path to the config file.
const defaultConfigPath = "config/.config.yaml"

func main() {
	log.Printf("🌿 PATH gateway starting...")

	// Initialize version metrics for Prometheus monitoring
	metrics.SetVersionInfo(Version, Commit, BuildDate)

	// Get the config path
	configPath, err := getConfigPath(defaultConfigPath)
	if err != nil {
		log.Fatalf("failed to get config path: %v", err)
	}

	// Load the config
	config, err := configpkg.LoadGatewayConfigFromYAML(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize the logger
	log.Printf("Initializing PATH logger with level: %s", config.Logger.Level)
	loggerOpts := []polylog.LoggerOption{
		polyzero.WithLevel(polyzero.ParseLevel(config.Logger.Level)),
	}
	logger := polyzero.NewLogger(loggerOpts...)

	// Log the config path
	logger.Info().Msgf("Starting PATH using config file: %s", configPath)

	// Create the protocol
	protocol, err := getProtocol(logger, config)
	if err != nil {
		log.Fatalf("failed to create protocol: %v", err)
	}

	// Prepare the QoS instances
	qosInstances, err := getServiceQoSInstances(logger, config, protocol)
	if err != nil {
		log.Fatalf("failed to setup QoS instances: %v", err)
	}

	// Setup metrics reporter, to be used by Gateway and Hydrator
	metricsReporter, err := setupMetricsServer(logger, prometheusMetricsServerAddr)
	if err != nil {
		log.Fatalf("failed to start metrics server: %v", err)
	}

	// Setup the pprof server
	setupPprofServer(context.TODO(), logger, pprofAddr)

	// Setup the data reporter
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

	// Setup the request parser which maps requests to the correct QoS instance.
	requestParser := &request.Parser{
		Logger:      logger,
		QoSServices: qosInstances,
	}

	// NOTE: the gateway uses the requestParser to get the correct QoS instance for any incoming request.
	gateway := &gateway.Gateway{
		Logger:            logger,
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

	// Convert qosInstances to DataReporter map to satisfy the QoSDisqualifiedEndpointsReporter interface.
	qosLevelReporters := make(map[protocolPkg.ServiceID]devtools.QoSDisqualifiedEndpointsReporter)
	for serviceID, qosService := range qosInstances {
		qosLevelReporters[serviceID] = qosService
	}

	// Create the disqualified endpoints reporter to report data on disqualified endpoints
	// through the `/disqualified_endpoints` route for real time QoS data on service endpoints.
	disqualifiedEndpointsReporter := &devtools.DisqualifiedEndpointReporter{
		Logger:                logger,
		ProtocolLevelReporter: protocol,
		QoSLevelReporters:     qosLevelReporters,
	}

	// Initialize the API router to serve requests to the PATH API.
	apiRouter := router.NewRouter(
		logger,
		gateway,
		disqualifiedEndpointsReporter,
		healthChecker,
		config.GetRouterConfig(),
	)

	// -------------------- Log PATH Startup Info --------------------

	// Log out some basic info about the running PATH instance
	configuredServiceIDs := make([]string, 0, len(protocol.ConfiguredServiceIDs()))
	for serviceID := range protocol.ConfiguredServiceIDs() {
		configuredServiceIDs = append(configuredServiceIDs, string(serviceID))
	}
	logger.Info().Msgf("🌿 PATH gateway starting on port %d for Protocol: %s with Configured Service IDs: %s",
		config.GetRouterConfig().Port, protocol.Name(), strings.Join(configuredServiceIDs, ", "))

	// -------------------- Start PATH API Router --------------------

	// This will block until the router is stopped.
	server, err := apiRouter.Start()
	if err != nil {
		logger.Error().Err(err).Msg("failed to start PATH API router")
	}

	// -------------------- PATH Shutdown --------------------
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info().Msg("Shutting down PATH...")

	// TODO_IMPROVE: Make shutdown timeout configurable and add graceful shutdown of dependencies
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("PATH forced to shutdown")
	}

	logger.Info().Msg("PATH exited properly")
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
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	configPath = filepath.Join(filepath.Dir(exeDir), defaultConfigPath)

	return configPath, nil
}

// getProtocol returns the protocol instance based on the config YAML.
//
// - If `shannon_config` is set it returns a Shannon protocol instance.
// - If `morse_config` is set it returns a Morse protocol instance.
// - If neither is set, it returns an error.
func getProtocol(logger polylog.Logger, config configpkg.GatewayConfig) (gateway.Protocol, error) {
	if shannonConfig := config.GetShannonConfig(); shannonConfig != nil {
		return getShannonProtocol(logger, shannonConfig)
	}

	if morseConfig := config.GetMorseConfig(); morseConfig != nil {
		return getMorseProtocol(logger, morseConfig)
	}

	return nil, fmt.Errorf("no protocol config set")
}
