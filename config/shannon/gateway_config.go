package shannon

import (
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"

	"github.com/buildwithgrove/path/protocol"
	shannonprotocol "github.com/buildwithgrove/path/protocol/shannon"
)

var websocketURLRegex = regexp.MustCompile("^(wss|ws)://.*$")

// Fields that are unmarshaled from the config YAML must be capitalized.
type ShannonGatewayConfig struct {
	FullNodeConfig shannonprotocol.FullNodeConfig `yaml:"full_node_config"`
	GatewayConfig  shannonprotocol.GatewayConfig  `yaml:"gateway_config"`

	// WebsocketEndpointURL is a TEMPORARY workaround to allow users of PATH to enable
	// websocket connections to a user-provided websocket-enabled endpoint URL.
	// It is placed in the ShannonGatewayConfig struct to indicate that websockets will
	// only be supported by the Shannon protocol, never on Morse.
	// TODO_FUTURE(@commoddity)[WebSockets]: Remove this field once the Shannon protocol supports websocket connections.
	WebsocketEndpointURLs map[protocol.ServiceID]string `yaml:"websocket_endpoint_urls"`
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
	if err := c.FullNodeConfig.Validate(); err != nil {
		return err
	}

	if err := c.GatewayConfig.Validate(); err != nil {
		return err
	}

	// Validate WebsocketEndpointURLs
	// TODO_FUTURE(@commoddity)[WebSockets]: Remove this validation once the Shannon protocol supports websocket connections.
	if err := c.validateWebsocketEndpointURLs(); err != nil {
		return err
	}

	return nil
}

// validateWebsocketEndpointURLs checks if the WebsocketEndpointURLs are valid.
func (c ShannonGatewayConfig) validateWebsocketEndpointURLs() error {
	for _, url := range c.WebsocketEndpointURLs {
		if !websocketURLRegex.MatchString(url) {
			return fmt.Errorf("invalid websocket endpoint URL: %s", url)
		}
	}
	return nil
}
