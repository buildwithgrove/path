package e2e

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/stretchr/testify/require"
	vegeta "github.com/tsenart/vegeta/lib"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/buildwithgrove/path/request"
)

/* -------------------- Vegeta Helper Functions -------------------- */

// createRPCTarget creates a vegeta.Targeter for the specified RPC method
func createRPCTarget(gatewayURL string, serviceID protocol.ServiceID, method jsonrpc.Method, params ...any) vegeta.Targeter {
	return func(tgt *vegeta.Target) error {
		req, err := buildJSONRPCReq(1, method, params...)
		if err != nil {
			return err
		}

		body, err := json.Marshal(req)
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

// buildJSONRPCReq builds a JSON-RPC request for the given method and parameters
func buildJSONRPCReq(id int, method jsonrpc.Method, params ...any) (jsonrpc.Request, error) {
	request := jsonrpc.Request{
		JSONRPC: jsonrpc.Version2,
		ID:      jsonrpc.IDFromInt(id),
		Method:  method,
	}

	if len(params) > 0 {
		jsonParams, err := json.Marshal(params)
		if err != nil {
			return jsonrpc.Request{}, err
		}
		request.Params = jsonrpc.NewParams(jsonParams)
	}

	return request, nil
}

// runAttack executes a load test for the given method
func runAttack(
	gatewayURL string,
	serviceID protocol.ServiceID,
	method jsonrpc.Method,
	methodDef methodDefinition,
	progressBar *pb.ProgressBar,
	params ...any,
) *MethodMetrics {
	// Initialize metrics for the method
	metrics := &MethodMetrics{
		method:      method,
		statusCodes: make(map[int]int),
		errors:      make(map[string]int),
		results:     make([]*vegeta.Result, 0, methodDef.totalRequests),
	}

	// Create target for the method
	target := createRPCTarget(gatewayURL, serviceID, method, params...)

	// Calculate max duration as a safety to prevent infinite runs
	// Allow 2x the theoretical time needed plus a 5 second buffer
	maxDuration := time.Duration(2*methodDef.totalRequests/methodDef.rps)*time.Second + 5*time.Second

	// Create an attacker
	attacker := vegeta.NewAttacker(
		vegeta.Timeout(5*time.Second),
		vegeta.KeepAlive(true),
		vegeta.Workers(methodDef.workers),
	)

	// Track exactly how many requests we've successfully processed
	processedCount := 0

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

				// Update progress bar exactly (prevent overflow)
				if progressBar.Current() < int64(methodDef.totalRequests) {
					progressBar.Increment()
				}
			}
		}

		// Ensure the progress bar shows exactly 100% at the end
		if progressBar.Current() < int64(methodDef.totalRequests) {
			remaining := int64(methodDef.totalRequests) - progressBar.Current()
			progressBar.Add64(remaining)
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

	// Update error counts if there's an error
	if result.Error != "" {
		m.errors[result.Error]++
	}

	// Process JSON-RPC validation if we have a successful HTTP response
	if result.Code >= 200 && result.Code < 300 && len(result.Body) > 0 {
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
	c := require.New(t)

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

	// Log top errors in red
	if len(m.errors) > 0 {
		fmt.Printf("\x1b[31mTop errors:\x1b[0m\n")
		count := 0
		for err, errCount := range m.errors {
			if count < 5 {
				fmt.Printf("  \x1b[31m%s\x1b[0m: %d\n", err, errCount)
				count++
			}
		}
		if len(m.errors) > 5 {
			fmt.Printf("  ... and \x1b[31m%d\x1b[0m more error types\n", len(m.errors)-5)
		}
	}

	// Perform all assertions
	assertHTTPSuccessRate(c, m, methodDef.successRate)
	assertJSONRPCRates(c, m, methodDef.successRate)
	assertLatency(c, m, methodDef)
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

// assertHTTPSuccessRate checks if the HTTP success rate meets requirements
func assertHTTPSuccessRate(c *require.Assertions, m *MethodMetrics, requiredRate float64) {
	msg := fmt.Sprintf("Method %s HTTP success rate %.2f%% should be >= %.2f%%",
		m.method, m.successRate*100, requiredRate*100)
	c.GreaterOrEqual(m.successRate, requiredRate, msg)
}

// assertJSONRPCRates checks if all JSON-RPC success rates meet requirements
func assertJSONRPCRates(c *require.Assertions, m *MethodMetrics, requiredRate float64) {
	// Skip if we don't have any JSON-RPC responses
	if m.jsonRPCResponses+m.jsonRPCUnmarshalErrors == 0 {
		return
	}

	// Check JSON-RPC unmarshal success rate
	msg := fmt.Sprintf("Method %s JSON-RPC unmarshal success rate %.2f%% should be >= %.2f%%",
		m.method, m.jsonRPCSuccessRate*100, requiredRate*100)
	c.GreaterOrEqual(m.jsonRPCSuccessRate, requiredRate, msg)

	// Skip the rest if we don't have valid JSON-RPC responses
	if m.jsonRPCResponses == 0 {
		return
	}

	// Check Error field absence rate
	msg = fmt.Sprintf("Method %s JSON-RPC error field absence rate %.2f%% should be >= %.2f%%",
		m.method, m.jsonRPCErrorFieldRate*100, requiredRate*100)
	c.GreaterOrEqual(m.jsonRPCErrorFieldRate, requiredRate, msg)

	// Check non-nil result rate
	msg = fmt.Sprintf("Method %s JSON-RPC non-nil result rate %.2f%% should be >= %.2f%%",
		m.method, m.jsonRPCResultRate*100, requiredRate*100)
	c.GreaterOrEqual(m.jsonRPCResultRate, requiredRate, msg)

	// Check validation success rate
	msg = fmt.Sprintf("Method %s JSON-RPC validation success rate %.2f%% should be >= %.2f%%",
		m.method, m.jsonRPCValidateRate*100, requiredRate*100)
	c.GreaterOrEqual(m.jsonRPCValidateRate, requiredRate, msg)
}

// assertLatency checks if the latency meets requirements
func assertLatency(c *require.Assertions, m *MethodMetrics, methodDef methodDefinition) {
	// P50 latency check
	msg := fmt.Sprintf("Method %s P50 latency %s should be <= %s",
		m.method, formatLatency(m.p50), formatLatency(methodDef.maxP50Latency))
	c.LessOrEqual(m.p50, methodDef.maxP50Latency, msg)

	// P95 latency check
	msg = fmt.Sprintf("Method %s P95 latency %s should be <= %s",
		m.method, formatLatency(m.p95), formatLatency(methodDef.maxP95Latency))
	c.LessOrEqual(m.p95, methodDef.maxP95Latency, msg)

	// P99 latency check
	msg = fmt.Sprintf("Method %s P99 latency %s should be <= %s",
		m.method, formatLatency(m.p99), formatLatency(methodDef.maxP99Latency))
	c.LessOrEqual(m.p99, methodDef.maxP99Latency, msg)
}

/* -------------------- Progress Bars -------------------- */

// progressBars holds and manages progress bars for all methods in a test
type progressBars struct {
	bars map[jsonrpc.Method]*pb.ProgressBar
	pool *pb.Pool
}

// newProgressBars creates a set of progress bars for all methods in a test
func newProgressBars(methods []jsonrpc.Method, methodDefs map[jsonrpc.Method]methodDefinition) (*progressBars, error) {
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

		// Create a bar with default template
		bar := pb.New(def.totalRequests)

		// Store the method name and padding for template use
		padding := longestLen - len(string(method))
		methodWithPadding := string(method) + strings.Repeat(" ", padding)
		bar.Set("prefix", methodWithPadding)

		// Set a simple template that just uses the prefix
		bar.SetTemplateString(`{{ string . "prefix" | blue }} {{ counters . }} {{ bar . "[" "=" ">" " " "]" | blue }} {{ percent . | green }}`)

		// Ensure we're not using byte formatting
		bar.Set(pb.Bytes, false)

		// Set max width for the bar
		bar.SetMaxWidth(100)

		bars[method] = bar
		barList = append(barList, bar)
	}

	// Create a pool with all the bars
	pool, err := pb.StartPool(barList...)
	if err != nil {
		return nil, err
	}

	return &progressBars{
		bars: bars,
		pool: pool,
	}, nil
}

// finish completes all progress bars
func (p *progressBars) finish() error {
	return p.pool.Stop()
}

// get returns the progress bar for a specific method
func (p *progressBars) get(method jsonrpc.Method) *pb.ProgressBar {
	return p.bars[method]
}

// formatLatency formats latency values to whole milliseconds
func formatLatency(d time.Duration) string {
	return fmt.Sprintf("%dms", d/time.Millisecond)
}
