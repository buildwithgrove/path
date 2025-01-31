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

	// Get the config path based on the path of the executable
	// If the `-config` flag is set, it takes precedence and its value is used.
	configPath, err := getConfigPath(defaultConfigPath)
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
	endpointStore, err := store.NewEndpointStore(ctx, logger, grpcClient)
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
		Logger: logger,

		EndpointStore:       endpointStore,
		APIKeyAuthorizer:    &auth.APIKeyAuthorizer{},
		JWTAuthorizer:       &auth.JWTAuthorizer{},
		EndpointIDExtractor: endpointIDExtractor,
		ServiceIDExtractor: &auth.ServiceIDExtractor{
			ServiceAliases: config.ServiceAliases,
		},
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

// getConfigPath returns the full path to the config file relative to the executable.
//
// Priority for determining config path:
// - If `-config` flag is set, use its value
// - Otherwise, use defaultConfigPath relative to executable directory
//
// Examples:
// - Executable in `/app` → config at `/app/config/.config.yaml`
// - Executable in `./bin` → config at `./bin/config/.config.yaml`
// - Executable in `./local/path` → config at `./local/path/config/.config.yaml`
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

// connectGRPC connects to the gRPC server for the GatewayEndpoints service
// and returns a gRPC client connection.
func connectGRPC(hostPort string, useInsecureCredentials bool) (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials
	if useInsecureCredentials {
		creds = insecure.NewCredentials()
	} else {
		creds = credentials.NewTLS(&tls.Config{})
	}
	return grpc.NewClient(hostPort, grpc.WithTransportCredentials(creds))
}

// getEndpointIDExtractor returns the endpoint ID extractor based on the config YAML.
func getEndpointIDExtractor(endpointIDExtractorType auth.EndpointIDExtractorType) auth.EndpointIDExtractor {
	switch endpointIDExtractorType {
	case auth.EndpointIDExtractorTypeURLPath:
		return &auth.URLPathExtractor{}
	case auth.EndpointIDExtractorTypeHeader:
		return &auth.HeaderExtractor{}
	default: // this should never happen
		panic(fmt.Sprintf("invalid endpoint ID extractor type: %v", endpointIDExtractorType))
	}
}
