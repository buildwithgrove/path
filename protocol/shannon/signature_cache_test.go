package shannon

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	"github.com/stretchr/testify/require"
)

func TestSignatureCache_GetOrCompute(t *testing.T) {
	logger := polyzero.NewLogger()
	cache, err := NewSignatureCache(logger, 100, 15*time.Minute, true)
	require.NoError(t, err)
	require.NotNil(t, cache)

	app := apptypes.Application{
		Address: "app1",
	}

	sessionHeader := &sessiontypes.SessionHeader{
		SessionId: "session1",
		ApplicationAddress: "app1",
	}

	req := &servicetypes.RelayRequest{
		Meta: servicetypes.RelayRequestMetadata{
			SessionHeader: sessionHeader,
			SupplierOperatorAddress: "supplier1",
		},
		Payload: []byte("test payload"),
	}

	computeCalls := 0
	computeFn := func() (*servicetypes.RelayRequest, error) {
		computeCalls++
		signedReq := *req
		signedReq.Meta.Signature = []byte("signature")
		return &signedReq, nil
	}

	// First call should compute
	result1, err := cache.GetOrCompute(req, app, computeFn)
	require.NoError(t, err)
	require.NotNil(t, result1)
	require.Equal(t, []byte("signature"), result1.Meta.Signature)
	require.Equal(t, 1, computeCalls)

	// Second call should hit cache
	result2, err := cache.GetOrCompute(req, app, computeFn)
	require.NoError(t, err)
	require.NotNil(t, result2)
	require.Equal(t, []byte("signature"), result2.Meta.Signature)
	require.Equal(t, 1, computeCalls) // No additional compute

	// Verify stats
	stats := cache.GetStats()
	require.Equal(t, uint64(1), stats.Hits)
	require.Equal(t, uint64(1), stats.Misses)
}

func TestSignatureCache_DifferentRequests(t *testing.T) {
	logger := polyzero.NewLogger()
	cache, err := NewSignatureCache(logger, 100, 15*time.Minute, true)
	require.NoError(t, err)

	app := apptypes.Application{
		Address: "app1",
	}

	sessionHeader := &sessiontypes.SessionHeader{
		SessionId: "session1",
		ApplicationAddress: "app1",
	}

	req1 := &servicetypes.RelayRequest{
		Meta: servicetypes.RelayRequestMetadata{
			SessionHeader: sessionHeader,
			SupplierOperatorAddress: "supplier1",
		},
		Payload: []byte("payload1"),
	}

	req2 := &servicetypes.RelayRequest{
		Meta: servicetypes.RelayRequestMetadata{
			SessionHeader: sessionHeader,
			SupplierOperatorAddress: "supplier1",
		},
		Payload: []byte("payload2"),
	}

	computeCalls := 0
	computeFn := func(req *servicetypes.RelayRequest) func() (*servicetypes.RelayRequest, error) {
		return func() (*servicetypes.RelayRequest, error) {
			computeCalls++
			signedReq := *req
			// Different signatures for different payloads
			hash := sha256.Sum256(req.Payload)
			signedReq.Meta.Signature = []byte("sig:" + hex.EncodeToString(hash[:]))
			return &signedReq, nil
		}
	}

	// First request
	result1, err := cache.GetOrCompute(req1, app, computeFn(req1))
	require.NoError(t, err)
	require.NotNil(t, result1)
	require.Equal(t, 1, computeCalls)

	// Different request should compute again
	result2, err := cache.GetOrCompute(req2, app, computeFn(req2))
	require.NoError(t, err)
	require.NotNil(t, result2)
	require.Equal(t, 2, computeCalls)

	// Different signatures
	require.NotEqual(t, result1.Meta.Signature, result2.Meta.Signature)

	// Same request should hit cache
	result1Again, err := cache.GetOrCompute(req1, app, computeFn(req1))
	require.NoError(t, err)
	require.Equal(t, result1.Meta.Signature, result1Again.Meta.Signature)
	require.Equal(t, 2, computeCalls) // No additional compute
}

