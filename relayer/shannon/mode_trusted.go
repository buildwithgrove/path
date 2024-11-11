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

var (
	// TrustedMode implements the OperationMode interface consumed by the Protocol struct.
	// See the file mode.go for a detailed description of the Operation Mode concept.
	_ OperationMode = &TrustedMode{}

	// trustedModeInstance implements the OperationModeInstance interface consumed by the Protocol struct.
	// It is built by the TrustedMode struct based on a specific request's details (HTTP headers as of now).
	_ OperationModeInstance = &trustedModeInstance{}
)

const (
	// headerAppAddress is the key of the entry in HTTP headers that holds the target app's address.
	// The target app will be used to sign the relay request.
	headerAppAddr = "app_address"
)

// NewTrustedMode initializes an instance of TrustedMode using the supplied list of private keys.
// The supplied account client is required to instantiate relay request signers corresponding to the apps held by the Trusted operation mode.
func NewTrustedMode(appsPrivateKeys []*secp256k1.PrivKey, accountClient sdk.AccountClient) (*TrustedMode, error) {
	appsSigners := make(map[string]signer)
	for _, appPrivateKey := range appsPrivateKeys {
		appAddr, err := getAddressFromPrivateKey(appPrivateKey)
		if err != nil {
			return nil, err
		}

		// TODO_TECHDEBT: update this code to drop the extra conversion to Hex once the shannon-sdk is udated to accept either a specific
		// private key type or the generic PrivKey from github.com/cosmos/cosmos-sdk/crypto/types.
		privateKeyHex := hex.EncodeToString(appPrivateKey.Bytes())
		signer := signer{
			accountClient: accountClient,
			privateKeyHex: privateKeyHex,
		}

		appsSigners[appAddr] = signer
	}

	return &TrustedMode{
		appsSigners: appsSigners,
	}, nil
}

// TrustedMode represent an operation mode which behaves as follows:
// 1. PATH (or more speicifcally the Shannon protocol integration instance) holds the private keys of users' app(s).
// 2. Each relay request is signed by and sent on behalf of one of the held apps.
// 3. Users need to select a specific app for each relay request, done using HTTP request's headers as of now.
//
// See the following link for more details on PATH's Trusted operation mode.
// https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5?pvs=4#122a36edfff680eea2fbd46c7696d845
type TrustedMode struct {
	// appsSigners is a map of application address to the corresponding signer, i.e. the signer that holds the app's private key.
	// application address is used as the key to allow quick lookup of an app's corresponding signer for signing a relay request.
	appsSigners map[string]signer
}

// BuildInstanceFromParams builds and returns a trustedModeInstance based on the supplied params map, which as of now is the HTTP request's headers.
// It implements the OperationMode interface consumed by the Protocol struct.
func (tm *TrustedMode) BuildInstanceFromHTTPRequest(httpReq *http.Request) (OperationModeInstance, error) {
	if httpReq == nil || len(httpReq.Header) == 0 {
		return nil, fmt.Errorf("TrustedMode BuildInstanceFromHTTPRequest: no HTTP headers supplied.")
	}

	selectedAppAddr := httpReq.Header.Get(headerAppAddr)
	if selectedAppAddr == "" {
		return nil, fmt.Errorf("TrustedMode BuildInstanceFromHTTPRequest: a target app must be supplied as HTTP header %s", headerAppAddr)
	}

	// TODO_DISCUSS: need a solid method of verifying the user that sent the HTTP request has access to the requested app's private key.
	// This is needed to ensure only the app's owner can request that the relay be signed with the app's private key.
	selectedAppSigner, ok := tm.appsSigners[selectedAppAddr]
	if !ok {
		return nil, fmt.Errorf("TrustedMode BuildInstanceFromHTTPRequest: specified app %s not configured", selectedAppAddr)
	}

	return &trustedModeInstance{
		selectedAppAddr:   selectedAppAddr,
		selectedAppSigner: selectedAppSigner,
	}, nil
}

// trustedModeInstance is an instance of the Trusted Operation Mode, configured for a single application, as is the designed behavior of the Trusted Operation Mode.
// As of now, the headers of the HTTP request used as a relay request determine the application used for signing and sending the corresponding relay.
type trustedModeInstance struct {
	selectedAppAddr   string
	selectedAppSigner signer
}

// GetAppFilterFn returns a filter function that determines whether an app can be used in the context of the trusted mode's instance.
// In Trusted Mode, an instance matches a single application explicitly specified through HTTP request's headers at the time of creating the instance.
// This function implements the OperationModeInstance interface consumed by the Protocol struct.
func (tmi *trustedModeInstance) GetAppFilterFn() IsAppPermittedFn {
	return func(app *apptypes.Application) bool {
		return app.Address == tmi.selectedAppAddr
	}
}

// This function implements the OperationModeInstance interface consumed by the Protocol struct.
func (tmi *trustedModeInstance) GetRelayRequestSigner() RelayRequestSigner {
	return tmi
}

func (tmi *trustedModeInstance) SignRelayRequest(relayReq *servicetypes.RelayRequest, app apptypes.Application) (*servicetypes.RelayRequest, error) {
	// Verify the relay request's metadata, specifically the session header.
	// Note: cannot use the RelayRequest's ValidateBasic() method here, as it looks for a signature in the struct, which has not been added yet at this point.
	meta := relayReq.GetMeta()

	if meta.GetSessionHeader() == nil {
		return nil, errors.New("TrustedMode relayRequestSigner: relay request is missing session header")
	}

	sessionHeader := meta.GetSessionHeader()
	if err := sessionHeader.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("TrustedMode relayRequestSigner: relay request session header is invalid: %w", err)
	}

	// Verify the selected app matches the target app.
	if app.Address != tmi.selectedAppAddr {
		return nil,
			fmt.Errorf("Trusted Mode relayRequestSigner: supplied app with address %s does not match the selected app with address %s",
				app.Address,
				tmi.selectedAppAddr,
			)
	}

	// Sign the relay request using the selected app's private key
	return tmi.selectedAppSigner.SignRequest(relayReq, app)
}

// getAddressFromPrivateKey returns the address of the provided private key
func getAddressFromPrivateKey(privKey *secp256k1.PrivKey) (string, error) {
	addressBz := privKey.PubKey().Address()
	return bech32.ConvertAndEncode("pokt", addressBz)
}
