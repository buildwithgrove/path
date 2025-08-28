package shannon

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"slices"
	"strings"
	"time"

	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"

	"github.com/buildwithgrove/path/network/grpc"
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

	// Default session rollover blocks is the default value for the session rollover blocks config.
	defaultSessionRolloverBlocks = 10
)

var (
	ErrShannonInvalidGatewayPrivateKey                = errors.New("invalid shannon gateway private key")
	ErrShannonInvalidGatewayAddress                   = errors.New("invalid shannon gateway address")
	ErrShannonInvalidNodeUrl                          = errors.New("invalid shannon node URL")
	ErrShannonInvalidGrpcHostPort                     = errors.New("invalid shannon grpc host:port")
	ErrShannonUnsupportedGatewayMode                  = errors.New("invalid shannon gateway mode")
	ErrShannonCentralizedGatewayModeRequiresOwnedApps = errors.New("shannon Centralized gateway mode requires at-least 1 owned app")
	ErrShannonCacheConfigSetForLazyMode               = errors.New("cache config cannot be set for lazy mode")
	ErrShannonInvalidServiceFallback                  = errors.New("invalid service fallback configuration")
	ErrShannonInvalidSessionRolloverBlocks            = errors.New("session_rollover_blocks must be positive")
)

type (
	FullNodeConfig struct {
		RpcURL     string          `yaml:"rpc_url"`
		GRPCConfig grpc.GRPCConfig `yaml:"grpc_config"`

		// LazyMode, if set to true, will disable all caching of onchain data. For
		// example, this disables caching of apps and sessions.
		LazyMode bool `yaml:"lazy_mode" default:"true"`

		// Configuration options for the cache when LazyMode is false
		CacheConfig CacheConfig `yaml:"cache_config"`

		// SessionRolloverBlocks is a temporary fix to handle session rollover issues.
		// TODO_TECHDEBT(@commoddity): Should be removed when the rollover issue is solved at the protocol level.
		SessionRolloverBlocks int64 `yaml:"session_rollover_blocks"`
	}

	CacheConfig struct {
		SessionTTL time.Duration `yaml:"session_ttl"`
	}

	GatewayConfig struct {
		GatewayMode             protocol.GatewayMode `yaml:"gateway_mode"`
		GatewayAddress          string               `yaml:"gateway_address"`
		GatewayPrivateKeyHex    string               `yaml:"gateway_private_key_hex"`
		OwnedAppsPrivateKeysHex []string             `yaml:"owned_apps_private_keys_hex"`
		ServiceFallback         []ServiceFallback    `yaml:"service_fallback"`
	}

	// TODO_TECHDEBT(@adshmh): Make configuration and implementation explicit:
	// - Criteria to decide whether the "fallback" URL should be used at all.
	// - Criteria to decide the order in which a Shannon endpoint vs. a fallback URL should be used.
	// - Support "weighted" distribution to Shannon endpoints vs. "fallback" URLs.
	//
	// ServiceFallback is a configuration struct for specifying fallback endpoints for a service.
	ServiceFallback struct {
		ServiceID         protocol.ServiceID  `yaml:"service_id"`
		FallbackEndpoints []map[string]string `yaml:"fallback_endpoints"`
		// If true, all traffic will be sent to the fallback endpoints for the service,
		// regardless of the health of the protocol endpoints.
		SendAllTraffic bool `yaml:"send_all_traffic"`
	}
)

// defaultURLKey is the key for the default URL in the fallback endpoints map.
//   - If a service only supports one RPC type, the default URL is used for all requests.
//   - If a service supports multiple RPC types, the default URL is not used for requests.
//   - In all cases, the default URL is used as an identifier in the EndpointAddr.
const defaultURLKey = "default_url"

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

	if err := gc.validateServiceFallback(); err != nil {
		return err
	}

	return nil
}

