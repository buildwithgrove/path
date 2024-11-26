// package crypto contains all the cryptographic functionality required by Morse and Shannon.
package crypto

import (
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

// GetSecp256k1PrivateKeyFromKeyHex returns a Secp256k1 private key from the supplied hex-encoded private key string.
// It allows any configuration-related code to build secp256k1 private keys from hex-encoded private keys.
func GetSecp256k1PrivateKeyFromKeyHex(privateKeyHex string) (*secp256k1.PrivKey, error) {
	privateKeyBz, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, err
	}

	return &secp256k1.PrivKey{Key: privateKeyBz}, nil
}

// GetAddressFromPrivateKey returns the address of the provided private key
func GetAddressFromPrivateKey(privKey *secp256k1.PrivKey) (string, error) {
	addressBz := privKey.PubKey().Address()
	return bech32.ConvertAndEncode("pokt", addressBz)
}
