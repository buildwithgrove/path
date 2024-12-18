package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/buildwithgrove/path/config/morse"
	"github.com/buildwithgrove/path/config/shannon"
	"github.com/buildwithgrove/path/protocol"
)

/* ---------------------------------  Gateway Config Struct -------------------------------- */

// GatewayConfig is the top level struct that contains configuration details
// that which are parsed from a YAML config file. It contains all the various
// configuration details that are needed to operate a gateway.
type (
	GatewayConfig struct {
		// Only one of the following configs may be set
		MorseConfig   *morse.MorseGatewayConfig     `yaml:"morse_config"`
		ShannonConfig *shannon.ShannonGatewayConfig `yaml:"shannon_config"`

		Services        map[protocol.ServiceID]ServiceConfig `yaml:"services"`
		Router          RouterConfig                         `yaml:"router_config"`
		HydratorConfig  EndpointHydratorConfig               `yaml:"hydrator_config"`
		MessagingConfig MessagingConfig                      `yaml:"messaging_config"`
	}
	ServiceConfig struct {
		RequestTimeout time.Duration `yaml:"request_timeout"`
	}
)

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

	// hydrate required fields and set defaults for optional fields
	config.hydrateRouterConfig()

	return config, config.validate()
}

/* --------------------------------- Gateway Config Methods -------------------------------- */

func (c GatewayConfig) GetShannonConfig() *shannon.ShannonGatewayConfig {
	return c.ShannonConfig
}

func (c GatewayConfig) GetMorseConfig() *morse.MorseGatewayConfig {
	return c.MorseConfig
}

func (c GatewayConfig) GetRouterConfig() RouterConfig {
	return c.Router
}

// GetEnabledServiceIDs() returns the list of enabled service IDs.
func (c GatewayConfig) GetEnabledServiceIDs() []protocol.ServiceID {
	var enabledServices []protocol.ServiceID
	for serviceID := range c.Services {
		enabledServices = append(enabledServices, serviceID)
	}
	return enabledServices
}

/* --------------------------------- Gateway Config Hydration Helpers -------------------------------- */

func (c *GatewayConfig) hydrateRouterConfig() {
	c.Router.hydrateRouterDefaults()
}

/* --------------------------------- Gateway Config Validation Helpers -------------------------------- */

func (c GatewayConfig) validate() error {
	if err := c.validateProtocolConfig(); err != nil {
		return err
	}
	if err := c.validateServiceConfig(); err != nil {
		return err
	}
	return nil
}

// validateProtocolConfig checks if the protocol configuration is valid, by both performing validation on the
// protocol specific config and ensuring that the correct protocol specific config is set.
func (c GatewayConfig) validateProtocolConfig() error {
	switch {
	case c.MorseConfig != nil && c.ShannonConfig != nil:
		return errors.New("only one of morse or shannon config may be set")
	case c.MorseConfig != nil:
		return c.MorseConfig.Validate()
	case c.ShannonConfig != nil:
		return c.ShannonConfig.Validate()
	default:
		return errors.New("no protocol configured")
	}
}

func (c GatewayConfig) validateServiceConfig() error {
	if len(c.Services) == 0 {
		return fmt.Errorf("at least one service must be configured")
	}
	return nil
}
