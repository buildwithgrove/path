package shannon

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	apptypes "github.com/pokt-network/poktroll/x/application/types"

	"github.com/buildwithgrove/path/protocol"
)

// TODO_DOCUMENT(@adshmh): Convert the following notion doc into a proper README.
//
// Gateway Mode defines the behavior of a specific mode of operation of PATH.
// See the following link for more details on PATH's different modes of operation.
// https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5
//
// SupportedGatewayModes returns the list of gateway modes supported by the Shannon protocol integration.
// This method implements the gateway.Protocol interface.
func (p *Protocol) SupportedGatewayModes() []protocol.GatewayMode {
	return supportedGatewayModes()
}

// getGatewayModePermittedApps returns the apps permitted under the supplied gateway mode.
// The permitted apps are determined as follows:
//   - Centralized mode: the gateway address and owned apps addresses are used to determine the permitted apps (specified in configs).
//   - Delegated mode: the gateway address and app address in the HTTP headers are used to determine the permitted apps.
func (p *Protocol) getGatewayModePermittedApps(
	ctx context.Context,
	serviceID protocol.ServiceID,
	req *http.Request,
) ([]*apptypes.Application, error) {
	switch p.gatewayMode {

	case protocol.GatewayModeCentralized:
		return p.getCentralizedGatewayModeApps(ctx, serviceID)

	case protocol.GatewayModeDelegated:
		return p.getDelegatedGatewayModeApps(ctx, req)

		// TODO_MVP(@adshmh): Uncomment the following code section once support for Permissionless Gateway mode is added to the shannon package.
		//case protocol.GatewayModePermissionless:
		//	return getPermissionlessGatewayModeApps(p.ownedAppsAddr), nil

	default:
		return nil, fmt.Errorf("unsupported gateway mode: %s", p.gatewayMode)
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

// supportedGatewayModes returns the list of gateway modes currently supported by the Shannon protocol integration.
func supportedGatewayModes() []protocol.GatewayMode {
	return []protocol.GatewayMode{
		protocol.GatewayModeCentralized,
		protocol.GatewayModeDelegated,
		// TODO_MVP(@adshmh): Uncomment this line once support for Permissionless Gateway mode is added to the shannon package.
		// protocol.GatewayModePermissionless,
	}
}

// gatewayHasDelegationsForApp returns true if the supplied application delegates to the supplied gateway address.
func gatewayHasDelegationForApp(gatewayAddr string, app *apptypes.Application) bool {
	return slices.Contains(app.DelegateeGatewayAddresses, gatewayAddr)
}
