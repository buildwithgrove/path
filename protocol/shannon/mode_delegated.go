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

	selectedAppAddr, err := getAppAddrFromHTTPReq(httpReq)
	if err != nil {
		// Wrap the context setup error: used for observations.
		err = fmt.Errorf("%w: %+v: %w. ", errProtocolContextSetupGetAppFromHTTPReq, httpReq, err)
		logger.Error().Err(err).Msg("error getting the app address from the HTTP request. Relay request will fail.")
		return nil, err
	}

	logger.Debug().Msgf("fetching the app with the selected address %s.", selectedAppAddr)

	// Retrieve the session for the selected app.
	selectedSession, err := p.FullNode.GetSession(ctx, serviceID, selectedAppAddr)
	if err != nil {
		// Wrap the context setup error: used for observations.
		err = fmt.Errorf("%w: app %s: %w", errProtocolContextSetupFetchSession, selectedAppAddr, err)
		logger.Error().Err(err).Msg("Relay request will fail because of an error fetching the session for the app.")
		return nil, err
	}

	selectedApp := selectedSession.Application

	logger.Debug().Msgf("fetched the app with the selected address %s.", selectedApp.Address)

	// Skip the session's app if it is not staked for the requested service.
	if !appIsStakedForService(serviceID, selectedApp) {
		err = fmt.Errorf("%w: app %s is not staked for the service", errProtocolContextSetupAppNotStaked, selectedApp.Address)
		logger.Error().Err(err).Msg("Relay request will fail because the app is not staked for the service.")
		return nil, err
	}

	if !gatewayHasDelegationForApp(p.gatewayAddr, selectedApp) {
		// Wrap the context setup error: used for observations.
		err = fmt.Errorf("%w: gateway %s app %s. Relay request will fail.", errProtocolContextSetupAppDoesNotDelegate, p.gatewayAddr, selectedApp.Address)
		logger.Error().Err(err).Msg("Relay request will fail because the gateway does not have delegation for the app.")
		return nil, err
	}

	logger.Debug().Msgf("successfully verified the gateway has delegation for the selected app with address %s.", selectedApp.Address)

	return []sessiontypes.Session{selectedSession}, nil
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

	selectedAppAddr := httpReq.Header.Get(request.HTTPHeaderAppAddress)
	if selectedAppAddr == "" {
		return "", fmt.Errorf("getAppAddrFromHTTPReq: a target app must be supplied as HTTP header %s", request.HTTPHeaderAppAddress)
	}

	return selectedAppAddr, nil
}
