package shannon

import (
	"testing"

	"github.com/stretchr/testify/require"

	shannonRelayer "github.com/buildwithgrove/path/relayer/shannon"
)

func Test_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ShannonGatewayConfig
		wantErr bool
	}{
		{
			name: "should pass with valid config",
			config: ShannonGatewayConfig{
				FullNodeConfig: shannonRelayer.FullNodeConfig{
					RpcURL: "https://rpc-url.io",
					GRPCConfig: shannonRelayer.GRPCConfig{
						HostPort: "grpc-url.io:443",
					},
					GatewayAddress:    "pokt1710ed9a8d0986d808e607c5815cc5a13f15dba",
					GatewayPrivateKey: "d5fcbfb894059a21e914a2d6bf1508319ce2b1b8878f15aa0c1cdf883feb018d",
				},
			},
			wantErr: false,
		},
		{
			name: "should fail with invalid URL",
			config: ShannonGatewayConfig{
				FullNodeConfig: shannonRelayer.FullNodeConfig{
					RpcURL: "invalid-url",
					GRPCConfig: shannonRelayer.GRPCConfig{
						HostPort: "grpc-url.io:443",
					},
					GatewayAddress:    "pokt1710ed9a8d0986d808e607c5815cc5a13f15dba",
					GatewayPrivateKey: "d5fcbfb894059a21e914a2d6bf1508319ce2b1b8878f15aa0c1cdf883feb018d",
				},
			},
			wantErr: true,
		},
		{
			name: "should fail with invalid gateway address",
			config: ShannonGatewayConfig{
				FullNodeConfig: shannonRelayer.FullNodeConfig{
					RpcURL: "https://rpc-url.io",
					GRPCConfig: shannonRelayer.GRPCConfig{
						HostPort: "grpc-url.io:443",
					},
					GatewayAddress:    "invalid_address",
					GatewayPrivateKey: "d5fcbfb894059a21e914a2d6bf1508319ce2b1b8878f15aa0c1cdf883feb018d",
				},
			},
			wantErr: true,
		},
		{
			name: "should fail with invalid gateway private key",
			config: ShannonGatewayConfig{
				FullNodeConfig: shannonRelayer.FullNodeConfig{
					RpcURL: "https://rpc-url.io",
					GRPCConfig: shannonRelayer.GRPCConfig{
						HostPort: "grpc-url.io:443",
					},
					GatewayAddress:    "pokt1710ed9a8d0986d808e607c5815cc5a13f15dba",
					GatewayPrivateKey: "invalid_private_key",
				},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.config.Validate()
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
