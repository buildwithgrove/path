package shannon

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"sync/atomic"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	
	shannonmetrics "github.com/buildwithgrove/path/metrics/protocol/shannon"
)

const (
	// Default cache size - can hold signatures for multiple sessions
	defaultSignatureCacheSize = 100000
	
	// Default TTL matches session duration (15 minutes)
	defaultSignatureCacheTTL = 15 * time.Minute
)

// SignatureCache caches ring signatures to avoid expensive cryptographic operations
// for repeated requests within the same session.
type SignatureCache struct {
	// cache is the underlying LRU cache with expiration
	cache *lru.Cache[string, *cachedSignature]
	
	// mutex protects concurrent access to the cache
	mu sync.RWMutex
	
	// inFlight tracks computations in progress to prevent duplicate work
	inFlight sync.Map
	
	// logger for debugging and monitoring
	logger polylog.Logger
	
	// Metrics for monitoring cache effectiveness
	hits   atomic.Uint64
	misses atomic.Uint64
	
	// Configuration
	enabled bool
	ttl     time.Duration
}

// cachedSignature holds a cached signature and its expiration time
type cachedSignature struct {
	signature  []byte
	expiresAt  time.Time
	
	// Store the original metadata to validate cache consistency
	sessionID    string
	supplierAddr string
	appAddr      string
}

// SignatureCacheKey represents the components used to generate a cache key
type SignatureCacheKey struct {
	SessionID    string
	SupplierAddr string
	AppAddr      string
	PayloadHash  [32]byte // SHA256 hash of the serialized payload
}

// NewSignatureCache creates a new signature cache with the specified size and TTL
func NewSignatureCache(logger polylog.Logger, size int, ttl time.Duration, enabled bool) (*SignatureCache, error) {
	if size <= 0 {
		size = defaultSignatureCacheSize
	}
	if ttl <= 0 {
		ttl = defaultSignatureCacheTTL
	}
	
	cache, err := lru.New[string, *cachedSignature](size)
	if err != nil {
		return nil, err
	}
	
	sc := &SignatureCache{
		cache:   cache,
		logger:  logger.With("component", "signature_cache"),
		enabled: enabled,
		ttl:     ttl,
	}
	
	// Start cleanup goroutine to remove expired entries
	if enabled {
		go sc.cleanupExpired()
	}
	
	return sc, nil
}

