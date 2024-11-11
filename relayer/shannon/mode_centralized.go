package shannon

import (
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sdk "github.com/pokt-network/shannon-sdk"
)

var (
	// CentralizedMode implements the OperationMode interface consumed by the Protocol struct.
	// See the file mode.go for a detailed description of the Operation Mode concept.
	_ OperationMode = &CentralizedMode{}

	// centralizedModeInstance implements the OperationModeInstance interface consumed by the Protocol struct.
	// It is built by the CentralizedMode struct.
	// Note that in the Centralized Mode, all operation mode instances are the same as user cannot direct the operation mode to use any specific apps (as of now).
	_ OperationModeInstance = &centralizedModeInstance{}
)

// NewCentralizedMode initializes an instance of CentralizedMode using the supplied list of private keys.
// The supplied account client is required to instantiate the single relay request signer used by all instance of the Centralized operation mode.
func NewCentralizedMode(gatewayAddr, gatewayPrivateKeyHex string, delegatingAppsPrivateKeys []*secp256k1.PrivKey, accountClient sdk.AccountClient) (*CentralizedMode, error) {
	configuredApps := make(map[string]struct{})
	for _, delegatingAppPrivateKey := range delegatingAppsPrivateKeys {
		appAddr, err := getAddressFromPrivateKey(delegatingAppPrivateKey)
		if err != nil {
			return nil, err
		}

		// No value is stored for the app at the address: the map entry is only used to verify the app has been configured for use.
		configuredApps[appAddr] = struct{}{}
	}

	// TODO_TECHDEBT: update this code to drop the extra conversion to Hex once the shannon-sdk is udated to accept either a specific
	// private key type or the generic PrivKey from github.com/cosmos/cosmos-sdk/crypto/types.
	gatewayKeySigner := signer{
		accountClient: accountClient,
		privateKeyHex: gatewayPrivateKeyHex,
	}

	return &CentralizedMode{
		configuredApps:   configuredApps,
		gatewayKeySigner: gatewayKeySigner,
		gatewayAddr:      gatewayAddr,
	}, nil
}

// CentralizedMode represent an operation mode which behaves as follows:
// 1. PATH (or more speicifcally the Shannon protocol integration instance) holds the private keys of the gateway operator's app(s).
// 2. All configured apps are owned by the gateway (PATH) operator.
// 3. All configured apps delegate (onchain) to the gateway address.
// 4. Each relay request is sent on behalf of one of the apps above (owned by the gateway operator)
// 5. Each relay request is signed by the gateway's private key (enabled by ring signatures supported by Shannon)
//
// See the following link for more details on PATH's Centralized operation mode.
// https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5?pvs=4#122a36edfff680d4a0fff3a40dea543e
type CentralizedMode struct {
	// gatewayAddr is the address of the the current instance of gateway (PATH).
	gatewayAddr string
	// gatewayKeySigner is the signer configured by the gateway (PATH) operator for signing all relay requests.
	// In the Centralized mode, all relay requests are signed by the gateway, as it has delegation from all the configured apps.
	gatewayKeySigner signer

	// configuredApps is a map of addresses of applications (owned) and configured for use by the gateway operator running PATH.
	// No value is needed for the entries: this map is only used to efficiently verify whether an app is configured for use by PATH.
	configuredApps map[string]struct{}
}

// BuildInstanceFromParams builds and returns a centralizedModeInstance.
// Note that, as of now, all instances of the Centralized mode are the same: the app for each relay is picked by the centralized mode.
// If there are multiple apps available for a specific relay/service request, the selected app depends on the endpoint picked by the user (the QoS for the service)
// This method implements the OperationMode interface consumed by the Protocol struct.
func (cm *CentralizedMode) BuildInstanceFromHTTPRequest(_ *http.Request) (OperationModeInstance, error) {
	return &centralizedModeInstance{
		configuredApps:   cm.configuredApps,
		gatewayKeySigner: cm.gatewayKeySigner,
		gatewayAddr:      cm.gatewayAddr,
	}, nil
}

// centralizedModeInstance is an instance of the Centralized Operation Mode.
// As of now, all instances of the Centralized operation mode are the same: they all use the same set of configured apps, and sign using the gateway's private key.
type centralizedModeInstance struct {
	// gatewayAddris the address of the the current instance of gateway (PATH).
	gatewayAddr string
	// gatewayKeySigner is the signer configured by the gateway (PATH) operator for signing all relay requests.
	// In the Centralized mode, all relay requests are signed by the gateway, as it has delegation from all the configured apps.
	gatewayKeySigner signer

	// configuredApps is a map of addresses of applications (owned) and configured for use by the gateway operator running PATH.
	// No value is needed for the entries: this map is only used to efficiently verify whether an app is configured for use by PATH.
	configuredApps map[string]struct{}
}

// GetAppFilterFn returns a filter function that determines whether an app can be used in the context of the Centralized mode of operation.
// In Centralized mode, an app can be used to pay for relays if it meets 2 conditions:
// 1. PATH has received its private key, and
// 2. The app delegates to the gateway address of the PATH instance.
//
// This function implements the OperationModeInstance interface consumed by the Protocol struct.
func (cmi *centralizedModeInstance) GetAppFilterFn() IsAppPermittedFn {
	return func(app *apptypes.Application) bool {
		return slices.Contains(app.DelegateeGatewayAddresses, cmi.gatewayAddr)
	}
}

// This function implements the OperationModeInstance interface consumed by the Protocol struct.
func (cmi *centralizedModeInstance) GetRelayRequestSigner() RelayRequestSigner {
	return cmi
}

func (cmi *centralizedModeInstance) SignRelayRequest(relayReq *servicetypes.RelayRequest, app apptypes.Application) (*servicetypes.RelayRequest, error) {
	// Verify the relay request's metadata, specifically the session header.
	// Note: cannot use the RelayRequest's ValidateBasic() method here, as it looks for a signature in the struct, which has not been added yet at this point.
	meta := relayReq.GetMeta()

	if meta.GetSessionHeader() == nil {
		return nil, errors.New("CentralizedMode relayRequestSigner: relay request is missing session header")
	}

	sessionHeader := meta.GetSessionHeader()
	if err := sessionHeader.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("CentralizedMode relayRequestSigner: relay request session header is invalid: %w", err)
	}

	// Verify the selected app is permitted for use by the gateway (PATH).
	if _, found := cmi.configuredApps[app.Address]; !found {
		return nil, fmt.Errorf("Centralized Mode relayRequestSigner: supplied app with address %s is not configured for use", app.Address)
	}

	// Sign the relay request using the gateway's private key.
	return cmi.gatewayKeySigner.SignRequest(relayReq, app)
}
