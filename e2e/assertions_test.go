//go:build e2e

// ===== Assertions Test File =====
//
// This file contains all assertion, validation, and reporting functionality
// for the E2E load tests. It was separated from vegeta_test.go to improve
// code organization and maintainability.
//
// Contents:
// - Test result validation and assertions
// - Colored terminal output and formatting
// - Service summary reporting
// - Latency and success rate calculations

package e2e

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/buildwithgrove/path/protocol"
)

// ===== ANSI Color Constants =====
//
// • Used for terminal output formatting
const (
	RED       = "\x1b[31m"
	GREEN     = "\x1b[32m"
	YELLOW    = "\x1b[33m"
	BLUE      = "\x1b[34m"
	CYAN      = "\x1b[36m"
	BOLD      = "\x1b[1m"
	BOLD_BLUE = "\x1b[1m\x1b[34m"
	BOLD_CYAN = "\x1b[1m\x1b[36m"
	RESET     = "\x1b[0m"
)

// ===== Assertion and Validation Functions =====

// validateResults performs assertions on test metrics and returns whether the method failed
func validateResults(t *testing.T, serviceId protocol.ServiceID, m *methodMetrics, serviceConfig ServiceConfig) bool {
	// Create a slice to collect all assertion failures
	var failures []string

	// Add a blank line before each test result for better readability
	fmt.Println()

	// Print metrics header with method name in blue
	fmt.Printf("%s~~~~~~~~~~ %s - %s ~~~~~~~~%s\n\n", BOLD_CYAN, serviceId, m.method, RESET)

	// Print success rate with color (green ≥90%, yellow ≥70%, red <70%)
	successColor := RED // Red by default
	if m.successRate >= 0.90 {
		successColor = GREEN // Green for ≥90%
	} else if m.successRate >= 0.70 {
		successColor = YELLOW // Yellow for ≥70%
	}
	fmt.Printf("%sHTTP Success Rate%s: %s%.2f%%%s (%d/%d requests)\n",
		BOLD, RESET, successColor, m.successRate*100, RESET, m.success, m.requestCount)

	// Print latencies (yellow if close to limit, green if well below)
	p50Color := getLatencyColor(m.p50, serviceConfig.MaxP50LatencyMS)
	p95Color := getLatencyColor(m.p95, serviceConfig.MaxP95LatencyMS)
	p99Color := getLatencyColor(m.p99, serviceConfig.MaxP99LatencyMS)
	fmt.Printf("%sLatency P50%s: %s%s%s\n", BOLD, RESET, p50Color, formatLatency(m.p50), RESET)
	fmt.Printf("%sLatency P95%s: %s%s%s\n", BOLD, RESET, p95Color, formatLatency(m.p95), RESET)
	fmt.Printf("%sLatency P99%s: %s%s%s\n", BOLD, RESET, p99Color, formatLatency(m.p99), RESET)

	// Print JSON-RPC metrics with coloring
	if m.jsonRPCResponses+m.jsonRPCUnmarshalErrors > 0 {
		fmt.Printf("%sJSON-RPC Metrics:%s\n", BOLD, RESET)

		if m.jsonRPCResponses > 0 {
			// Unmarshal success rate
			color := getRateColor(m.jsonRPCSuccessRate, serviceConfig.SuccessRate)
			fmt.Printf("  Unmarshal Success: %s%.2f%%%s (%d/%d responses)\n",
				color, m.jsonRPCSuccessRate*100, RESET, m.jsonRPCResponses, m.jsonRPCResponses+m.jsonRPCUnmarshalErrors)
			// Validation success rate
			color = getRateColor(m.jsonRPCValidateRate, serviceConfig.SuccessRate)
			fmt.Printf("  Validation Success: %s%.2f%%%s (%d/%d responses)\n",
				color, m.jsonRPCValidateRate*100, RESET, m.jsonRPCResponses-m.jsonRPCValidateErrors, m.jsonRPCResponses)
			// Non-nil result rate
			color = getRateColor(m.jsonRPCResultRate, serviceConfig.SuccessRate)
			fmt.Printf("  Has Result: %s%.2f%%%s (%d/%d responses)\n",
				color, m.jsonRPCResultRate*100, RESET, m.jsonRPCResponses-m.jsonRPCNilResult, m.jsonRPCResponses)
			// Error field absent rate
			color = getRateColor(m.jsonRPCErrorFieldRate, serviceConfig.SuccessRate)
			fmt.Printf("  Does Not Have Error: %s%.2f%%%s (%d/%d responses)\n",
				color, m.jsonRPCErrorFieldRate*100, RESET, m.jsonRPCResponses-m.jsonRPCErrorField, m.jsonRPCResponses)
		}
	}

	// Log status codes
	if len(m.statusCodes) > 0 {
		statusText := "Status Codes:"
		for code, count := range m.statusCodes {
			codeColor := GREEN // Green for 2xx
			if code >= 400 {
				codeColor = RED // Red for 4xx/5xx
			} else if code >= 300 {
				codeColor = YELLOW // Yellow for 3xx
			}
			statusText += fmt.Sprintf("\n  %s%d%s: %d", codeColor, code, RESET, count)
		}
		fmt.Println(statusText)
	}

	// Determine if the test passed based on our metrics
	testPassed := m.successRate >= serviceConfig.SuccessRate &&
		m.p50 <= serviceConfig.MaxP50LatencyMS &&
		m.p95 <= serviceConfig.MaxP95LatencyMS &&
		m.p99 <= serviceConfig.MaxP99LatencyMS

	// Choose error color based on test passing status
	errorColor := YELLOW // Yellow for warnings (test passed despite errors)
	if !testPassed {
		errorColor = RED // Red for critical errors (test failed)
	}

	// Log top errors with appropriate color and include detailed JSON-RPC errors
	if len(m.errors) > 0 {
		fmt.Println("") // Add a new line before logging errors
		fmt.Printf("%sTop errors:%s\n", errorColor, RESET)

		// Sort errors by count (descending) to show most frequent first
		type errorEntry struct {
			message string
			count   int
		}
		var sortedErrors []errorEntry
		for errMsg, count := range m.errors {
			sortedErrors = append(sortedErrors, errorEntry{message: errMsg, count: count})
		}
		sort.Slice(sortedErrors, func(i, j int) bool {
			return sortedErrors[i].count > sortedErrors[j].count
		})

		// Display top 5 errors
		for i, err := range sortedErrors {
			if i >= 5 {
				break
			}
			fmt.Printf("  %d. %s%s%s: %d\n", i+1, errorColor, err.message, RESET, err.count)
		}

		if len(sortedErrors) > 5 {
			fmt.Printf("  ... and %s%d%s more error types\n", errorColor, len(sortedErrors)-5, RESET)
		}
	}

	// Collect assertion failures
	failures = append(failures, collectHTTPSuccessRateFailures(m, serviceConfig.SuccessRate)...)
	failures = append(failures, collectJSONRPCRatesFailures(m, serviceConfig.SuccessRate)...)
	failures = append(failures, collectLatencyFailures(m, serviceConfig)...)

	// If there are failures, report them all at once at the end
	if len(failures) > 0 {
		fmt.Printf("\n%s❌ Assertion failures for %s:%s\n", RED, m.method, RESET)
		for i, failure := range failures {
			fmt.Printf("   %s%d. %s%s\n", RED, i+1, failure, RESET)
		}
		// Mark the test as failed but continue execution
		t.Fail()
		return true // Method failed
	} else {
		fmt.Printf("\n%s✅ Method %s passed all assertions%s\n", GREEN, m.method, RESET)
		return false // Method passed
	}
}

