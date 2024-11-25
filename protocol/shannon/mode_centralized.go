package shannon

import (
	"fmt"

	apptypes "github.com/pokt-network/poktroll/x/application/types"

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

// getCentralizedGatewayModeAppFilter returns a permittedAppsFilter for the Centralized gateway mode.
func getCentralizedGatewayModeAppFilter(gatewayAddr string, ownedAppsAddr map[string]struct{}) permittedAppFilter {
	return func(app *apptypes.Application) error {
		if _, found := ownedAppsAddr[app.Address]; !found {
			return fmt.Errorf("Centralized GatewayMode: app with address %s is not owned by the gateway", app.Address)
		}

		if !gatewayHasDelegationForApp(gatewayAddr, app) {
			return fmt.Errorf("Centralized GatewayMode: app with address %s does not delegate to gateway address: %s", app.Address, gatewayAddr)
		}

		return nil
	}
}
