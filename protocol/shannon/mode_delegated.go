package shannon

// - TODO(@Olshansk): Revisit the security specification & requirements for how the paying app is selected.
// - TODO_DOCUMENT(@Olshansk): Convert the Notion doc into a proper README.
// - For more details, see:
//   https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5?pvs=4#122a36edfff680eea2fbd46c7696d845

import (
	"context"
	"fmt"
	"net/http"

	apptypes "github.com/pokt-network/poktroll/x/application/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/request"
)

// Delegated Gateway Mode - Shannon Protocol Integration
//
// - Represents a gateway operation mode with the following behavior:
// - Each relay request is signed by the gateway key and sent on behalf of an app selected by the user.
// - Users must select a specific app for each relay request (currently via HTTP request headers).
//
// getDelegatedGatewayModeActiveSession returns active sessions for the selected app under Delegated gateway mode, for the supplied HTTP request.
func (p *Protocol) getDelegatedGatewayModeActiveSession(
	ctx context.Context,
	serviceID protocol.ServiceID,
	httpReq *http.Request,
) ([]sessiontypes.Session, error) {
	logger := p.logger.With("method", "getDelegatedGatewayModeActiveSession")

	extractedAppAddr, err := getAppAddrFromHTTPReq(httpReq)
	if err != nil {
		// Wrap the context setup error: used for observations.
		err = fmt.Errorf("%w: %+v: %w. ", errProtocolContextSetupGetAppFromHTTPReq, httpReq, err)
		logger.Error().Err(err).Msg("error getting the app address from the HTTP request. Relay request will fail.")
		return nil, err
	}

	logger.Debug().Msgf("fetching the app with the selected address %s.", extractedAppAddr)

	// Retrieve the session for the owned app, without grace period logic.
	selectedSessions := make([]sessiontypes.Session, 0)
	sessionLatest, err := p.GetSession(ctx, serviceID, extractedAppAddr)
	if err != nil {
		err = fmt.Errorf("%w: app: %s, error: %w", errProtocolContextSetupCentralizedAppFetchErr, extractedAppAddr, err)
		logger.Warn().Err(err).Msgf("SHOULD NEVER HAPPEN: Error getting the current session from the full node for app: %s. Cannot continue.", extractedAppAddr)
		return nil, err
	}
	selectedSessions = append(selectedSessions, sessionLatest)

	// Retrieve the session for the owned app, considering grace period logic.
	sessionPreviousExtended, err := p.GetSessionWithExtendedValidity(ctx, serviceID, extractedAppAddr)
	if err != nil {
		err = fmt.Errorf("%w: app: %s, error: %w", errProtocolContextSetupCentralizedAppFetchErr, extractedAppAddr, err)
		logger.Warn().Err(err).Msgf("SHOULD RARELY HAPPEN: Error getting the previous extended session from the full node for app: %s. Going to use the latest session only", extractedAppAddr)
	} else {
		if sessionLatest.Header.SessionId != sessionPreviousExtended.Header.SessionId {
			if sessionLatest.Application.Address != sessionPreviousExtended.Application.Address {
				logger.Warn().Msg("SHOULD NEVER HAPPEN: The current session app address and the previous session app address are different. Only using the latest session.")
			} else {
				// Append the previous session to the list if the session IDs are different.
				// selectedSessions = append(selectedSessions, sessionPreviousExtended)
				// TODO_IN_THIS_PR: Revert this change.
				// We are experimenting by forcing it to always use the previous session.
				selectedSessions = []sessiontypes.Session{sessionPreviousExtended}
				logger.Info().Msg("EXPERIMENT: Overriding the latest session with the previous session for the app.")
			}
		}
	}

	// Select the first session in the list.
	selectedApp := selectedSessions[0].Application
	logger.Debug().Msgf("fetched the app with the selected address %s.", selectedApp.Address)

	// Skip the session's app if it is not staked for the requested service.
	if !appIsStakedForService(serviceID, selectedApp) {
		err = fmt.Errorf("%w: Trying to use app %s that is not staked for the service %s", errProtocolContextSetupAppNotStaked, selectedApp.Address, serviceID)
		logger.Error().Err(err).Msgf("SHOULD NEVER HAPPEN: Trying to use an app that is not staked for the service. Relay request will fail.")
		return nil, err
	}

	if !gatewayHasDelegationForApp(p.gatewayAddr, selectedApp) {
		// Wrap the context setup error: used for observations.
		err = fmt.Errorf("%w: Trying to use app %s that is not delegated to the gateway %s", errProtocolContextSetupAppDoesNotDelegate, selectedApp.Address, p.gatewayAddr)
		logger.Error().Err(err).Msgf("SHOULD NEVER HAPPEN: Trying to use an app that is not delegated to the gateway. Relay request will fail.")
		return nil, err
	}

	logger.Debug().Msgf("successfully verified the gateway (%s) has delegation for the selected app (%s) for service (%s).", p.gatewayAddr, selectedApp.Address, serviceID)

	return selectedSessions, nil
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

// getAppAddrFromHTTPReq extracts the application address specified by the supplied HTTP request's headers.
func getAppAddrFromHTTPReq(httpReq *http.Request) (string, error) {
	if httpReq == nil || len(httpReq.Header) == 0 {
		return "", fmt.Errorf("getAppAddrFromHTTPReq: no HTTP headers supplied")
	}

	extractedAppAddr := httpReq.Header.Get(request.HTTPHeaderAppAddress)
	if extractedAppAddr == "" {
		return "", fmt.Errorf("getAppAddrFromHTTPReq: a target app must be supplied as HTTP header %s", request.HTTPHeaderAppAddress)
	}

	return extractedAppAddr, nil
}
