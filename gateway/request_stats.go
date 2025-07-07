package gateway

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
)

// RequestStatsCollector collects and periodically logs request statistics
type RequestStatsCollector struct {
	mu     sync.RWMutex
	logger polylog.Logger

	// Request counts
	totalRequests      atomic.Uint64
	parallelRequests   atomic.Uint64
	singleRequests     atomic.Uint64
	successfulRequests atomic.Uint64
	failedRequests     atomic.Uint64

	// TLD diversity tracking
	tldUsage   map[string]uint64
	tldSuccess map[string]uint64

	// Service breakdown
	serviceRequests map[protocol.ServiceID]uint64
	serviceSuccess  map[protocol.ServiceID]uint64

	// Timing metrics
	totalLatency   atomic.Uint64 // in milliseconds
	requestCount   atomic.Uint64 // for average calculation
	maxLatency     atomic.Uint64
	minLatency     atomic.Uint64
	winningIndexes map[int]uint64 // which endpoint index wins most often

	// Start time for rate calculation
	startTime time.Time
}

// NewRequestStatsCollector creates a new request statistics collector
func NewRequestStatsCollector(logger polylog.Logger) *RequestStatsCollector {
	rsc := &RequestStatsCollector{
		logger:          logger,
		tldUsage:        make(map[string]uint64),
		tldSuccess:      make(map[string]uint64),
		serviceRequests: make(map[protocol.ServiceID]uint64),
		serviceSuccess:  make(map[protocol.ServiceID]uint64),
		winningIndexes:  make(map[int]uint64),
		startTime:       time.Now(),
	}
	
	// Initialize min latency to a high value
	rsc.minLatency.Store(^uint64(0))
	
	return rsc
}

// RecordRequest records a request attempt
func (rsc *RequestStatsCollector) RecordRequest(serviceID protocol.ServiceID, isParallel bool) {
	rsc.totalRequests.Add(1)
	
	if isParallel {
		rsc.parallelRequests.Add(1)
	} else {
		rsc.singleRequests.Add(1)
	}
	
	rsc.mu.Lock()
	rsc.serviceRequests[serviceID]++
	rsc.mu.Unlock()
}

// RecordSuccess records a successful request with TLD and timing information
func (rsc *RequestStatsCollector) RecordSuccess(serviceID protocol.ServiceID, tld string, latencyMs uint64, winningIndex int) {
	rsc.successfulRequests.Add(1)
	
	// Update latency metrics
	rsc.totalLatency.Add(latencyMs)
	rsc.requestCount.Add(1)
	
	// Update max latency
	for {
		max := rsc.maxLatency.Load()
		if latencyMs <= max || rsc.maxLatency.CompareAndSwap(max, latencyMs) {
			break
		}
	}
	
	// Update min latency
	for {
		min := rsc.minLatency.Load()
		if latencyMs >= min || rsc.minLatency.CompareAndSwap(min, latencyMs) {
			break
		}
	}
	
	rsc.mu.Lock()
	if tld != "" {
		rsc.tldUsage[tld]++
		rsc.tldSuccess[tld]++
	}
	rsc.serviceSuccess[serviceID]++
	if winningIndex >= 0 {
		rsc.winningIndexes[winningIndex]++
	}
	rsc.mu.Unlock()
}

// RecordFailure records a failed request
func (rsc *RequestStatsCollector) RecordFailure(serviceID protocol.ServiceID) {
	rsc.failedRequests.Add(1)
}

// RecordTLDUsage records TLD usage for tracking diversity
func (rsc *RequestStatsCollector) RecordTLDUsage(tld string) {
	if tld == "" {
		return
	}
	
	rsc.mu.Lock()
	rsc.tldUsage[tld]++
	rsc.mu.Unlock()
}

