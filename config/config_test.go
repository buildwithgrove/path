package config

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/buildwithgrove/path/config/morse"
	"github.com/buildwithgrove/path/config/shannon"
	"github.com/buildwithgrove/path/protocol"
	morseprotocol "github.com/buildwithgrove/path/protocol/morse"
	shannonprotocol "github.com/buildwithgrove/path/protocol/shannon"
)

func Test_LoadGatewayConfigFromYAML(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		yamlData string
		want     GatewayConfig
		wantErr  bool
	}{
		{
			name:     "should load valid morse config without error",
			filePath: "./examples/config.morse_example.yaml",
			want: GatewayConfig{
				MorseConfig: &morse.MorseGatewayConfig{
					FullNodeConfig: morseprotocol.FullNodeConfig{
						URL:             "https://pocket-rpc.liquify.com",
						RelaySigningKey: "40af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d38840af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d388",
						HttpConfig: morseprotocol.HttpConfig{
							Retries: 3,
							Timeout: 5000 * time.Millisecond,
						},
					},
					SignedAATs: map[string]morse.SignedAAT{
						"40af4e7e1b311c76a573610fe115cd2adf1eeade": {
							ClientPublicKey:      "40af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d388",
							ApplicationPublicKey: "40af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d388",
							ApplicationSignature: "40af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d38840af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d388",
						},
					},
				},
				Router: RouterConfig{
					Port:               defaultPort,
					MaxRequestBodySize: defaultMaxRequestBodySize,
					ReadTimeout:        defaultHTTPServerReadTimeout,
					WriteTimeout:       defaultHTTPServerWriteTimeout,
					IdleTimeout:        defaultHTTPServerIdleTimeout,
				},
				Logger: LoggerConfig{
					Level: defaultLogLevel,
				},
			},
			wantErr: false,
		},
		{
			name:     "should load valid shannon config without error",
			filePath: "./examples/config.shannon_example.yaml",
			want: GatewayConfig{
				ShannonConfig: &shannon.ShannonGatewayConfig{
					FullNodeConfig: shannonprotocol.FullNodeConfig{
						RpcURL: "https://shannon-testnet-grove-rpc.beta.poktroll.com",
						GRPCConfig: shannonprotocol.GRPCConfig{
							HostPort: "shannon-testnet-grove-grpc.beta.poktroll.com:443",
						},
						LazyMode: false,
						CacheConfig: shannonprotocol.CacheConfig{
							AppTTL:     5 * time.Minute,
							SessionTTL: 5 * time.Minute,
						},
					},
					GatewayConfig: shannonprotocol.GatewayConfig{
						GatewayMode:          protocol.GatewayModeCentralized,
						GatewayAddress:       "pokt1up7zlytnmvlsuxzpzvlrta95347w322adsxslw",
						GatewayPrivateKeyHex: "40af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d388",
						OwnedAppsPrivateKeysHex: []string{
							"40af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d388",
						},
					},
				},
				Router: RouterConfig{
					Port:               defaultPort,
					MaxRequestBodySize: defaultMaxRequestBodySize,
					ReadTimeout:        defaultHTTPServerReadTimeout,
					WriteTimeout:       defaultHTTPServerWriteTimeout,
					IdleTimeout:        defaultHTTPServerIdleTimeout,
				},
				Logger: LoggerConfig{
					Level: defaultLogLevel,
				},
			},
			wantErr: false,
		},
		{
			name:     "should return error for invalid full node URL",
			filePath: "invalid_full_node_url.yaml",
			yamlData: `
			shannon_config:
			  full_node_config:
			    rpc_url: "invalid-url"
			    grpc_url: "grpcs://grpc-url.io"
			`,
			wantErr: true,
		},
		{
			name:     "should return error for invalid gateway address",
			filePath: "invalid_gateway_address.yaml",
			yamlData: `
			shannon_config:
			  full_node_config:
			    rpc_url: "https://rpc-url.io"
			    grpc_url: "grpcs://grpc-url.io"
			  gateway_config:
			    gateway_address: "invalid_gateway_address"
			    gateway_private_key_hex: "d5fcbfb894059a21e914a2d6bf1508319ce2b1b8878f15aa0c1cdf883feb018d"
			    gateway_mode: "delegated"
			`,
			wantErr: true,
		},
		{
			name:     "should return error for invalid service ID",
			filePath: "invalid_service_id.yaml",
			yamlData: `
			morse_config:
			  signed_aats:
			    invalid_service_id:
			      client_public_key: "a6258b46ecad0628b72099f91e87eef1b040a8747ed2d476f56ad359372bf619"
			      application_public_key: "5a8c62e4701f349a3b9288cfbd825db230a8ec74fd234e7cb0849e915bc6d6ce"
			      application_signature: "57d73225f83383e93571d0178f01368f26af9e552aaf073233d54600b60464043ba7013633d082b05d03ac7271667b307b09f47b8ac04000b19205cc1f99555d"
			`,
			wantErr: true,
		},
		{
			name:     "should return error for invalid client public key",
			filePath: "invalid_client_public_key.yaml",
			yamlData: `
			morse_config:
			  signed_aats:
			    f9076ec39b2a495883eb59740d566d5fa2e2b222:
			      client_public_key: "invalid_client_public_key"
			      application_public_key: "5a8c62e4701f349a3b9288cfbd825db230a8ec74fd234e7cb0849e915bc6d6ce"
			      application_signature: "57d73225f83383e93571d0178f01368f26af9e552aaf073233d54600b60464043ba7013633d082b05d03ac7271667b307b09f47b8ac04000b19205cc1f99555d"
			`,
			wantErr: true,
		},
		{
			name:     "should return error for invalid application signature",
			filePath: "invalid_application_signature.yaml",
			yamlData: `
			morse_config:
			  signed_aats:
			    f9076ec39b2a495883eb59740d566d5fa2e2b222:
			      client_public_key: "8604213b0c1ec52b5ae43eb854ce486a3756ec97cc194f3afe518947766aac11"
			      application_public_key: "71dd0e166022f1665dbba91b223998b0f328e9af2193a363456412a8eb4272e4"
			      application_signature: "invalid_application_signature"
			`,
			wantErr: true,
		},
		{
			name:     "should return error for invalid http_config timeout",
			filePath: "invalid_timeout.yaml",
			yamlData: `
			morse_config:
			  full_node_config:
			    url: "http://full-node-url"
			    http_config:
			      retries: 3
			      timeout: 5000
			    relay_signing_key: "gateway-private-key"
			  signed_aats:
			    f9076ec39b2a495883eb59740d566d5fa2e2b222:
			      client_public_key: "8604213b0c1ec52b5ae43eb854ce486a3756ec97cc194f3afe518947766aac11"
			      application_public_key: "71dd0e166022f1665dbba91b223998b0f328e9af2193a363456412a8eb4272e4"
			      application_signature: "bb04cb9cb34ea6e2d57fb679f7b1e73ff77992e0f39a1e7db0c8ed2a91aed3668d0b6399ea70614a0f51b714a3ad3bd3bd3ca2bc4a75302c14ce207d44c738cdbbf"
			`,
			wantErr: true,
		},
		{
			name:     "should return error for non-existent file",
			filePath: "non_existent.yaml",
			yamlData: "",
			wantErr:  true,
		},
		{
			name:     "should return error for invalid YAML",
			filePath: "invalid_config.yaml",
			yamlData: "invalid_yaml: [",
			wantErr:  true,
		},
		{
			name:     "should load config with valid logger level",
			filePath: "valid_logger.yaml",
			yamlData: `morse_config:
  full_node_config:
    url: "https://pocket-rpc.liquify.com"
    relay_signing_key: "40af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d38840af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d388"
    http_config:
      retries: 3
      timeout: "5000ms"
logger_config:
  level: "debug"`,
			want: GatewayConfig{
				MorseConfig: &morse.MorseGatewayConfig{
					FullNodeConfig: morseprotocol.FullNodeConfig{
						URL:             "https://pocket-rpc.liquify.com",
						RelaySigningKey: "40af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d38840af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d388",
						HttpConfig: morseprotocol.HttpConfig{
							Retries: 3,
							Timeout: 5000 * time.Millisecond,
						},
					},
				},
				Router: RouterConfig{
					Port:               defaultPort,
					MaxRequestBodySize: defaultMaxRequestBodySize,
					ReadTimeout:        defaultHTTPServerReadTimeout,
					WriteTimeout:       defaultHTTPServerWriteTimeout,
					IdleTimeout:        defaultHTTPServerIdleTimeout,
				},
				Logger: LoggerConfig{
					Level: "debug",
				},
			},
			wantErr: false,
		},
		{
			name:     "should return error for invalid logger level",
			filePath: "invalid_logger_level.yaml",
			yamlData: `
			morse_config:
			  full_node_config:
			    url: "https://pocket-rpc.liquify.com"
			    relay_signing_key: "40af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d38840af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d388"
			logger_config:
			  level: "invalid_level"
			`,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			if test.yamlData != "" {
				err := os.WriteFile(test.filePath, []byte(test.yamlData), 0644)
				defer os.Remove(test.filePath)
				c.NoError(err)
			}

			got, err := LoadGatewayConfigFromYAML(test.filePath)
			if test.wantErr {
				c.Error(err)
			} else {
				c.NoError(err)
				compareConfigs(c, test.want, got)
			}
		})
	}
}

