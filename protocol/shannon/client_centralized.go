package shannon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"
	"github.com/pokt-network/shannon-sdk/client"
	"github.com/pokt-network/shannon-sdk/crypto"
)

var (
	// Centralized gateway mode: Error getting onchain data for app
	errProtocolContextSetupCentralizedAppFetchErr = errors.New("error getting onchain data for app owned by the gateway")
	// Centralized gateway mode app does not delegate to the gateway.
	errProtocolContextSetupCentralizedAppDelegation = errors.New("centralized gateway mode app does not delegate to the gateway")
	// Centralized gateway mode: no active sessions could be retrieved for the service.
	errProtocolContextSetupCentralizedNoSessions = errors.New("no active sessions could be retrieved for the service")
)

// centralizedGatewayClient implements the GatewayClient interface for Centralized Gateway Mode.
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
	*client.GatewayClient
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

// GetGatewayModeActiveSessions implements GatewayClient interface.
//   - Returns the set of permitted sessions under the Centralized gateway mode.
//   - Gateway address and owned apps addresses (specified in configs) are used to retrieve active sessions.
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
		return nil, fmt.Errorf("no owned apps for service %s", serviceID)
	}

	gatewayAddr := c.GetGatewayAddress()

	var ownedAppSessions []sessiontypes.Session

	// Loop over the address of apps owned by the gateway in Centralized gateway mode.
	for _, ownedAppAddr := range ownedAppsForService {
		logger.Info().Msgf("About to get a session for  owned app %s for service %s", ownedAppAddr, serviceID)

		// Retrieve the session for the owned app.
		session, err := c.GetSession(ctx, serviceID, ownedAppAddr)
		if err != nil {
			// Wrap the protocol context setup error.
			err = fmt.Errorf("%w: app: %s, error: %w",
				errProtocolContextSetupCentralizedAppFetchErr,
				ownedAppAddr,
				err,
			)
			logger.Error().Err(err).Msg(err.Error())
			return nil, err
		}

		app := session.Application

		// Verify the app delegates to the gateway	.
		if !c.gatewayHasDelegationForApp(app) {
			// Wrap the protocol context setup error.
			err := fmt.Errorf("%w: app: %s, gateway: %s",
				errProtocolContextSetupCentralizedAppDelegation,
				app.Address,
				gatewayAddr,
			)
			logger.Error().Msg(err.Error())
			return nil, err
		}

		ownedAppSessions = append(ownedAppSessions, session)
	}

	// If no sessions were found, return an error.
	if len(ownedAppSessions) == 0 {
		err := fmt.Errorf("%w: service %s.",
			errProtocolContextSetupCentralizedNoSessions,
			serviceID,
		)
		logger.Error().Msg(err.Error())
		return nil, err
	}

	logger.Info().Msgf("Successfully fetched %d sessions for %d owned apps for service %s.", len(ownedAppSessions), len(ownedAppsForService), serviceID)

	return ownedAppSessions, nil
}

// GetConfiguredServiceIDs returns the service IDs configured for the gateway.
func (c *centralizedGatewayClient) GetConfiguredServiceIDs() map[sdk.ServiceID]struct{} {
	servicesIDs := make(map[sdk.ServiceID]struct{})
	for serviceID := range c.ownedApps {
		servicesIDs[serviceID] = struct{}{}
	}
	return servicesIDs
}

// gatewayHasDelegationForApp returns true if the supplied application delegates to the supplied gateway address.
func (c *centralizedGatewayClient) gatewayHasDelegationForApp(app *apptypes.Application) bool {
	return slices.Contains(app.DelegateeGatewayAddresses, c.GetGatewayAddress())
}

// getOwnedApps:
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
func getOwnedApps(logger polylog.Logger, ownedAppsPrivateKeysHex []string, gatewayClient *client.GatewayClient) (map[sdk.ServiceID][]string, error) {
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
		// GetApp passthrough to the underlying full node to ensure the request
		// is a remote request and not using cached data.
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
