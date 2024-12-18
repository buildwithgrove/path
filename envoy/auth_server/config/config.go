package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/buildwithgrove/path/envoy/auth_server/auth"
)

const (
	defaultEndpointIDExtractorType = auth.EndpointIDExtractorTypeURLPath
	defaultPort                    = 10003
)

type GatewayConfig struct {
	AuthServerConfig AuthServerConfig `yaml:"auth_server_config"`
}

type AuthServerConfig struct {
	GRPCHostPort               string                       `yaml:"grpc_host_port"`
	GRPCUseInsecureCredentials bool                         `yaml:"grpc_use_insecure_credentials"`
	EndpointIDExtractorType    auth.EndpointIDExtractorType `yaml:"endpoint_id_extractor_type"`
	Port                       int                          `yaml:"port"`
}

// LoadAuthServerConfigFromYAML reads a YAML configuration file from the specified path
// and unmarshals its content into a AuthServerConfig instance.
func LoadAuthServerConfigFromYAML(path string) (AuthServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AuthServerConfig{}, err
	}

	var config GatewayConfig
	if err = yaml.Unmarshal(data, &config); err != nil {
		return AuthServerConfig{}, err
	}

	authServerConfig := config.AuthServerConfig

	// Validate the loaded configuration
	if err = authServerConfig.validate(); err != nil {
		return AuthServerConfig{}, err
	}

	// Hydrate the configuration with default values
	authServerConfig.hydrateDefaults()

	return authServerConfig, nil
}

func (c *AuthServerConfig) hydrateDefaults() {
	if !c.EndpointIDExtractorType.IsValid() {
		fmt.Printf("invalid endpoint ID extractor type: %s, using default: %s\n",
			c.EndpointIDExtractorType, defaultEndpointIDExtractorType,
		)
		c.EndpointIDExtractorType = defaultEndpointIDExtractorType
	}
	if c.Port == 0 {
		c.Port = defaultPort
	}
}

func (c *AuthServerConfig) validate() error {
	if c.GRPCHostPort == "" {
		return fmt.Errorf("grpc_host_port is not set in the configuration")
	}
	return nil
}