// LogSummary logs a comprehensive summary of request patterns
func (rsc *RequestStatsCollector) LogSummary() {
	rsc.mu.RLock()
	defer rsc.mu.RUnlock()
	
	totalReqs := rsc.totalRequests.Load()
	if totalReqs == 0 {
		return
	}
	
	duration := time.Since(rsc.startTime)
	requestsPerSecond := float64(totalReqs) / duration.Seconds()
	
	successReqs := rsc.successfulRequests.Load()
	failedReqs := rsc.failedRequests.Load()
	successRate := float64(successReqs) * 100.0 / float64(totalReqs)
	
	parallelReqs := rsc.parallelRequests.Load()
	singleReqs := rsc.singleRequests.Load()
	parallelPercentage := float64(parallelReqs) * 100.0 / float64(totalReqs)
	
	// Calculate average latency
	avgLatency := uint64(0)
	count := rsc.requestCount.Load()
	if count > 0 {
		avgLatency = rsc.totalLatency.Load() / count
	}
	
	// Find most used TLDs
	topTLDs := rsc.getTopTLDs(5)
	
	// Find most successful services
	topServices := rsc.getTopServices(5)
	
	// Find winning index distribution
	winningDistribution := rsc.getWinningIndexDistribution()
	
	rsc.logger.Info().
		Uint64("total_requests", totalReqs).
		Uint64("successful_requests", successReqs).
		Uint64("failed_requests", failedReqs).
		Float64("success_rate_percent", successRate).
		Float64("requests_per_second", requestsPerSecond).
		Uint64("parallel_requests", parallelReqs).
		Uint64("single_requests", singleReqs).
		Float64("parallel_percentage", parallelPercentage).
		Uint64("avg_latency_ms", avgLatency).
		Uint64("min_latency_ms", rsc.minLatency.Load()).
		Uint64("max_latency_ms", rsc.maxLatency.Load()).
		Int("unique_tlds", len(rsc.tldUsage)).
		Int("unique_services", len(rsc.serviceRequests)).
		Str("top_tlds", topTLDs).
		Str("top_services", topServices).
		Str("winning_indexes", winningDistribution).
		Str("duration", duration.String()).
		Msg("Request patterns summary")
}

// getTopTLDs returns the top N most used TLDs
func (rsc *RequestStatsCollector) getTopTLDs(n int) string {
	type tldCount struct {
		tld   string
		count uint64
	}
	
	var tlds []tldCount
	for tld, count := range rsc.tldUsage {
		tlds = append(tlds, tldCount{tld, count})
	}
	
	// Sort by count
	for i := 0; i < len(tlds); i++ {
		for j := i + 1; j < len(tlds); j++ {
			if tlds[j].count > tlds[i].count {
				tlds[i], tlds[j] = tlds[j], tlds[i]
			}
		}
	}
	
	result := ""
	limit := n
	if len(tlds) < limit {
		limit = len(tlds)
	}
	
	for i := 0; i < limit; i++ {
		if i > 0 {
			result += ", "
		}
		successRate := float64(0)
		if successCount, ok := rsc.tldSuccess[tlds[i].tld]; ok && tlds[i].count > 0 {
			successRate = float64(successCount) * 100.0 / float64(tlds[i].count)
		}
		result += fmt.Sprintf("%s=%d (%.0f%% success)", tlds[i].tld, tlds[i].count, successRate)
	}
	
	return result
}

// getTopServices returns the top N most requested services
func (rsc *RequestStatsCollector) getTopServices(n int) string {
	type serviceCount struct {
		service protocol.ServiceID
		count   uint64
	}
	
	var services []serviceCount
	for service, count := range rsc.serviceRequests {
		services = append(services, serviceCount{service, count})
	}
	
	// Sort by count
	for i := 0; i < len(services); i++ {
		for j := i + 1; j < len(services); j++ {
			if services[j].count > services[i].count {
				services[i], services[j] = services[j], services[i]
			}
		}
	}
	
	result := ""
	limit := n
	if len(services) < limit {
		limit = len(services)
	}
	
	for i := 0; i < limit; i++ {
		if i > 0 {
			result += ", "
		}
		successRate := float64(0)
		if successCount, ok := rsc.serviceSuccess[services[i].service]; ok && services[i].count > 0 {
			successRate = float64(successCount) * 100.0 / float64(services[i].count)
		}
		result += fmt.Sprintf("%s=%d (%.0f%% success)", services[i].service, services[i].count, successRate)
	}
	
	return result
}

// getWinningIndexDistribution returns the distribution of winning endpoint indexes
func (rsc *RequestStatsCollector) getWinningIndexDistribution() string {
	if len(rsc.winningIndexes) == 0 {
		return "none"
	}
	
	result := ""
	total := uint64(0)
	for _, count := range rsc.winningIndexes {
		total += count
	}
	
	// Sort by index
	for i := 0; i < 10; i++ { // Assume max 10 endpoints
		if count, ok := rsc.winningIndexes[i]; ok {
			if result != "" {
				result += ", "
			}
			percentage := float64(count) * 100.0 / float64(total)
			result += fmt.Sprintf("index_%d=%.0f%%", i, percentage)
		}
	}
	
	return result
}

// StartPeriodicSummary starts a goroutine that logs summary statistics periodically
func (rsc *RequestStatsCollector) StartPeriodicSummary(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for range ticker.C {
			rsc.LogSummary()
		}
	}()
}