package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/buildwithgrove/path/config/shannon"
)

/* ---------------------------------  Gateway Config Struct -------------------------------- */

// GatewayConfig contains all configuration details needed to operate a gateway,
// parsed from a YAML config file.
type GatewayConfig struct {
	ShannonConfig      *shannon.ShannonGatewayConfig `yaml:"shannon_config"`
	Router             RouterConfig                  `yaml:"router_config"`
	Logger             LoggerConfig                  `yaml:"logger_config"`
	HydratorConfig     EndpointHydratorConfig        `yaml:"hydrator_config"`
	MessagingConfig    MessagingConfig               `yaml:"messaging_config"`
	DataReporterConfig HTTPDataReporterConfig        `yaml:"data_reporter_config"`
}

// LoadGatewayConfigFromYAML reads a YAML configuration file from the specified path
// and unmarshals its content into a GatewayConfig instance.
func LoadGatewayConfigFromYAML(path string) (GatewayConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return GatewayConfig{}, err
	}

	var config GatewayConfig
	if err = yaml.Unmarshal(data, &config); err != nil {
		return GatewayConfig{}, err
	}

	config.hydrateDefaults()

	return config, config.validate()
}

/* --------------------------------- Gateway Config Methods -------------------------------- */

func (c GatewayConfig) GetGatewayConfig() *shannon.ShannonGatewayConfig {
	return c.ShannonConfig
}

func (c GatewayConfig) GetRouterConfig() RouterConfig {
	return c.Router
}

/* --------------------------------- Gateway Config Hydration Helpers -------------------------------- */

func (c *GatewayConfig) hydrateDefaults() {
	c.Router.hydrateRouterDefaults()
	c.Logger.hydrateLoggerDefaults()
	c.HydratorConfig.hydrateHydratorDefaults()
	c.ShannonConfig.FullNodeConfig.HydrateDefaults()
}

/* --------------------------------- Gateway Config Validation Helpers -------------------------------- */

func (c GatewayConfig) validate() error {
	if err := c.validateProtocolConfig(); err != nil {
		return err
	}
	if err := c.Logger.Validate(); err != nil {
		return err
	}
	return nil
}

// validateProtocolConfig checks if the protocol configuration is valid, by both performing validation on the
// protocol specific config and ensuring that the correct protocol specific config is set.
func (c GatewayConfig) validateProtocolConfig() error {
	if c.ShannonConfig == nil {
		return fmt.Errorf("protocol configuration is required")
	}
	return c.ShannonConfig.Validate()
}
