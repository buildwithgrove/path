//go:build e2e

// ===== Vegeta Load Testing Engine =====
//
// This file contains the core Vegeta-based load testing functionality.
// Assertion and reporting code has been moved to assertions_test.go
// for better code organization.
//
// Contents:
// - Vegeta attack execution and coordination
// - Request/response processing
// - Progress bar management
// - Metrics collection and calculation

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

// ===== Vegeta Helper Functions =====

// runServiceTest runs the E2E test for a single EVM service in a test case.
func runServiceTest(t *testing.T, ctx context.Context, ts *TestService) (serviceTestFailed bool) {
	results := make(map[string]*methodMetrics)
	var resultsMutex sync.Mutex

	progBars, err := newProgressBars(ts.testMethodsMap)
	if err != nil {
		t.Fatalf("Failed to create progress bars: %v", err)
	}
	defer func() {
		if err := progBars.finish(); err != nil {
			fmt.Printf("Error stopping progress bars: %v", err)
		}
	}()

	var methodWg sync.WaitGroup
	for method := range ts.testMethodsMap {
		methodWg.Add(1)

		go func(ctx context.Context, method string, methodConfig testMethodConfig) {
			defer methodWg.Done()

			metrics := runMethodAttack(ctx, method, ts, progBars.get(method))

			resultsMutex.Lock()
			results[method] = metrics
			resultsMutex.Unlock()

		}(ctx, method, ts.testMethodsMap[method])
	}
	methodWg.Wait()

	if err := progBars.finish(); err != nil {
		fmt.Printf("Error stopping progress bars: %v", err)
	}

	return calculateServiceSummary(t, ts, results)
}

// runMethodAttack executes the attack for a single JSON-RPC method and returns metrics.
func runMethodAttack(ctx context.Context, method string, ts *TestService, progBar *pb.ProgressBar) *methodMetrics {
	select {
	case <-ctx.Done():
		fmt.Printf("Method %s canceled", method)
		return nil
	default:
	}

	// We don't need to extract or modify the target anymore, just pass it through
	metrics := runAttack(ctx, method, ts, progBar)

	return metrics
}

// runAttack
// • Executes a load test for a given method
// • Sends `serviceConfig.totalRequests` requests at `serviceConfig.rps` requests/sec
// • DEV_NOTE: "Attack" is Vegeta's term for a single request
// • See: https://github.com/tsenart/vegeta
func runAttack(ctx context.Context, method string, ts *TestService, progressBar *pb.ProgressBar) *methodMetrics {
	methodConfig := ts.testMethodsMap[method]

	// Calculate RPS per method, rounding up and ensuring at least 1 RPS
	attackRPS := max((methodConfig.serviceConfig.GlobalRPS+len(ts.testMethodsMap)-1)/len(ts.testMethodsMap), 1)

	// Initialize the method metrics
	metrics := initMethodMetrics(method, methodConfig.serviceConfig.RequestsPerMethod)

	// Use the target directly, no need to recreate it
	targeter := func(tgt *vegeta.Target) error {
		*tgt = methodConfig.target
		return nil
	}

	maxDuration := time.Duration(2*methodConfig.serviceConfig.RequestsPerMethod/attackRPS)*time.Second + 5*time.Second

	// Vegeta timeout is set to the 99th percentile latency of the method + 5 seconds
	// This is because the P99 latency is the highest latency band for test assertions.
	// We add 5 seconds to account for any unexpected delays.
	attacker := createVegetaAttacker(attackRPS, methodConfig.serviceConfig.MaxP99LatencyMS+5*time.Second)

	if progressBar == nil {
		fmt.Printf("Starting test for method %s (%d requests at %d GlobalRPS)...\n",
			method, methodConfig.serviceConfig.RequestsPerMethod, attackRPS,
		)
	}

	// Create a channel to collect results
	resultsChan := make(chan *vegeta.Result, methodConfig.serviceConfig.RequestsPerMethod)

	// Start a goroutine to process results
	var resultsWg sync.WaitGroup
	startResultsCollector(
		ts,
		method,
		methodConfig,
		metrics,
		resultsChan,
		&resultsWg,
		progressBar,
	)

	// Run the Vegeta attack
	attackCh := attacker.Attack(
		makeTargeter(methodConfig, targeter),
		vegeta.Rate{Freq: attackRPS, Per: time.Second},
		maxDuration,
		method,
	)

	// Run the attack loop, sending results to the channel and handling cancellation
	runVegetaAttackLoop(ctx, attackCh, resultsChan)

	close(resultsChan)
	resultsWg.Wait()

	calculateSuccessRate(metrics)
	calculatePercentiles(metrics)
	return metrics
}