func TestSignatureCache_TTLExpiration(t *testing.T) {
	logger := polyzero.NewLogger()
	// Use very short TTL for testing
	cache, err := NewSignatureCache(logger, 100, 10*time.Millisecond, true)
	require.NoError(t, err)

	app := apptypes.Application{
		Address: "app1",
	}

	sessionHeader := &sessiontypes.SessionHeader{
		SessionId: "session1",
		ApplicationAddress: "app1",
	}

	req := &servicetypes.RelayRequest{
		Meta: servicetypes.RelayRequestMetadata{
			SessionHeader: sessionHeader,
			SupplierOperatorAddress: "supplier1",
		},
		Payload: []byte("test payload"),
	}

	computeCalls := 0
	computeFn := func() (*servicetypes.RelayRequest, error) {
		computeCalls++
		signedReq := *req
		signedReq.Meta.Signature = []byte("signature")
		return &signedReq, nil
	}

	// First call should compute
	_, err = cache.GetOrCompute(req, app, computeFn)
	require.NoError(t, err)
	require.Equal(t, 1, computeCalls)

	// Wait for TTL to expire
	time.Sleep(15 * time.Millisecond)

	// Should compute again due to expiration
	_, err = cache.GetOrCompute(req, app, computeFn)
	require.NoError(t, err)
	require.Equal(t, 2, computeCalls)
}

func TestSignatureCache_Disabled(t *testing.T) {
	logger := polyzero.NewLogger()
	cache, err := NewSignatureCache(logger, 100, 15*time.Minute, false)
	require.NoError(t, err)

	app := apptypes.Application{
		Address: "app1",
	}

	sessionHeader := &sessiontypes.SessionHeader{
		SessionId: "session1",
		ApplicationAddress: "app1",
	}

	req := &servicetypes.RelayRequest{
		Meta: servicetypes.RelayRequestMetadata{
			SessionHeader: sessionHeader,
			SupplierOperatorAddress: "supplier1",
		},
		Payload: []byte("test payload"),
	}

	computeCalls := 0
	computeFn := func() (*servicetypes.RelayRequest, error) {
		computeCalls++
		signedReq := *req
		signedReq.Meta.Signature = []byte("signature")
		return &signedReq, nil
	}

	// Should always compute when disabled
	_, err = cache.GetOrCompute(req, app, computeFn)
	require.NoError(t, err)
	require.Equal(t, 1, computeCalls)

	_, err = cache.GetOrCompute(req, app, computeFn)
	require.NoError(t, err)
	require.Equal(t, 2, computeCalls) // Should compute again

	// Stats should show no cache activity
	stats := cache.GetStats()
	require.Equal(t, uint64(0), stats.Hits)
	require.Equal(t, uint64(0), stats.Misses)
}

func TestSignatureCache_MissingSessionHeader(t *testing.T) {
	logger := polyzero.NewLogger()
	cache, err := NewSignatureCache(logger, 100, 15*time.Minute, true)
	require.NoError(t, err)

	app := apptypes.Application{
		Address: "app1",
	}

	req := &servicetypes.RelayRequest{
		Meta: servicetypes.RelayRequestMetadata{
			SupplierOperatorAddress: "supplier1",
			// Missing SessionHeader
		},
		Payload: []byte("test payload"),
	}

	computeFn := func() (*servicetypes.RelayRequest, error) {
		signedReq := *req
		signedReq.Meta.Signature = []byte("signature")
		return &signedReq, nil
	}

	// Should compute directly without caching
	result, err := cache.GetOrCompute(req, app, computeFn)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Stats should show no cache activity
	stats := cache.GetStats()
	require.Equal(t, uint64(0), stats.Hits)
	require.Equal(t, uint64(0), stats.Misses)
}

