package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_LoadAuthServerConfigFromYAML(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     AuthServerConfig
		wantErr  bool
	}{
		{
			name:     "should load valid auth server config without error",
			filePath: "../../../config/examples/config.shannon_example.yaml",
			want: AuthServerConfig{
				GRPCHostPort:               "path-auth-data-server:50051",
				GRPCUseInsecureCredentials: true,
				EndpointIDExtractorType:    defaultEndpointIDExtractorType,
				Port:                       defaultPort,
			},
			wantErr: false,
		},
		{
			name:     "should return error for missing grpc_host_port",
			filePath: "./examples/missing_grpc_host_port.yaml",
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			got, err := LoadAuthServerConfigFromYAML(test.filePath)
			if test.wantErr {
				c.Error(err)
			} else {
				c.NoError(err)
				c.Equal(test.want, got)
			}
		})
	}
}
