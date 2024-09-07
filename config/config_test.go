package config

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/pokt-foundation/pocket-go/provider"
	"github.com/stretchr/testify/require"

	"github.com/buildwithgrove/path/config/morse"
	"github.com/buildwithgrove/path/config/shannon"
	"github.com/buildwithgrove/path/relayer"
	morseRelayer "github.com/buildwithgrove/path/relayer/morse"
	shannonRelayer "github.com/buildwithgrove/path/relayer/shannon"
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
			filePath: "./testdata/morse.example.yaml",
			want: GatewayConfig{
				MorseConfig: &morse.MorseGatewayConfig{
					FullNodeConfig: morseRelayer.FullNodeConfig{
						URL:             "https://full-node-url.io",
						RelaySigningKey: "05d126124d35fd7c645b78bf3128b989d03fa2c38cd69a81742b0dedbf9ca05aab35ab6f5137076136d0ef926a37fb3ac70249c3b0266b95d4b5db85a11fef8e",
						HttpConfig: morseRelayer.HttpConfig{
							Retries: 3,
							Timeout: 5000 * time.Millisecond,
						},
						RequestConfig: provider.RequestConfigOpts{
							Retries: 3,
						},
					},
					SignedAATs: map[relayer.AppAddr]morse.SignedAAT{
						"af929e588bb37d8e6bbc8cb25ba4b4d9383f9238": {
							ClientPublicKey:      "a6258b46ecad0628b72099f91e87eef1b040a8747ed2d476f56ad359372bf619",
							ApplicationPublicKey: "5a8c62e4701f349a3b9288cfbd825db230a8ec74fd234e7cb0849e915bc6d6ce",
							ApplicationSignature: "57d73225f83383e93571d0178f01368f26af9e552aaf073233d54600b60464043ba7013633d082b05d03ac7271667b307b09f47b8ac04000b19205cc1f99555d",
						},
						"f9076ec39b2a495883eb59740d566d5fa2e2b222": {
							ClientPublicKey:      "8604213b0c1ec52b5ae43eb854ce486a3756ec97cc194f3afe518947766aac11",
							ApplicationPublicKey: "71dd0e166022f1665dbba91b223998b0f328e9af2193a363456412a8eb4272e4",
							ApplicationSignature: "bb04cb9cb34ea6e2d57fb679f7b1e73ff77992e0f39a1e7db0c8ed2a91aed3668d0b6399ea70614a0f51b714a3ad3bd3ca2bc4a75302c14ce207d44c738cdbbf",
						},
					},
				},
				Services: map[relayer.ServiceID]ServiceConfig{
					"0021": {
						Alias:          "eth-mainnet",
						RequestTimeout: 3000 * time.Millisecond,
					},
					"0001": {}, // Example of a service with no additional configuration
				},
				Router: RouterConfig{
					Port:               8080,
					MaxRequestBodySize: 512000,
					ReadTimeout:        5000 * time.Millisecond,
					WriteTimeout:       5000 * time.Millisecond,
					IdleTimeout:        5000 * time.Millisecond,
				},
				UserData: UserDataConfig{
					DBConnectionString:   "postgres://user:password@localhost:5432/database",
					CacheRefreshInterval: defaultCacheRefreshInterval,
				},
				serviceAliases: map[string]relayer.ServiceID{
					"eth-mainnet": "0021",
				},
			},
			wantErr: false,
		},
		{
			name:     "should load valid shannon config without error",
			filePath: "./testdata/shannon.example.yaml",
			want: GatewayConfig{
				ShannonConfig: &shannon.ShannonGatewayConfig{
					FullNodeConfig: shannonRelayer.FullNodeConfig{
						RpcURL: "https://rpc-url.io",
						GRPCConfig: shannonRelayer.GRPCConfig{
							HostPort: "grpc-url.io:443",
						},
						GatewayPrivateKey: "d5fcbfb894059a21e914a2d6bf1508319ce2b1b8878f15aa0c1cdf883feb018d",
						GatewayAddress:    "pokt1710ed9a8d0986d808e607c5815cc5a13f15dba",
						DelegatedApps: []string{
							"pokt1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p7q8r9s0",
							"pokt1u2v3w4x5y6z7a8b9c0d1e2f3g4h5i6j7k8l9m0",
						},
					},
				},
				Services: map[relayer.ServiceID]ServiceConfig{
					"0021": {
						Alias:          "eth-mainnet",
						RequestTimeout: 3000 * time.Millisecond,
					},
					"0001": {}, // Example of a service with no additional configuration
				},
				Router: RouterConfig{
					Port:               8080,
					MaxRequestBodySize: 512000,
					ReadTimeout:        5000 * time.Millisecond,
					WriteTimeout:       5000 * time.Millisecond,
					IdleTimeout:        5000 * time.Millisecond,
				},
				UserData: UserDataConfig{
					DBConnectionString:   "postgres://user:password@localhost:5432/database",
					CacheRefreshInterval: defaultCacheRefreshInterval,
				},
				serviceAliases: map[string]relayer.ServiceID{
					"eth-mainnet": "0021",
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
			    gateway_address: "pokt1710ed9a8d0986d808e607c5815cc5a13f15dba"
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
			    gateway_address: "invalid_gateway_address"
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
			name:     "should return error for invalid service alias",
			filePath: "invalid_service_alias.yaml",
			yamlData: `
			morse_config:
			  serviceAliases:
			    invalid_alias!@#:
			      id: "0001"
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
			name:     "should return error for invalid fallback URL",
			filePath: "invalid_fallback_url.yaml",
			yamlData: `
			morse_config:
			  services:
			    0001:
			      id: "0001"
			      config:
			        alias: "pokt-mainnet"
			        fallback_url: "invalid-url"
			        service_type_override: "REST"
			        request_timeout_seconds: 30
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
			    request_config:
			      retries: 3
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

func Test_GetServiceIDFromAlias(t *testing.T) {
	test := []struct {
		name   string
		config GatewayConfig
		alias  string
		want   relayer.ServiceID
		ok     bool
	}{
		{
			name: "should return service ID for existing alias",
			config: GatewayConfig{
				serviceAliases: map[string]relayer.ServiceID{
					"eth-mainnet": "0021",
				},
			},
			alias: "eth-mainnet",
			want:  "0021",
			ok:    true,
		},
		{
			name: "should return false for non-existing alias",
			config: GatewayConfig{
				serviceAliases: map[string]relayer.ServiceID{
					"eth-mainnet": "0021",
				},
			},
			alias: "btc-mainnet",
			want:  "",
			ok:    false,
		},
	}

	for _, test := range test {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			got, ok := test.config.GetServiceIDFromAlias(test.alias)
			c.Equal(test.ok, ok)
			if ok {
				c.Equal(test.want, got)
			}
		})
	}
}

func compareConfigs(c *require.Assertions, want, got GatewayConfig) {
	c.Equal(want.Router, got.Router)
	c.Equal(want.UserData, got.UserData)
	if want.MorseConfig != nil {
		compareMorseFullNodeConfig(c, want.MorseConfig.FullNodeConfig, got.MorseConfig.FullNodeConfig)
		c.Equal(want.MorseConfig.SignedAATs, got.MorseConfig.SignedAATs)
	}
	if want.ShannonConfig != nil {
		c.Equal(want.ShannonConfig, got.ShannonConfig)
	}
	c.Equal(want.serviceAliases, got.serviceAliases)
}

func compareMorseFullNodeConfig(c *require.Assertions, want, got morseRelayer.FullNodeConfig) {
	c.Equal(want.URL, got.URL)
	c.Equal(want.RelaySigningKey, got.RelaySigningKey)
	c.Equal(want.RequestConfig, got.RequestConfig)
	compareHTTPConfig(c, want.HttpConfig, got.HttpConfig)
}

func compareHTTPConfig(c *require.Assertions, want, got morseRelayer.HttpConfig) {
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
