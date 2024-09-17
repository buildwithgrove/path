package shannon

import (
	"gopkg.in/yaml.v3"

	shannonRelayer "github.com/buildwithgrove/path/relayer/shannon"
)

// Fields that are unmarshaled from the config YAML must be capitalized.
type ShannonGatewayConfig struct {
	FullNodeConfig shannonRelayer.FullNodeConfig `yaml:"full_node_config"`
}

// UnmarshalYAML is a custom unmarshaller for GatewayConfig.
// It performs validation after unmarshalling the config.
func (c *ShannonGatewayConfig) UnmarshalYAML(value *yaml.Node) error {
	// Temp alias to avoid recursion; this is the recommend pattern for Go YAML custom unmarshalers
	type temp ShannonGatewayConfig
	var val struct {
		temp `yaml:",inline"`
	}
	if err := value.Decode(&val); err != nil {
		return err
	}
	*c = ShannonGatewayConfig(val.temp)
	return nil
}

// validate checks if the configuration is valid after loading it from the YAML file.
func (c ShannonGatewayConfig) Validate() error {
	// TODO_IMPROVE: implement YAML validation for all fields in the config,
	// including regex for GRPC host/port, etc.
	return c.FullNodeConfig.Validate()
}