// validateServiceFallback validates the service fallback configuration.
// It checks for duplicate service IDs, at-least one fallback URL, and valid fallback URLs.
func (gc GatewayConfig) validateServiceFallback() error {
	seenServiceIDs := make(map[protocol.ServiceID]struct{})

	for _, serviceFallback := range gc.ServiceFallback {
		if serviceFallback.ServiceID == "" {
			return fmt.Errorf("%w: service ID is required", ErrShannonInvalidServiceFallback)
		}

		// Check for duplicate service IDs
		if _, exists := seenServiceIDs[serviceFallback.ServiceID]; exists {
			return fmt.Errorf("%w: duplicate service ID '%s' found in service_fallback configuration",
				ErrShannonInvalidServiceFallback, serviceFallback.ServiceID)
		}
		seenServiceIDs[serviceFallback.ServiceID] = struct{}{}

		// Check that at least one fallback endpoint is defined
		if len(serviceFallback.FallbackEndpoints) == 0 {
			return fmt.Errorf("%w: at-least one fallback endpoint is required for service '%s'", ErrShannonInvalidServiceFallback, serviceFallback.ServiceID)
		}

		// Validate all fallback endpoints
		for i, endpointMap := range serviceFallback.FallbackEndpoints {
			if len(endpointMap) == 0 {
				return fmt.Errorf("%w: fallback endpoint %d is empty for service '%s'", ErrShannonInvalidServiceFallback, i, serviceFallback.ServiceID)
			}

			for rpcType, url := range endpointMap {
				// Skip default_url as it's not an RPC type
				if rpcType == defaultURLKey {
					if url == "" {
						return fmt.Errorf("%w: default_url is required for service '%s' fallback endpoint %d",
							ErrShannonInvalidServiceFallback, serviceFallback.ServiceID, i)
					}
					if !isValidURL(url) {
						return fmt.Errorf("%w: invalid default_url '%s' for service '%s' fallback endpoint %d",
							ErrShannonInvalidServiceFallback, url, serviceFallback.ServiceID, i)
					}
					continue
				}

				// Validate RPC type
				_, err := sharedtypes.GetRPCTypeFromConfig(rpcType)
				if err != nil {
					return fmt.Errorf("%w: invalid RPC type '%s' for service '%s' fallback endpoint %d",
						ErrShannonInvalidServiceFallback, rpcType, serviceFallback.ServiceID, i)
				}

				// Validate URL
				if !isValidURL(url) {
					return fmt.Errorf("%w: invalid %s fallback endpoint URL '%s' for service '%s' fallback endpoint %d",
						ErrShannonInvalidServiceFallback, rpcType, url, serviceFallback.ServiceID, i)
				}
			}
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
	if c.SessionRolloverBlocks <= 0 {
		return ErrShannonInvalidSessionRolloverBlocks
	}
	if err := c.CacheConfig.validate(c.LazyMode); err != nil {
		return err
	}
	return nil
}

// getServiceFallbackMap returns the fallback endpoint information for each
// service ID from the YAML config, including the SendAllTraffic setting.
func (gc GatewayConfig) getServiceFallbackMap() map[protocol.ServiceID]serviceFallback {
	configs := make(map[protocol.ServiceID]serviceFallback, len(gc.ServiceFallback))

	for _, serviceFallbackConfig := range gc.ServiceFallback {
		endpoints := make(map[protocol.EndpointAddr]endpoint, len(serviceFallbackConfig.FallbackEndpoints))

		// Create fallback endpoints from the configuration
		for _, endpointMap := range serviceFallbackConfig.FallbackEndpoints {
			rpcTypeURLs := make(map[sharedtypes.RPCType]string, len(endpointMap))

			for rpcTypeStr, url := range endpointMap {
				// Convert string keys to RPC types
				rpcType, err := sharedtypes.GetRPCTypeFromConfig(rpcTypeStr)
				if err != nil {
					// This should not happen if validation passed, but skip invalid RPC types
					continue
				}
				rpcTypeURLs[rpcType] = url
			}

			// Create fallback endpoint struct from the configuration and add
			// it to the map of endpoints for the service by its EndpointAddr.
			fallbackEndpoint := fallbackEndpoint{
				defaultURL:  endpointMap[defaultURLKey],
				rpcTypeURLs: rpcTypeURLs,
			}
			endpoints[fallbackEndpoint.Addr()] = fallbackEndpoint
		}

		configs[serviceFallbackConfig.ServiceID] = serviceFallback{
			SendAllTraffic: serviceFallbackConfig.SendAllTraffic,
			Endpoints:      endpoints,
		}
	}

	return configs
}

// Session TTL should match the protocol's session length.
// TODO_NEXT(@commoddity): Session refresh handling should be significantly reworked as part of the next changes following PATH PR #297.
// The proposed change is to align session refreshes with actual session expiry time,
// using the session expiry block and the Shannon SDK's block client.
// When this is done, session cache TTL can be removed altogether.
const defaultSessionCacheTTL = 20 * time.Second

func (c *CacheConfig) validate(lazyMode bool) error {
	// Cannot set both lazy mode and cache configuration.
	if lazyMode && c.SessionTTL != 0 {
		return ErrShannonCacheConfigSetForLazyMode
	}
	return nil
}

func (c *CacheConfig) hydrateDefaults() CacheConfig {
	if c.SessionTTL == 0 {
		c.SessionTTL = defaultSessionCacheTTL
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

// HydrateDefaults applies default values to FullNodeConfig
func (fnc *FullNodeConfig) HydrateDefaults() {
	fnc.GRPCConfig = fnc.GRPCConfig.HydrateDefaults()
	fnc.CacheConfig = fnc.CacheConfig.hydrateDefaults()
	if fnc.SessionRolloverBlocks == 0 {
		fnc.SessionRolloverBlocks = defaultSessionRolloverBlocks
	}
}
