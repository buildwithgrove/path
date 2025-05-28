//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cheggaaa/pb/v3"
	vegeta "github.com/tsenart/vegeta/lib"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
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

// ===== Vegeta Helper Functions =====

// runServiceTest runs the E2E test for a single EVM service in a test case.
func runServiceTest(
	t *testing.T,
	ctx context.Context,
	targets []vegeta.Target,
	serviceConfig ServiceConfig,
	serviceType serviceType,
	methodCount int,
	summary *serviceSummary,
) (serviceTestFailed bool) {
	results := make(map[string]*methodMetrics)
	var resultsMutex sync.Mutex

	// Get methods to test from the summary
	methods := summary.methodsToTest

	// Create a map for progress bars and summary calculation (with the same config for all methods)
	methodConfigMap := make(map[string]ServiceConfig)
	// Create a map to associate methods with their targets
	methodTargets := make(map[string]vegeta.Target)

	for i, method := range methods {
		methodConfigMap[method] = serviceConfig
		if i < len(targets) {
			methodTargets[method] = targets[i]
		}
	}

	progBars, err := newProgressBars(methods, methodConfigMap)
	if err != nil {
		t.Fatalf("Failed to create progress bars: %v", err)
	}
	defer func() {
		if err := progBars.finish(); err != nil {
			fmt.Printf("Error stopping progress bars: %v", err)
		}
	}()

	var methodWg sync.WaitGroup
	for _, method := range methods {
		methodWg.Add(1)

		target, exists := methodTargets[method]
		if !exists {
			fmt.Printf("Warning: No target found for method %s, skipping...\n", method)
			methodWg.Done()
			continue
		}

		go func(ctx context.Context, method string, serviceConfig ServiceConfig, target vegeta.Target) {
			defer methodWg.Done()

			metrics := runMethodAttack(
				ctx,
				method,
				serviceConfig,
				serviceType,
				methodCount,
				target,
				progBars.get(method),
			)

			resultsMutex.Lock()
			results[method] = metrics
			resultsMutex.Unlock()

		}(ctx, method, serviceConfig, target)
	}
	methodWg.Wait()

	if err := progBars.finish(); err != nil {
		fmt.Printf("Error stopping progress bars: %v", err)
	}

	// We use the same methodConfigMap we created earlier for the summary calculation

	calculateServiceSummary(t, methodConfigMap, results, summary, &serviceTestFailed)
	return serviceTestFailed
}

// runMethodAttack executes the attack for a single JSON-RPC method and returns metrics.
func runMethodAttack(
	ctx context.Context,
	method string,
	serviceConfig ServiceConfig,
	serviceType serviceType,
	methodCount int,
	target vegeta.Target,
	progBar *pb.ProgressBar,
) *methodMetrics {
	select {
	case <-ctx.Done():
		fmt.Printf("Method %s cancelled", method)
		return nil
	default:
	}

	// We don't need to extract or modify the target anymore, just pass it through
	metrics := runAttack(
		ctx,
		target,
		string(method), // Pass method name as string for metrics
		serviceConfig,
		serviceType,
		methodCount,
		progBar,
	)

	return metrics
}

