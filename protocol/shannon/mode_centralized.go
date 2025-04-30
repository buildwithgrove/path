package shannon

import (
	"context"
	"fmt"

	apptypes "github.com/pokt-network/poktroll/x/application/types"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/protocol/crypto"
)

// Own
type ownedApp struct {
	appAddr         string
	stakedServiceID protocol.ServiceID
}

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
// getCentralizedModeOwnedApps returns the list of apps owned by the gateway, built using the supplied private keys.
//
// The following fields are populated for each owned app:
//   - appAddr: the address of the app
//   - stakedServiceID: the service ID for which the app is staked
//
// The ONLY use of the supplied apps' private keys by the centralized mode is to build apps' addresses on behalf of which relays are sent.
func (p *Protocol) getCentralizedModeOwnedApps(ownedAppsPrivateKeysHex []string) ([]ownedApp, error) {
	var ownedApps []ownedApp
	for _, ownedAppPrivateKeyHex := range ownedAppsPrivateKeysHex {
		ownedAppPrivateKey, err := crypto.GetSecp256k1PrivateKeyFromKeyHex(ownedAppPrivateKeyHex)
		if err != nil {
			return nil, err
		}

		appAddr, err := crypto.GetAddressFromPrivateKey(ownedAppPrivateKey)
		if err != nil {
			return nil, err
		}

		application, err := p.FullNode.GetApp(context.Background(), appAddr)
		if err != nil {
			return nil, err
		}

		appServiceConfigs := application.GetServiceConfigs()
		if len(appServiceConfigs) != 1 {
			return nil, fmt.Errorf("centralized GatewayMode: app with address %s is not staked for exactly one service", appAddr)
		}

		stakedServiceID := appServiceConfigs[0].GetServiceId()
		if stakedServiceID == "" {
			return nil, fmt.Errorf("centralized GatewayMode: app with address %s is not staked for any service", appAddr)
		}

		ownedApps = append(ownedApps, ownedApp{
			appAddr:         appAddr,
			stakedServiceID: protocol.ServiceID(stakedServiceID),
		})
	}

	return ownedApps, nil
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
		"method", "getCentralizedGatewayModeApps",
		"service_id", string(serviceID),
		"gateway_addr", p.gatewayAddr,
		"gateway_mode", protocol.GatewayModeCentralized,
		"num_owned_apps", len(p.ownedApps),
	)

	var permittedApps []*apptypes.Application

	// Loop over the address of apps owned by the gateway in Centralized gateway mode.
	for _, ownedApp := range p.ownedApps {
		ownedAppAddr := ownedApp.appAddr

		logger.Info().Msgf("Centralized GatewayMode: checking app owned by the gateway with address: %s", ownedAppAddr)

		onchainApp, err := p.FullNode.GetApp(ctx, ownedAppAddr)
		if err != nil {
			return nil, fmt.Errorf("centralized GatewayMode: error getting onchain data for app %s owned by the gateway: %w", ownedAppAddr, err)
		}

		// Skip the app if it is not staked for the requested service.
		if !appIsStakedForService(serviceID, onchainApp) {
			logger.With("app_addr", ownedAppAddr).Debug().Msg("owned app is not staked for the service. Skipping.")
			continue
		}

		// Verify the app delegates to the gateway.
		if !gatewayHasDelegationForApp(p.gatewayAddr, onchainApp) {
			return nil, fmt.Errorf("centralized GatewayMode: app with address %s does not delegate to gateway address: %s", onchainApp.Address, p.gatewayAddr)
		}

		permittedApps = append(permittedApps, onchainApp)
	}

	if len(permittedApps) == 0 {
		logger.Info().Msg("No owned apps matched the request.")
		return nil, fmt.Errorf("centralized GatewayMode: no owned apps matched the request")
	}

	return permittedApps, nil
}
