//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cheggaaa/pb/v3"
	vegeta "github.com/tsenart/vegeta/lib"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/buildwithgrove/path/request"
)

/* -------------------- Vegeta Helper Functions -------------------- */

// createRPCTarget creates a vegeta.Targeter for the specified RPC method
func createRPCTarget(
	gatewayURL string,
	serviceID protocol.ServiceID,
	jsonrpcReq jsonrpc.Request,
) vegeta.Targeter {
	return func(tgt *vegeta.Target) error {
		body, err := json.Marshal(jsonrpcReq)
		if err != nil {
			return err
		}

		tgt.Method = http.MethodPost
		tgt.URL = gatewayURL
		tgt.Body = body
		tgt.Header = http.Header{
			"Content-Type":                    []string{"application/json"},
			request.HTTPHeaderTargetServiceID: []string{string(serviceID)},
		}

		return nil
	}
}

// runAttack executes a load test for the given method
func runAttack(
	gatewayURL string,
	serviceID protocol.ServiceID,
	method jsonrpc.Method,
	methodDef methodDefinition,
	progressBar *pb.ProgressBar,
	jsonrpcReq jsonrpc.Request,
) *MethodMetrics {
	// Initialize metrics for the method
	metrics := &MethodMetrics{
		method:      method,
		statusCodes: make(map[int]int),
		errors:      make(map[string]int),
		results:     make([]*vegeta.Result, 0, methodDef.totalRequests),
	}

	// Create target for the method
	target := createRPCTarget(gatewayURL, serviceID, jsonrpcReq)

	// Calculate max duration as a safety to prevent infinite runs
	// Allow 2x the theoretical time needed plus a 5 second buffer
	maxDuration := time.Duration(2*methodDef.totalRequests/methodDef.rps)*time.Second + 5*time.Second

	// Create an attacker
	attacker := vegeta.NewAttacker(
		vegeta.Timeout(5*time.Second),
		vegeta.KeepAlive(true),
		vegeta.Workers(3),
		vegeta.MaxWorkers(5),
	)

	// Track exactly how many requests we've successfully processed
	processedCount := 0

	// Log test start info if progress bars are disabled
	if progressBar == nil {
		fmt.Printf("Starting test for method %s (%d requests at %d RPS)...\n",
			method, methodDef.totalRequests, methodDef.rps)
	}

	// Use channels to control exactly how many requests are processed
	resultsChan := make(chan *vegeta.Result, methodDef.totalRequests)

	// Start collection goroutine
	var resultsWg sync.WaitGroup
	resultsWg.Add(1)
	go func() {
		defer resultsWg.Done()
		for res := range resultsChan {
			// Skip "no targets to attack" errors
			if res.Error == "no targets to attack" {
				continue
			}

			// Only process up to the exact request count
			if processedCount < methodDef.totalRequests {
				processResult(metrics, res)
				processedCount++

				// Update progress bar if we have one
				if progressBar != nil && progressBar.Current() < int64(methodDef.totalRequests) {
					progressBar.Increment()
				}

				// If progress bar is disabled, print periodic status updates
				if progressBar == nil && processedCount%50 == 0 {
					percent := float64(processedCount) / float64(methodDef.totalRequests) * 100
					fmt.Printf("  %s: %d/%d requests completed (%.1f%%)\n",
						method, processedCount, methodDef.totalRequests, percent)
				}
			}
		}

		// Ensure the progress bar shows exactly 100% at the end if we have one
		if progressBar != nil && progressBar.Current() < int64(methodDef.totalRequests) {
			remaining := int64(methodDef.totalRequests) - progressBar.Current()
			progressBar.Add64(remaining)
		}

		// Final status update if progress bar is disabled
		if progressBar == nil {
			fmt.Printf("  %s: test completed (%d/%d requests)\n",
				method, processedCount, methodDef.totalRequests)
		}
	}()

	// Create exactly the number of request slots we want to process
	requestSlots := methodDef.totalRequests

	// Use a targeting function that limits the number of requests
	targeter := func(tgt *vegeta.Target) error {
		// Atomically decrement the counter, and if it's <= 0, return no targets
		if requestSlots <= 0 {
			return vegeta.ErrNoTargets
		}
		requestSlots--

		return target(tgt)
	}

	// Run the attack until we hit the total number of requests
	for res := range attacker.Attack(
		targeter,
		vegeta.Rate{
			Freq: methodDef.rps,
			Per:  time.Second,
		},
		maxDuration,
		string(method),
	) {
		resultsChan <- res
	}

	close(resultsChan)
	resultsWg.Wait()

	// Calculate success rate and percentiles
	calculateSuccessRate(metrics)
	calculatePercentiles(metrics)

	return metrics
}