// runAttack
// • Executes a load test for a given method
// • Sends `serviceConfig.totalRequests` requests at `serviceConfig.rps` requests/sec
// • DEV_NOTE: "Attack" is Vegeta's term for a single request
// • See: https://github.com/tsenart/vegeta
func runAttack(
	ctx context.Context,
	target vegeta.Target,
	methodName string, // Method name used for metrics and reporting
	serviceConfig ServiceConfig,
	serviceType serviceType,
	methodCount int,
	progressBar *pb.ProgressBar,
) *methodMetrics {
	// Calculate RPS per method, rounding up and ensuring at least 1 RPS
	attackRPS := max((serviceConfig.GlobalRPS+methodCount-1)/methodCount, 1)

	metrics := initMethodMetrics(string(methodName), serviceConfig.RequestsPerMethod)

	// Use the target directly, no need to recreate it
	targeter := func(tgt *vegeta.Target) error {
		*tgt = target
		return nil
	}

	maxDuration := time.Duration(2*serviceConfig.RequestsPerMethod/attackRPS)*time.Second + 5*time.Second

	// Vegeta timeout is set to the 99th percentile latency of the method + 5 seconds
	// This is because the P99 latency is the highest latency band for test assertions.
	// We add 5 seconds to account for any unexpected delays.
	attacker := createVegetaAttacker(attackRPS, serviceConfig.MaxP99LatencyMS+5*time.Second)

	if progressBar == nil {
		fmt.Printf("Starting test for method %s (%d requests at %d GlobalRPS)...\n",
			methodName, serviceConfig.RequestsPerMethod, attackRPS,
		)
	}

	resultsChan := make(chan *vegeta.Result, serviceConfig.RequestsPerMethod)
	var resultsWg sync.WaitGroup
	processedCount := 0
	startResultsCollector(
		&resultsWg,
		resultsChan,
		metrics,
		string(methodName),
		serviceConfig,
		serviceType,
		progressBar,
		&processedCount,
	)

	requestSlots := serviceConfig.RequestsPerMethod
	targeterWithLimit := makeTargeter(&requestSlots, targeter)
	attackCh := attacker.Attack(
		targeterWithLimit,
		vegeta.Rate{
			Freq: attackRPS,
			Per:  time.Second,
		},
		maxDuration,
		methodName,
	)
	runVegetaAttackLoop(ctx, attackCh, resultsChan)

	close(resultsChan)
	resultsWg.Wait()

	calculateSuccessRate(metrics)
	calculatePercentiles(metrics)
	return metrics
}

// initMethodMetrics
// • Initializes methodMetrics struct for a method
func initMethodMetrics(method string, totalRequests int) *methodMetrics {
	return &methodMetrics{
		method:      method,
		statusCodes: make(map[int]int),
		errors:      make(map[string]int),
		results:     make([]*vegeta.Result, 0, totalRequests),
	}
}

// createVegetaAttacker
// • Sets up a vegeta attacker with fixed options
func createVegetaAttacker(rps int, timeout time.Duration) *vegeta.Attacker {
	return vegeta.NewAttacker(
		vegeta.Timeout(timeout),
		vegeta.KeepAlive(true),
		vegeta.Workers(uint64(rps/2)),
		vegeta.MaxWorkers(uint64(rps)),
	)
}

// startResultsCollector
// • Launches a goroutine to process results, update progress bar, print status
func startResultsCollector(
	resultsWg *sync.WaitGroup,
	resultsChan <-chan *vegeta.Result,
	metrics *methodMetrics,
	method string,
	serviceConfig ServiceConfig,
	serviceType serviceType,
	progressBar *pb.ProgressBar,
	processedCount *int,
) {
	resultsWg.Add(1)
	go func() {
		defer resultsWg.Done()
		for res := range resultsChan {
			if res.Error == "no targets to attack" {
				continue
			}
			if *processedCount < serviceConfig.RequestsPerMethod {
				processResult(metrics, res, serviceType)
				(*processedCount)++
				if progressBar != nil && progressBar.Current() < int64(serviceConfig.RequestsPerMethod) {
					progressBar.Increment()
				}
				if progressBar == nil && *processedCount%50 == 0 {
					percent := float64(*processedCount) / float64(serviceConfig.RequestsPerMethod) * 100
					fmt.Printf("  %s: %d/%d requests completed (%.1f%%)\n",
						method, *processedCount, serviceConfig.RequestsPerMethod, percent)
				}
			}
		}
		if progressBar != nil && progressBar.Current() < int64(serviceConfig.RequestsPerMethod) {
			remaining := int64(serviceConfig.RequestsPerMethod) - progressBar.Current()
			progressBar.Add64(remaining)
		}
		if progressBar == nil {
			fmt.Printf("  %s: test completed (%d/%d requests)\n",
				method, *processedCount, serviceConfig.RequestsPerMethod)
		}
	}()
}

