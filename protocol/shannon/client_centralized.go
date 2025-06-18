package shannon

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"
	"github.com/pokt-network/shannon-sdk/client"
	"github.com/pokt-network/shannon-sdk/crypto"
)

var errCentralizedGatewayClientNoOwnedAppsForService = errors.New("no owned apps for service")

// centralizedGatewayClient implements the GatewayClient interface for Centralized Gateway Mode.
//
// It embeds the GatewayClient interface from the Shannon SDK package, which provides the
// functionality needed by the gateway package for handling service requests.
//
// Centralized Gateway Mode - Shannon Protocol Integration
//
//   - PATH (Shannon instance) holds private keys for gateway operator's apps.
//   - All apps are owned by the gateway (PATH) operator.
//   - All apps delegate (onchain) to the gateway address.
//   - Each relay request is sent for one of these apps (owned by the gateway operator).
//   - Each relay is signed by the gateway's private key (via Shannon ring signatures)
//
// More details: https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5?pvs=4#122a36edfff680d4a0fff3a40dea543e
type centralizedGatewayClient struct {
	logger polylog.Logger

	// Embeds the GatewayClient interface from the Shannon SDK package, which provides the
	// functionality needed by the gateway package for handling service requests.
	*client.GatewayClient

	// In centralized mode, we need to know which apps are owned by the gateway operator.
	// This is used to retrieve the active sessions for the gateway operator's apps.
	ownedApps map[sdk.ServiceID][]string
}

// NewCentralizedGatewayClient creates a new centralizedGatewayClient instance.
func NewCentralizedGatewayClient(
	logger polylog.Logger,
	gatewayClient *client.GatewayClient,
	ownedAppsPrivateKeysHex []string,
) (*centralizedGatewayClient, error) {
	logger = logger.With("client_type", "centralized")

	// Build the owned apps map from the provided private keys
	ownedApps, err := getOwnedApps(logger, ownedAppsPrivateKeysHex, gatewayClient)
	if err != nil {
		return nil, fmt.Errorf("failed to build owned apps: %w", err)
	}

	return &centralizedGatewayClient{
		logger:        logger,
		GatewayClient: gatewayClient,
		ownedApps:     ownedApps,
	}, nil
}

// getGatewayModeActiveSessions implements GatewayClient interface.
//   - Returns the set of permitted sessions under the Centralized gateway mode.
//   - Owned app addresses are used to retrieve active sessions for a service.
func (c *centralizedGatewayClient) GetGatewayModeActiveSessions(
	ctx context.Context,
	serviceID sdk.ServiceID,
	httpReq *http.Request,
) ([]sessiontypes.Session, error) {
	logger := c.logger.With(
		"method", "GetActiveGatewaySessions",
		"service_id", string(serviceID),
	)
	logger.Debug().Msgf("fetching active sessions for the service %s.", serviceID)

	// TODO_TECHDEBT(@commoddity): if an owned app is re-staked for a
	// different service, PATH must be restarted for changes to take effect.
	ownedAppsForService, ok := c.ownedApps[serviceID]
	if !ok || len(ownedAppsForService) == 0 {
		return nil, fmt.Errorf("%w: service %s",
			errCentralizedGatewayClientNoOwnedAppsForService,
			serviceID,
		)
	}

	return c.GetActiveSessions(ctx, serviceID, ownedAppsForService)
}

// GetConfiguredServiceIDs returns the service IDs configured for the gateway.
func (c *centralizedGatewayClient) GetConfiguredServiceIDs() map[sdk.ServiceID]struct{} {
	servicesIDs := make(map[sdk.ServiceID]struct{})
	for serviceID := range c.ownedApps {
		servicesIDs[serviceID] = struct{}{}
	}
	return servicesIDs
}

// getOwnedApps is called only once when initializing the centralized gateway client.
//
//   - Returns list of apps owned by the gateway, built from supplied private keys
//   - Supplied private keys are ONLY used to build app addresses for relay signing
//   - Populates `appAddr` and `stakedServiceID` for each app
//
// ownedAppsMap maps service IDs to a list of app addresses owned by the gateway operator in Centralized Gateway Mode.
// Note that one service ID can have multiple apps owned by the gateway operator.
//
// Example:
//
//	{
//	  "anvil": ["pokt1...", "pokt2..."],
//	  "eth": ["pokt3...", "pokt4..."],
//	}
func getOwnedApps(
	logger polylog.Logger,
	ownedAppsPrivateKeysHex []string,
	gatewayClient *client.GatewayClient,
) (map[sdk.ServiceID][]string, error) {
	logger = logger.With("method", "getCentralizedModeOwnedApps")
	logger.Debug().Msg("Building the list of owned apps.")

	ownedApps := make(map[sdk.ServiceID][]string)
	for _, ownedAppPrivateKeyHex := range ownedAppsPrivateKeysHex {
		// Retrieve the app's secp256k1 private key from the hex string.
		ownedAppPrivateKey, err := crypto.GetSecp256k1PrivateKeyFromKeyHex(ownedAppPrivateKeyHex)
		if err != nil {
			logger.Error().Err(err).Msgf("error getting app private key from hex for app with private key %s", ownedAppPrivateKeyHex)
			return nil, err
		}

		// Retrieve the app's address from the private key.
		appAddr, err := crypto.GetAddressFromPrivateKey(ownedAppPrivateKey)
		if err != nil {
			logger.Error().Err(err).Msgf("error getting app address from private key for app with private key %s", ownedAppPrivateKeyHex)
			return nil, err
		}

		// Retrieve the app's onchain data.
		// GetApp passes through to the underlying full node
		//to ensure the request is not using cached data.
		app, err := gatewayClient.GetApp(context.Background(), appAddr)
		if err != nil {
			logger.Error().Err(err).Msgf("error getting onchain data for app with address %s", appAddr)
			return nil, err
		}

		// Retrieve the app's service configs.
		appServiceConfigs := app.GetServiceConfigs()
		if len(appServiceConfigs) != 1 {
			logger.Error().Msgf("should never happen: app with address %s is not staked for exactly one service but %d instead", appAddr, len(appServiceConfigs))
			return nil, fmt.Errorf("app with address %s is not staked for exactly one service", appAddr)
		}

		// Retrieve the app's service ID; each app is staked for exactly one service.
		appServiceConfig := appServiceConfigs[0]
		serviceID := sdk.ServiceID(appServiceConfig.GetServiceId())
		if serviceID == "" {
			logger.Error().Msgf("should never happen: app with address %s is staked for service with an empty ID", appAddr)
			return nil, fmt.Errorf("app with address %s is staked for service with an empty ID", appAddr)
		}

		// Add the app to the list of owned apps.
		ownedApps[serviceID] = append(ownedApps[serviceID], appAddr)
	}

	logger.Debug().Msgf("Successfully built the list of %d owned apps.", len(ownedApps))
	return ownedApps, nil
}
