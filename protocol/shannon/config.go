package shannon

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/buildwithgrove/path/protocol"
)

const (
	// Shannon uses secp256k1 key schemes (the cosmos default)
	// secp256k1 keys are 32 bytes -> 64 hexadecimal characters
	// Ref: https://docs.cosmos.network/v0.45/basics/accounts.html
	shannonPrivateKeyLengthHex = 64
	// secp256k1 keys are 20 bytes, but are then bech32 encoded -> 43 bytes
	// Ref: https://docs.cosmos.network/main/build/spec/addresses/bech32
	shannonAddressLengthBech32 = 43
)

var (
	ErrShannonInvalidGatewayPrivateKey                = errors.New("invalid shannon gateway private key")
	ErrShannonInvalidGatewayAddress                   = errors.New("invalid shannon gateway address")
	ErrShannonInvalidNodeUrl                          = errors.New("invalid shannon node URL")
	ErrShannonInvalidGrpcHostPort                     = errors.New("invalid shannon grpc host:port")
	ErrShannonUnsupportedGatewayMode                  = errors.New("invalid shannon gateway mode")
	ErrShannonCentralizedGatewayModeRequiresOwnedApps = errors.New("shannon Centralized gateway mode requires at-least 1 owned app")
)

type (
	FullNodeConfig struct {
		RpcURL     string     `yaml:"rpc_url"`
		GRPCConfig GRPCConfig `yaml:"grpc_config"`

		// LazyMode, if set, will disable all caching of onchain data, specifically apps and sessions.
		// This enables supporting short block times, e.g. when running E2E tests on LocalNet.
		LazyMode bool `yaml:"lazy_mode"`
	}

	GatewayConfig struct {
		GatewayMode             protocol.GatewayMode `yaml:"gateway_mode"`
		GatewayAddress          string               `yaml:"gateway_address"`
		GatewayPrivateKeyHex    string               `yaml:"gateway_hex_private_key"`
		OwnedAppsPrivateKeysHex []string             `yaml:"owned_apps_hex_private_keys"`
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

func (gc GatewayConfig) Validate() error {
	if len(gc.GatewayPrivateKeyHex) != shannonPrivateKeyLengthHex {
		return ErrShannonInvalidGatewayPrivateKey
	}
	if len(gc.GatewayAddress) != shannonAddressLengthBech32 {
		return ErrShannonInvalidGatewayAddress
	}
	if !strings.HasPrefix(gc.GatewayAddress, "pokt1") {
		return ErrShannonInvalidGatewayAddress
	}

	if !slices.Contains(supportedGatewayModes(), gc.GatewayMode) {
		return fmt.Errorf("%w: %s", ErrShannonUnsupportedGatewayMode, gc.GatewayMode)
	}

	if gc.GatewayMode == protocol.GatewayModeCentralized && len(gc.OwnedAppsPrivateKeysHex) == 0 {
		return ErrShannonCentralizedGatewayModeRequiresOwnedApps
	}

	for index, privKey := range gc.OwnedAppsPrivateKeysHex {
		if len(privKey) != shannonPrivateKeyLengthHex {
			return fmt.Errorf("%w: invalid owned app private key at index: %d", ErrShannonInvalidGatewayPrivateKey, index)
		}
	}

	return nil
}

func (c FullNodeConfig) Validate() error {
	if !isValidURL(c.RpcURL) {
		return ErrShannonInvalidNodeUrl
	}
	if !isValidHostPort(c.GRPCConfig.HostPort) {
		return ErrShannonInvalidGrpcHostPort
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

// isValidURL returns true if the supplied URL string can be parsed into a valid URL accepted by the Shannon SDK.
func isValidURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	if u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

// isValidHostPort returns true if the supplied string can be parsed into a host and port combination.
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