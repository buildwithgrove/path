package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/buildwithgrove/path/config/morse"
	"github.com/buildwithgrove/path/config/shannon"
	"github.com/buildwithgrove/path/config/utils"
	"github.com/buildwithgrove/path/relayer"
	"github.com/buildwithgrove/path/request"
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

		Services map[relayer.ServiceID]ServiceConfig `yaml:"services"`
		Router   RouterConfig                        `yaml:"router_config"`
		// UserDataConfig is optional and only used if user data handling is enabled
		// for the gateway by setting the 'user_data_config' field in the config YAML file.
		UserData *UserDataConfig `yaml:"user_data_config"`

		// A map from human readable aliases (e.g. eth-mainnet) to service ID (e.g. 0021)
		serviceAliases map[string]relayer.ServiceID
	}
	ServiceConfig struct {
		Alias          string        `yaml:"alias"`
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
	config.hydrateServiceAliases()
	config.hydrateRouterConfig()
	if config.UserDataEnabled() {
		config.hydrateUserDataConfig()
	}

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

// UserDataEnabled returns true if user data handling is enabled for the Gateway.
func (c GatewayConfig) UserDataEnabled() bool {
	return c.UserData != nil
}

func (c GatewayConfig) GetUserDataConfig() *UserDataConfig {
	return c.UserData
}

// GetServiceIDFromAlias retrieves the ServiceID associated with a given service alias.
//
// This method allows for the use of a user-friendly string service alias in the
// URL subdomain, enabling more user-friendly URLs. For example, instead of
// using a ServiceID like "0021", an alias such as "eth-mainnet" can be used,
// resulting in a URL like "eth-mainnet.rpc.gateway.io" instead of "0021.rpc.gateway.io".
func (c GatewayConfig) GetServiceIDFromAlias(alias string) (relayer.ServiceID, bool) {
	serviceID, ok := c.serviceAliases[alias]
	return serviceID, ok
}

// GetEnabledServiceConfigs returns all services enabled in the Gateway. This is used to
// construct the map of QoS services to be used by the QoS service provider.
func (c GatewayConfig) GetEnabledServiceConfigs() map[relayer.ServiceID]request.QoSServiceConfig {
	serviceConfigs := make(map[relayer.ServiceID]request.QoSServiceConfig, len(c.Services))
	for serviceID, service := range c.Services {
		serviceConfigs[serviceID] = request.QoSServiceConfig{
			RequestTimeout: service.RequestTimeout,
		}
	}
	return serviceConfigs
}

/* --------------------------------- Gateway Config Hydration Helpers -------------------------------- */

func (c *GatewayConfig) hydrateServiceAliases() {
	if c.serviceAliases == nil {
		c.serviceAliases = make(map[string]relayer.ServiceID)
	}
	for serviceID, service := range c.Services {
		if service.Alias != "" {
			c.serviceAliases[service.Alias] = serviceID
		}
	}
}

func (c *GatewayConfig) hydrateRouterConfig() {
	c.Router.hydrateRouterDefaults()
}

func (c *GatewayConfig) hydrateUserDataConfig() {
	c.UserData.hydrateDefaults()
}

/* --------------------------------- Gateway Config Validation Helpers -------------------------------- */

func (c GatewayConfig) validate() error {
	if err := c.validateProtocolConfig(); err != nil {
		return err
	}
	if err := c.validateServiceConfig(); err != nil {
		return err
	}
	if c.UserData != nil {
		if err := c.UserData.validate(); err != nil {
			return err
		}
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

	for _, service := range c.Services {
		if service.Alias != "" {
			if !utils.IsValidSubdomain(service.Alias) {
				return fmt.Errorf("invalid service alias %s: must be a valid URL subdomain", service.Alias)
			}
		}
		if err := c.validateProtocolConfig(); err != nil {
			return err
		}

	}

	return nil
}
