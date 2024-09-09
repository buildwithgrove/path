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
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/relayer"
	"github.com/buildwithgrove/path/relayer/morse"
	"github.com/buildwithgrove/path/relayer/shannon"
	"github.com/buildwithgrove/path/request"
	"github.com/buildwithgrove/path/router"
	"github.com/buildwithgrove/path/user"
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

	requestParser, err := request.NewParser(config, logger)
	if err != nil {
		log.Fatalf("failed to create request parser: %v", err)
	}

	relayer := &relayer.Relayer{Protocol: protocol}

	gateway := &gateway.Gateway{
		HTTPRequestParser: requestParser,
		Relayer:           relayer,
	}
	if config.UserDataEnabled() {
		gateway.UserRequestAuthenticator = getUserReqAuthenticator(config, logger)
	}

	apiRouter := router.NewRouter(gateway, config.GetRouterConfig(), config.UserDataEnabled(), logger)
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

func getUserReqAuthenticator(config config.GatewayConfig, logger polylog.Logger) gateway.UserRequestAuthenticator {
	if userDataConfig := config.GetUserDataConfig(); userDataConfig != nil {
		cache, cleanup, err := db.NewCache(*userDataConfig, logger)
		if err != nil {
			log.Fatalf("failed to create user data cache: %v", err)
		}
		defer cleanup()

		return &user.RequestAuthenticator{Cache: cache}
	}
	return nil
}