// makeTargeter
// • Returns a vegeta.Targeter that enforces the request limit
func makeTargeter(requestSlots *int, target vegeta.Targeter) vegeta.Targeter {
	return func(tgt *vegeta.Target) error {
		if *requestSlots <= 0 {
			return vegeta.ErrNoTargets
		}
		(*requestSlots)--
		return target(tgt)
	}
}

// runVegetaAttackLoop
// • Runs the attack loop, sending results to the channel and handling cancellation
func runVegetaAttackLoop(
	ctx context.Context,
	attackCh <-chan *vegeta.Result,
	resultsChan chan<- *vegeta.Result,
) {
attackLoop:
	for {
		select {
		case <-ctx.Done():
			break attackLoop
		case res, ok := <-attackCh:
			if !ok {
				break attackLoop
			}
			resultsChan <- res
		}
	}
}

// processResult
// • Updates metrics based on a single result
func processResult(m *methodMetrics, result *vegeta.Result, serviceType serviceType) {
	// Skip "no targets to attack" errors (not actual requests)
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
		m.jsonRPCUnmarshalErrors++
	} else {
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
		if err := rpcResponse.Validate(getExpectedID(serviceType)); err != nil {
			m.jsonRPCValidateErrors++
		}
	}
}

// ===== Assertions and Calculation Helpers =====

// ===== Metrics Types =====

