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
	metrics *MethodMetrics,
) error {
	var rpcResponse jsonrpc.Response

	// Try to unmarshal the response
	if err := json.Unmarshal(responseBody, &rpcResponse); err != nil {
		metrics.JSONRPCUnmarshalErrors++

		// Create response preview for parse errors
		preview := createResponsePreview(responseBody, 100)
		errorMsg := fmt.Sprintf("JSON parse error: %v (response preview: %s)", err, preview)
		metrics.JSONRPCParseErrors[errorMsg]++
		metrics.Errors[errorMsg]++
		return fmt.Errorf("failed to unmarshal JSON-RPC response: %w", err)
	}

	metrics.JSONRPCResponses++

	// Validate the response structure
	validationErr := rpcResponse.Validate(expectedID)

	// Check if Error field is nil (good)
	if rpcResponse.Error != nil {
		metrics.JSONRPCErrorField++
		// Only track the error field message if there's no validation error
		// (to avoid duplicate tracking when validation fails due to error field)
		if validationErr == nil {
			metrics.Errors[rpcResponse.Error.Message]++
		}
	}

	// Check if Result field is not nil (good)
	if rpcResponse.Result == nil {
		metrics.JSONRPCNilResult++
	}

	// Process validation error
	if validationErr != nil {
		metrics.JSONRPCValidateErrors++

		// Create response preview for validation errors
		preview := createResponsePreview(responseBody, 100)
		errorMsg := fmt.Sprintf("JSON-RPC validation error: %v (response preview: %s)", validationErr, preview)
		metrics.JSONRPCValidationErrors[errorMsg]++
		metrics.Errors[errorMsg]++
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
	metrics *MethodMetrics,
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
		fmt.Printf("\n%s❌ Assertion failures for %s:%s\n", RED, metrics.Method, RESET)
		for i, failure := range failures {
			fmt.Printf("   %s%d. %s%s\n", RED, i+1, failure, RESET)
		}
		t.Fail()
		return false
	}

	fmt.Printf("\n%s✅ Method %s passed all assertions%s\n", GREEN, metrics.Method, RESET)
	return true
}

// printMethodMetrics prints formatted metrics for a method
func printMethodMetrics(serviceID protocol.ServiceID, m *MethodMetrics, config ServiceConfig) {
	// Add a blank line before each test result for better readability
	fmt.Println()

	// Print metrics header with method name in blue
	fmt.Printf("%s~~~~~~~~~~ %s - %s ~~~~~~~~%s\n\n", BOLD_CYAN, serviceID, m.Method, RESET)

	// Print HTTP success rate with color
	successColor := getSuccessRateColor(m.SuccessRate)
	fmt.Printf("%sHTTP Success Rate%s: %s%.2f%%%s (%d/%d requests)\n",
		BOLD, RESET, successColor, m.SuccessRate*100, RESET, m.Success, m.RequestCount)

	// Print latencies with color
	p50Color := getLatencyColor(m.P50, config.MaxP50LatencyMS)
	p95Color := getLatencyColor(m.P95, config.MaxP95LatencyMS)
	p99Color := getLatencyColor(m.P99, config.MaxP99LatencyMS)
	fmt.Printf("%sLatency P50%s: %s%s%s\n", BOLD, RESET, p50Color, formatLatency(m.P50), RESET)
	fmt.Printf("%sLatency P95%s: %s%s%s\n", BOLD, RESET, p95Color, formatLatency(m.P95), RESET)
	fmt.Printf("%sLatency P99%s: %s%s%s\n", BOLD, RESET, p99Color, formatLatency(m.P99), RESET)

	// Print JSON-RPC metrics if applicable
	printJSONRPCMetrics(m, config)

	// Print status codes
	printStatusCodes(m)

	// Print errors if any
	printErrors(m, config)
}

// printJSONRPCMetrics prints JSON-RPC specific metrics
func printJSONRPCMetrics(m *MethodMetrics, config ServiceConfig) {
	if m.JSONRPCResponses+m.JSONRPCUnmarshalErrors == 0 {
		return
	}

	fmt.Printf("%sJSON-RPC Metrics:%s\n", BOLD, RESET)

	if m.JSONRPCResponses > 0 {
		// Unmarshal success rate
		color := getRateColor(m.JSONRPCSuccessRate, config.SuccessRate)
		fmt.Printf("  Unmarshal Success: %s%.2f%%%s (%d/%d responses)\n",
			color, m.JSONRPCSuccessRate*100, RESET, m.JSONRPCResponses, m.JSONRPCResponses+m.JSONRPCUnmarshalErrors)

		// Validation success rate
		color = getRateColor(m.JSONRPCValidateRate, config.SuccessRate)
		fmt.Printf("  Validation Success: %s%.2f%%%s (%d/%d responses)\n",
			color, m.JSONRPCValidateRate*100, RESET, m.JSONRPCResponses-m.JSONRPCValidateErrors, m.JSONRPCResponses)

		// Non-nil result rate
		color = getRateColor(m.JSONRPCResultRate, config.SuccessRate)
		fmt.Printf("  Has Result: %s%.2f%%%s (%d/%d responses)\n",
			color, m.JSONRPCResultRate*100, RESET, m.JSONRPCResponses-m.JSONRPCNilResult, m.JSONRPCResponses)

		// Error field absent rate
		color = getRateColor(m.JSONRPCErrorFieldRate, config.SuccessRate)
		fmt.Printf("  Does Not Have Error: %s%.2f%%%s (%d/%d responses)\n",
			color, m.JSONRPCErrorFieldRate*100, RESET, m.JSONRPCResponses-m.JSONRPCErrorField, m.JSONRPCResponses)
	}
}