// initMethodMetrics
// • Initializes serviceConfig struct for a method
func initMethodMetrics(method string, totalRequests int) *methodMetrics {
	return &methodMetrics{
		method:      method,
		statusCodes: make(map[int]int),
		errors:      make(map[string]int),
		results:     make([]*vegeta.Result, 0, totalRequests),
		// Initialize the new error tracking fields
		jsonRPCParseErrors:      make(map[string]int),
		jsonRPCValidationErrors: make(map[string]int),
	}
}

// createVegetaAttacker
// • Sets up a vegeta attacker with fixed options
func createVegetaAttacker(rps int, timeout time.Duration) *vegeta.Attacker {
	return vegeta.NewAttacker(
		vegeta.Timeout(timeout),
		vegeta.KeepAlive(true),
		vegeta.Workers(uint64(max(rps/2, 1))),  // Ensure at least 1 worker
		vegeta.MaxWorkers(uint64(max(rps, 1))), // Ensure at least 1 max worker
	)
}

// startResultsCollector
// • Launches a goroutine to process results, update progress bar, print status
func startResultsCollector(
	ts *TestService,
	method string,
	methodConfig testMethodConfig,
	metrics *methodMetrics,
	resultsChan <-chan *vegeta.Result,
	resultsWg *sync.WaitGroup,
	progressBar *pb.ProgressBar,
) {
	processedCount := 0
	resultsWg.Add(1)
	go func() {
		defer resultsWg.Done()
		for res := range resultsChan {
			if res.Error == "no targets to attack" {
				continue
			}
			if processedCount < methodConfig.serviceConfig.RequestsPerMethod {
				processResult(metrics, res, ts.serviceType)
				processedCount++
				if progressBar != nil && progressBar.Current() < int64(methodConfig.serviceConfig.RequestsPerMethod) {
					progressBar.Increment()
				}
				if progressBar == nil && processedCount%50 == 0 {
					percent := float64(processedCount) / float64(methodConfig.serviceConfig.RequestsPerMethod) * 100
					fmt.Printf("  %s: %d/%d requests completed (%.1f%%)\n",
						method, processedCount, methodConfig.serviceConfig.RequestsPerMethod, percent)
				}
			}
		}
		if progressBar != nil && progressBar.Current() < int64(methodConfig.serviceConfig.RequestsPerMethod) {
			remaining := int64(methodConfig.serviceConfig.RequestsPerMethod) - progressBar.Current()
			progressBar.Add64(remaining)
		}
		if progressBar == nil {
			fmt.Printf("  %s: test completed (%d/%d requests)\n",
				method, processedCount, methodConfig.serviceConfig.RequestsPerMethod,
			)
		}
	}()
}

