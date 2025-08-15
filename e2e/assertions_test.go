//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// ===== JSON-RPC Response Validation =====

// validateJSONRPCResponse validates a JSON-RPC response and updates metrics
// This is decoupled from HTTP/WebSocket transport and can be used for both
func validateJSONRPCResponse(
	responseBody []byte,
	expectedID jsonrpc.ID,
	metrics *methodMetrics,
) error {
	var rpcResponse jsonrpc.Response

	// Try to unmarshal the response
	if err := json.Unmarshal(responseBody, &rpcResponse); err != nil {
		metrics.jsonrpcUnmarshalErrors++

		// Create response preview for parse errors
		preview := createResponsePreview(responseBody, 100)
		errorMsg := fmt.Sprintf("JSON parse error: %v (response preview: %s)", err, preview)
		metrics.jsonrpcParseErrors[errorMsg]++
		metrics.errors[errorMsg]++
		return fmt.Errorf("failed to unmarshal JSON-RPC response: %w", err)
	}

	metrics.jsonrpcResponses++

	// Validate the response structure
	validationErr := rpcResponse.Validate(expectedID)

	// Check if Error field is nil (good)
	if rpcResponse.Error != nil {
		metrics.jsonrpcErrorField++
		// Only track the error field message if there's no validation error
		// (to avoid duplicate tracking when validation fails due to error field)
		if validationErr == nil {
			metrics.errors[rpcResponse.Error.Message]++
		}
	}

	// Check if Result field is not nil (good)
	if rpcResponse.Result == nil {
		metrics.jsonrpcNilResult++
	}

	// Process validation error
	if validationErr != nil {
		metrics.jsonrpcValidateErrors++

		// Create response preview for validation errors
		preview := createResponsePreview(responseBody, 100)
		errorMsg := fmt.Sprintf("JSON-RPC validation error: %v (response preview: %s)", validationErr, preview)
		metrics.jsonrpcValidationErrors[errorMsg]++
		metrics.errors[errorMsg]++
		return validationErr
	}

	return nil
}

// createResponsePreview creates a sanitized preview of the response body for error logging
func createResponsePreview(body []byte, maxLen int) string {
	if len(body) == 0 {
		return "(empty)"
	}

	// Convert to string and normalize whitespace
	bodyStr := string(body)

	// Replace all whitespace characters with single spaces
	bodyStr = strings.ReplaceAll(bodyStr, "\n", " ")
	bodyStr = strings.ReplaceAll(bodyStr, "\r", " ")
	bodyStr = strings.ReplaceAll(bodyStr, "\t", " ")

	// Collapse multiple spaces into single spaces
	bodyStr = strings.Join(strings.Fields(bodyStr), " ")

	// Truncate if needed
	if len(bodyStr) <= maxLen {
		return bodyStr
	}
	if maxLen <= 3 {
		return bodyStr[:maxLen]
	}
	return bodyStr[:maxLen-3] + "..."
}

// ===== Test Assertions =====

// validateMethodResults performs all assertions on test metrics for a single method
func validateMethodResults(
	t *testing.T,
	serviceID protocol.ServiceID,
	metrics *methodMetrics,
	config ServiceConfig,
) bool {
	// Create a slice to collect all assertion failures
	var failures []string

	// Print metrics with formatting
	printMethodMetrics(serviceID, metrics, config)

	// Collect all assertion failures
	failures = append(failures, collectHTTPSuccessRateFailures(metrics, config.SuccessRate)...)
	failures = append(failures, collectJSONRPCRatesFailures(metrics, config.SuccessRate)...)
	failures = append(failures, collectLatencyFailures(metrics, config)...)

	// Report failures if any
	if len(failures) > 0 {
		fmt.Printf("\n%s❌ Assertion failures for %s:%s\n", RED, metrics.method, RESET)
		for i, failure := range failures {
			fmt.Printf("   %s%d. %s%s\n", RED, i+1, failure, RESET)
		}
		t.Fail()
		return false
	}

	fmt.Printf("\n%s✅ Method %s passed all assertions%s\n", GREEN, metrics.method, RESET)
	return true
}

