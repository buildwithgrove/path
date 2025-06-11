package shannon

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	apptypes "github.com/pokt-network/poktroll/x/application/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/protocol"
)

// TODO_DOCUMENT(@adshmh): Convert the following notion doc into a proper README.
//
// Gateway Mode defines the behavior of a specific mode of operation of PATH.
// See the following link for more details on PATH's different modes of operation.
// https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5
//
// SupportedGatewayModes returns the list of gateway modes supported by the Shannon protocol integration.
// Implements the gateway.Protocol interface.
func (p *Protocol) SupportedGatewayModes() []protocol.GatewayMode {
	return supportedGatewayModes()
}

// TODO_NEXT(@commoddity): Most of the functionality in this file should be moved to the Shannon SDK.
// A `GatewayClient` should be created for each GatewayMode that embeds a `FullNode`.
//
// For Example:
// - Centralized mode: `CentralizedGatewayClient`
// - Delegated mode: `DelegatedGatewayClient`
// etc...
//
// With each `GatewayClient` containing the below logic that is specific to the GatewayMode.
//
// Issue:
// https://github.com/buildwithgrove/path/issues/291

// TODO_NEXT(@commoddity): This function should be moved to the Shannon SDK and
// and handled in the specific GatewayClient structs for each GatewayMode.
//
// getGatewayModePermittedSessions returns the sessions permitted under the supplied gateway mode.
// The permitted sessions are determined as follows:
//   - Centralized mode: the gateway address and owned apps addresses are used to determine the permitted sessions (specified in configs).
//   - Delegated mode: the gateway address and app address in the HTTP headers are used to determine the permitted sessions.
func (p *Protocol) getGatewayModePermittedSessions(
	ctx context.Context,
	serviceID sdk.ServiceID,
	httpReq *http.Request,
) ([]sessiontypes.Session, error) {
	p.logger.With(
		"service_ID", serviceID,
		"gateway_mode", p.gatewayMode,
	).Debug().Msg("fetching permitted apps under the current gateway mode.")

	switch p.gatewayMode {

	case protocol.GatewayModeCentralized:
		return p.getCentralizedGatewayModeSessions(ctx, serviceID)

	case protocol.GatewayModeDelegated:
		return p.getDelegatedGatewayModeSession(ctx, serviceID, httpReq)

	// TODO_MVP(@adshmh): Uncomment the following code section once support for Permissionless Gateway mode is added to the shannon package.
	//case protocol.GatewayModePermissionless:
	//	return getPermissionlessGatewayModeApps(p.ownedAppsAddr), nil

	default:
		return nil, fmt.Errorf("%w: %s", errProtocolContextSetupUnsupportedGatewayMode, p.gatewayMode)
	}
}

// TODO_NEXT(@commoddity): This function should be moved to the Shannon SDK and
// and handled in the specific GatewayClient structs for each GatewayMode.
//
// getGatewayModePermittedRelaySigner returns the relay request signer matching the supplied gateway mode.
func (p *Protocol) getGatewayModePermittedRelaySigner(
	gatewayMode protocol.GatewayMode,
) (RelayRequestSigner, error) {
	switch gatewayMode {
	case protocol.GatewayModeCentralized:
		return &sdk.Signer{
			PrivateKeyHex:    p.gatewayPrivateKeyHex,
			PublicKeyFetcher: p.FullNode,
		}, nil
	case protocol.GatewayModeDelegated:
		return &sdk.Signer{
			PrivateKeyHex:    p.gatewayPrivateKeyHex,
			PublicKeyFetcher: p.FullNode,
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