// processResult updates metrics based on a single result
func processResult(m *MethodMetrics, result *vegeta.Result) {
	// Skip "no targets to attack" errors as these aren't actual requests
	if result.Error == "no targets to attack" {
		return
	}

	// Store the raw result
	m.results = append(m.results, result)

	// Process HTTP result
	if result.Code >= 200 && result.Code < 300 && result.Error == "" {
		m.success++
	} else {
		m.failed++
	}

	// Update status code counts
	m.statusCodes[int(result.Code)]++

	// Process JSON-RPC validation if we have a successful HTTP response
	var rpcResponse jsonrpc.Response
	if err := json.Unmarshal(result.Body, &rpcResponse); err != nil {
		// Failed to unmarshal as JSON-RPC
		m.jsonRPCUnmarshalErrors++
	} else {
		// Successfully unmarshaled as JSON-RPC
		m.jsonRPCResponses++

		// Check if Error field is nil (good)
		if rpcResponse.Error != nil {
			m.jsonRPCErrorField++
			m.errors[rpcResponse.Error.Message]++
		}

		// Check if Result field is not nil (good)
		if rpcResponse.Result == nil {
			m.jsonRPCNilResult++
		}

		// Validate the response
		expectedID := jsonrpc.IDFromInt(1) // Expected ID from our request
		if err := rpcResponse.Validate(expectedID); err != nil {
			m.jsonRPCValidateErrors++
		}
	}
}

/* -------------------- Assertions and Calculation Helpers -------------------- */

// MethodMetrics stores metrics for each method
type MethodMetrics struct {
	method       jsonrpc.Method   // RPC method name
	success      int              // Number of successful requests
	failed       int              // Number of failed requests
	statusCodes  map[int]int      // Count of each status code
	errors       map[string]int   // Count of each error type
	results      []*vegeta.Result // All raw results for this method
	requestCount int              // Total number of requests
	successRate  float64          // Success rate as a ratio (0-1)
	p50          time.Duration    // 50th percentile latency
	p95          time.Duration    // 95th percentile latency
	p99          time.Duration    // 99th percentile latency

	// JSON-RPC specific validation metrics
	jsonRPCResponses       int // Count of responses we could unmarshal as JSON-RPC
	jsonRPCUnmarshalErrors int // Count of responses we couldn't unmarshal
	jsonRPCErrorField      int // Count of responses with non-nil Error field
	jsonRPCNilResult       int // Count of responses with nil Result field
	jsonRPCValidateErrors  int // Count of responses that fail validation

	// Success rates for specific checks
	jsonRPCSuccessRate    float64 // Success rate for JSON-RPC unmarshaling
	jsonRPCErrorFieldRate float64 // Error field absent rate (success = no error)
	jsonRPCResultRate     float64 // Non-nil result rate
	jsonRPCValidateRate   float64 // Validation success rate
}

// calculateSuccessRate computes all success rates
func calculateSuccessRate(m *MethodMetrics) {
	// Overall HTTP success rate
	m.requestCount = m.success + m.failed
	if m.requestCount > 0 {
		m.successRate = float64(m.success) / float64(m.requestCount)
	}

	// JSON-RPC unmarshal success rate
	totalJSONAttempts := m.jsonRPCResponses + m.jsonRPCUnmarshalErrors
	if totalJSONAttempts > 0 {
		m.jsonRPCSuccessRate = float64(m.jsonRPCResponses) / float64(totalJSONAttempts)
	}

	// Only calculate these if we have valid JSON-RPC responses
	if m.jsonRPCResponses > 0 {
		// Error field absence rate (success = no error field)
		m.jsonRPCErrorFieldRate = float64(m.jsonRPCResponses-m.jsonRPCErrorField) / float64(m.jsonRPCResponses)

		// Non-nil result rate
		m.jsonRPCResultRate = float64(m.jsonRPCResponses-m.jsonRPCNilResult) / float64(m.jsonRPCResponses)

		// Validation success rate
		m.jsonRPCValidateRate = float64(m.jsonRPCResponses-m.jsonRPCValidateErrors) / float64(m.jsonRPCResponses)
	}
}

