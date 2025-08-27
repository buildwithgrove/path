// Package shannon provides functionality for exporting Shannon protocol metrics to Prometheus.
package shannon

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Signature Cache Metrics Documentation
//
// The signature cache dramatically reduces CPU utilization by caching expensive ring signature
// computations. Within a 15-minute session, the same requests (same payload + supplier + app)
// will reuse cached signatures instead of recomputing them.
//
// Expected Impact:
// - 70-80% reduction in CPU usage for cryptographic operations
// - Signature computation reduced from ~10-50ms to <100μs for cache hits
// - Memory usage: ~50-55MB for 100k cache entries
//
// Cache Effectiveness:
// Even though sessions have high cardinality, the cache is highly effective because:
// - Sessions last 15 minutes with hundreds/thousands of requests per session
// - Many requests within a session are identical (e.g., repeated eth_blockNumber calls)
// - Popular RPC methods create natural request patterns that benefit from caching
//
// Key Metrics to Monitor:
//
// 1. Cache Hit Rate (target >90%):
//    rate(shannon_signature_cache_hits_total) / 
//    (rate(shannon_signature_cache_hits_total) + rate(shannon_signature_cache_misses_total))
//
// 2. Time Saved by Caching:
//    histogram_quantile(0.95, shannon_signature_cache_compute_time_seconds{cache_status="computed"})
//    vs
//    histogram_quantile(0.95, shannon_signature_cache_compute_time_seconds{cache_status="cached"})
//
// 3. Cache Saturation:
//    shannon_signature_cache_size / 100000
//
// 4. Eviction Pressure:
//    rate(shannon_signature_cache_evictions_total{reason="lru"})
//    High LRU evictions indicate cache size should be increased
//
// 5. Cache Efficiency by Service:
//    rate(shannon_signature_cache_hits_total) by (service_id)
//    Shows which chains benefit most from caching

const (
	// Signature cache metrics
	signatureCacheHitsTotalMetric    = "shannon_signature_cache_hits_total"
	signatureCacheMissesTotalMetric  = "shannon_signature_cache_misses_total"
	signatureCacheSizeMetric         = "shannon_signature_cache_size"
	signatureCacheEvictionsMetric    = "shannon_signature_cache_evictions_total"
	signatureCacheComputeTimeMetric  = "shannon_signature_cache_compute_time_seconds"
)

func init() {
	// Register signature cache metrics
	prometheus.MustRegister(signatureCacheHitsTotal)
	prometheus.MustRegister(signatureCacheMissesTotal)
	prometheus.MustRegister(signatureCacheSize)
	prometheus.MustRegister(signatureCacheEvictions)
	prometheus.MustRegister(signatureCacheComputeTime)
}

var (
	// signatureCacheHitsTotal tracks the total number of cache hits.
	// A cache hit occurs when a previously computed signature is found in cache.
	// Labels:
	//   - service_id: The service/chain identifier (e.g., "eth", "polygon")
	//
	// Use to analyze:
	//   - Cache effectiveness (hit rate = hits / (hits + misses))
	//   - Service-specific cache utilization
	//   - Which chains benefit most from caching
	signatureCacheHitsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      signatureCacheHitsTotalMetric,
			Help:      "Total number of signature cache hits",
		},
		[]string{"service_id"},
	)

	// signatureCacheMissesTotal tracks the total number of cache misses.
	// A cache miss occurs when a signature needs to be computed (not found in cache).
	// Labels:
	//   - service_id: The service/chain identifier
	//   - reason: Reason for miss ("not_found", "expired", "disabled")
	//
	// Use to analyze:
	//   - Cache miss patterns by service
	//   - TTL effectiveness (expired vs not_found)
	//   - Whether cache is enabled/disabled per service
	signatureCacheMissesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      signatureCacheMissesTotalMetric,
			Help:      "Total number of signature cache misses",
		},
		[]string{"service_id", "reason"},
	)

	// signatureCacheSize tracks the current number of entries in the cache.
	// This is a gauge metric that shows the instantaneous cache size.
	// No labels needed as we only have one signature cache instance.
	//
	// Use to analyze:
	//   - Cache utilization (size / 100000)
	//   - Memory usage patterns (~500 bytes per entry)
	//   - Cache growth over time
	signatureCacheSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: pathProcess,
			Name:      signatureCacheSizeMetric,
			Help:      "Current number of entries in the signature cache (max 100k)",
		},
	)

	// signatureCacheEvictions tracks the total number of cache evictions.
	// Evictions occur when the cache reaches capacity or entries expire.
	// Labels:
	//   - reason: Reason for eviction ("lru", "ttl_expired", "manual_clear")
	//
	// Use to analyze:
	//   - Cache capacity issues (high LRU evictions)
	//   - TTL configuration effectiveness
	//   - Cache clearing frequency
	signatureCacheEvictions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      signatureCacheEvictionsMetric,
			Help:      "Total number of signature cache evictions",
		},
		[]string{"reason"},
	)

	// signatureCacheComputeTime tracks the time taken to compute signatures.
	// This histogram measures the duration of cryptographic signature operations.
	// Labels:
	//   - service_id: The service/chain identifier
	//   - cache_status: Whether computation was needed ("computed" or "cached")
	//
	// Use to analyze:
	//   - Signature computation performance
	//   - Time saved by caching (compare "computed" vs "cached" latencies)
	//   - Performance impact of ring signature operations
	//   - P50, P95, P99 compute times
	signatureCacheComputeTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: pathProcess,
			Name:      signatureCacheComputeTimeMetric,
			Help:      "Histogram of signature computation times in seconds",
			// Buckets optimized for cryptographic operations (typically 1-100ms)
			Buckets: []float64{
				0.001, // 1ms
				0.005, // 5ms
				0.01,  // 10ms
				0.025, // 25ms
				0.05,  // 50ms
				0.1,   // 100ms
				0.25,  // 250ms
				0.5,   // 500ms
				1.0,   // 1s
			},
		},
		[]string{"service_id", "cache_status"},
	)
)

// PublishSignatureCacheSize updates the cache size gauge metric
func PublishSignatureCacheSize(size int) {
	signatureCacheSize.Set(float64(size))
}

// RecordSignatureCacheHit records a cache hit event
func RecordSignatureCacheHit(serviceID string) {
	signatureCacheHitsTotal.WithLabelValues(serviceID).Inc()
}

// RecordSignatureCacheMiss records a cache miss event
func RecordSignatureCacheMiss(serviceID, reason string) {
	signatureCacheMissesTotal.WithLabelValues(serviceID, reason).Inc()
}

// RecordSignatureCacheEviction records a cache eviction event
func RecordSignatureCacheEviction(reason string) {
	signatureCacheEvictions.WithLabelValues(reason).Inc()
}

// RecordSignatureComputeTime records the time taken for signature computation
func RecordSignatureComputeTime(serviceID string, cacheStatus string, duration float64) {
	signatureCacheComputeTime.WithLabelValues(serviceID, cacheStatus).Observe(duration)
}