// collectHTTPSuccessRateFailures checks HTTP success rate and returns failure message if not met
func collectHTTPSuccessRateFailures(m *methodMetrics, requiredRate float64) []string {
	var failures []string

	if m.successRate < requiredRate {
		msg := fmt.Sprintf("HTTP success rate %.2f%% is below required %.2f%%", m.successRate*100, requiredRate*100)
		failures = append(failures, msg)
	}

	return failures
}

// collectJSONRPCRatesFailures checks all JSON-RPC success rates and returns failure messages
func collectJSONRPCRatesFailures(m *methodMetrics, requiredRate float64) []string {
	var failures []string

	// Skip if we don't have any JSON-RPC responses
	if m.jsonRPCResponses+m.jsonRPCUnmarshalErrors == 0 {
		return failures
	}

	// Check JSON-RPC unmarshal success rate
	if m.jsonRPCSuccessRate < requiredRate {
		msg := fmt.Sprintf("JSON-RPC unmarshal success rate %.2f%% is below required %.2f%%", m.jsonRPCSuccessRate*100, requiredRate*100)
		failures = append(failures, msg)
	}

	// Skip the rest if we don't have valid JSON-RPC responses
	if m.jsonRPCResponses == 0 {
		return failures
	}

	// Check Error field absence rate
	if m.jsonRPCErrorFieldRate < requiredRate {
		msg := fmt.Sprintf("JSON-RPC error field absence rate %.2f%% is below required %.2f%%", m.jsonRPCErrorFieldRate*100, requiredRate*100)
		failures = append(failures, msg)
	}

	// Check non-nil result rate
	if m.jsonRPCResultRate < requiredRate {
		msg := fmt.Sprintf("JSON-RPC non-nil result rate %.2f%% is below required %.2f%%", m.jsonRPCResultRate*100, requiredRate*100)
		failures = append(failures, msg)
	}

	// Check validation success rate
	if m.jsonRPCValidateRate < requiredRate {
		msg := fmt.Sprintf("JSON-RPC validation success rate %.2f%% is below required %.2f%%", m.jsonRPCValidateRate*100, requiredRate*100)
		failures = append(failures, msg)
	}

	return failures
}

