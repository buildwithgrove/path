//go:build auth_server

package main

import (
	"fmt"
	"net"
	"os"

	auth_pb "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"google.golang.org/grpc"

	"github.com/buildwithgrove/auth-server/config"
	"github.com/buildwithgrove/auth-server/db"
	"github.com/buildwithgrove/auth-server/db/postgres"
	"github.com/buildwithgrove/auth-server/server"
)

// filterName is the name of the filter that Envoy will use to identify and load the server
// If must match the `http_filters.typed_config.library_id` field for the Go filter in envoy.yaml
const filterName = "auth-server"

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

	cache, err := db.NewUserDataCache(dbDriver, config.CacheRefreshInterval, logger)
	if err != nil {
		panic(err)
	}

	endPoint := fmt.Sprintf("localhost:%d", 3001)
	listen, err := net.Listen("tcp", endPoint)
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()

	// register envoy proto server
	server := &server.AuthServer{
		JWTParser: &server.JWTParser{
			Issuer:   config.JWTIssuer,
			Audience: config.JWTAudience,
			JWKSURL:  config.JWTJWKSURL,
		},
		Cache:  cache,
		Logger: logger,
	}
	auth_pb.RegisterAuthorizationServer(grpcServer, server)

	grpcServer.Serve(listen)
}
