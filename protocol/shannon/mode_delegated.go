package shannon

import (
	"context"
	"fmt"
	"net/http"

	apptypes "github.com/pokt-network/poktroll/x/application/types"

	"github.com/buildwithgrove/path/request"
)

// Delegated Gateway Mode represents an gateway operation mode which behaves as follows:
// 1. Each relay request is signed by the gateway key, and sent on behalf of an app selected by the user.
// 2. Users need to select a specific app for each relay request, done using HTTP request's headers as of now.
// TODO(@Olshansk): Revisit the security specification & requirements for how the paying app is selected.
//
// TODO_DOCUMENT(@Olshansk): Convert the notion doc into a proper README.
// See the following link for more details on PATH's centralized (i.e. trusted) operation mode.
// https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5?pvs=4#122a36edfff680eea2fbd46c7696d845

// getDelegatedGatewayModeApps returns the set of permitted apps under Delegated gateway mode, for the supplied HTTP request.
func (p *Protocol) getDelegatedGatewayModeApps(ctx context.Context, httpReq *http.Request) ([]*apptypes.Application, error) {
	logger := p.logger.With(
		"gateway_mode", p.gatewayMode,
		"gateway_addr", p.gatewayAddr,
	)

	selectedAppAddr, err := getAppAddrFromHTTPReq(httpReq)
	if err != nil {
		logger.Error().Err(err).Msg("error extracting the selected app from the HTTP request: relay request will fail.")
		return nil, fmt.Errorf("delegated GatewayMode: error getting the selected app from the HTTP request: %w", err)
	}

	logger = logger.With("selected_app_addr", selectedAppAddr)
	logger.Debug().Msg("fetching the app with the selected address")

	// TODO_TECHDEBT(@adshmh): Pass a context with deadline to the protocol.
	// This is necessary to ensure the HTTP handling goroutine does not timeout waiting for the protocol.
	//
	selectedApp, err := p.FullNode.GetApp(ctx, selectedAppAddr)
	if err != nil {
		logger.Error().Err(err).Msg("error fetching the selected app: relay request will fail.")
		return nil, fmt.Errorf("delegated GatewayMode: error getting the selected app %s data from the SDK: %w", selectedAppAddr, err)
	}

	// log successful app fetch message.
	logger = logger.With("fetched_app_addr", selectedApp.Address)
	logger.Debug().Msg("fetched the app with the selected address")

	if !gatewayHasDelegationForApp(p.gatewayAddr, selectedApp) {
		logger.Error().Msg("gateway does not have delegation for the selected app: relay request will fail.")
		return nil, fmt.Errorf("delegated GatewayMode: app with address %s does not delegate to gateway address: %s", selectedApp.Address, p.gatewayAddr)
	}

	// log success message.
	logger.Debug().Msg("successfully verified the gateway has delegation for the selected app.")

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