// makeTargeter
// • Returns a vegeta.Targeter that enforces the request limit
func makeTargeter(methodConfig testMethodConfig, target vegeta.Targeter) vegeta.Targeter {
	requestSlots := methodConfig.serviceConfig.RequestsPerMethod

	return func(tgt *vegeta.Target) error {
		if requestSlots <= 0 {
			return vegeta.ErrNoTargets
		}
		requestSlots--
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

// createResponsePreview creates a sanitized preview of the response body for error logging
func createResponsePreview(body []byte, maxLen int) string {
	if len(body) == 0 {
		return "(empty)"
	}

	// Convert to string and normalize whitespace using strings lib
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

// processJSONRPCResponse handles both single and batch JSON-RPC responses
func processJSONRPCResponse(body []byte, m *methodMetrics, serviceType serviceType) error {
	// Check if response is a batch (array) by looking at the first non-whitespace character
	trimmed := strings.TrimSpace(string(body))
	if len(trimmed) == 0 {
		return fmt.Errorf("empty response body")
	}

	if trimmed[0] == '[' {
		// Batch response - unmarshal as array
		var batchResponse []jsonrpc.Response
		if err := json.Unmarshal(body, &batchResponse); err != nil {
			return err
		}

		// For batch responses, validate that all expected IDs are present
		expectedIDs := []int{1, 2, 3} // Expected IDs for our batch request (eth_blockNumber, eth_chainId, eth_gasPrice)
		processBatchJSONRPCResponse(batchResponse, m, expectedIDs)
	} else {
		// Single response
		var rpcResponse jsonrpc.Response
		if err := json.Unmarshal(body, &rpcResponse); err != nil {
			return err
		}

		m.jsonRPCResponses++
		expectedID := getExpectedID(serviceType)
		processSingleJSONRPCResponseWithID(&rpcResponse, m, expectedID, "single")
	}

	return nil
}

// processBatchJSONRPCResponse validates and processes a batch of JSON-RPC responses
// It validates that all expected IDs are present regardless of order
func processBatchJSONRPCResponse(batchResponse []jsonrpc.Response, m *methodMetrics, expectedIDs []int) {
	// Create maps for tracking
	responsesByID := make(map[int]*jsonrpc.Response)
	expectedIDsMap := make(map[int]bool)

	// Build expected IDs map
	for _, id := range expectedIDs {
		expectedIDsMap[id] = true
	}

	// Parse responses and group by ID
	for i := range batchResponse {
		response := &batchResponse[i]
		m.jsonRPCResponses++

		// Extract ID as int - check if it's an integer ID
		if !response.ID.IsEmpty() {
			// Access the private intID field via the String() method and parse
			idStr := response.ID.String()
			if idStr != "null" {
				// Try to parse as integer for our batch validation
				var responseID int
				if n, err := fmt.Sscanf(idStr, "%d", &responseID); err == nil && n == 1 {
					responsesByID[responseID] = response
				} else {
					// Handle non-integer IDs as validation errors
					m.jsonRPCValidateErrors++
					errorMsg := fmt.Sprintf("[batch] Invalid ID type: expected integer, got %s", idStr)
					m.jsonRPCValidationErrors[errorMsg]++
					m.errors[errorMsg]++
					continue
				}
			} else {
				// Handle null ID
				m.jsonRPCValidateErrors++
				errorMsg := "[batch] Null ID in batch response"
				m.jsonRPCValidationErrors[errorMsg]++
				m.errors[errorMsg]++
				continue
			}
		} else {
			// Handle missing/empty ID
			m.jsonRPCValidateErrors++
			errorMsg := "[batch] Missing ID in response"
			m.jsonRPCValidationErrors[errorMsg]++
			m.errors[errorMsg]++
			continue
		}
	}

	// Validate that all expected IDs are present
	for _, expectedID := range expectedIDs {
		if response, found := responsesByID[expectedID]; found {
			// Process the response with the correct expected ID
			expectedIDObj := jsonrpc.IDFromInt(expectedID)
			processSingleJSONRPCResponseWithID(response, m, expectedIDObj, fmt.Sprintf("batch[id=%d]", expectedID))
		} else {
			// Missing expected ID
			m.jsonRPCValidateErrors++
			errorMsg := fmt.Sprintf("[batch] Missing expected ID %d in batch response", expectedID)
			m.jsonRPCValidationErrors[errorMsg]++
			m.errors[errorMsg]++
		}
	}

	// Check for unexpected IDs (extra responses)
	for responseID := range responsesByID {
		if !expectedIDsMap[responseID] {
			m.jsonRPCValidateErrors++
			errorMsg := fmt.Sprintf("[batch] Unexpected ID %d in batch response", responseID)
			m.jsonRPCValidationErrors[errorMsg]++
			m.errors[errorMsg]++
		}
	}
}

// processSingleJSONRPCResponseWithID validates and processes a single JSON-RPC response with explicit expected ID
func processSingleJSONRPCResponseWithID(rpcResponse *jsonrpc.Response, m *methodMetrics, expectedID jsonrpc.ID, context string) {
	// Validate the response first
	validationErr := rpcResponse.Validate(expectedID)

	// Check if Error field is nil (good)
	if rpcResponse.Error != nil {
		m.jsonRPCErrorField++
		// Only track the error field message if there's no validation error
		// (to avoid duplicate tracking when validation fails due to error field)
		if validationErr == nil {
			errorMsg := fmt.Sprintf("[%s] RPC Error: %s", context, rpcResponse.Error.Message)
			m.errors[errorMsg]++
		}
	}

	// Check if Result field is not nil (good)
	if rpcResponse.Result == nil {
		m.jsonRPCNilResult++
	}

	// Process validation error
	if validationErr != nil {
		m.jsonRPCValidateErrors++

		errorMsg := fmt.Sprintf("[%s] JSON-RPC validation error: %v", context, validationErr)
		m.jsonRPCValidationErrors[errorMsg]++
		m.errors[errorMsg]++
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

	// Update status code counts
	m.statusCodes[int(result.Code)]++

	// Track if this request should be considered successful
	httpSuccess := result.Code >= 200 && result.Code < 300 && result.Error == ""

	// Track validation error counts before processing
	preValidationErrors := m.jsonRPCUnmarshalErrors + m.jsonRPCValidateErrors + m.jsonRPCErrorField

	// Process JSON-RPC validation if we have a successful HTTP response
	if httpSuccess {
		if err := processJSONRPCResponse(result.Body, m, serviceType); err != nil {
			m.jsonRPCUnmarshalErrors++

			// Create response preview for parse errors
			preview := createResponsePreview(result.Body, 100)
			errorMsg := fmt.Sprintf("JSON parse error: %v (response preview: %s)", err, preview)
			m.jsonRPCParseErrors[errorMsg]++
			m.errors[errorMsg]++
		}
	}

	// Check if any validation errors occurred during processing
	postValidationErrors := m.jsonRPCUnmarshalErrors + m.jsonRPCValidateErrors + m.jsonRPCErrorField
	jsonRPCSuccess := httpSuccess && (postValidationErrors == preValidationErrors)

	// Count as success only if both HTTP and JSON-RPC validation succeed
	if jsonRPCSuccess {
		m.success++
	} else {
		m.failed++
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

	// New fields for detailed error tracking with response previews
	jsonRPCParseErrors      map[string]int // Parse errors with response previews
	jsonRPCValidationErrors map[string]int // Validation errors with response previews

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

	avgP50Latency  time.Duration
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

func newServiceSummary(serviceID protocol.ServiceID, serviceConfig ServiceConfig, methodsToTest []string) *serviceSummary {
	return &serviceSummary{
		serviceID:     serviceID,
		serviceConfig: serviceConfig,
		methodsToTest: methodsToTest,
		methodErrors:  make(map[string]map[string]int),
		methodCount:   len(methodsToTest),
	}
}

// calculateServiceSummary validates method results, aggregates summary metrics, and updates the service summary.
func calculateServiceSummary(
	t *testing.T,
	ts *TestService,
	results map[string]*methodMetrics,
) bool {
	var serviceTestFailed bool = false
	var totalLatency time.Duration
	var totalP50Latency time.Duration
	var totalP90Latency time.Duration
	var totalSuccessRate float64
	var methodsWithResults int
	var failedMethods []string

	methodConfigs := ts.testMethodsMap
	summary := ts.summary
	serviceId := ts.ServiceID

	// Track service totals
	summary.totalRequests = 0
	summary.totalSuccess = 0
	summary.totalFailure = 0

	// Validate results for each method and collect summary data
	for method := range methodConfigs {
		serviceConfig := results[method]

		// Skip methods with no data
		if serviceConfig == nil || len(serviceConfig.results) == 0 {
			continue
		}

		// Convert ServiceConfig to methodTestConfig for validation
		methodDef := methodConfigs[method]
		methodTestConfig := ServiceConfig{
			RequestsPerMethod: methodDef.serviceConfig.RequestsPerMethod,
			GlobalRPS:         methodDef.serviceConfig.GlobalRPS,
			SuccessRate:       methodDef.serviceConfig.SuccessRate,
			MaxP50LatencyMS:   methodDef.serviceConfig.MaxP50LatencyMS,
			MaxP95LatencyMS:   methodDef.serviceConfig.MaxP95LatencyMS,
			MaxP99LatencyMS:   methodDef.serviceConfig.MaxP99LatencyMS,
		}

		methodFailed := validateResults(t, serviceId, serviceConfig, methodTestConfig)

		// Track failed methods and set service failure flag
		if methodFailed {
			serviceTestFailed = true
			failedMethods = append(failedMethods, fmt.Sprintf("%s.%s", serviceId, method))
		}

		// Accumulate totals for the service summary
		summary.totalRequests += serviceConfig.requestCount
		summary.totalSuccess += serviceConfig.success
		summary.totalFailure += serviceConfig.failed

		// Extract latencies for P90 calculation
		var latencies []time.Duration
		for _, res := range serviceConfig.results {
			latencies = append(latencies, res.Latency)
		}

		// Calculate p50 and p90 latencies for this method
		p50 := calculateP50(latencies)
		p90 := calculateP90(latencies)
		avgLatency := calculateAvgLatency(latencies)

		// Add to summary totals
		totalLatency += avgLatency
		totalP50Latency += p50
		totalP90Latency += p90
		totalSuccessRate += serviceConfig.successRate
		methodsWithResults++

		// Collect errors for the summary
		if len(serviceConfig.errors) > 0 {
			// Initialize method errors map if not already created
			if summary.methodErrors[method] == nil {
				summary.methodErrors[method] = make(map[string]int)
			}

			// Copy errors to summary
			for errMsg, count := range serviceConfig.errors {
				summary.methodErrors[method][errMsg] = count
				summary.totalErrors += count
			}
		}
	}

	// Calculate averages if we have methods with results
	if methodsWithResults > 0 {
		summary.avgLatency = time.Duration(int64(totalLatency) / int64(methodsWithResults))
		summary.avgP50Latency = time.Duration(int64(totalP50Latency) / int64(methodsWithResults))
		summary.avgP90Latency = time.Duration(int64(totalP90Latency) / int64(methodsWithResults))
		summary.avgSuccessRate = totalSuccessRate / float64(methodsWithResults)
	}

	// Print clear failure summary if there were any failures
	printTestFailureSummary(t, failedMethods, summary.totalErrors)

	return serviceTestFailed
}

// ===== Metric Calculation Helpers =====

// calculateSuccessRate
// • Computes all success rates for a serviceConfig struct
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
func newProgressBars(testMethodsMap map[string]testMethodConfig) (*progressBars, error) {
	// Check if we're running in CI or non-interactive environment
	if isCIEnv() {
		fmt.Println("Running in CI environment - progress bars disabled")
		return &progressBars{
			bars:    make(map[string]*pb.ProgressBar),
			enabled: false,
		}, nil
	}

	// Sort methods for consistent display order
	var sortedMethods []string
	for method := range testMethodsMap {
		sortedMethods = append(sortedMethods, method)
	}
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
	barList := make([]*pb.ProgressBar, 0, len(testMethodsMap))

	for _, method := range sortedMethods {
		def := testMethodsMap[method]

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
		bar := pb.ProgressBarTemplate(tmpl).New(def.serviceConfig.RequestsPerMethod)

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