// printStatusCodes prints HTTP status code distribution
func printStatusCodes(m *MethodMetrics) {
	if len(m.StatusCodes) == 0 {
		return
	}

	statusText := "Status Codes:"
	for code, count := range m.StatusCodes {
		codeColor := getStatusCodeColor(code)
		statusText += fmt.Sprintf("\n  %s%d%s: %d", codeColor, code, RESET, count)
	}
	fmt.Println(statusText)
}

// printErrors prints top errors with appropriate coloring
func printErrors(m *MethodMetrics, config ServiceConfig) {
	if len(m.Errors) == 0 {
		return
	}

	// Determine if the test passed
	testPassed := m.SuccessRate >= config.SuccessRate &&
		m.P50 <= config.MaxP50LatencyMS &&
		m.P95 <= config.MaxP95LatencyMS &&
		m.P99 <= config.MaxP99LatencyMS

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
	for errMsg, count := range m.Errors {
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
func collectHTTPSuccessRateFailures(m *MethodMetrics, requiredRate float64) []string {
	var failures []string

	if m.SuccessRate < requiredRate {
		msg := fmt.Sprintf("HTTP success rate %.2f%% is below required %.2f%%",
			m.SuccessRate*100, requiredRate*100)
		failures = append(failures, msg)
	}

	return failures
}

// collectJSONRPCRatesFailures checks all JSON-RPC success rates and returns failure messages
func collectJSONRPCRatesFailures(m *MethodMetrics, requiredRate float64) []string {
	var failures []string

	// Skip if we don't have any JSON-RPC responses
	if m.JSONRPCResponses+m.JSONRPCUnmarshalErrors == 0 {
		return failures
	}

	// Check JSON-RPC unmarshal success rate
	if m.JSONRPCSuccessRate < requiredRate {
		msg := fmt.Sprintf("JSON-RPC unmarshal success rate %.2f%% is below required %.2f%%",
			m.JSONRPCSuccessRate*100, requiredRate*100)
		failures = append(failures, msg)
	}

	// Skip the rest if we don't have valid JSON-RPC responses
	if m.JSONRPCResponses == 0 {
		return failures
	}

	// Check Error field absence rate
	if m.JSONRPCErrorFieldRate < requiredRate {
		msg := fmt.Sprintf("JSON-RPC error field absence rate %.2f%% is below required %.2f%%",
			m.JSONRPCErrorFieldRate*100, requiredRate*100)
		failures = append(failures, msg)
	}

	// Check non-nil result rate
	if m.JSONRPCResultRate < requiredRate {
		msg := fmt.Sprintf("JSON-RPC non-nil result rate %.2f%% is below required %.2f%%",
			m.JSONRPCResultRate*100, requiredRate*100)
		failures = append(failures, msg)
	}

	// Check validation success rate
	if m.JSONRPCValidateRate < requiredRate {
		msg := fmt.Sprintf("JSON-RPC validation success rate %.2f%% is below required %.2f%%",
			m.JSONRPCValidateRate*100, requiredRate*100)
		failures = append(failures, msg)
	}

	return failures
}

// collectLatencyFailures checks latency metrics and returns failure messages
func collectLatencyFailures(m *MethodMetrics, config ServiceConfig) []string {
	var failures []string

	// P50 latency check
	if m.P50 > config.MaxP50LatencyMS {
		msg := fmt.Sprintf("P50 latency %s exceeds maximum allowed %s",
			formatLatency(m.P50), formatLatency(config.MaxP50LatencyMS))
		failures = append(failures, msg)
	}

	// P95 latency check
	if m.P95 > config.MaxP95LatencyMS {
		msg := fmt.Sprintf("P95 latency %s exceeds maximum allowed %s",
			formatLatency(m.P95), formatLatency(config.MaxP95LatencyMS))
		failures = append(failures, msg)
	}

	// P99 latency check
	if m.P99 > config.MaxP99LatencyMS {
		msg := fmt.Sprintf("P99 latency %s exceeds maximum allowed %s",
			formatLatency(m.P99), formatLatency(config.MaxP99LatencyMS))
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