// GetOrCompute tries to get a signature from cache, or computes it if not found
func (sc *SignatureCache) GetOrCompute(
	req *servicetypes.RelayRequest,
	app apptypes.Application,
	computeFn func() (*servicetypes.RelayRequest, error),
) (*servicetypes.RelayRequest, error) {
	// If caching is disabled, always compute
	if !sc.enabled {
		return computeFn()
	}
	
	// Build cache key
	key, err := sc.buildCacheKey(req, app)
	if err != nil {
		// If we can't build a cache key, fall back to computing
		sc.logger.Warn().Err(err).Msg("failed to build cache key, computing signature")
		return computeFn()
	}
	
	keyStr := sc.keyToString(key)
	
	// Try to get from cache
	if cached := sc.get(key); cached != nil {
		// Clone the request and apply the cached signature
		signedReq := sc.applyCachedSignature(req, cached)
		sc.hits.Add(1)
		
		// Record Prometheus metric for cache hit
		serviceID := "" // TODO: Get service ID from context
		shannonmetrics.RecordSignatureCacheHit(serviceID)
		shannonmetrics.RecordSignatureComputeTime(serviceID, "cached", 0.0001) // ~100 microseconds for cache hit
		
		sc.logger.Debug().
			Str("session_id", cached.sessionID).
			Str("app_addr", cached.appAddr).
			Uint64("hits", sc.hits.Load()).
			Float64("hit_rate", sc.getHitRate()).
			Msg("signature cache hit")
		
		return signedReq, nil
	}
	
	// Check if another goroutine is already computing this signature
	type result struct {
		req *servicetypes.RelayRequest
		err error
	}
	
	// Use a channel to coordinate concurrent computations
	ch := make(chan result)
	actual, loaded := sc.inFlight.LoadOrStore(keyStr, ch)
	if loaded {
		// Another goroutine is already computing, wait for result
		resultCh := actual.(chan result)
		res, ok := <-resultCh
		if !ok {
			// Channel was closed, the computation is complete, try cache again
			if cached := sc.get(key); cached != nil {
				signedReq := sc.applyCachedSignature(req, cached)
				sc.hits.Add(1)
				sc.logger.Debug().
					Str("session_id", req.Meta.SessionHeader.SessionId).
					Str("app_addr", app.Address).
					Uint64("hits", sc.hits.Load()).
					Float64("hit_rate", sc.getHitRate()).
					Msg("signature cache hit (after waiting)")
				return signedReq, nil
			}
			// Should not happen, but fall through to compute
			sc.misses.Add(1)
			return computeFn()
		}
		
		if res.err == nil {
			sc.hits.Add(1)
			sc.logger.Debug().
				Str("session_id", req.Meta.SessionHeader.SessionId).
				Str("app_addr", app.Address).
				Uint64("hits", sc.hits.Load()).
				Float64("hit_rate", sc.getHitRate()).
				Msg("signature cache hit (waited for in-flight computation)")
		}
		
		return res.req, res.err
	}
	
	// We are responsible for computing
	sc.misses.Add(1)
	
	// Measure computation time
	startTime := time.Now()
	signedReq, err := computeFn()
	computeDuration := time.Since(startTime).Seconds()
	
	// Record Prometheus metrics for cache miss and computation
	serviceID := "" // TODO: Get service ID from context
	shannonmetrics.RecordSignatureCacheMiss(serviceID, "not_found")
	shannonmetrics.RecordSignatureComputeTime(serviceID, "computed", computeDuration)
	
	// Store in cache first before notifying waiters
	if err == nil && signedReq != nil && signedReq.Meta.Signature != nil {
		sc.set(key, signedReq, app)
	}
	
	// Close channel to signal completion
	close(ch)
	
	// Remove from in-flight map
	sc.inFlight.Delete(keyStr)
	
	if err != nil {
		return nil, err
	}
	
	// Log the cache miss
	if signedReq != nil && signedReq.Meta.Signature != nil {
		sc.logger.Debug().
			Str("session_id", req.Meta.SessionHeader.SessionId).
			Str("app_addr", app.Address).
			Uint64("misses", sc.misses.Load()).
			Float64("hit_rate", sc.getHitRate()).
			Msg("signature cache miss - computed and cached")
	}
	
	return signedReq, nil
}

// buildCacheKey creates a deterministic cache key from the request and app
func (sc *SignatureCache) buildCacheKey(req *servicetypes.RelayRequest, app apptypes.Application) (*SignatureCacheKey, error) {
	if req.Meta.SessionHeader == nil {
		return nil, ErrMissingSessionHeader
	}
	
	// Hash the payload for the cache key
	payloadHash := sha256.Sum256(req.Payload)
	
	return &SignatureCacheKey{
		SessionID:    req.Meta.SessionHeader.SessionId,
		SupplierAddr: req.Meta.SupplierOperatorAddress,
		AppAddr:      app.Address,
		PayloadHash:  payloadHash,
	}, nil
}

// get retrieves a cached signature if it exists and hasn't expired
func (sc *SignatureCache) get(key *SignatureCacheKey) *cachedSignature {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	keyStr := sc.keyToString(key)
	cached, ok := sc.cache.Get(keyStr)
	if !ok {
		return nil
	}
	
	// Check if expired
	if time.Now().After(cached.expiresAt) {
		// Remove expired entry
		sc.cache.Remove(keyStr)
		return nil
	}
	
	// Validate cache consistency
	if cached.sessionID != key.SessionID ||
		cached.appAddr != key.AppAddr ||
		cached.supplierAddr != key.SupplierAddr {
		sc.logger.Warn().
			Str("cached_session", cached.sessionID).
			Str("key_session", key.SessionID).
			Msg("cache key mismatch - invalidating entry")
		sc.cache.Remove(keyStr)
		return nil
	}
	
	return cached
}

