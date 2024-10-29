package shannon

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

const (
	gatewayPrivateKeyLength = 64
	addressLength           = 43
)

var (
	ErrShannonInvalidGatewayPrivateKey = errors.New("invalid shannon gateway private key")
	ErrShannonInvalidGatewayAddress    = errors.New("invalid shannon gateway address")
	ErrShannonInvalidNodeUrl           = errors.New("invalid shannon node URL")
	ErrShannonInvalidGrpcHostPort      = errors.New("invalid shannon grpc host:port")
)

type (
	FullNodeConfig struct {
		RpcURL            string     `yaml:"rpc_url"`
		GRPCConfig        GRPCConfig `yaml:"grpc_config"`
		GatewayAddress    string     `yaml:"gateway_address"`
		GatewayPrivateKey string     `yaml:"gateway_private_key"`
		// TODO_UPNEXT(@adshmh): use private keys of owned apps in the configuration, and only use an app if it
		// can be verified, i.e. if the public key derived from the stored private key matches the onchain app data.
		// A list of addresses of onchain Applications delegated to the Gateway.
		DelegatedApps []string `yaml:"delegated_app_addresses"`

		// LazyMode, if set, will disable all caching of onchain data, specifically apps and sessions.
		// This enables supporting short block times, e.g. when running E2E tests on LocalNet.
		LazyMode bool `yaml:"lazy_mode"`
	}

	// TODO_TECHDEBT(@adshmh): Move this and related helpers into a new `grpc` package.
	GRPCConfig struct {
		HostPort          string        `yaml:"host_port"`
		Insecure          bool          `yaml:"insecure"`
		BackoffBaseDelay  time.Duration `yaml:"backoff_base_delay"`
		BackoffMaxDelay   time.Duration `yaml:"backoff_max_delay"`
		MinConnectTimeout time.Duration `yaml:"min_connect_timeout"`
		KeepAliveTime     time.Duration `yaml:"keep_alive_time"`
		KeepAliveTimeout  time.Duration `yaml:"keep_alive_timeout"`
	}
)

func (c FullNodeConfig) Validate() error {
	if len(c.GatewayPrivateKey) != gatewayPrivateKeyLength {
		return ErrShannonInvalidGatewayPrivateKey
	}
	if len(c.GatewayAddress) != addressLength {
		return ErrShannonInvalidGatewayAddress
	}
	if !strings.HasPrefix(c.GatewayAddress, "pokt1") {
		return ErrShannonInvalidGatewayAddress
	}
	if !isValidURL(c.RpcURL) {
		return ErrShannonInvalidNodeUrl
	}
	if !isValidURLWithPort(c.GRPCConfig.HostPort) {
		return ErrShannonInvalidGrpcHostPort
	}
	for _, addr := range c.DelegatedApps {
		if len(addr) != addressLength {
			return fmt.Errorf("invalid delegated app address: %s", addr)
		}
	}
	return nil
}

// TODO_TECHDEBT(@adshmh): add a new `grpc` package to handle all GRPC related functionality and configuration.
// The config package is not a good fit for this, because it is designed to build the configuration structs for other packages,
// and so it has dependencies on all other packages, including `relayer/shannon`. Therefore, no packages except `cmd` can have a dependency
// on the `config` package.
const (
	defaultBackoffBaseDelay  = 1 * time.Second
	defaultBackoffMaxDelay   = 120 * time.Second
	defaultMinConnectTimeout = 20 * time.Second
	defaultKeepAliveTime     = 20 * time.Second
	defaultKeepAliveTimeout  = 20 * time.Second
)

func (c *GRPCConfig) hydrateDefaults() GRPCConfig {
	if c.BackoffBaseDelay == 0 {
		c.BackoffBaseDelay = defaultBackoffBaseDelay
	}
	if c.BackoffMaxDelay == 0 {
		c.BackoffMaxDelay = defaultBackoffMaxDelay
	}
	if c.MinConnectTimeout == 0 {
		c.MinConnectTimeout = defaultMinConnectTimeout
	}
	if c.KeepAliveTime == 0 {
		c.KeepAliveTime = defaultKeepAliveTime
	}
	if c.KeepAliveTimeout == 0 {
		c.KeepAliveTimeout = defaultKeepAliveTimeout
	}
	return *c
}

// isValidUrl checks whether the provided string is a formatted as the poktroll SDK expects
// The gRPC url requires a port
func isValidUrl(urlToCheck string, needPort bool) bool {
	u, err := url.Parse(urlToCheck)
	if err != nil {
		return false
	}

	if u.Scheme == "" || u.Host == "" {
		return false
	}

	if !needPort {
		return true
	}

	_, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return false
	}

	if port == "" {
		return false
	}

	return true
}

// isValidURL returns true if the supplied URL string can be parsed into a valid URL.
func isValidURL(url string) bool {
	_, isValid := parseURL(url)
	return isValid
}

// isValidURLWithPort returns true if the supplied URL string can be parsed into a valid URL and a port.
func isValidURLWithPort(url string) bool {
	parsedURL, isValid := parseURL(url)
	if !isValid {
		return false
	}

	return isValidHostPort(parsedURL.Host)
}

// parseURL parses a string into a URL, and returns the parsed value, and a boolean indicating whether the URL string is valid.
func parseURL(urlStr string) (*url.URL, bool) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, false
	}

	if u.Scheme == "" || u.Host == "" {
		return u, false
	}

	return u, true
}

// isValidHostPort returns true if the supplied string can be parsed into a host and port combination.
// The input string can be taken from the `Host` field of a parsed net/url.URL struct.
func isValidHostPort(hostPort string) bool {
	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return false
	}

	if host == "" || port == "" {
		return false
	}

	return true
}
