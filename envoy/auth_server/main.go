//go:build auth_server

package main

import (
	"fmt"
	"net"
	"os"

	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"google.golang.org/grpc"

	"github.com/buildwithgrove/auth-server/config"
	"github.com/buildwithgrove/auth-server/db"
	"github.com/buildwithgrove/auth-server/db/postgres"
	"github.com/buildwithgrove/auth-server/server"
)

// CONFIG_PATH is set in the Envoy Docker image during the build process.
// It points to the mounted `.config.auth_server.yaml` file. See `Dockerfile.envoy`.
const envVarConfigPath = "CONFIG_PATH"

func main() {
	logger := polyzero.NewLogger()

	configPath := os.Getenv(envVarConfigPath)
	if configPath == "" {
		panic(fmt.Sprintf("%s is not set in the environment", envVarConfigPath))
	}

	config, err := config.LoadAuthServerConfigFromYAML(configPath)
	if err != nil {
		panic(err)
	}

	dbDriver, _, err := postgres.NewPostgresDriver(config.PostgresConnectionString)
	if err != nil {
		panic(err)
	}

	cache, err := db.NewEndpointDataCache(dbDriver, config.CacheRefreshInterval, logger)
	if err != nil {
		panic(err)
	}

	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		panic(err)
	}

	authServer := &server.AuthServer{
		Cache: cache,
		// TODO_IMPROVE: make the authorizers configurable from the plugin config YAML
		Authorizers: []server.Authorizer{
			&server.AccountUserIDAuthorizer{},
		},
		Logger: logger,
	}

	grpcServer := grpc.NewServer()

	// register envoy proto server
	envoy_auth.RegisterAuthorizationServer(grpcServer, authServer)

	fmt.Printf("Auth server starting on %s:%d\n", config.Host, config.Port)
	err = grpcServer.Serve(listen)
	if err != nil {
		panic(err)
	}

}
