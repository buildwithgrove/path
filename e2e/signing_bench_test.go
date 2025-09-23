//go:build bench

package e2e

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"
	sdkcrypto "github.com/pokt-network/shannon-sdk/crypto"
)

// BenchmarkShannonSigningDirect benchmarks the shannon-sdk signing operations using
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

	// Use the app private key for signing (convert to hex)
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
func BenchmarkShannonKeyOperations(b *testing.B) {
	testPrivateKeyHex := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	b.ResetTimer()

	for range b.N {
		start := time.Now()

		// Benchmark key creation/parsing
		_, err := sdkcrypto.NewCryptoSigner(testPrivateKeyHex)
		if err != nil {
			b.Fatalf("Failed to create signer: %v", err)
		}

		duration := time.Since(start)
		b.ReportMetric(float64(duration.Nanoseconds()), "ns/op")
	}
}

// Benchmark for the complete signing pipeline
func BenchmarkShannonCompleteSigningPipeline(b *testing.B) {
	testPrivateKeyHex := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	// Pre-create signer to isolate signing performance
	sdkSigner, err := sdkcrypto.NewCryptoSigner(testPrivateKeyHex)
	if err != nil {
		b.Fatalf("Failed to create signer: %v", err)
	}

	// Pre-create test application and ring setup like the optimized benchmark
	appPrivKey := secp256k1.GenPrivKey()
	supplierPrivKey1 := secp256k1.GenPrivKey()
	supplierPrivKey2 := secp256k1.GenPrivKey()

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

	for range b.N {
		start := time.Now()

		// Just benchmark the actual signing operation
		_, err := sdkSigner.Sign(context.Background(), relayRequest, ring)
		if err != nil {
			b.Fatalf("Failed to sign request: %v", err)
		}

		duration := time.Since(start)
		b.ReportMetric(float64(duration.Nanoseconds()), "ns/op")
	}
}