// calculatePercentiles computes P50, P95, and P99 latency percentiles
func calculatePercentiles(m *MethodMetrics) {
	if len(m.results) == 0 {
		return
	}

	// Extract latencies
	latencies := make([]time.Duration, 0, len(m.results))
	for _, res := range m.results {
		latencies = append(latencies, res.Latency)
	}

	// Sort latencies
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	// Calculate percentiles
	m.p50 = percentile(latencies, 50)
	m.p95 = percentile(latencies, 95)
	m.p99 = percentile(latencies, 99)
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

// validateResults performs assertions on test metrics
func validateResults(t *testing.T, m *MethodMetrics, methodDef methodDefinition) {
	// Create a slice to collect all assertion failures
	var failures []string

	// Add a blank line before each test result for better readability
	fmt.Println()

	// Print metrics header with method name in blue
	fmt.Printf("\x1b[1m\x1b[34m========= Test results for %s =========\x1b[0m\n", m.method)

	// Print success rate with color (green ≥99%, yellow ≥95%, red <95%)
	successColor := "\x1b[31m" // Red by default
	if m.successRate >= 0.99 {
		successColor = "\x1b[32m" // Green for ≥99%
	} else if m.successRate >= 0.95 {
		successColor = "\x1b[33m" // Yellow for ≥95%
	}
	fmt.Printf("HTTP Success Rate: %s%.2f%%\x1b[0m (%d/%d requests)\n",
		successColor, m.successRate*100, m.success, m.requestCount)

	// Print latencies (yellow if close to limit, green if well below)
	p50Color := getLatencyColor(m.p50, methodDef.maxP50Latency)
	p95Color := getLatencyColor(m.p95, methodDef.maxP95Latency)
	p99Color := getLatencyColor(m.p99, methodDef.maxP99Latency)
	fmt.Printf("Latency P50: %s%s\x1b[0m, P95: %s%s\x1b[0m, P99: %s%s\x1b[0m\n",
		p50Color, formatLatency(m.p50), p95Color, formatLatency(m.p95), p99Color, formatLatency(m.p99))

	// Print JSON-RPC metrics with coloring
	if m.jsonRPCResponses+m.jsonRPCUnmarshalErrors > 0 {
		fmt.Printf("\x1b[1mJSON-RPC Metrics:\x1b[0m\n")

		if m.jsonRPCResponses > 0 {
			// Unmarshal success rate
			color := getRateColor(m.jsonRPCSuccessRate, methodDef.successRate)
			fmt.Printf("  Unmarshal Success: %s%.2f%%\x1b[0m (%d/%d responses)\n",
				color, m.jsonRPCSuccessRate*100, m.jsonRPCResponses, m.jsonRPCResponses+m.jsonRPCUnmarshalErrors)
			// Validation success rate
			color = getRateColor(m.jsonRPCValidateRate, methodDef.successRate)
			fmt.Printf("  Validation Success: %s%.2f%%\x1b[0m (%d/%d responses)\n",
				color, m.jsonRPCValidateRate*100, m.jsonRPCResponses-m.jsonRPCValidateErrors, m.jsonRPCResponses)
			// Non-nil result rate
			color = getRateColor(m.jsonRPCResultRate, methodDef.successRate)
			fmt.Printf("  Has Result: %s%.2f%%\x1b[0m (%d/%d responses)\n",
				color, m.jsonRPCResultRate*100, m.jsonRPCResponses-m.jsonRPCNilResult, m.jsonRPCResponses)
			// Error field absent rate
			color = getRateColor(m.jsonRPCErrorFieldRate, methodDef.successRate)
			fmt.Printf("  Does Not Have Error: %s%.2f%%\x1b[0m (%d/%d responses)\n",
				color, m.jsonRPCErrorFieldRate*100, m.jsonRPCResponses-m.jsonRPCErrorField, m.jsonRPCResponses)
		}
	}

	// Log status codes
	if len(m.statusCodes) > 0 {
		statusText := "Status Codes: "
		for code, count := range m.statusCodes {
			codeColor := "\x1b[32m" // Green for 2xx
			if code >= 400 {
				codeColor = "\x1b[31m" // Red for 4xx/5xx
			} else if code >= 300 {
				codeColor = "\x1b[33m" // Yellow for 3xx
			}
			statusText += fmt.Sprintf("%s%d\x1b[0m:%d ", codeColor, code, count)
		}
		fmt.Println(statusText)
	}

	// Determine if the test passed based on our metrics
	testPassed := m.successRate >= methodDef.successRate &&
		m.p50 <= methodDef.maxP50Latency &&
		m.p95 <= methodDef.maxP95Latency &&
		m.p99 <= methodDef.maxP99Latency

	// Choose error color based on test passing status
	errorColor := "\x1b[33m" // Yellow for warnings (test passed despite errors)
	if !testPassed {
		errorColor = "\x1b[31m" // Red for critical errors (test failed)
	}

	// Log top errors with appropriate color
	if len(m.errors) > 0 {
		fmt.Printf("%sTop errors:\x1b[0m\n", errorColor)
		count := 0
		for err, errCount := range m.errors {
			if count < 5 {
				fmt.Printf("  %s%s\x1b[0m: %d\n", errorColor, err, errCount)
				count++
			}
		}
		if len(m.errors) > 5 {
			fmt.Printf("  ... and %s%d\x1b[0m more error types\n", errorColor, len(m.errors)-5)
		}
	}

	// Collect assertion failures
	failures = append(failures, collectHTTPSuccessRateFailures(m, methodDef.successRate)...)
	failures = append(failures, collectJSONRPCRatesFailures(m, methodDef.successRate)...)
	failures = append(failures, collectLatencyFailures(m, methodDef)...)

	// If there are failures, report them all at once at the end
	if len(failures) > 0 {
		fmt.Printf("\n\x1b[31m❌ Method %s has %d assertion failures:\x1b[0m\n", m.method, len(failures))
		for i, failure := range failures {
			fmt.Printf("   \x1b[31m%d. %s\x1b[0m\n", i+1, failure)
		}
		// Mark the test as failed but continue execution
		t.Fail()
	} else {
		fmt.Printf("\n\x1b[32m✅ Method %s passed all assertions\x1b[0m\n", m.method)
	}
}

// collectHTTPSuccessRateFailures checks HTTP success rate and returns failure message if not met
func collectHTTPSuccessRateFailures(m *MethodMetrics, requiredRate float64) []string {
	var failures []string

	if m.successRate < requiredRate {
		msg := fmt.Sprintf("HTTP success rate %.2f%% is below required %.2f%% (%d/%d requests)",
			m.successRate*100, requiredRate*100, m.success, m.requestCount)
		failures = append(failures, msg)
	}

	return failures
}

// collectJSONRPCRatesFailures checks all JSON-RPC success rates and returns failure messages
func collectJSONRPCRatesFailures(m *MethodMetrics, requiredRate float64) []string {
	var failures []string

	// Skip if we don't have any JSON-RPC responses
	if m.jsonRPCResponses+m.jsonRPCUnmarshalErrors == 0 {
		return failures
	}

	// Check JSON-RPC unmarshal success rate
	if m.jsonRPCSuccessRate < requiredRate {
		msg := fmt.Sprintf("JSON-RPC unmarshal success rate %.2f%% is below required %.2f%% (%d/%d responses)",
			m.jsonRPCSuccessRate*100, requiredRate*100, m.jsonRPCResponses, m.jsonRPCResponses+m.jsonRPCUnmarshalErrors)
		failures = append(failures, msg)
	}

	// Skip the rest if we don't have valid JSON-RPC responses
	if m.jsonRPCResponses == 0 {
		return failures
	}

	// Check Error field absence rate
	if m.jsonRPCErrorFieldRate < requiredRate {
		msg := fmt.Sprintf("JSON-RPC error field absence rate %.2f%% is below required %.2f%% (%d/%d responses)",
			m.jsonRPCErrorFieldRate*100, requiredRate*100, m.jsonRPCResponses-m.jsonRPCErrorField, m.jsonRPCResponses)
		failures = append(failures, msg)
	}

	// Check non-nil result rate
	if m.jsonRPCResultRate < requiredRate {
		msg := fmt.Sprintf("JSON-RPC non-nil result rate %.2f%% is below required %.2f%% (%d/%d responses)",
			m.jsonRPCResultRate*100, requiredRate*100, m.jsonRPCResponses-m.jsonRPCNilResult, m.jsonRPCResponses)
		failures = append(failures, msg)
	}

	// Check validation success rate
	if m.jsonRPCValidateRate < requiredRate {
		msg := fmt.Sprintf("JSON-RPC validation success rate %.2f%% is below required %.2f%% (%d/%d responses)",
			m.jsonRPCValidateRate*100, requiredRate*100, m.jsonRPCResponses-m.jsonRPCValidateErrors, m.jsonRPCResponses)
		failures = append(failures, msg)
	}

	return failures
}

// collectLatencyFailures checks latency metrics and returns failure messages
func collectLatencyFailures(m *MethodMetrics, methodDef methodDefinition) []string {
	var failures []string

	// P50 latency check
	if m.p50 > methodDef.maxP50Latency {
		msg := fmt.Sprintf("P50 latency %s exceeds maximum allowed %s",
			formatLatency(m.p50), formatLatency(methodDef.maxP50Latency))
		failures = append(failures, msg)
	}

	// P95 latency check
	if m.p95 > methodDef.maxP95Latency {
		msg := fmt.Sprintf("P95 latency %s exceeds maximum allowed %s",
			formatLatency(m.p95), formatLatency(methodDef.maxP95Latency))
		failures = append(failures, msg)
	}

	// P99 latency check
	if m.p99 > methodDef.maxP99Latency {
		msg := fmt.Sprintf("P99 latency %s exceeds maximum allowed %s",
			formatLatency(m.p99), formatLatency(methodDef.maxP99Latency))
		failures = append(failures, msg)
	}

	return failures
}

// Helper function to get color for success rates
func getRateColor(rate, requiredRate float64) string {
	if rate >= requiredRate {
		return "\x1b[32m" // Green for meeting requirement
	} else if rate >= requiredRate*0.95 {
		return "\x1b[33m" // Yellow for close
	}
	return "\x1b[31m" // Red for failing
}

// Helper function to get color for latency values
func getLatencyColor(actual, maxAllowed time.Duration) string {
	if float64(actual) <= float64(maxAllowed)*0.7 {
		return "\x1b[32m" // Green if well under limit (≤70%)
	} else if float64(actual) <= float64(maxAllowed) {
		return "\x1b[33m" // Yellow if close to limit (70-100%)
	}
	return "\x1b[31m" // Red if over limit
}

/* -------------------- Progress Bars -------------------- */

// progressBars holds and manages progress bars for all methods in a test
type progressBars struct {
	bars    map[jsonrpc.Method]*pb.ProgressBar
	pool    *pb.Pool
	enabled bool
}

// newProgressBars creates a set of progress bars for all methods in a test
func newProgressBars(methods []jsonrpc.Method, methodDefs map[jsonrpc.Method]methodDefinition) (*progressBars, error) {
	// Check if we're running in CI or non-interactive environment
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		fmt.Println("Running in CI environment - progress bars disabled")
		return &progressBars{
			bars:    make(map[jsonrpc.Method]*pb.ProgressBar),
			enabled: false,
		}, nil
	}

	// Sort methods for consistent display order
	sortedMethods := make([]jsonrpc.Method, len(methods))
	copy(sortedMethods, methods)
	sort.Slice(sortedMethods, func(i, j int) bool {
		return string(sortedMethods[i]) < string(sortedMethods[j])
	})

	// Calculate the longest method name for padding
	longestLen := 0
	for _, method := range sortedMethods {
		if len(string(method)) > longestLen {
			longestLen = len(string(method))
		}
	}

	// Create a progress bar for each method
	bars := make(map[jsonrpc.Method]*pb.ProgressBar)
	barList := make([]*pb.ProgressBar, 0, len(methods))

	for _, method := range sortedMethods {
		def := methodDefs[method]

		// Store the method name with padding for display
		padding := longestLen - len(string(method))
		methodWithPadding := string(method) + strings.Repeat(" ", padding)

		// Create a custom format for counters with padding for consistent spacing
		// Format: current/total with padding to make 3 digits minimum
		// This formats as "  1/300" or "010/300" for consistent width
		customCounterFormat := `{{ printf "%3d/%3d" .Current .Total }}`

		// Create a colored template with padded counters
		tmpl := fmt.Sprintf(`{{ blue "%s" }} %s {{ bar . "[" "=" ">" " " "]" | blue }} {{ green (percent .) }}`,
			methodWithPadding, customCounterFormat)

		// Create the bar with the template and start it
		bar := pb.ProgressBarTemplate(tmpl).New(def.totalRequests)

		// Ensure we're not using byte formatting
		bar.Set(pb.Bytes, false)

		// Set max width for the bar
		bar.SetMaxWidth(100)

		bars[method] = bar
		barList = append(barList, bar)
	}

	// Try to create a pool with all the bars
	pool, err := pb.StartPool(barList...)
	if err != nil {
		// If we fail to create progress bars, fall back to simple output
		fmt.Printf("Warning: Could not create progress bars: %v\n", err)
		return &progressBars{
			bars:    make(map[jsonrpc.Method]*pb.ProgressBar),
			enabled: false,
		}, nil
	}

	return &progressBars{
		bars:    bars,
		pool:    pool,
		enabled: true,
	}, nil
}

// finish completes all progress bars
func (p *progressBars) finish() error {
	if !p.enabled || p.pool == nil {
		return nil
	}
	return p.pool.Stop()
}

// get returns the progress bar for a specific method
func (p *progressBars) get(method jsonrpc.Method) *pb.ProgressBar {
	if !p.enabled {
		return nil
	}
	return p.bars[method]
}

// formatLatency formats latency values to whole milliseconds
func formatLatency(d time.Duration) string {
	return fmt.Sprintf("%dms", d/time.Millisecond)
}
