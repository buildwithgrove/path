package shannon

import (
	"context"
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/protocol/crypto"
)

// ownedApp represents a single app owned by the gateway operator in Centralized Gateway Mode.
type OwnedApp struct {
	// The address of the app. E.g. "pokt1..."
	AppAddr string
	// The service ID for which the app is staked. E.g. "anvil"
	StakedServiceID protocol.ServiceID
}

// Centralized Gateway Mode - Shannon Protocol Integration
//
// - PATH (Shannon instance) holds private keys for gateway operator's apps
// - All apps are owned by the gateway (PATH) operator
// - All apps delegate (onchain) to the gateway address
// - Each relay request is sent for one of these apps (owned by the gateway operator)
// - Each relay is signed by the gateway's private key (via Shannon ring signatures)
//
// More details: https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5?pvs=4#122a36edfff680d4a0fff3a40dea543e
//
// GetCentralizedModeOwnedApps:
//   - Returns list of apps owned by the gateway, built from supplied private keys
//   - Supplied private keys are ONLY used to build app addresses for relay signing
//   - Populates `appAddr` and `stakedServiceID` for each app
func GetCentralizedModeOwnedApps(
	logger polylog.Logger,
	ownedAppsPrivateKeysHex []string,
	lazyFullNode *LazyFullNode,
) ([]OwnedApp, error) {
	logger = logger.With("method", "getCentralizedModeOwnedApps")
	logger.Debug().Msg("Building the list of owned apps.")

	var ownedApps []OwnedApp
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

		// Retrieve the app's onchain data using the lazy full node to ensure the request
		// is a remote request and not attempting to use cached data.
		app, err := lazyFullNode.GetApp(context.Background(), appAddr)
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
		ownedApps = append(ownedApps, OwnedApp{
			AppAddr:         appAddr,
			StakedServiceID: serviceID,
		})
	}

	logger.Debug().Msgf("Successfully built the list of %d owned apps.", len(ownedApps))
	return ownedApps, nil
}

// appIsStakedForService returns true if the supplied application is staked for the supplied service ID.
func appIsStakedForService(serviceID protocol.ServiceID, app *apptypes.Application) bool {
	for _, svcCfg := range app.ServiceConfigs {
		if protocol.ServiceID(svcCfg.ServiceId) == serviceID {
			return true
		}
	}
	return false
}

// TODO_IMPROVE(@commoddity, @adshmh): This function currently loops through all apps owned by the gateway.
// Without a caching FullNode, this results in extremely slow behaviour. We should look into improving the
// efficiency of this lookup to get the list of apps owned by the gateway.
//
// getCentralizedGatewayModeSessions returns the set of permitted sessions under the Centralized gateway mode.
func (p *Protocol) getCentralizedGatewayModeSessions(
	ctx context.Context,
	serviceID protocol.ServiceID,
) ([]sessiontypes.Session, error) {
	logger := p.logger.With(
		"method", "getCentralizedGatewayModeSessions",
		"service_id", string(serviceID),
	)
	logger.Debug().Msg("fetching the list of owned apps.")

	var permittedSessions []sessiontypes.Session

	// Loop over the address of apps owned by the gateway in Centralized gateway mode.
	for _, ownedApp := range p.ownedApps {
		if ownedApp.StakedServiceID != serviceID {
			continue
		}

		ownedAppAddr := ownedApp.AppAddr

		logger.Info().Msgf("checking app %s owned by the gateway", ownedAppAddr)

		session, err := p.FullNode.GetSession(ctx, serviceID, ownedAppAddr)
		if err != nil {
			// Wrap the protocol context setup error.
			err = fmt.Errorf("%w: app: %s, error: %w", errProtocolContextSetupCentralizedAppFetchErr, ownedAppAddr, err)
			logger.Error().Err(err).Msg(err.Error())
			return nil, err
		}

		app := session.Application

		// Skip the session's app if it is not staked for the requested service.
		if !appIsStakedForService(serviceID, app) {
			logger.Debug().Msgf("owned app %s is not staked for the service. Skipping.", ownedAppAddr)
			continue
		}

		// Verify the app delegates to the gateway.
		if !gatewayHasDelegationForApp(p.gatewayAddr, app) {
			// Wrap the protocol context setup error.
			err := fmt.Errorf("%w: app: %s, gateway: %s", errProtocolContextSetupCentralizedAppDelegation, app.Address, p.gatewayAddr)
			logger.Error().Msg(err.Error())
			return nil, err
		}

		permittedSessions = append(permittedSessions, session)
	}

	// If no sessions matched the request, return an error.
	if len(permittedSessions) == 0 {
		err := fmt.Errorf("%w: service %s.", errProtocolContextSetupCentralizedNoApps, serviceID)
		logger.Error().Msg(err.Error())
		return nil, err
	}

	logger.Debug().Msgf("Successfully fetched the list of sessions for %d owned apps for service %s.", len(permittedSessions), serviceID)

	return permittedSessions, nil
}
