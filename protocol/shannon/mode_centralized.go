package shannon

import (
	"context"
	"fmt"

	apptypes "github.com/pokt-network/poktroll/x/application/types"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/protocol/crypto"
)

// In Centralized Gateway Mode, the Shannon protocol integration behaves as follows:
// 1. PATH (or more specifically the Shannon protocol integration instance) holds the private keys of the gateway operator's app(s).
// 2. All configured apps are owned by the gateway (PATH) operator.
// 3. All configured apps delegate (onchain) to the gateway address.
// 4. Each relay request is sent on behalf of one of the apps above (owned by the gateway operator)
// 5. Each relay request is signed by the gateway's private key (enabled by ring signatures supported by Shannon)
//
// See the following link for more details on PATH's Centralized operation mode.
// https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5?pvs=4#122a36edfff680d4a0fff3a40dea543e
//
// getCentralizedModeOwnedAppsAddr returns the list of addresses of apps owned by the gateway, built using the supplied private keys.
// The ONLY use of the supplied apps' private keys by the centralized mode is to build apps' addresses on behalf of which relays are sent.
func getCentralizedModeOwnedAppsAddr(ownedAppsPrivateKeysHex []string) ([]string, error) {
	var ownedAppsAddr []string
	for _, ownedAppPrivateKeyHex := range ownedAppsPrivateKeysHex {
		ownedAppPrivateKey, err := crypto.GetSecp256k1PrivateKeyFromKeyHex(ownedAppPrivateKeyHex)
		if err != nil {
			return nil, err
		}

		appAddr, err := crypto.GetAddressFromPrivateKey(ownedAppPrivateKey)
		if err != nil {
			return nil, err
		}

		ownedAppsAddr = append(ownedAppsAddr, appAddr)
	}

	return ownedAppsAddr, nil
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

// getCentralizedGatewayModeApps returns the set of permitted apps under the Centralized gateway mode.
func (p *Protocol) getCentralizedGatewayModeApps(ctx context.Context, serviceID protocol.ServiceID) ([]*apptypes.Application, error) {
	logger := p.logger.With(
		"service_id", string(serviceID),
		"gateway_addr", p.gatewayAddr,
		"gateway_mode", protocol.GatewayModeCentralized,
		"num_owned_apps", len(p.ownedAppsAddr),
	)

	var permittedApps []*apptypes.Application

	// Loop over the address of apps owned by the gateway in Centralized gateway mode.
	for ownedAppAddr := range p.ownedAppsAddr {
		logger.Debug().Msgf("Centralized GatewayMode: checking app %s owned by the gateway", ownedAppAddr)

		onchainApp, err := p.FullNode.GetApp(ctx, ownedAppAddr)
		if err != nil {
			return nil, fmt.Errorf("Centralized GatewayMode: error getting onchain data for app %s owned by the gateway: %w", ownedAppAddr, err)
		}

		// Skip the app if it is not staked for the requested service.
		if !appIsStakedForService(serviceID, onchainApp) {
			logger.With("app_addr", ownedAppAddr).Debug().Msg("owned app is not staked for the service. Skipping.")
			continue
		}

		// Verify the app delegates to the gateway.
		if !gatewayHasDelegationForApp(p.gatewayAddr, onchainApp) {
			return nil, fmt.Errorf("Centralized GatewayMode: app with address %s does not delegate to gateway address: %s", onchainApp.Address, p.gatewayAddr)
		}

		permittedApps = append(permittedApps, onchainApp)
	}

	if len(permittedApps) == 0 {
		logger.Info().Msg("No owned apps matched the request.")
		return nil, fmt.Errorf("Centralized GatewayMode: no owned apps matched the request.")
	}

	return permittedApps, nil
}
