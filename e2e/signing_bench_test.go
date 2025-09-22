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

// BenchmarkShannonSigningDirect benchmarks the shannon-sdk signing operations directly
// without the full E2E infrastructure. This provides a focused measurement of the
// ethereum_secp256k1 build tag performance impact.
//
// Usage:
//
//	go test -bench=BenchmarkShannonSigningDirect -benchtime=30s -tags="bench"
//	go test -bench=BenchmarkShannonSigningDirect -benchtime=30s -tags="bench,ethereum_secp256k1"
func BenchmarkShannonSigningDirect(b *testing.B) {
	// Test private key (example key, not used in production)
	testPrivateKeyHex := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	// Create test relay request - simplified for benchmarking
	relayRequest := &servicetypes.RelayRequest{
		Meta: servicetypes.RelayRequestMetadata{
			SessionHeader: &sessiontypes.SessionHeader{
				ApplicationAddress:      "test_app_address",
				ServiceId:               "eth",
				SessionId:               "test_session_id",
				SessionStartBlockHeight: 1000,
				SessionEndBlockHeight:   2000,
			},
		},
		Payload: []byte(`{"jsonrpc":"2.0","method":"eth_blockNumber","id":1}`),
	}

	// Create test application
	testApp := apptypes.Application{
		Address: "test_app_address",
	}

	// Create a mock account client (minimal implementation for benchmarking)
	mockAccountClient := createMockAccountClient()

	b.ResetTimer()

	// Run the benchmark
	for range b.N {
		// Measure the signing operation
		start := time.Now()

		// Create application ring (part of signing setup)
		ring := sdk.NewApplicationRing(
			testApp,
			mockAccountClient,
		)

		// Create signer (includes key parsing/setup)
		sdkSigner, err := sdkcrypto.NewCryptoSigner(testPrivateKeyHex)
		if err != nil {
			b.Fatalf("Failed to create signer: %v", err)
		}

		// Perform the signing operation
		_, err = sdkSigner.Sign(context.Background(), relayRequest, ring)
		if err != nil {
			b.Fatalf("Failed to sign request: %v", err)
		}

		// Record the duration
		duration := time.Since(start)

		// Report metrics
		b.ReportMetric(float64(duration.Nanoseconds()), "ns/op")
	}
}

// mockAccountClient provides a minimal AccountClient implementation for benchmarking
type mockAccountClient struct {
	pubKey cryptotypes.PubKey
}

func (m *mockAccountClient) GetPubKeyFromAddress(ctx context.Context, address string) (cryptotypes.PubKey, error) {
	// Return a real secp256k1 public key for benchmarking
	return m.pubKey, nil
}

// createMockAccountClient creates a mock account client with a proper secp256k1 public key
func createMockAccountClient() *mockAccountClient {
	// Use the same private key to derive the public key
	privateKeyBytes, _ := hex.DecodeString("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	privKey := &secp256k1.PrivKey{Key: privateKeyBytes}
	pubKey := privKey.PubKey()

	return &mockAccountClient{
		pubKey: pubKey,
	}
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

	testApp := apptypes.Application{Address: "test_app_address"}
	mockAccountClient := createMockAccountClient()
	ring := sdk.NewApplicationRing(testApp, mockAccountClient)

	relayRequest := &servicetypes.RelayRequest{
		Meta: servicetypes.RelayRequestMetadata{
			SessionHeader: &sessiontypes.SessionHeader{
				ApplicationAddress:      "test_app_address",
				ServiceId:               "eth",
				SessionId:               "test_session_id",
				SessionStartBlockHeight: 1000,
				SessionEndBlockHeight:   2000,
			},
		},
		Payload: []byte(`{"jsonrpc":"2.0","method":"eth_blockNumber","id":1}`),
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
