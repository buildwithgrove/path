package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strconv"

	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	_ "github.com/joho/godotenv/autoload" // autoload env vars
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/buildwithgrove/path/envoy/auth_server/auth"
	store "github.com/buildwithgrove/path/envoy/auth_server/endpoint_store"
	"github.com/buildwithgrove/path/envoy/auth_server/proto"
)

// The auth server runs on port 10003.
// This matches the port used by the Envoy gRPC filter as defined in `envoy.yaml`.
// TODO_CONSIDER(@commoddity): Make this configurable. See thread here: https://github.com/buildwithgrove/path/pull/52/files/1a3e7a11f159f5b8d3c414f2417f7879bcfab410..258136504608c1269a27047bb9bded1ab4fefcc8#r1859409934
const port = 10003

const (
	envVarGRPCHostPort                = "GRPC_HOST_PORT"
	envVarGRPCUseInsecure             = "GRPC_USE_INSECURE"
	defaultGRPCUseInsecureCredentials = false
)

type options struct {
	grpcHostPort               string
	grpcUseInsecureCredentials bool
}

func gatherOptions() options {
	grpcHostPort := os.Getenv(envVarGRPCHostPort)
	if grpcHostPort == "" {
		panic(fmt.Sprintf("%s is not set in the environment", envVarGRPCHostPort))
	}

	grpcUseInsecureCredentials := defaultGRPCUseInsecureCredentials
	if insecureStr := os.Getenv(envVarGRPCUseInsecure); insecureStr != "" {
		if insecure, err := strconv.ParseBool(insecureStr); err == nil {
			grpcUseInsecureCredentials = insecure
		}
	}

	return options{
		grpcHostPort:               grpcHostPort,
		grpcUseInsecureCredentials: grpcUseInsecureCredentials,
	}
}

func connectGRPC(hostPort string, useInsecureCredentials bool) (*grpc.ClientConn, error) {
	var transport grpc.DialOption
	if useInsecureCredentials {
		transport = grpc.WithTransportCredentials(insecure.NewCredentials())
	} else {
		transport = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	}
	return grpc.NewClient(hostPort, transport)
}

func main() {

	// Initialize new polylog logger
	logger := polyzero.NewLogger()

	// Gather options from environment variables
	opts := gatherOptions()

	// Connect to the gRPC server for the GatewayEndpoints service
	conn, err := connectGRPC(opts.grpcHostPort, opts.grpcUseInsecureCredentials)
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
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	// Create a new AuthHandler to handle the request auth
	authHandler := &auth.AuthHandler{
		EndpointStore:    endpointStore,
		APIKeyAuthorizer: &auth.APIKeyAuthorizer{},
		JWTAuthorizer:    &auth.JWTAuthorizer{},
		Logger:           logger,
	}

	// Create a new gRPC server for handling auth requests from Envoy
	grpcServer := grpc.NewServer()

	// Register envoy proto server
	envoy_auth.RegisterAuthorizationServer(grpcServer, authHandler)

	fmt.Printf("Auth server starting on port %d...\n", port)
	if err = grpcServer.Serve(listen); err != nil {
		panic(err)
	}
}
