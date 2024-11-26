package morse

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pokt-foundation/pocket-go/provider"
	"github.com/pokt-foundation/pocket-go/relayer"
	"github.com/pokt-foundation/pocket-go/signer"
)

// The FullNode interface defined by the Morse protocol struct is fulfilled by the fullNode struct below.
var _ FullNode = &fullNode{}

// NewFullNode returns a Morse full node to be used in the Morse relayer.
func NewFullNode(config FullNodeConfig) (FullNode, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("NewSdk: invalid SDK config: %w", err)
	}

	sdkProvider := provider.NewProvider(config.URL)
	if config.HttpConfig != (HttpConfig{}) {
		sdkProvider.UpdateRequestConfig(provider.RequestConfigOpts{
			Retries:   config.HttpConfig.Retries,
			Timeout:   config.HttpConfig.Timeout,
			Transport: config.HttpConfig.Transport,
		})
	}

	s, err := signer.NewSignerFromPrivateKey(config.RelaySigningKey)
	if err != nil {
		return nil, fmt.Errorf("NewSdk: error getting a signer: %w", err)
	}

	return &fullNode{
		Provider: sdkProvider,
		Relayer:  relayer.NewRelayer(s, sdkProvider),
	}, nil
}

// fullNode groups a pocket-go provider and a pocket-go relayer to offer a FullNode interface
// required by the Morse relayer.
type fullNode struct {
	*provider.Provider
	*relayer.Relayer
}

type HttpConfig struct {
	Retries   int
	Timeout   time.Duration
	Transport http.RoundTripper
}

type FullNodeConfig struct {
	URL        string     `yaml:"url"`
	HttpConfig HttpConfig `yaml:"http_config"`

	// RelaySigningKey is either:
	// A. An application's private key, or
	// B. Private key of the gateway that is a client for the application (as specificed by the AAT)
	RelaySigningKey string `yaml:"relay_signing_key"`
}

func (c FullNodeConfig) Validate() error {
	if c.RelaySigningKey == "" {
		return fmt.Errorf("Morse full node config invalid: no signing key specified")
	}

	return nil
}

// TODO_UPNEXT: add output, status code, and Morse-specific errors
// Evidence Sealed error, etc.
