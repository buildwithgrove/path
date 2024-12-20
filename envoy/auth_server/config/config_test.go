package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_LoadAuthServerConfigFromYAML(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		yamlData string
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
			name: "should return error for missing grpc_host_port",
			yamlData: `
auth_server_config:
  grpc_use_insecure_credentials: true
  endpoint_id_extractor_type: "url_path"
  port: 10003
`,
			wantErr: true,
		},
		{
			name: "should return error when grpc_host_port does not match pattern",
			yamlData: `
auth_server_config:
  grpc_host_port: "invalid_host_port"
  grpc_use_insecure_credentials: true
  endpoint_id_extractor_type: "url_path"
  port: 10003
`,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			var filePath string
			if test.yamlData != "" {
				tmpFile, err := os.CreateTemp("", "config-*.yaml")
				if err != nil {
					t.Fatalf("failed to create temp file: %v", err)
				}
				defer os.Remove(tmpFile.Name())

				if _, err := tmpFile.Write([]byte(test.yamlData)); err != nil {
					t.Fatalf("failed to write to temp file: %v", err)
				}
				tmpFile.Close()
				filePath = tmpFile.Name()
			} else {
				filePath = test.filePath
			}

			got, err := LoadAuthServerConfigFromYAML(filePath)
			if test.wantErr {
				c.Error(err)
			} else {
				c.NoError(err)
				c.Equal(test.want, got)
			}
		})
	}
}
