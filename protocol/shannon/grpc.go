package shannon

import (
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// connectGRPC creates a new gRPC connection.
// Backoff configuration may be customized using the config YAML fields
// under `grpc_config`. TLS is enabled by default, unless overridden by
// the `grpc_config.insecure` field.
// TODO_TECHDEBT: use an enhanced grpc connection with reconnect logic.
// All GRPC settings have been disabled to focus the E2E tests on the
// gateway functionality rather than GRPC settings.
func connectGRPC(config GRPCConfig) (*grpc.ClientConn, error) {
	if config.Insecure {
		transport := grpc.WithTransportCredentials(insecure.NewCredentials())
		dialOptions := []grpc.DialOption{transport}
		return grpc.NewClient(
			config.HostPort,
			dialOptions...,
		)
	}

	// TODO_TECHDEBT: make the necessary changes to allow using grpc.NewClient here.
	// Currently using the grpc.NewClient method fails the E2E tests.
	return grpc.Dial( //nolint:all
		config.HostPort,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
	)
}
