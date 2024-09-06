package morse

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/pokt-foundation/pocket-go/provider"
	"github.com/pokt-foundation/portal-middleware/config/utils"
	"github.com/pokt-foundation/portal-middleware/relayer"
	morseRelayer "github.com/pokt-foundation/portal-middleware/relayer/morse"
)

const (
	defaultMaxConnsPerHost     = 100
	defaultMaxIdleConnsPerHost = 100
	defaultMaxIdleConns        = 10_000
	defaultIdleConnTimeout     = 90 * time.Second
	defaultDialTimeout         = 3 * time.Second
	defaultKeepAlive           = 30 * time.Second
)

const (
	relaySigningKeyLength = 128
	appAddressLength      = 40
)

// Fields that are unmarshaled from the config YAML must be capitalized.
type (
	MorseGatewayConfig struct {
		FullNodeConfig morseRelayer.FullNodeConfig   `yaml:"full_node_config"`
		Transport      FullNodeHTTPTransportConfig   `yaml:"transport_config"`
		SignedAATs     map[relayer.AppAddr]SignedAAT `yaml:"signed_aats"`
	}
	FullNodeHTTPTransportConfig struct {
		MaxConnsPerHost     int           `yaml:"max_conns_per_host"`
		MaxIdleConnsPerHost int           `yaml:"max_idle_conns_per_host"`
		MaxIdleConns        int           `yaml:"max_idle_conns"`
		IdleConnTimeout     time.Duration `yaml:"idle_conn_timeout"`
		DialTimeout         time.Duration `yaml:"dial_timeout"`
		KeepAlive           time.Duration `yaml:"keep_alive"`
	}
)

// UnmarshalYAML is a custom unmarshaller for MorseGatewayConfig.
// It populates the serviceAliases map, sets the transport, and performs validation after unmarshalling the config.
func (c *MorseGatewayConfig) UnmarshalYAML(value *yaml.Node) error {
	// Temp alias to avoid recursion; this is the recommend pattern for Go YAML custom unmarshalers
	type temp MorseGatewayConfig
	var val struct {
		temp `yaml:",inline"`
	}
	if err := value.Decode(&val); err != nil {
		return err
	}
	*c = MorseGatewayConfig(val.temp)
	c.hydrateTransport()
	return nil
}

// GetSignedAAT retrieves the Pocket Application Authentication Token (PocketAAT) for a given application ID.
//
// Parameters:
//   - appID: The unique public key of the application for which the PocketAAT is being requested.
//
// Returns:
//   - provider.PocketAAT: The PocketAAT associated with the given application ID.
//   - bool: A boolean indicating whether the PocketAAT was successfully retrieved.
func (c MorseGatewayConfig) GetSignedAAT(appID relayer.AppAddr) (provider.PocketAAT, bool) {
	application, ok := c.SignedAATs[appID]
	if !ok {
		return provider.PocketAAT{}, false
	}

	return application.AAT(), true
}

// GetFullNodeConfig returns the full node configuration.
func (c MorseGatewayConfig) GetFullNodeConfig() morseRelayer.FullNodeConfig {
	return c.FullNodeConfig
}

// validate checks if the configuration is valid after loading it from the YAML file.
func (c MorseGatewayConfig) Validate() error {
	if !utils.IsValidURL(c.FullNodeConfig.URL) {
		return fmt.Errorf("invalid full node URL %s: must be a valid URL", c.FullNodeConfig.URL)
	}
	if !utils.IsValidHex(c.FullNodeConfig.RelaySigningKey, relaySigningKeyLength) {
		return fmt.Errorf("invalid relay signing key %s: must be a %d character hex code", c.FullNodeConfig.RelaySigningKey, relaySigningKeyLength)
	}

	for appAddress, app := range c.SignedAATs {
		if !utils.IsValidHex(string(appAddress), appAddressLength) {
			return fmt.Errorf("invalid application address %s: must be a %d character hex code", appAddress, appAddressLength)
		}
		if err := app.validate(); err != nil {
			return fmt.Errorf("invalid application %s: %w", appAddress, err)
		}
	}

	return nil
}

// hydrateTransport prepares the http transport structure of the full node configuration.
// It sets default values where appropriate if omitted, and applies the transport configuration.
func (c *MorseGatewayConfig) hydrateTransport() {
	if c.Transport.MaxConnsPerHost == 0 {
		c.Transport.MaxConnsPerHost = defaultMaxConnsPerHost
	}
	if c.Transport.MaxIdleConnsPerHost == 0 {
		c.Transport.MaxIdleConnsPerHost = defaultMaxIdleConnsPerHost
	}
	if c.Transport.MaxIdleConns == 0 {
		c.Transport.MaxIdleConns = defaultMaxIdleConns
	}
	if c.Transport.IdleConnTimeout == 0 {
		c.Transport.IdleConnTimeout = defaultIdleConnTimeout
	}
	if c.Transport.DialTimeout == 0 {
		c.Transport.DialTimeout = defaultDialTimeout
	}
	if c.Transport.KeepAlive == 0 {
		c.Transport.KeepAlive = defaultKeepAlive
	}

	c.FullNodeConfig.HttpConfig.Transport = &http.Transport{
		MaxConnsPerHost:     c.Transport.MaxConnsPerHost,
		MaxIdleConnsPerHost: c.Transport.MaxIdleConnsPerHost,
		MaxIdleConns:        c.Transport.MaxIdleConns,
		IdleConnTimeout:     c.Transport.IdleConnTimeout,
		DialContext: (&net.Dialer{
			Timeout:   c.Transport.DialTimeout,
			KeepAlive: c.Transport.KeepAlive,
			DualStack: true,
		}).DialContext,
	}
	c.Transport = FullNodeHTTPTransportConfig{}
}