// printMethodMetrics prints formatted metrics for a method
func printMethodMetrics(
	serviceID protocol.ServiceID,
	m *methodMetrics,
	config ServiceConfig,
) {
	// Add a blank line before each test result for better readability
	fmt.Println()

	// Print metrics header with method name in blue
	fmt.Printf("%s~~~~~~~~~~ %s - %s ~~~~~~~~%s\n\n", BOLD_CYAN, serviceID, m.method, RESET)

	// Print HTTP success rate with color
	successColor := getSuccessRateColor(m.successRate)
	fmt.Printf("%sHTTP Success Rate%s: %s%.2f%%%s (%d/%d requests)\n",
		BOLD, RESET, successColor, m.successRate*100, RESET, m.success, m.requestCount)

	// Print latencies with color
	p50Color := getLatencyColor(m.p50, config.MaxP50LatencyMS)
	p95Color := getLatencyColor(m.p95, config.MaxP95LatencyMS)
	p99Color := getLatencyColor(m.p99, config.MaxP99LatencyMS)
	fmt.Printf("%sLatency P50%s: %s%s%s\n", BOLD, RESET, p50Color, formatLatency(m.p50), RESET)
	fmt.Printf("%sLatency P95%s: %s%s%s\n", BOLD, RESET, p95Color, formatLatency(m.p95), RESET)
	fmt.Printf("%sLatency P99%s: %s%s%s\n", BOLD, RESET, p99Color, formatLatency(m.p99), RESET)

	// Print JSON-RPC metrics if applicable
	printJSONRPCMetrics(m, config)

	// Print status codes
	printStatusCodes(m)

	// Print errors if any
	printErrors(m, config)
}

// printJSONRPCMetrics prints JSON-RPC specific metrics
func printJSONRPCMetrics(m *methodMetrics, config ServiceConfig) {
	if m.jsonrpcResponses+m.jsonrpcUnmarshalErrors == 0 {
		return
	}

	fmt.Printf("%sJSON-RPC Metrics:%s\n", BOLD, RESET)

	if m.jsonrpcResponses > 0 {
		// Unmarshal success rate
		color := getRateColor(m.jsonrpcSuccessRate, config.SuccessRate)
		fmt.Printf("  Unmarshal Success: %s%.2f%%%s (%d/%d responses)\n",
			color, m.jsonrpcSuccessRate*100, RESET, m.jsonrpcResponses, m.jsonrpcResponses+m.jsonrpcUnmarshalErrors)

		// Validation success rate
		color = getRateColor(m.jsonrpcValidateRate, config.SuccessRate)
		fmt.Printf("  Validation Success: %s%.2f%%%s (%d/%d responses)\n",
			color, m.jsonrpcValidateRate*100, RESET, m.jsonrpcResponses-m.jsonrpcValidateErrors, m.jsonrpcResponses)

		// Non-nil result rate
		color = getRateColor(m.jsonrpcResultRate, config.SuccessRate)
		fmt.Printf("  Has Result: %s%.2f%%%s (%d/%d responses)\n",
			color, m.jsonrpcResultRate*100, RESET, m.jsonrpcResponses-m.jsonrpcNilResult, m.jsonrpcResponses)

		// Error field absent rate
		color = getRateColor(m.jsonrpcErrorFieldRate, config.SuccessRate)
		fmt.Printf("  Does Not Have Error: %s%.2f%%%s (%d/%d responses)\n",
			color, m.jsonrpcErrorFieldRate*100, RESET, m.jsonrpcResponses-m.jsonrpcErrorField, m.jsonrpcResponses)
	}
}

// printStatusCodes prints HTTP status code distribution
func printStatusCodes(m *methodMetrics) {
	if len(m.statusCodes) == 0 {
		return
	}

	statusText := "Status Codes:"
	for code, count := range m.statusCodes {
		codeColor := getStatusCodeColor(code)
		statusText += fmt.Sprintf("\n  %s%d%s: %d", codeColor, code, RESET, count)
	}
	fmt.Println(statusText)
}