// set stores a signature in the cache
func (sc *SignatureCache) set(key *SignatureCacheKey, signedReq *servicetypes.RelayRequest, app apptypes.Application) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	cached := &cachedSignature{
		signature:    signedReq.Meta.Signature,
		expiresAt:    time.Now().Add(sc.ttl),
		sessionID:    key.SessionID,
		supplierAddr: key.SupplierAddr,
		appAddr:      key.AppAddr,
	}
	
	keyStr := sc.keyToString(key)
	sc.cache.Add(keyStr, cached)
}

// applyCachedSignature creates a new signed request using the cached signature
func (sc *SignatureCache) applyCachedSignature(req *servicetypes.RelayRequest, cached *cachedSignature) *servicetypes.RelayRequest {
	// Create a shallow copy of the request
	signedReq := &servicetypes.RelayRequest{
		Meta:    req.Meta,
		Payload: req.Payload,
	}
	
	// Apply the cached signature
	signedReq.Meta.Signature = cached.signature
	
	return signedReq
}

// keyToString converts a cache key to a string for use in the LRU cache
func (sc *SignatureCache) keyToString(key *SignatureCacheKey) string {
	return key.SessionID + ":" +
		key.SupplierAddr + ":" +
		key.AppAddr + ":" +
		hex.EncodeToString(key.PayloadHash[:])
}

// cleanupExpired periodically removes expired entries from the cache
func (sc *SignatureCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		sc.mu.Lock()
		
		// Get all keys and check for expiration
		keys := sc.cache.Keys()
		now := time.Now()
		removed := 0
		
		for _, key := range keys {
			if cached, ok := sc.cache.Peek(key); ok {
				if now.After(cached.expiresAt) {
					sc.cache.Remove(key)
					removed++
					// Record TTL expiration eviction
					shannonmetrics.RecordSignatureCacheEviction("ttl_expired")
				}
			}
		}
		
		sc.mu.Unlock()
		
		if removed > 0 {
			sc.logger.Debug().
				Int("removed", removed).
				Int("remaining", sc.cache.Len()).
				Msg("cleaned up expired cache entries")
		}
	}
}

// getHitRate returns the cache hit rate as a percentage
func (sc *SignatureCache) getHitRate() float64 {
	hits := sc.hits.Load()
	misses := sc.misses.Load()
	total := hits + misses
	
	if total == 0 {
		return 0
	}
	
	return float64(hits) / float64(total) * 100
}

// GetStats returns cache statistics for monitoring
func (sc *SignatureCache) GetStats() SignatureCacheStats {
	return SignatureCacheStats{
		Hits:    sc.hits.Load(),
		Misses:  sc.misses.Load(),
		HitRate: sc.getHitRate(),
		Size:    sc.cache.Len(),
		Enabled: sc.enabled,
	}
}

// SignatureCacheStats holds cache statistics
type SignatureCacheStats struct {
	Hits    uint64
	Misses  uint64
	HitRate float64
	Size    int
	Enabled bool
}

// Clear removes all entries from the cache
func (sc *SignatureCache) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	oldSize := sc.cache.Len()
	sc.cache.Purge()
	sc.hits.Store(0)
	sc.misses.Store(0)
	
	// Record evictions due to manual clear
	if oldSize > 0 {
		for i := 0; i < oldSize; i++ {
			shannonmetrics.RecordSignatureCacheEviction("manual_clear")
		}
	}
	
	sc.logger.Info().Msg("signature cache cleared")
}