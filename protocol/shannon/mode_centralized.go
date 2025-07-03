package shannon

// Centralized Gateway Mode - Shannon Protocol Integration
//
// - PATH (Shannon instance) holds private keys for gateway operator's apps
// - All apps are owned by the gateway (PATH) operator
// - All apps delegate (onchain) to the gateway address
// - Each relay request is sent for one of these apps (owned by the gateway operator)
// - Each relay is signed by the gateway's private key (via Shannon ring signatures)
// More details: https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5?pvs=4#122a36edfff680d4a0fff3a40dea543e

import (
	"context"
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/protocol/crypto"
)

// getCentralizedModeOwnedApps:
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
func getCentralizedModeOwnedApps(
	logger polylog.Logger,
	ownedAppsPrivateKeysHex []string,
	fullNode FullNode,
) (map[protocol.ServiceID][]string, error) {
	logger = logger.With("method", "getCentralizedModeOwnedApps")
	logger.Debug().Msg("Building the list of owned apps.")

	ownedAppsMap := make(map[protocol.ServiceID][]string)
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

		// Retrieve the app's onchain data using the full node to ensure the request
		// is a remote request and not attempting to use cached data.
		app, err := fullNode.GetApp(context.Background(), appAddr)
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
		serviceID := protocol.ServiceID(appServiceConfig.GetServiceId())
		if serviceID == "" {
			logger.Error().Msgf("should never happen: app with address %s is staked for service with an empty ID", appAddr)
			return nil, fmt.Errorf("app with address %s is staked for service with an empty ID", appAddr)
		}

		// Add the app to the list of owned apps.
		ownedAppsMap[serviceID] = append(ownedAppsMap[serviceID], appAddr)
	}

	logger.Debug().Msgf("Successfully built the list of %d owned apps.", len(ownedAppsMap))
	return ownedAppsMap, nil
}

// getCentralizedGatewayModeActiveSessions returns the set of active sessions under the Centralized gateway mode.
func (p *Protocol) getCentralizedGatewayModeActiveSessions(
	ctx context.Context,
	serviceID protocol.ServiceID,
) ([]sessiontypes.Session, error) {
	logger := p.logger.With(
		"method", "getCentralizedGatewayModeActiveSessions",
		"service_id", string(serviceID),
	)
	logger.Debug().Msgf("fetching active sessions for the service %s.", serviceID)

	// TODO_CRITICAL(@commoddity): if an owned app is changed (i.e. re-staked) for
	// a different service, PATH must be restarted for changes to take effect.
	ownedAppsForService, ok := p.ownedApps[serviceID]
	if !ok || len(ownedAppsForService) == 0 {
		err := fmt.Errorf("%s: %s", errProtocolContextSetupCentralizedNoAppsForService, serviceID)
		logger.Error().Err(err).Msgf("🚨 MISCONFIGURATION: ❌ ZERO owned apps found for service: %s", serviceID)
		return nil, err
	}

	var ownedAppSessions []sessiontypes.Session

	// Loop over the address of apps owned by the gateway in Centralized gateway mode.
	for _, ownedAppAddr := range ownedAppsForService {
		logger.Info().Msgf("About to get a session for  owned app %s for service %s", ownedAppAddr, serviceID)

		// Retrieve the session for the owned app.
		session, err := p.FullNode.GetSession(ctx, serviceID, ownedAppAddr)
		if err != nil {
			// Wrap the protocol context setup error.
			err = fmt.Errorf("%w: app: %s, error: %w", errProtocolContextSetupCentralizedAppFetchErr, ownedAppAddr, err)
			logger.Error().Err(err).Msg(err.Error())
			return nil, err
		}

		app := session.Application

		// Verify the app delegates to the gateway	.
		if !gatewayHasDelegationForApp(p.gatewayAddr, app) {
			// Wrap the protocol context setup error.
			err := fmt.Errorf("%w: app: %s, gateway: %s", errProtocolContextSetupCentralizedAppDelegation, app.Address, p.gatewayAddr)
			logger.Error().Msg(err.Error())
			return nil, err
		}

		ownedAppSessions = append(ownedAppSessions, session)
	}

	// If no sessions were found, return an error.
	if len(ownedAppSessions) == 0 {
		err := fmt.Errorf("%w: service %s.", errProtocolContextSetupCentralizedNoSessions, serviceID)
		logger.Error().Msg(err.Error())
		return nil, err
	}

	logger.Info().Msgf("Successfully fetched %d sessions for %d owned apps for service %s.", len(ownedAppSessions), len(ownedAppsForService), serviceID)

	return ownedAppSessions, nil
}