// printErrors prints top errors with appropriate coloring
func printErrors(m *methodMetrics, config ServiceConfig) {
	if len(m.errors) == 0 {
		return
	}

	// Determine if the test passed
	testPassed := m.successRate >= config.SuccessRate &&
		m.p50 <= config.MaxP50LatencyMS &&
		m.p95 <= config.MaxP95LatencyMS &&
		m.p99 <= config.MaxP99LatencyMS

	// Choose error color based on test status
	errorColor := YELLOW // Yellow for warnings
	if !testPassed {
		errorColor = RED // Red for critical errors
	}

	fmt.Println("") // Add a new line before logging errors
	fmt.Printf("%sTop errors:%s\n", errorColor, RESET)

	// Sort errors by count (descending)
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

// ===== Failure Collection Functions =====

// collectHTTPSuccessRateFailures checks HTTP success rate and returns failure messages
func collectHTTPSuccessRateFailures(m *methodMetrics, requiredRate float64) []string {
	var failures []string

	if m.successRate < requiredRate {
		msg := fmt.Sprintf("HTTP success rate %.2f%% is below required %.2f%%",
			m.successRate*100, requiredRate*100)
		failures = append(failures, msg)
	}

	return failures
}

// collectJSONRPCRatesFailures checks all JSON-RPC success rates and returns failure messages
func collectJSONRPCRatesFailures(m *methodMetrics, requiredRate float64) []string {
	var failures []string

	// Skip if we don't have any JSON-RPC responses
	if m.jsonrpcResponses+m.jsonrpcUnmarshalErrors == 0 {
		return failures
	}

	// Check JSON-RPC unmarshal success rate
	if m.jsonrpcSuccessRate < requiredRate {
		msg := fmt.Sprintf("JSON-RPC unmarshal success rate %.2f%% is below required %.2f%%",
			m.jsonrpcSuccessRate*100, requiredRate*100)
		failures = append(failures, msg)
	}

	// Skip the rest if we don't have valid JSON-RPC responses
	if m.jsonrpcResponses == 0 {
		return failures
	}

	// Check Error field absence rate
	if m.jsonrpcErrorFieldRate < requiredRate {
		msg := fmt.Sprintf("JSON-RPC error field absence rate %.2f%% is below required %.2f%%",
			m.jsonrpcErrorFieldRate*100, requiredRate*100)
		failures = append(failures, msg)
	}

	// Check non-nil result rate
	if m.jsonrpcResultRate < requiredRate {
		msg := fmt.Sprintf("JSON-RPC non-nil result rate %.2f%% is below required %.2f%%",
			m.jsonrpcResultRate*100, requiredRate*100)
		failures = append(failures, msg)
	}

	// Check validation success rate
	if m.jsonrpcValidateRate < requiredRate {
		msg := fmt.Sprintf("JSON-RPC validation success rate %.2f%% is below required %.2f%%",
			m.jsonrpcValidateRate*100, requiredRate*100)
		failures = append(failures, msg)
	}

	return failures
}

// collectLatencyFailures checks latency metrics and returns failure messages
func collectLatencyFailures(m *methodMetrics, config ServiceConfig) []string {
	var failures []string

	// P50 latency check
	if m.p50 > config.MaxP50LatencyMS {
		msg := fmt.Sprintf("P50 latency %s exceeds maximum allowed %s",
			formatLatency(m.p50), formatLatency(config.MaxP50LatencyMS))
		failures = append(failures, msg)
	}

	// P95 latency check
	if m.p95 > config.MaxP95LatencyMS {
		msg := fmt.Sprintf("P95 latency %s exceeds maximum allowed %s",
			formatLatency(m.p95), formatLatency(config.MaxP95LatencyMS))
		failures = append(failures, msg)
	}

	// P99 latency check
	if m.p99 > config.MaxP99LatencyMS {
		msg := fmt.Sprintf("P99 latency %s exceeds maximum allowed %s",
			formatLatency(m.p99), formatLatency(config.MaxP99LatencyMS))
		failures = append(failures, msg)
	}

	return failures
}

// ===== Service Summary Printing =====

// serviceRow represents a row in the service summary table
type serviceRow struct {
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

// printServiceSummaries prints a summary table of all services after tests complete
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

	rows := make([]serviceRow, 0, len(serviceIDs))

	for _, svcID := range serviceIDs {
		summary := summaries[svcID]
		serviceConfig := summary.ServiceConfig

		successEmoji := getRateEmoji(summary.AvgSuccessRate, serviceConfig.SuccessRate)

		maxFailureRate := 1.0 - serviceConfig.SuccessRate
		maxAllowedFailures := int(float64(summary.TotalRequests) * maxFailureRate)
		failuresStr := fmt.Sprintf("%d", summary.TotalFailure)
		if summary.TotalFailure > 0 {
			failuresStr += fmt.Sprintf(" (max allowed: %d)", maxAllowedFailures)
		}

		rows = append(rows, serviceRow{
			service:      string(svcID),
			status:       successEmoji,
			totalReq:     summary.TotalRequests,
			totalSuccess: summary.TotalSuccess,
			failuresStr:  failuresStr,
			successRate:  fmt.Sprintf("%.2f%%", summary.AvgSuccessRate*100),
			p50Latency:   formatLatency(summary.AvgP50Latency),
			p90Latency:   formatLatency(summary.AvgP90Latency),
			avgLatency:   formatLatency(summary.AvgLatency),
		})
	}

	// Sort rows by success rate, then latency
	sortServiceRows(rows)

	// Print table
	printServiceTable(rows)

	fmt.Printf("\n%s===== END SERVICE SUMMARY =====%s\n", BOLD_CYAN, RESET)
}

// sortServiceRows sorts service summary rows by success rate and latency
func sortServiceRows(rows []serviceRow) {
	sort.Slice(rows, func(i, j int) bool {
		// Parse success rates
		srI, _ := strconv.ParseFloat(strings.TrimSuffix(rows[i].successRate, "%"), 64)
		srJ, _ := strconv.ParseFloat(strings.TrimSuffix(rows[j].successRate, "%"), 64)
		if srI != srJ {
			return srI > srJ // Descending success rate
		}
		// Parse P50 latency
		p50I, _ := strconv.Atoi(strings.TrimSuffix(rows[i].p50Latency, "ms"))
		p50J, _ := strconv.Atoi(strings.TrimSuffix(rows[j].p50Latency, "ms"))
		if p50I != p50J {
			return p50I < p50J // Ascending P50 latency
		}
		// Parse P90 latency
		p90I, _ := strconv.Atoi(strings.TrimSuffix(rows[i].p90Latency, "ms"))
		p90J, _ := strconv.Atoi(strings.TrimSuffix(rows[j].p90Latency, "ms"))
		if p90I != p90J {
			return p90I < p90J // Ascending P90 latency
		}
		// Parse avg latency
		avgI, _ := strconv.Atoi(strings.TrimSuffix(rows[i].avgLatency, "ms"))
		avgJ, _ := strconv.Atoi(strings.TrimSuffix(rows[j].avgLatency, "ms"))
		return avgI < avgJ // Ascending avg latency
	})
}

// printServiceTable prints the formatted service summary table
func printServiceTable(rows []serviceRow) {
	fmt.Printf("| %-16s | %-6s | %-8s | %-9s | %-20s | %-12s | %-11s | %-11s | %-11s |\n",
		"Service", "Status", "Requests", "Successes", "Failures", "Success Rate", "P50 Latency", "P90 Latency", "Avg Latency")
	fmt.Printf("|------------------|--------|----------|-----------|----------------------|--------------|-------------|-------------|-------------|\n")
	for _, r := range rows {
		fmt.Printf("| %-16s | %-6s | %-8d | %-9d | %-20s | %-12s | %-11s | %-11s | %-11s |\n",
			r.service, r.status, r.totalReq, r.totalSuccess, r.failuresStr, r.successRate, r.p50Latency, r.p90Latency, r.avgLatency)
	}
}
