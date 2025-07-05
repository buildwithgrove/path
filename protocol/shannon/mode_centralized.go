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

	sessiontypes "github.com/pokt-network/poktroll/x/session/types"

	"github.com/buildwithgrove/path/protocol"
)

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
		logger.Error().Err(err).Msg("🚨 MISCONFIGURATION: ❌ ZERO owned apps found for service.")
		return nil, err
	}

	// Loop over the address of apps owned by the gateway in Centralized gateway mode.
	var ownedAppSessions []sessiontypes.Session
	for _, ownedAppAddr := range ownedAppsForService {
		logger.Info().Msgf("About to get a session for  owned app %s for service %s", ownedAppAddr, serviceID)

		// Retrieve the session for the owned app, without grace period logic.
		sessionLatest, err := p.GetSession(ctx, serviceID, ownedAppAddr)
		if err != nil {
			err = fmt.Errorf("%w: app: %s, error: %w", errProtocolContextSetupFetchSession, ownedAppAddr, err)
			logger.Error().Err(err).Msgf("Error getting the current session from the full node for app: %s", ownedAppAddr)
			return nil, err
		}
		// Verify both apps delegate to the gateway
		if !gatewayHasDelegationForApp(p.gatewayAddr, sessionLatest.Application) {
			err := fmt.Errorf("%w: app: %s, gateway: %s", errProtocolContextSetupCentralizedAppDelegation, sessionLatest.Application.Address, p.gatewayAddr)
			logger.Error().Msg(err.Error())
			return nil, err
		}
		ownedAppSessions = append(ownedAppSessions, sessionLatest)

		// Retrieve the session for the owned app, considering grace period logic.
		sessionPreviousExtended, err := p.GetSessionWithExtendedValidity(ctx, serviceID, ownedAppAddr)
		if err != nil {
			err = fmt.Errorf("%w: app: %s, error: %w", errProtocolContextSetupFetchSession, ownedAppAddr, err)
			logger.Warn().Err(err).Msgf("Error getting the previous extended session from the full node for app: %s. Skipping it", ownedAppAddr)
			continue
		}

		// Compare session IDs - if they're different, return both sessions
		if sessionLatest.Header.SessionId != sessionPreviousExtended.Header.SessionId {
			if !gatewayHasDelegationForApp(p.gatewayAddr, sessionLatest.Application) {
				err := fmt.Errorf("%w: app: %s, gateway: %s", errProtocolContextSetupCentralizedAppDelegation, sessionLatest.Application.Address, p.gatewayAddr)
				logger.Error().Msg(err.Error())
				continue
			}
			ownedAppSessions = append(ownedAppSessions, sessionPreviousExtended, sessionLatest)
		}
	}

	// If no sessions were found, return an error.
	if len(ownedAppSessions) == 0 {
		err := fmt.Errorf("%w: service %s", errProtocolContextSetupCentralizedNoSessions, serviceID)
		logger.Error().Msg(err.Error())
		return nil, err
	}

	logger.Info().Msgf("Successfully fetched %d sessions for %d owned apps for service %s.",
		len(ownedAppSessions), len(ownedAppsForService), serviceID)

	return ownedAppSessions, nil
}