func TestSignatureCache_Clear(t *testing.T) {
	logger := polyzero.NewLogger()
	cache, err := NewSignatureCache(logger, 100, 15*time.Minute, true)
	require.NoError(t, err)

	app := apptypes.Application{
		Address: "app1",
	}

	sessionHeader := &sessiontypes.SessionHeader{
		SessionId: "session1",
		ApplicationAddress: "app1",
	}

	req := &servicetypes.RelayRequest{
		Meta: servicetypes.RelayRequestMetadata{
			SessionHeader: sessionHeader,
			SupplierOperatorAddress: "supplier1",
		},
		Payload: []byte("test payload"),
	}

	computeCalls := 0
	computeFn := func() (*servicetypes.RelayRequest, error) {
		computeCalls++
		signedReq := *req
		signedReq.Meta.Signature = []byte("signature")
		return &signedReq, nil
	}

	// First call should compute
	_, err = cache.GetOrCompute(req, app, computeFn)
	require.NoError(t, err)
	require.Equal(t, 1, computeCalls)

	// Clear cache
	cache.Clear()

	// Should compute again after clear
	_, err = cache.GetOrCompute(req, app, computeFn)
	require.NoError(t, err)
	require.Equal(t, 2, computeCalls)
}

func TestSignatureCache_ConcurrentAccess(t *testing.T) {
	logger := polyzero.NewLogger()
	cache, err := NewSignatureCache(logger, 100, 15*time.Minute, true)
	require.NoError(t, err)

	app := apptypes.Application{
		Address: "app1",
	}

	sessionHeader := &sessiontypes.SessionHeader{
		SessionId: "session1",
		ApplicationAddress: "app1",
	}

	req := &servicetypes.RelayRequest{
		Meta: servicetypes.RelayRequestMetadata{
			SessionHeader: sessionHeader,
			SupplierOperatorAddress: "supplier1",
		},
		Payload: []byte("test payload"),
	}

	computeCalls := 0
	computeFn := func() (*servicetypes.RelayRequest, error) {
		computeCalls++
		time.Sleep(10 * time.Millisecond) // Simulate computation time
		signedReq := *req
		signedReq.Meta.Signature = []byte("signature")
		return &signedReq, nil
	}

	// Run multiple goroutines accessing cache concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			result, err := cache.GetOrCompute(req, app, computeFn)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, []byte("signature"), result.Meta.Signature)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have computed only once despite concurrent access
	require.Equal(t, 1, computeCalls)

	// Verify stats
	stats := cache.GetStats()
	require.Equal(t, uint64(9), stats.Hits) // 9 hits
	require.Equal(t, uint64(1), stats.Misses) // 1 miss (first call)
}

func BenchmarkSignatureCache_Hit(b *testing.B) {
	logger := polyzero.NewLogger()
	cache, err := NewSignatureCache(logger, 1000, 15*time.Minute, true)
	require.NoError(b, err)

	app := apptypes.Application{
		Address: "app1",
	}

	sessionHeader := &sessiontypes.SessionHeader{
		SessionId: "session1",
		ApplicationAddress: "app1",
	}

	req := &servicetypes.RelayRequest{
		Meta: servicetypes.RelayRequestMetadata{
			SessionHeader: sessionHeader,
			SupplierOperatorAddress: "supplier1",
		},
		Payload: []byte("test payload"),
	}

	computeFn := func() (*servicetypes.RelayRequest, error) {
		signedReq := *req
		signedReq.Meta.Signature = []byte("signature")
		return &signedReq, nil
	}

	// Prime the cache
	_, _ = cache.GetOrCompute(req, app, computeFn)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.GetOrCompute(req, app, computeFn)
	}
}

func BenchmarkSignatureCache_Miss(b *testing.B) {
	logger := polyzero.NewLogger()
	cache, err := NewSignatureCache(logger, 1000, 15*time.Minute, true)
	require.NoError(b, err)

	app := apptypes.Application{
		Address: "app1",
	}

	sessionHeader := &sessiontypes.SessionHeader{
		SessionId: "session1",
		ApplicationAddress: "app1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := &servicetypes.RelayRequest{
			Meta: servicetypes.RelayRequestMetadata{
				SessionHeader: sessionHeader,
				SupplierOperatorAddress: "supplier1",
			},
			Payload: []byte("test payload " + string(rune(i))),
		}

		computeFn := func() (*servicetypes.RelayRequest, error) {
			signedReq := *req
			signedReq.Meta.Signature = []byte("signature")
			return &signedReq, nil
		}

		_, _ = cache.GetOrCompute(req, app, computeFn)
	}
}