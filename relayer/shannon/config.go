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
	// TODO_DISCUSS: move this (and the morse FullNodeConfig) to the config package?
	FullNodeConfig struct {
		RpcURL            string     `yaml:"rpc_url"`
		GRPCConfig        GRPCConfig `yaml:"grpc_config"`
		GatewayAddress    string     `yaml:"gateway_address"`
		GatewayPrivateKey string     `yaml:"gateway_private_key"`
		// A list of addresses of onchain Applications delegated to the Gateway.
		DelegatedApps []string `yaml:"delegated_app_addresses"`
	}

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

// TODO_IMPROVE: move this to the config package?
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
	if !isValidUrl(c.RpcURL, false) {
		return ErrShannonInvalidNodeUrl
	}
	if !isValidGrpcHostPort(c.GRPCConfig.HostPort) {
		return ErrShannonInvalidGrpcHostPort
	}
	for _, addr := range c.DelegatedApps {
		if len(addr) != addressLength {
			return fmt.Errorf("invalid delegated app address: %s", addr)
		}
	}
	return nil
}

// TODO_IMPROVE: move this to the config package?
const (
	defaultBackoffBaseDelay  = 1 * time.Second
	defaultBackoffMaxDelay   = 120 * time.Second
	defaultMinConnectTimeout = 20 * time.Second
	defaultKeepAliveTime     = 20 * time.Second
	defaultKeepAliveTimeout  = 20 * time.Second
)

// TODO_IMPROVE: move this to the config package?
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

func isValidGrpcHostPort(hostPort string) bool {
	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return false
	}

	if host == "" || port == "" {
		return false
	}

	return true
}
