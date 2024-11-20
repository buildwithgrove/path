package shannon

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sdk "github.com/pokt-network/shannon-sdk"
)

// Delegated Gateway Mode represents an gateway operation mode which behaves as follows:
// 1. Each relay request is signed by the gateway key, and sent on behalf of an app selected by the user.
// 3. Users need to select a specific app for each relay request, done using HTTP request's headers as of now.
//
// See the following link for more details on PATH's Trusted operation mode.
// https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5?pvs=4#122a36edfff680eea2fbd46c7696d845
const (
	// headerAppAddress is the key of the entry in HTTP headers that holds the target app's address in delegated mode.
	// The target app will be used for sending the relay request.
	headerAppAddr = "X-App-Address"
)

// getDelegatedGatewayModeAppFilter returns a permittedAppsFilter for the Delegated gateway mode.
func getDelegatedGatewayModeAppFilter(gatewayAddr string, req *http.Request) permittedAppsFilter {
	return func(app *apptypes.Application) bool {
		selectedAppAddr, err := getAppAddrFromHTTPReq(req)
		if err != nil {
			return false
		}

		if app.Address != selectedAppAddr {
			return false
		}

		if !gatewayHasDelegationForApp(gatewayaddr, app) {
			return false
		}

		return true
	}
}

// getAppAddrFromHTTPReq extracts the application address specified by the supplied HTTP request's headers.
func getAppAddrFromHTTPReq(httpReq *http.Request) (string, error) {
	if httpReq == nil || len(httpReq.Header) == 0 {
		return nil, fmt.Errorf("getAppAddrFromHTTPReq: no HTTP headers supplied.")
	}

	selectedAppAddr := httpReq.Header.Get(headerAppAddr)
	if selectedAppAddr == "" {
		return nil, fmt.Errorf("getAppAddrFromHTTPReq: a target app must be supplied as HTTP header %s", headerAppAddr)
	}

	return selectedAppAddr, nil
}
