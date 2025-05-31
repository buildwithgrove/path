package shannon

import (
	"context"
	"fmt"
	"net/http"

	apptypes "github.com/pokt-network/poktroll/x/application/types"

	"github.com/buildwithgrove/path/request"
)

// Delegated Gateway Mode:
// - Represents a gateway operation mode with the following behavior:
//   - Each relay request is signed by the gateway key and sent on behalf of an app selected by the user.
//   - Users must select a specific app for each relay request (currently via HTTP request headers).
// - TODO(@Olshansk): Revisit the security specification & requirements for how the paying app is selected.
// - TODO_DOCUMENT(@Olshansk): Convert the Notion doc into a proper README.
// - For more details, see:
//   https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5?pvs=4#122a36edfff680eea2fbd46c7696d845

// getDelegatedGatewayModeApps returns the set of permitted apps under Delegated gateway mode, for the supplied HTTP request.
func (p *Protocol) getDelegatedGatewayModeApps(ctx context.Context, httpReq *http.Request) ([]*apptypes.Application, error) {
	logger := p.logger.With("method", "getDelegatedGatewayModeApps")

	selectedAppAddr, err := getAppAddrFromHTTPReq(httpReq)
	if err != nil {
		// Wrap the context setup error: used for observations.
		err = fmt.Errorf("%w: %+v: %w. ", errProtocolContextSetupGetAppFromHTTPReq, httpReq, err)
		logger.Error().Err(err).Msg("error getting the app address from the HTTP request. Relay request will fail.")
		return nil, err
	}

	logger.Debug().Msgf("fetching the app with the selected address %s.", selectedAppAddr)

	selectedApp, err := p.FullNode.GetApp(ctx, selectedAppAddr)
	if err != nil {
		// Wrap the context setup error: used for observations.
		err = fmt.Errorf("%w: app %s: %w. Relay request will fail.", errProtocolContextSetupFetchApp, selectedAppAddr, err)
		logger.Error().Err(err).Msg("error fetching the app. Relay request will fail.")
		return nil, err
	}

	logger.Debug().Msgf("fetched the app with the selected address %s.", selectedApp.Address)

	if !gatewayHasDelegationForApp(p.gatewayAddr, selectedApp) {
		// Wrap the context setup error: used for observations.
		err = fmt.Errorf("%w: gateway %s app %s. Relay request will fail.", errProtocolContextSetupAppDoesNotDelegate, p.gatewayAddr, selectedApp.Address)
		logger.Error().Err(err).Msg("Gateway does noth ave deletation for the app. Relay request will fail.")
		return nil, err
	}

	logger.Debug().Msgf("successfully verified the gateway has delegation for the selected app with address %s.", selectedApp.Address)

	return []*apptypes.Application{selectedApp}, nil
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