// methodMetrics
// • Stores metrics for each method
// • Tracks HTTP and JSON-RPC results and derived rates
// • Used for assertion and reporting
type methodMetrics struct {
	method       string           // RPC method name
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

// serviceSummary
// • Holds aggregated metrics for a service
// • Used for service-level reporting
type serviceSummary struct {
	serviceID protocol.ServiceID

	avgP90Latency  time.Duration
	avgLatency     time.Duration
	avgSuccessRate float64

	totalRequests int
	totalSuccess  int
	totalFailure  int

	serviceConfig ServiceConfig
	methodsToTest []string
	methodErrors  map[string]map[string]int
	methodCount   int
	totalErrors   int
}

// calculateServiceSummary validates method results, aggregates summary metrics, and updates the service summary.
func calculateServiceSummary(
	t *testing.T,
	methodConfigs map[string]ServiceConfig,
	results map[string]*methodMetrics,
	summary *serviceSummary,
	serviceTestFailed *bool,
) {
	var totalLatency time.Duration
	var totalP90Latency time.Duration
	var totalSuccessRate float64
	var methodsWithResults int

	// Track service totals
	summary.totalRequests = 0
	summary.totalSuccess = 0
	summary.totalFailure = 0

	// Validate results for each method and collect summary data
	for method := range methodConfigs {
		methodMetrics := results[method]

		// Skip methods with no data
		if methodMetrics == nil || len(methodMetrics.results) == 0 {
			continue
		}

		// Convert ServiceConfig to methodTestConfig for validation
		methodDef := methodConfigs[method]
		serviceConfig := ServiceConfig{
			RequestsPerMethod: methodDef.RequestsPerMethod,
			GlobalRPS:         methodDef.GlobalRPS,
			SuccessRate:       methodDef.SuccessRate,
			MaxP50LatencyMS:   methodDef.MaxP50LatencyMS,
			MaxP95LatencyMS:   methodDef.MaxP95LatencyMS,
			MaxP99LatencyMS:   methodDef.MaxP99LatencyMS,
		}

		validateResults(t, methodMetrics, serviceConfig)

		// If the test has failed after validation, set the service failure flag
		if t.Failed() {
			*serviceTestFailed = true
		}

		// Accumulate totals for the service summary
		summary.totalRequests += methodMetrics.requestCount
		summary.totalSuccess += methodMetrics.success
		summary.totalFailure += methodMetrics.failed

		// Extract latencies for P90 calculation
		var latencies []time.Duration
		for _, res := range methodMetrics.results {
			latencies = append(latencies, res.Latency)
		}

		// Calculate P90 for this method
		p90 := calculateP90(latencies)
		avgLatency := calculateAvgLatency(latencies)

		// Add to summary totals
		totalLatency += avgLatency
		totalP90Latency += p90
		totalSuccessRate += methodMetrics.successRate
		methodsWithResults++

		// Collect errors for the summary
		if len(methodMetrics.errors) > 0 {
			// Initialize method errors map if not already created
			if summary.methodErrors[method] == nil {
				summary.methodErrors[method] = make(map[string]int)
			}

			// Copy errors to summary
			for errMsg, count := range methodMetrics.errors {
				summary.methodErrors[method][errMsg] = count
				summary.totalErrors += count
			}
		}
	}

	// Calculate averages if we have methods with results
	if methodsWithResults > 0 {
		summary.avgLatency = time.Duration(int64(totalLatency) / int64(methodsWithResults))
		summary.avgP90Latency = time.Duration(int64(totalP90Latency) / int64(methodsWithResults))
		summary.avgSuccessRate = totalSuccessRate / float64(methodsWithResults)
	}
}

// ===== Metric Calculation Helpers =====

// calculateSuccessRate
// • Computes all success rates for a methodMetrics struct
func calculateSuccessRate(m *methodMetrics) {
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
func calculatePercentiles(m *methodMetrics) {
	if len(m.results) == 0 {
		return
	}

	// Extract latencies
	latencies := make([]time.Duration, 0, len(m.results))
	for _, res := range m.results {
		latencies = append(latencies, res.Latency)
	}

	// Sort latencies
	slices.Sort(latencies)

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
func validateResults(t *testing.T, m *methodMetrics, serviceConfig ServiceConfig) {
	// Create a slice to collect all assertion failures
	var failures []string

	// Add a blank line before each test result for better readability
	fmt.Println()

	// Print metrics header with method name in blue
	fmt.Printf("%s====================== %s ======================%s\n", BOLD_BLUE, m.method, RESET)

	// Print success rate with color (green ≥99%, yellow ≥95%, red <95%)
	successColor := RED // Red by default
	if m.successRate >= 0.99 {
		successColor = GREEN // Green for ≥99%
	} else if m.successRate >= 0.95 {
		successColor = YELLOW // Yellow for ≥95%
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

	// Log top errors with appropriate color
	if len(m.errors) > 0 {
		fmt.Println("") // Add a new line before logging errors
		fmt.Printf("%sTop errors:%s\n", errorColor, RESET)
		count := 0
		num := 1
		for err, errCount := range m.errors {
			if count < 5 {
				fmt.Printf("  %d. %s%s%s: %d\n", num, errorColor, err, RESET, errCount)
				count++
				num++
			}
		}
		if len(m.errors) > 5 {
			fmt.Printf("  ... and %s%d%s more error types\n", errorColor, len(m.errors)-5, RESET)
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
	} else {
		fmt.Printf("\n%s✅ Method %s passed all assertions%s\n", GREEN, m.method, RESET)
	}
	fmt.Printf("%s====================== %s ======================%s\n", BOLD_BLUE, m.method, RESET)
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

// Helper function to get color for success rates
func getRateColor(rate, requiredRate float64) string {
	if rate >= requiredRate {
		return GREEN // Green for meeting requirement
	} else if rate >= requiredRate*0.95 {
		return YELLOW // Yellow for close
	}
	return RED // Red for failing
}

// Helper function to get color for latency values
func getLatencyColor(actual, maxAllowed time.Duration) string {
	if float64(actual) <= float64(maxAllowed)*0.7 {
		return GREEN // Green if well under limit (≤70%)
	} else if float64(actual) <= float64(maxAllowed) {
		return YELLOW // Yellow if close to limit (70-100%)
	}
	return RED // Red if over limit
}

// formatLatency formats latency values to whole milliseconds
func formatLatency(d time.Duration) string {
	return fmt.Sprintf("%dms", d/time.Millisecond)
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

// getAnyMethodKey returns an arbitrary method key from the map (first found)

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

	// Print summary for each service
	for _, svcID := range serviceIDs {
		summary := summaries[svcID]
		// TODO_TECHDEBT: Using a random key for now to avoid the effort of computing a mean (there are nuances involved).
		serviceConfig := summary.serviceConfig

		// Header with service ID
		fmt.Printf("\n%s⛓️  Service: %s%s\n", BOLD_BLUE, svcID, RESET)

		// Use helpers for coloring based on method config
		successColor := getRateColor(summary.avgSuccessRate, serviceConfig.SuccessRate)
		p90Color := getLatencyColor(summary.avgP90Latency, serviceConfig.MaxP95LatencyMS) // P90 closest to P95
		avgColor := getLatencyColor(summary.avgLatency, serviceConfig.MaxP50LatencyMS)    // Avg closest to P50

		// Calculate allowed failure rate
		maxFailureRate := 1.0 - serviceConfig.SuccessRate
		maxAllowedFailures := int(float64(summary.totalRequests) * maxFailureRate)

		// Color for failures based on whether they exceed the allowed threshold
		failureColor := GREEN
		if summary.totalFailure > maxAllowedFailures {
			failureColor = RED
		}

		// Print request stats
		fmt.Printf("  • Total Requests: %d\n", summary.totalRequests)
		fmt.Printf("  • Total Successes: %s%d%s\n", GREEN, summary.totalSuccess, RESET)
		fmt.Printf("  • Total Failures: %s%d%s", failureColor, summary.totalFailure, RESET)

		// Add context about allowed failures if there are any failures
		if summary.totalFailure > 0 {
			fmt.Printf(" (max allowed: %d)", maxAllowedFailures)
		}
		fmt.Println()

		fmt.Printf("  • Average Success Rate: %s%.2f%%%s\n",
			successColor, summary.avgSuccessRate*100, RESET)
		fmt.Printf("  • Average P90 Latency: %s%s%s\n",
			p90Color, formatLatency(summary.avgP90Latency), RESET)
		fmt.Printf("  • Average Latency: %s%s%s\n",
			avgColor, formatLatency(summary.avgLatency), RESET)
	}

	fmt.Printf("\n%s===== END SERVICE SUMMARY =====%s\n", BOLD_CYAN, RESET)
}

// ===== Progress Bars =====

// progressBars
// • Holds and manages progress bars for all methods in a test
// • Used to visualize test progress interactively
type progressBars struct {
	bars    map[string]*pb.ProgressBar
	pool    *pb.Pool
	enabled bool
}

// newProgressBars
// • Creates a set of progress bars for all methods in a test
// • Disables progress bars in CI/non-interactive environments
func newProgressBars(methods []string, testConfigs map[string]ServiceConfig) (*progressBars, error) {
	// Check if we're running in CI or non-interactive environment
	if isCIEnv() {
		fmt.Println("Running in CI environment - progress bars disabled")
		return &progressBars{
			bars:    make(map[string]*pb.ProgressBar),
			enabled: false,
		}, nil
	}

	// Sort methods for consistent display order
	sortedMethods := make([]string, len(methods))
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
	bars := make(map[string]*pb.ProgressBar)
	barList := make([]*pb.ProgressBar, 0, len(methods))

	for _, method := range sortedMethods {
		def := testConfigs[method]

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
		bar := pb.ProgressBarTemplate(tmpl).New(def.RequestsPerMethod)

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
			bars:    make(map[string]*pb.ProgressBar),
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
func (p *progressBars) get(method string) *pb.ProgressBar {
	if !p.enabled {
		return nil
	}
	return p.bars[method]
}

// showWaitBar shows a progress bar for the optional for hydrator checks to complete
func showWaitBar(secondsToWait int) {
	// Create a progress bar for the optional wait time
	waitBar := pb.ProgressBarTemplate(`{{ blue "Waiting" }} {{ printf "%2d/%2d" .Current .Total }} {{ bar . "[" "=" ">" " " "]" | blue }} {{ green (percent .) }}`).New(secondsToWait)
	waitBar.Set(pb.Bytes, false)
	waitBar.SetMaxWidth(100)
	waitBar.Start()

	// Wait for specified seconds, updating the progress bar every second
	for range secondsToWait {
		waitBar.Increment()
		<-time.After(1 * time.Second)
	}

	waitBar.Finish()
}
