package shannon

import (
	"fmt"
	"net/http"

	apptypes "github.com/pokt-network/poktroll/x/application/types"

	"github.com/buildwithgrove/path/protocol"
)

// TODO_DOCUMENT(@adshmh): Convert the following notion doc into a proper README.
//
// Gateway Mode defines the behavior of a specific mode of operation of PATH.
// See the following link for more details on PATH's different modes of operation.
// https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5
//
// TODO_MVP(@adshmh): Add the GatewayModePermissionless to the output once it is supported.
// SupportedGatewayModes returns the list of gateway modes supported by the Shannon protocol integration.
// This method implements the gateway.Protocol interface.
func (p Protocol) SupportedGatewayModes() []protocol.GatewayMode {
	return []protocol.GatewayMode{
		protocol.GatewayModeCentralized,
		protocol.GatewayModeDelegated,
	}
}

// TODO_TECHDEBT: once Shannon supports querying the applications based on one more criteria, this function's name and signature should be updated to
// build and return the query criteria.
//
// permittedAppFilter represents any function that can be used to filter an onchain app based on its attributes.
// It is used by different gateway modes to select app(s) that are permitted for use by the gateway for sending relay requests.
type permittedAppFilter func(*apptypes.Application) error

// getGatewayModePermittedAppFilter returns the app filter matching the supplied gateway mode.
// As of now, the HTTP request that initiates a relay request can also be used to adjust the app filter, e.g. in the Delegated gateway mode.
func (p *Protocol) getGatewayModePermittedAppFilter(
	gatewayMode protocol.GatewayMode,
	req *http.Request,
) (permittedAppFilter, error) {
	switch gatewayMode {
	case protocol.GatewayModeCentralized:
		return getCentralizedGatewayModeAppFilter(p.gatewayAddr, p.ownedAppsAddr), nil
	case protocol.GatewayModeDelegated:
		return getDelegatedGatewayModeAppFilter(p.gatewayAddr, req), nil
	default:
		return nil, fmt.Errorf("unsupported gateway mode: %s", gatewayMode)
	}
}

// getGatewayModePermittedRelaySigner returns the relay request signer matching the supplied gateway mode.
func (p *Protocol) getGatewayModePermittedRelaySigner(
	gatewayMode protocol.GatewayMode,
) (RelayRequestSigner, error) {
	switch gatewayMode {
	case protocol.GatewayModeCentralized:
		return &signer{
			accountClient: *p.FullNode.GetAccountClient(),
			//  Centralized gateway mode uses the gateway's private key to sign the relay requests.
			privateKeyHex: p.gatewayPrivateKeyHex,
		}, nil
	case protocol.GatewayModeDelegated:
		return &signer{
			accountClient: *p.FullNode.GetAccountClient(),
			//  Delegated gateway mode uses the gateway's private key to sign the relay requests (i.e. the same as the Centralized gateway mode)
			privateKeyHex: p.gatewayPrivateKeyHex,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported gateway mode: %s", gatewayMode)
	}
}
