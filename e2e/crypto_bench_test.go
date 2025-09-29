//go:build bench

package e2e

import (
	"context"
	"encoding/hex"
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"
)

// BenchmarkShannonSigningDirect benchmarks the shannon-sdk crypto operations using
// the same approach as Shannon SDK's own benchmarks for direct comparison.
//
// Usage:
//
//	go test -bench=BenchmarkShannonSigningDirect -benchtime=30s -tags="bench"
//	go test -bench=BenchmarkShannonSigningDirect -benchtime=30s -tags="bench,ethereum_secp256k1"
func BenchmarkShannonSigningDirect(b *testing.B) {
	// Use Shannon SDK's EXACT approach: generate keys like their benchmark
	appPrivKey := secp256k1.GenPrivKey()
	supplierPrivKey1 := secp256k1.GenPrivKey()
	supplierPrivKey2 := secp256k1.GenPrivKey()

	// Use the app private key for crypto (convert to hex)
	privateKeyHex := hex.EncodeToString(appPrivKey.Bytes())

	signer, err := sdk.NewSignerFromHex(privateKeyHex)
	if err != nil {
		b.Fatalf("Failed to create signer: %v", err)
	}

	// Create a mock public key fetcher with corresponding public keys
	pubKeyFetcher := &mockPublicKeyFetcher{
		publicKeys: map[string]cryptotypes.PubKey{
			"pokt1app1":      appPrivKey.PubKey(),
			"pokt1supplier1": supplierPrivKey1.PubKey(),
			"pokt1supplier2": supplierPrivKey2.PubKey(),
		},
	}

	// Create an application
	app := apptypes.Application{
		Address: "pokt1app1",
	}

	appRing := sdk.NewApplicationRing(app, pubKeyFetcher)

	// Create a relay request exactly like Shannon SDK
	relayRequest := &servicetypes.RelayRequest{
		Meta: servicetypes.RelayRequestMetadata{
			SessionHeader: &sessiontypes.SessionHeader{
				ApplicationAddress:      "pokt1app1",
				ServiceId:               "test-service",
				SessionStartBlockHeight: 1,
				SessionEndBlockHeight:   10,
			},
			SupplierOperatorAddress: "pokt1supplier1",
		},
		Payload: []byte(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := signer.Sign(context.Background(), relayRequest, appRing)
		if err != nil {
			b.Fatalf("Sign failed: %v", err)
		}
	}
}

// mockPublicKeyFetcher is a test implementation of PublicKeyFetcher that matches Shannon SDK's approach
type mockPublicKeyFetcher struct {
	publicKeys map[string]cryptotypes.PubKey
}

func (m *mockPublicKeyFetcher) GetPubKeyFromAddress(ctx context.Context, address string) (cryptotypes.PubKey, error) {
	if pubKey, exists := m.publicKeys[address]; exists {
		return pubKey, nil
	}
	return nil, nil
}

// Additional benchmark for just the signer creation (key operations)
var benchSink any

func withSilencedOutput(fn func()) {
	// Best-effort: silence stdout/stderr during noisy operations
	oldStdout, oldStderr := os.Stdout, os.Stderr
	devnull, err := os.Open(os.DevNull)
	if err == nil {
		os.Stdout = devnull
		os.Stderr = devnull
		defer func() {
			os.Stdout = oldStdout
			os.Stderr = oldStderr
			_ = devnull.Close()
		}()
	}
	fn()
}

func BenchmarkShannonKeyOperations(b *testing.B) {
	testPrivateKeyHex := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		withSilencedOutput(func() {
			s, err := sdk.NewSignerFromHex(testPrivateKeyHex)
			if err != nil {
				b.Fatalf("Failed to create signer: %v", err)
			}
			benchSink = s
		})
	}
}

// Benchmark for the complete crypto pipeline
func BenchmarkShannonCompleteSigningPipeline(b *testing.B) {
	// Use a consistent keypair between signer and ring
	appPrivKey := secp256k1.GenPrivKey()
	supplierPrivKey1 := secp256k1.GenPrivKey()
	supplierPrivKey2 := secp256k1.GenPrivKey()

	privateKeyHex := hex.EncodeToString(appPrivKey.Bytes())

	// Pre-create signer to isolate crypto performance
	sdkSigner, err := sdk.NewSignerFromHex(privateKeyHex)
	if err != nil {
		b.Fatalf("Failed to create signer: %v", err)
	}

	pubKeyFetcher := &mockPublicKeyFetcher{
		publicKeys: map[string]cryptotypes.PubKey{
			"pokt1app1":      appPrivKey.PubKey(),
			"pokt1supplier1": supplierPrivKey1.PubKey(),
			"pokt1supplier2": supplierPrivKey2.PubKey(),
		},
	}

	testApp := apptypes.Application{Address: "pokt1app1"}
	ring := sdk.NewApplicationRing(testApp, pubKeyFetcher)

	relayRequest := &servicetypes.RelayRequest{
		Meta: servicetypes.RelayRequestMetadata{
			SessionHeader: &sessiontypes.SessionHeader{
				ApplicationAddress:      "pokt1app1",
				ServiceId:               "test-service",
				SessionStartBlockHeight: 1,
				SessionEndBlockHeight:   10,
			},
			SupplierOperatorAddress: "pokt1supplier1",
		},
		Payload: []byte(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := sdkSigner.Sign(context.Background(), relayRequest, ring)
		if err != nil {
			b.Fatalf("Failed to sign request: %v", err)
		}
	}
}
