package shannon

import (
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

// GetSec256k1PrivateKeyFromKeyHex returns a Sec256k1 private key from the supplied hex-encoded private key string.
// It allows any configuration-related code to build sec256k1 private keys from hex-encoded private keys.
func GetSec256k1PrivateKeyFromKeyHex(privateKeyHex string) (*secp256k1.PrivKey, error) {
	privateKeyBz, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, err
	}

	return &secp256k1.PrivKey{Key: privateKeyBz}, nil
}
