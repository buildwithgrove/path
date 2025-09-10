//go:build e2e

// Package e2e provides metrics calculation functions for PATH E2E and load testing.
//
// This file contains all mathematical and statistical calculation logic used to process
// test results from both HTTP (Vegeta) and WebSocket test executions. It handles success
// rate calculations, latency percentile computations, and service-level metric aggregation.
//
// CALCULATION CATEGORIES:
// - Success Rate Calculations: HTTP success rates and JSON-RPC validation success rates
// - Latency Calculations: P50, P90, P95, P99 percentiles and average latency computation
// - Service Summary Calculations: Cross-method aggregation for service-level reporting
// - Error Collection: Aggregation of error counts and types across test methods
//
// STATISTICAL FUNCTIONS:
// - Percentile calculation using ceiling-based indexing for sorted latency arrays
// - Average latency computation across all test results
// - Success rate calculations for HTTP responses and JSON-RPC validation steps
// - Service-level metric aggregation from individual method results
//
// TRANSPORT AGNOSTIC:
// All calculation functions work with the shared methodMetrics and VegetaResult structures,
// making them usable for both HTTP (Vegeta) and WebSocket test result processing.
// This ensures consistent metric calculation regardless of the underlying transport protocol.

package e2e

import (
	"math"
	"slices"
	"time"
)

// ===== Success Rate Calculations =====

// calculateHTTPSuccessRate computes the HTTP success rate for metrics
func calculateHTTPSuccessRate(m *methodMetrics) {
	m.requestCount = m.success + m.failed
	if m.requestCount > 0 {
		m.successRate = float64(m.success) / float64(m.requestCount)
	}
}

// calculateJSONRPCSuccessRates computes all JSON-RPC related success rates
func calculateJSONRPCSuccessRates(m *methodMetrics) {
	// JSON-RPC unmarshal success rate
	totalJSONAttempts := m.jsonrpcResponses + m.jsonrpcUnmarshalErrors
	if totalJSONAttempts > 0 {
		m.jsonrpcSuccessRate = float64(m.jsonrpcResponses) / float64(totalJSONAttempts)
	}

	// Only calculate these if we have valid JSON-RPC responses
	if m.jsonrpcResponses > 0 {
		// Error field absence rate (success = no error field)
		m.jsonrpcErrorFieldRate = float64(m.jsonrpcResponses-m.jsonrpcErrorField) / float64(m.jsonrpcResponses)

		// Non-nil result rate
		m.jsonrpcResultRate = float64(m.jsonrpcResponses-m.jsonrpcNilResult) / float64(m.jsonrpcResponses)

		// Validation success rate
		m.jsonrpcValidateRate = float64(m.jsonrpcResponses-m.jsonrpcValidateErrors) / float64(m.jsonrpcResponses)
	}
}

// calculateAllSuccessRates computes all success rates for a methodMetrics struct
func calculateAllSuccessRates(m *methodMetrics) {
	calculateHTTPSuccessRate(m)
	calculateJSONRPCSuccessRates(m)
}

// ===== Latency Calculations =====

// calculatePercentiles computes P50, P95, and P99 latency percentiles
func calculatePercentiles(m *methodMetrics) {
	if len(m.results) == 0 {
		return
	}

	// Extract latencies
	latencies := extractLatencies(m.results)

	// Sort latencies
	slices.Sort(latencies)

	// Calculate percentiles
	m.p50 = percentile(latencies, 50)
	m.p95 = percentile(latencies, 95)
	m.p99 = percentile(latencies, 99)
}

// extractLatencies extracts latency values from results
func extractLatencies(results []*VegetaResult) []time.Duration {
	latencies := make([]time.Duration, 0, len(results))
	for _, res := range results {
		latencies = append(latencies, res.Latency)
	}
	return latencies
}

// percentile calculates the p-th percentile of the given sorted slice
func percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}

	if p <= 0 {
		return sorted[0]
	}

	if p >= 100 {
		return sorted[len(sorted)-1]
	}

	// Calculate the index
	idx := int(math.Ceil(float64(p)/100.0*float64(len(sorted)))) - 1
	// Use max to ensure idx is at least 0
	idx = max(idx, 0)

	return sorted[idx]
}

// calculateP50 computes the 50th percentile latency
func calculateP50(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	// Sort latencies if they aren't already sorted
	slices.Sort(latencies)
	return percentile(latencies, 50)
}

// calculateP90 computes the 90th percentile latency
func calculateP90(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	// Sort latencies if they aren't already sorted
	slices.Sort(latencies)
	return percentile(latencies, 90)
}

// calculateP95 computes the 95th percentile latency
func calculateP95(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	// Sort latencies if they aren't already sorted
	slices.Sort(latencies)
	return percentile(latencies, 95)
}

// calculateP99 computes the 99th percentile latency
func calculateP99(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	// Sort latencies if they aren't already sorted
	slices.Sort(latencies)
	return percentile(latencies, 99)
}

// calculateAvgLatency computes the average latency
func calculateAvgLatency(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	var sum time.Duration
	for _, lat := range latencies {
		sum += lat
	}

	return time.Duration(int64(sum) / int64(len(latencies)))
}

// ===== Service Summary Calculations =====

// calculateServiceAverages calculates average metrics across all methods for a service
func calculateServiceAverages(summary *serviceSummary, methodResults map[string]*methodMetrics) {
	var totalLatency time.Duration
	var totalP50Latency time.Duration
	var totalP90Latency time.Duration
	var totalSuccessRate float64
	var methodsWithResults int

	for _, metrics := range methodResults {
		// Skip methods with no data
		if metrics == nil || len(metrics.results) == 0 {
			continue
		}

		// Extract latencies for calculations
		latencies := extractLatencies(metrics.results)

		// Calculate percentiles for this method
		p50 := calculateP50(latencies)
		p90 := calculateP90(latencies)
		avgLatency := calculateAvgLatency(latencies)

		// Add to summary totals
		totalLatency += avgLatency
		totalP50Latency += p50
		totalP90Latency += p90
		totalSuccessRate += metrics.successRate
		methodsWithResults++

		// Accumulate totals for the service summary
		summary.TotalRequests += metrics.requestCount
		summary.TotalSuccess += metrics.success
		summary.TotalFailure += metrics.failed
	}

	// Calculate averages if we have methods with results
	if methodsWithResults > 0 {
		summary.AvgLatency = time.Duration(int64(totalLatency) / int64(methodsWithResults))
		summary.AvgP50Latency = time.Duration(int64(totalP50Latency) / int64(methodsWithResults))
		summary.AvgP90Latency = time.Duration(int64(totalP90Latency) / int64(methodsWithResults))
		summary.AvgSuccessRate = totalSuccessRate / float64(methodsWithResults)
	}
}

// collectServiceErrors collects all errors from method metrics into the service summary
func collectServiceErrors(summary *serviceSummary, methodResults map[string]*methodMetrics) {
	summary.MethodErrors = make(map[string]map[string]int)
	summary.TotalErrors = 0

	for method, metrics := range methodResults {
		if metrics == nil || len(metrics.errors) == 0 {
			continue
		}

		// Initialize method errors map if not already created
		if summary.MethodErrors[method] == nil {
			summary.MethodErrors[method] = make(map[string]int)
		}

		// Copy errors to summary
		for errMsg, count := range metrics.errors {
			summary.MethodErrors[method][errMsg] = count
			summary.TotalErrors += count
		}
	}
}