func compareConfigs(c *require.Assertions, want, got GatewayConfig) {
	c.Equal(want.Router, got.Router)
	c.Equal(want.Logger, got.Logger)
	if want.MorseConfig != nil {
		compareMorseFullNodeConfig(c, want.MorseConfig.FullNodeConfig, got.MorseConfig.FullNodeConfig)
		c.Equal(want.MorseConfig.SignedAATs, got.MorseConfig.SignedAATs)
	}
	if want.ShannonConfig != nil {
		c.Equal(want.ShannonConfig, got.ShannonConfig)
	}
}

func compareMorseFullNodeConfig(c *require.Assertions, want, got morseprotocol.FullNodeConfig) {
	c.Equal(want.URL, got.URL)
	c.Equal(want.RelaySigningKey, got.RelaySigningKey)
	compareHTTPConfig(c, want.HttpConfig, got.HttpConfig)
}

func compareHTTPConfig(c *require.Assertions, want, got morseprotocol.HttpConfig) {
	c.Equal(want.Retries, got.Retries)
	c.Equal(want.Timeout, got.Timeout)
	if want.Transport != nil {
		wantTransport := want.Transport.(*http.Transport)
		gotTransport := got.Transport.(*http.Transport)
		c.Equal(wantTransport.MaxConnsPerHost, gotTransport.MaxConnsPerHost)
		c.Equal(wantTransport.MaxIdleConnsPerHost, gotTransport.MaxIdleConnsPerHost)
		c.Equal(wantTransport.MaxIdleConns, gotTransport.MaxIdleConns)
		c.Equal(wantTransport.IdleConnTimeout, gotTransport.IdleConnTimeout)
		// Compare DialContext properties instead of the function itself
		c.Equal(wantTransport.DialContext != nil, gotTransport.DialContext != nil)
	}
}
