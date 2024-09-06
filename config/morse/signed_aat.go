package morse

import (
	"fmt"

	sdk "github.com/pokt-foundation/pocket-go/provider"

	"github.com/buildwithgrove/path/config/utils"
)

// AATVersion is the current version of the Application Authentication Token (AAT).
// For details, see this PR: https://github.com/pokt-network/pocket-core/pull/1598
const AATVersion = "0.0.1"

const (
	applicationPublicKeyLength = 64
	clientPublicKeyLength      = 64
	applicationSignatureLength = 128
)

// SignedAAT represents the signed Application Authentication Token (AAT) necessary for signing relays.
type SignedAAT struct {
	ClientPublicKey      string `yaml:"client_public_key"`
	ApplicationPublicKey string `yaml:"application_public_key"`
	ApplicationSignature string `yaml:"application_signature"`
}

// AAT returns the PocketAAT representation of the signed AAT.
func (a SignedAAT) AAT() sdk.PocketAAT {
	return configToPocketAAT(a)
}

// validate checks if the signed AAT fields are valid.
func (a SignedAAT) validate() error {
	if !utils.IsValidHex(a.ApplicationPublicKey, applicationPublicKeyLength) {
		return fmt.Errorf("invalid application public key: must be a %d character hex code", applicationPublicKeyLength)
	}
	if !utils.IsValidHex(a.ClientPublicKey, clientPublicKeyLength) {
		return fmt.Errorf("invalid client public key: must be a %d character hex code", clientPublicKeyLength)
	}
	if !utils.IsValidHex(a.ApplicationSignature, applicationSignatureLength) {
		return fmt.Errorf("invalid application signature: must be a %d character hex code", applicationSignatureLength)
	}
	return nil
}

// configToPocketAAT converts a config AAT to a PocketAAT.
// Necessary to allow parsing of the AAT from the YAML config file.
func configToPocketAAT(a SignedAAT) sdk.PocketAAT {
	return sdk.PocketAAT{
		ClientPubKey: a.ClientPublicKey,
		AppPubKey:    a.ApplicationPublicKey,
		Signature:    a.ApplicationSignature,
		Version:      AATVersion,
	}
}