// collectLatencyFailures checks latency metrics and returns failure messages
func collectLatencyFailures(m *methodMetrics, serviceConfig ServiceConfig) []string {
	var failures []string

	// P50 latency check
	if m.p50 > serviceConfig.MaxP50LatencyMS {
		msg := fmt.Sprintf("P50 latency %s exceeds maximum allowed %s",
			formatLatency(m.p50), formatLatency(serviceConfig.MaxP50LatencyMS))
		failures = append(failures, msg)
	}

	// P95 latency check
	if m.p95 > serviceConfig.MaxP95LatencyMS {
		msg := fmt.Sprintf("P95 latency %s exceeds maximum allowed %s",
			formatLatency(m.p95), formatLatency(serviceConfig.MaxP95LatencyMS))
		failures = append(failures, msg)
	}

	// P99 latency check
	if m.p99 > serviceConfig.MaxP99LatencyMS {
		msg := fmt.Sprintf("P99 latency %s exceeds maximum allowed %s",
			formatLatency(m.p99), formatLatency(serviceConfig.MaxP99LatencyMS))
		failures = append(failures, msg)
	}

	return failures
}

// ===== Display Helper Functions =====

// getRateColor returns color for success rates
func getRateColor(rate, requiredRate float64) string {
	if rate >= requiredRate {
		return GREEN // Green for meeting requirement
	} else if rate >= requiredRate*0.50 {
		return YELLOW // Yellow for close
	}
	return RED // Red for failing
}

// getRateEmoji returns emoji for success rates
func getRateEmoji(rate, requiredRate float64) string {
	if rate >= requiredRate {
		return "🟢" // Green for meeting requirement
	} else if rate >= requiredRate*0.50 {
		return "🟡" // Yellow for close
	}
	return "🔴" // Red for failing
}

// getLatencyColor returns color for latency values
func getLatencyColor(actual, maxAllowed time.Duration) string {
	if float64(actual) <= float64(maxAllowed)*0.5 {
		return GREEN // Green if well under limit (≤50%)
	} else if float64(actual) <= float64(maxAllowed) {
		return YELLOW // Yellow if close to limit (50-100%)
	}
	return RED // Red if over limit
}

// formatLatency formats latency values to whole milliseconds
func formatLatency(d time.Duration) string {
	return fmt.Sprintf("%dms", d/time.Millisecond)
}

// ===== Calculation Helper Functions =====

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

// ===== Service Summary Reporting =====

