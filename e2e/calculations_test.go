//go:build e2e

package e2e

import (
	"math"
	"slices"
	"time"
)

// ===== Success Rate Calculations =====

// calculateHTTPSuccessRate computes the HTTP success rate for metrics
func calculateHTTPSuccessRate(m *MethodMetrics) {
	m.RequestCount = m.Success + m.Failed
	if m.RequestCount > 0 {
		m.SuccessRate = float64(m.Success) / float64(m.RequestCount)
	}
}

// calculateJSONRPCSuccessRates computes all JSON-RPC related success rates
func calculateJSONRPCSuccessRates(m *MethodMetrics) {
	// JSON-RPC unmarshal success rate
	totalJSONAttempts := m.JSONRPCResponses + m.JSONRPCUnmarshalErrors
	if totalJSONAttempts > 0 {
		m.JSONRPCSuccessRate = float64(m.JSONRPCResponses) / float64(totalJSONAttempts)
	}

	// Only calculate these if we have valid JSON-RPC responses
	if m.JSONRPCResponses > 0 {
		// Error field absence rate (success = no error field)
		m.JSONRPCErrorFieldRate = float64(m.JSONRPCResponses-m.JSONRPCErrorField) / float64(m.JSONRPCResponses)

		// Non-nil result rate
		m.JSONRPCResultRate = float64(m.JSONRPCResponses-m.JSONRPCNilResult) / float64(m.JSONRPCResponses)

		// Validation success rate
		m.JSONRPCValidateRate = float64(m.JSONRPCResponses-m.JSONRPCValidateErrors) / float64(m.JSONRPCResponses)
	}
}

// calculateAllSuccessRates computes all success rates for a MethodMetrics struct
func calculateAllSuccessRates(m *MethodMetrics) {
	calculateHTTPSuccessRate(m)
	calculateJSONRPCSuccessRates(m)
}

// ===== Latency Calculations =====

// calculatePercentiles computes P50, P95, and P99 latency percentiles
func calculatePercentiles(m *MethodMetrics) {
	if len(m.Results) == 0 {
		return
	}

	// Extract latencies
	latencies := extractLatencies(m.Results)

	// Sort latencies
	slices.Sort(latencies)

	// Calculate percentiles
	m.P50 = percentile(latencies, 50)
	m.P95 = percentile(latencies, 95)
	m.P99 = percentile(latencies, 99)
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
func calculateServiceAverages(summary *serviceSummary, methodResults map[string]*MethodMetrics) {
	var totalLatency time.Duration
	var totalP50Latency time.Duration
	var totalP90Latency time.Duration
	var totalSuccessRate float64
	var methodsWithResults int

	for _, metrics := range methodResults {
		// Skip methods with no data
		if metrics == nil || len(metrics.Results) == 0 {
			continue
		}

		// Extract latencies for calculations
		latencies := extractLatencies(metrics.Results)

		// Calculate percentiles for this method
		p50 := calculateP50(latencies)
		p90 := calculateP90(latencies)
		avgLatency := calculateAvgLatency(latencies)

		// Add to summary totals
		totalLatency += avgLatency
		totalP50Latency += p50
		totalP90Latency += p90
		totalSuccessRate += metrics.SuccessRate
		methodsWithResults++

		// Accumulate totals for the service summary
		summary.TotalRequests += metrics.RequestCount
		summary.TotalSuccess += metrics.Success
		summary.TotalFailure += metrics.Failed
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
func collectServiceErrors(summary *serviceSummary, methodResults map[string]*MethodMetrics) {
	summary.MethodErrors = make(map[string]map[string]int)
	summary.TotalErrors = 0

	for method, metrics := range methodResults {
		if metrics == nil || len(metrics.Errors) == 0 {
			continue
		}

		// Initialize method errors map if not already created
		if summary.MethodErrors[method] == nil {
			summary.MethodErrors[method] = make(map[string]int)
		}

		// Copy errors to summary
		for errMsg, count := range metrics.Errors {
			summary.MethodErrors[method][errMsg] = count
			summary.TotalErrors += count
		}
	}
}
