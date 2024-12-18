package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	_ "github.com/joho/godotenv/autoload" // autoload env vars
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/buildwithgrove/path/envoy/auth_server/auth"
	"github.com/buildwithgrove/path/envoy/auth_server/config"
	store "github.com/buildwithgrove/path/envoy/auth_server/endpoint_store"
	"github.com/buildwithgrove/path/envoy/auth_server/proto"
)

// defaultConfigPath will be appended to the location of
// the executable to get the full path to the config file.
const defaultConfigPath = "config/.config.yaml"

func main() {
	// Initialize new polylog logger
	logger := polyzero.NewLogger()

	configPath, err := getConfigPath()
	if err != nil {
		log.Fatalf("failed to get config path: %v", err)
	}
	logger.Info().Msgf("Starting Envoy Auth Server using config file: %s", configPath)

	config, err := config.LoadAuthServerConfigFromYAML(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Connect to the gRPC server for the GatewayEndpoints service
	conn, err := connectGRPC(config.GRPCHostPort, config.GRPCUseInsecureCredentials)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to gRPC server: %v", err))
	}
	defer conn.Close()

	// Create a new gRPC client for the GatewayEndpoints service
	grpcClient := proto.NewGatewayEndpointsClient(conn)

	// Create a new endpoint store
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	endpointStore, err := store.NewEndpointStore(ctx, grpcClient, logger)
	if err != nil {
		panic(err)
	}

	// Create a new listener to listen for requests from Envoy
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		panic(err)
	}

	// Determine which gateway endpoint ID extractor to use
	// If the extractor is not set in the config, the default "url_path" extractor is used
	endpointIDExtractor := getEndpointIDExtractor(config.EndpointIDExtractorType)

	// Create a new AuthHandler to handle the request auth
	authHandler := &auth.AuthHandler{
		EndpointStore:       endpointStore,
		APIKeyAuthorizer:    &auth.APIKeyAuthorizer{},
		JWTAuthorizer:       &auth.JWTAuthorizer{},
		EndpointIDExtractor: endpointIDExtractor,
		Logger:              logger,
	}

	// Create a new gRPC server for handling auth requests from Envoy
	grpcServer := grpc.NewServer()

	// Register envoy proto server
	envoy_auth.RegisterAuthorizationServer(grpcServer, authHandler)

	fmt.Printf("Auth server starting on port %d...\n", config.Port)
	if err = grpcServer.Serve(listen); err != nil {
		panic(err)
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

// connectGRPC connects to the gRPC server for the GatewayEndpoints service
// and returns a gRPC client connection.
func connectGRPC(hostPort string, useInsecureCredentials bool) (*grpc.ClientConn, error) {
	var transport grpc.DialOption
	if useInsecureCredentials {
		transport = grpc.WithTransportCredentials(insecure.NewCredentials())
	} else {
		transport = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	}
	return grpc.NewClient(hostPort, transport)
}

// getEndpointIDExtractor returns the endpoint ID extractor based on the config YAML.
func getEndpointIDExtractor(endpointIDExtractorType auth.EndpointIDExtractorType) auth.EndpointIDExtractor {
	switch endpointIDExtractorType {
	case auth.EndpointIDExtractorTypeURLPath:
		return &auth.URLPathExtractor{}
	case auth.EndpointIDExtractorTypeHeader:
		return &auth.HeaderExtractor{}
	}
	return nil // this should never happen
}
