package shannon

import (
	"context"
	"fmt"
	"net/http"

	apptypes "github.com/pokt-network/poktroll/x/application/types"
)

// Delegated Gateway Mode represents an gateway operation mode which behaves as follows:
// 1. Each relay request is signed by the gateway key, and sent on behalf of an app selected by the user.
// 2. Users need to select a specific app for each relay request, done using HTTP request's headers as of now.
// TODO(@Olshansk): Revisit the security specification & requirements for how the paying app is selected.
//
// TODO_DOCUMENT(@Olshansk): Convert the notion doc into a proper README.
// See the following link for more details on PATH's centralized (i.e. trusted) operation mode.
// https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5?pvs=4#122a36edfff680eea2fbd46c7696d845
const (
	// TODO_DOCUMENT(@adshmh): Update the docs at https://path.grove.city/ to reflect this usage pattern.
	// headerAppAddress is the key of the entry in HTTP headers that holds the target app's address in delegated mode.
	// The target app will be used for sending the relay request.
	// TODO_TECHDEBT(@commoddity): remove deprecated `x-` prefix from the header name.
	headerAppAddr = "X-App-Address"
)

// getDelegatedGatewayModeApps returns the set of permitted apps under Delegated gateway mode, for the supplied HTTP request.
func (p *Protocol) getDelegatedGatewayModeApps(ctx context.Context, req *http.Request) ([]*apptypes.Application, error) {
	selectedAppAddr, err := getAppAddrFromHTTPReq(req)
	if err != nil {
		return nil, fmt.Errorf("Delegated GatewayMode: error getting the selected app from the HTTP request: %w", err)
	}

	selectedApp, err := p.FullNode.GetApp(context.TODO(), selectedAppAddr)
	if err != nil {
		return nil, fmt.Errorf("Delegated GatewayMode: error getting the selected app %s data from the SDK: %w", selectedAppAddr, err)
	}

	if !gatewayHasDelegationForApp(p.gatewayAddr, selectedApp) {
		return nil, fmt.Errorf("Delegated GatewayMode: app with address %s does not delegate to gateway address: %s", selectedApp.Address, p.gatewayAddr)
	}

	return []*apptypes.Application{selectedApp}, nil
}

// getAppAddrFromHTTPReq extracts the application address specified by the supplied HTTP request's headers.
func getAppAddrFromHTTPReq(httpReq *http.Request) (string, error) {
	if httpReq == nil || len(httpReq.Header) == 0 {
		return "", fmt.Errorf("getAppAddrFromHTTPReq: no HTTP headers supplied.")
	}

	selectedAppAddr := httpReq.Header.Get(headerAppAddr)
	if selectedAppAddr == "" {
		return "", fmt.Errorf("getAppAddrFromHTTPReq: a target app must be supplied as HTTP header %s", headerAppAddr)
	}

	return selectedAppAddr, nil
}
