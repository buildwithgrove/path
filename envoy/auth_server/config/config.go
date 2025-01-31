// config package handles loading and validating the auth server configuration
// from PATH's .config.yaml file to set up the auth servers' configurable values.
package config

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"

	"github.com/buildwithgrove/path/envoy/auth_server/auth"
)

const (
	defaultEndpointIDExtractorType = auth.EndpointIDExtractorTypeURLPath
	defaultPort                    = 10003
)

var grpcHostPortPattern = "^[^:]+:[0-9]+$"

type (
	// TODO_DISCUSS/TODO_MVP(@commoddity, @olshansk, @adshmh): See the discussion
	// in this thread to make a decision if `auth_server_config` should be moved
	// out into a separate configuration.
	// https://github.com/buildwithgrove/path/pull/108#discussion_r1893326146

	// GatewayConfig is the top level struct that contains configuration details
	// that which are parsed from a YAML config file. For the purposes of this
	// files, only the `auth_server` configurable needs to be loaded from the YAML.
	GatewayConfig struct {
		AuthServerConfig AuthServerConfig `yaml:"auth_server_config"`
	}

	// AuthServerConfig contains all the configurable values for the auth server.
	AuthServerConfig struct {
		// GRPCHostPort is the host and port of the Remote gRPC Server (eg. PADS)
		// that the auth server reads GatewayEndpoint data from.
		GRPCHostPort string `yaml:"grpc_host_port"`
		// GRPCUseInsecureCredentials should be set only if the gRPC server does not
		// use a TLS-enabled connection. Defaults to false.
		GRPCUseInsecureCredentials bool `yaml:"grpc_use_insecure_credentials"`
		// EndpointIDExtractorType determines which method the auth server will use
		// to extract the endpoint ID from the request. Options are:
		// - URLPath: Extracts the endpoint ID from the URL path.
		// - Header: Extracts the endpoint ID from the request header.
		EndpointIDExtractorType auth.EndpointIDExtractorType `yaml:"endpoint_id_extractor_type"`
		// ServiceAliases is a map of service IDs to their aliases.
		ServiceAliases map[string]string `yaml:"service_aliases"`
		// Port is the port that the auth server will listen on.
		Port int `yaml:"port"`
	}
)

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

	// Perform regex validation on GRPCHostPort
	matched, err := regexp.MatchString(grpcHostPortPattern, c.GRPCHostPort)
	if err != nil {
		return fmt.Errorf("failed to validate grpc_host_port: %v", err)
	}
	if !matched {
		return fmt.Errorf("grpc_host_port does not match the required pattern")
	}

	return nil
}
