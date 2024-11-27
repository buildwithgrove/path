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
			filePath: "./examples/.config.morse_example.yaml",
			want: GatewayConfig{
				MorseConfig: &morse.MorseGatewayConfig{
					FullNodeConfig: morseprotocol.FullNodeConfig{
						URL:             "https://pocket-network-full-full-node-url.io",
						RelaySigningKey: "ea5768843c5ad06897d43faa70b9ad39f143f3c7be8930d2b380022dbfbde3858b96695a0acc3458f84c2062c6d04f0dc44cf012d4244c688a60060f87940ef5",
						HttpConfig: morseprotocol.HttpConfig{
							Retries: 3,
							Timeout: 5000 * time.Millisecond,
						},
					},
					SignedAATs: map[string]morse.SignedAAT{
						"d39c0468509264eb455f8b7593b1967180058c89": {
							ClientPublicKey:      "22632c2191c3cd27b0c4c42509f5c349be7a1d3f5258fc34986df543f0e5a82c",
							ApplicationPublicKey: "e9b360514aca1f544801dc25353e9e5d73783e41562f34f99d2b130301ab00b1",
							ApplicationSignature: "4c6abff229a13f9b0f1915f6d1353f77bc96a2e317645c235ea0bb635e2e554b4cb384c516550f86535e683be3e0fba69e39ae608aa51428abbb8b28036d02ae",
						},
					},
				},
				Services: map[protocol.ServiceID]ServiceConfig{
					"F00C": {
						Alias:          "eth",
						RequestTimeout: 3000 * time.Millisecond,
					},
					"0001": {}, // Example of a service with no additional configuration
				},
				HydratorConfig: EndpointHydratorConfig{
					ServiceIDs: []protocol.ServiceID{"F00C"},
				},
				Router: RouterConfig{
					Port:               defaultPort,
					MaxRequestBodySize: defaultMaxRequestBodySize,
					ReadTimeout:        defaultReadTimeout,
					WriteTimeout:       defaultWriteTimeout,
					IdleTimeout:        defaultIdleTimeout,
				},
				serviceAliases: map[string]protocol.ServiceID{
					"eth": "F00C",
				},
			},
			wantErr: false,
		},
		{
			name:     "should load valid shannon config without error",
			filePath: "./examples/.config.shannon_example.yaml",
			want: GatewayConfig{
				ShannonConfig: &shannon.ShannonGatewayConfig{
					FullNodeConfig: shannonprotocol.FullNodeConfig{
						RpcURL: "https://testnet-validated-validator-rpc.poktroll.com",
						GRPCConfig: shannonprotocol.GRPCConfig{
							HostPort: "testnet-validated-validator-grpc.poktroll.com:443",
						},
					},
					GatewayConfig: shannonprotocol.GatewayConfig{
						GatewayPrivateKeyHex: "cf09805c952fa999e9a63a9f434147b0a5abfd10f268879694c6b5a70e1ae177",
						GatewayAddress:       "pokt1up7zlytnmvlsuxzpzvlrta95347w322adsxslw",
						GatewayMode:          protocol.GatewayModeCentralized,
						OwnedAppsPrivateKeysHex: []string{
							"cf09805c952fa999e9a63a9f434147b0a5abfd10f268879694c6b5a70e1ae177",
						},
					},
				},
				Services: map[protocol.ServiceID]ServiceConfig{
					"gatewaye2e": {
						Alias: "test-service",
					},
					"0021": {
						Alias:          "eth",
						RequestTimeout: 3000 * time.Millisecond,
					},
					"0001": {}, // Example of a service with no additional configuration
				},
				HydratorConfig: EndpointHydratorConfig{
					ServiceIDs: []protocol.ServiceID{"0021"},
				},
				Router: RouterConfig{
					Port:               defaultPort,
					MaxRequestBodySize: defaultMaxRequestBodySize,
					ReadTimeout:        defaultReadTimeout,
					WriteTimeout:       defaultWriteTimeout,
					IdleTimeout:        defaultIdleTimeout,
				},
				serviceAliases: map[string]protocol.ServiceID{
					"test-service": "gatewaye2e",
					"eth":          "0021",
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
			name:     "should return error for duplicate service alias",
			filePath: "duplicate_service_alias.yaml",
			yamlData: `
			morse_config:
			  serviceAliases:
			    eth: "0021"
			    eth: "0022"
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

func Test_GetServiceIDFromAlias(t *testing.T) {
	test := []struct {
		name   string
		config GatewayConfig
		alias  string
		want   protocol.ServiceID
		ok     bool
	}{
		{
			name: "should return service ID for existing alias",
			config: GatewayConfig{
				serviceAliases: map[string]protocol.ServiceID{
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
				serviceAliases: map[string]protocol.ServiceID{
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
	if want.MorseConfig != nil {
		compareMorseFullNodeConfig(c, want.MorseConfig.FullNodeConfig, got.MorseConfig.FullNodeConfig)
		c.Equal(want.MorseConfig.SignedAATs, got.MorseConfig.SignedAATs)
	}
	if want.ShannonConfig != nil {
		c.Equal(want.ShannonConfig, got.ShannonConfig)
	}
	c.Equal(want.serviceAliases, got.serviceAliases)
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
