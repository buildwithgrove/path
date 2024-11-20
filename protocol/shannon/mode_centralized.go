package shannon

import (
	"fmt"
	"slices"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
)

// In Centralized Gateway Mode, the Shannon protocol integration behaves as follows:
// 1. PATH (or more speicifcally the Shannon protocol integration instance) holds the private keys of the gateway operator's app(s).
// 2. All configured apps are owned by the gateway (PATH) operator.
// 3. All configured apps delegate (onchain) to the gateway address.
// 4. Each relay request is sent on behalf of one of the apps above (owned by the gateway operator)
// 5. Each relay request is signed by the gateway's private key (enabled by ring signatures supported by Shannon)
//
// See the following link for more details on PATH's Centralized operation mode.
// https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5?pvs=4#122a36edfff680d4a0fff3a40dea543e

// getCentralizedModeOwnedAppsAddr returns the list of addresses of apps owned by the gateway, built using the supplied private keys.
func getCentralizedModeOwnedAppsAddr(ownedAppsPrivateKeys []*secp256k1.PrivKey) ([]string, error) {
	var ownedAppsAddr []string
	for _, ownedAppPrivateKey := range ownedAppsPrivateKeys {
		appAddr, err := getAddressFromPrivateKey(ownedAppPrivateKey)
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

func gatewayHasDelegationForApp(gatewayAddr string, app *apptypes.Application) bool {
	return slices.Contains(app.DelegateeGatewayAddresses, gatewayAddr)
}

// getAddressFromPrivateKey returns the address of the provided private key
func getAddressFromPrivateKey(privKey *secp256k1.PrivKey) (string, error) {
	addressBz := privKey.PubKey().Address()
	return bech32.ConvertAndEncode("pokt", addressBz)
}
