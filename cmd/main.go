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
	"github.com/buildwithgrove/path/db"
	"github.com/buildwithgrove/path/db/driver"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/relayer"
	"github.com/buildwithgrove/path/relayer/morse"
	"github.com/buildwithgrove/path/relayer/shannon"
	"github.com/buildwithgrove/path/request"
	"github.com/buildwithgrove/path/router"
	"github.com/buildwithgrove/path/user/authorizer"
)

const configPath = ".config.yaml"

func main() {
	logger := polyzero.NewLogger()

	config, err := config.LoadGatewayConfigFromYAML(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	protocol, err := getProtocol(config, logger)
	if err != nil {
		panic(fmt.Sprintf("failed to create protocol: %v", err))
	}

	requestParser, err := request.NewParser(config, logger)
	if err != nil {
		panic(fmt.Sprintf("failed to create request parser: %v", err))
	}

	relayer := &relayer.Relayer{Protocol: protocol}

	gateway := &gateway.Gateway{
		HTTPRequestParser: requestParser,
		Relayer:           relayer,
	}
	if config.IsUserDataEnabled() {
		userReqAuthorizer, cleanup, err := getUserReqAuthorizer(config.GetUserDataConfig(), logger)
		if err != nil {
			panic(fmt.Sprintf("failed to create user request authorizer: %v", err))
		}
		defer cleanup()
		gateway.UserRequestAuthorizer = userReqAuthorizer
	}

	// Until all components are ready, the `/healthz` endpoint will return a 503 Service
	// Unavailable status; once all components are ready, it will return a 200 OK status.
	// health check components must implement the router.HealthCheckComponent
	// interface to be able to signal they are ready to service requests.
	healthCheckComponents := []router.HealthCheckComponent{
		protocol,
	}

	apiRouter := router.NewRouter(router.RouterParams{
		Gateway:               gateway,
		HealthCheckComponents: healthCheckComponents,
		Config:                config.GetRouterConfig(),
		UserDataEnabled:       config.IsUserDataEnabled(),
		Logger:                logger,
	})
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

func getUserReqAuthorizer(config config.UserDataConfig, logger polylog.Logger) (gateway.UserRequestAuthorizer, func() error, error) {
	dbDriver, cleanup, err := driver.NewPostgresDriver(config.DBConnectionString)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create postgres driver: %v", err)
	}

	cache, err := db.NewCache(dbDriver, config.CacheRefreshInterval, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create user data cache: %v", err)
	}

	return authorizer.NewRequestAuthorizer(cache, config.RedisHostPort, logger), cleanup, nil
}