// printServiceSummaries prints a summary of all services after tests are complete
func printServiceSummaries(summaries map[protocol.ServiceID]*serviceSummary) {
	fmt.Printf("\n\n%s===== SERVICE SUMMARY =====%s\n", BOLD_CYAN, RESET)

	// Sort service IDs for consistent output
	serviceIDs := make([]protocol.ServiceID, 0, len(summaries))
	for svcID := range summaries {
		serviceIDs = append(serviceIDs, svcID)
	}
	sort.Slice(serviceIDs, func(i, j int) bool {
		return string(serviceIDs[i]) < string(serviceIDs[j])
	})

	type row struct {
		service      string
		status       string
		totalReq     int
		totalSuccess int
		failuresStr  string
		successRate  string
		p50Latency   string
		p90Latency   string
		avgLatency   string
	}

	rows := make([]row, 0, len(serviceIDs))

	for _, svcID := range serviceIDs {
		summary := summaries[svcID]
		serviceConfig := summary.serviceConfig

		successEmoji := getRateEmoji(summary.avgSuccessRate, serviceConfig.SuccessRate)

		maxFailureRate := 1.0 - serviceConfig.SuccessRate
		maxAllowedFailures := int(float64(summary.totalRequests) * maxFailureRate)
		failuresStr := fmt.Sprintf("%d", summary.totalFailure)
		if summary.totalFailure > 0 {
			failuresStr += fmt.Sprintf(" (max allowed: %d)", maxAllowedFailures)
		}

		rows = append(rows, row{
			service:      string(svcID),
			status:       successEmoji,
			totalReq:     summary.totalRequests,
			totalSuccess: summary.totalSuccess,
			failuresStr:  failuresStr,
			successRate:  fmt.Sprintf("%.2f%%", summary.avgSuccessRate*100),
			p50Latency:   formatLatency(summary.avgP50Latency),
			p90Latency:   formatLatency(summary.avgP90Latency),
			avgLatency:   formatLatency(summary.avgLatency),
		})
	}

	// Sort rows by descending success rate, then ascending P90 latency, then ascending avg latency
	sort.Slice(rows, func(i, j int) bool {
		// Parse success rates
		srI, _ := strconv.ParseFloat(strings.TrimSuffix(rows[i].successRate, "%"), 64)
		srJ, _ := strconv.ParseFloat(strings.TrimSuffix(rows[j].successRate, "%"), 64)
		if srI != srJ {
			return srI > srJ // Descending success rate
		}
		// Parse P50 latency (remove "ms")
		p50I, _ := strconv.Atoi(strings.TrimSuffix(rows[i].p50Latency, "ms"))
		p50J, _ := strconv.Atoi(strings.TrimSuffix(rows[j].p50Latency, "ms"))
		if p50I != p50J {
			return p50I < p50J // Ascending P50 latency
		}
		// Parse P90 latency (remove "ms")
		p90I, _ := strconv.Atoi(strings.TrimSuffix(rows[i].p90Latency, "ms"))
		p90J, _ := strconv.Atoi(strings.TrimSuffix(rows[j].p90Latency, "ms"))
		if p90I != p90J {
			return p90I < p90J // Ascending P90 latency
		}
		// Parse avg latency (remove "ms")
		avgI, _ := strconv.Atoi(strings.TrimSuffix(rows[i].avgLatency, "ms"))
		avgJ, _ := strconv.Atoi(strings.TrimSuffix(rows[j].avgLatency, "ms"))
		return avgI < avgJ // Ascending avg latency
	})

	// Print table header
	fmt.Printf("| %-16s | %-6s | %-8s | %-9s | %-20s | %-12s | %-11s | %-11s | %-11s |\n",
		"Service", "Status", "Requests", "Successes", "Failures", "Success Rate", "P50 Latency", "P90 Latency", "Avg Latency")
	fmt.Printf("|------------------|--------|----------|-----------|----------------------|--------------|-------------|-------------|-------------|\n")
	for _, r := range rows {
		fmt.Printf("| %-16s | %-6s | %-8d | %-9d | %-20s | %-12s | %-11s | %-11s | %-11s |\n",
			r.service, r.status, r.totalReq, r.totalSuccess, r.failuresStr, r.successRate, r.p50Latency, r.p90Latency, r.avgLatency)
	}

	fmt.Printf("\n%s===== END SERVICE SUMMARY =====%s\n", BOLD_CYAN, RESET)
}

// printTestFailureSummary prints a prominent failure summary when tests fail
func printTestFailureSummary(t *testing.T, failedMethods []string, totalErrors int) {
	if !t.Failed() {
		// Print success message when all tests pass
		fmt.Printf("\n%s🎉 ALL TESTS PASSED! 🎉%s\n", GREEN, RESET)
		fmt.Printf("%s✅ All services met their performance and reliability requirements%s\n\n", GREEN, RESET)
		return
	}

	// Print prominent failure summary
	fmt.Printf("\n\n%s"+strings.Repeat("=", 80)+"%s\n", RED, RESET)
	fmt.Printf("%s❌ TEST FAILURE SUMMARY ❌%s\n", RED, RESET)
	fmt.Printf("%s"+strings.Repeat("=", 80)+"%s\n", RED, RESET)

	if len(failedMethods) > 0 {
		fmt.Printf("\n%s🚨 FAILED METHODS (%d):%s\n", RED, len(failedMethods), RESET)
		for i, method := range failedMethods {
			fmt.Printf("   %s%d. %s%s\n", RED, i+1, method, RESET)
		}
	}

	if totalErrors > 0 {
		fmt.Printf("\n%s📊 Total Errors Encountered: %s%d%s\n", YELLOW, RED, totalErrors, RESET)
	}

	fmt.Printf("\n%s💡 TIP: Scroll up to see detailed failure reasons for each method%s\n", CYAN, RESET)
	fmt.Printf("%s🔍 Look for ❌ symbols and red text in the method results above%s\n", CYAN, RESET)

	fmt.Printf("\n%s"+strings.Repeat("=", 80)+"%s\n", RED, RESET)
	fmt.Printf("%s🔴 TESTS FAILED - FIX REQUIRED 🔴%s\n", RED, RESET)
	fmt.Printf("%s"+strings.Repeat("=", 80)+"%s\n\n", RED, RESET)
}
