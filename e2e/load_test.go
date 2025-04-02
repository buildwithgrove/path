package e2e

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/stretchr/testify/require"
	vegeta "github.com/tsenart/vegeta/lib"
)

// RPCRequest represents a JSON-RPC request
type RPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  Method `json:"method"`
	Params  []any  `json:"params,omitempty"`
}

// MethodDefinition combines configuration and requirements for a method
type MethodDefinition struct {
	// Configuration
	TotalRequests int    // Total number of requests to send
	RPS           int    // Requests per second
	Workers       uint64 // Number of workers to use

	// Requirements
	SuccessRate   float64       // Minimum success rate (0-1)
	MaxP50Latency time.Duration // Maximum P50 latency
	MaxP95Latency time.Duration // Maximum P95 latency
	MaxP99Latency time.Duration // Maximum P99 latency
}

// MethodMetrics stores metrics for each method
type MethodMetrics struct {
	Method       Method           // RPC method name
	Success      int              // Number of successful requests
	Failed       int              // Number of failed requests
	StatusCodes  map[int]int      // Count of each status code
	Errors       map[string]int   // Count of each error type
	Results      []*vegeta.Result // All raw results for this method
	RequestCount int              // Total number of requests
	SuccessRate  float64          // Success rate as a ratio (0-1)
	P50          time.Duration    // 50th percentile latency
	P95          time.Duration    // 95th percentile latency
	P99          time.Duration    // 99th percentile latency
}

type Method string

const (
	ethBlockNumber Method = "eth_blockNumber"
)

// Set default gateway URL
var gatewayURL = "http://localhost:3069/v1"

// init initializes the gateway URL with an optional override
func init() {
	if gatewayURLOverride := os.Getenv("GATEWAY_URL"); gatewayURLOverride != "" {
		gatewayURL = gatewayURLOverride
	}
}

var methodDefinitions = map[Method]MethodDefinition{
	ethBlockNumber: {
		// Configuration
		TotalRequests: 300,
		RPS:           50,
		Workers:       20,

		// Requirements
		SuccessRate:   0.99,
		MaxP50Latency: 75 * time.Millisecond,
		MaxP95Latency: 150 * time.Millisecond,
		MaxP99Latency: 300 * time.Millisecond,
	},
}

// Test_PATHLoad runs load tests against JSON-RPC endpoints
func Test_PATHLoad(t *testing.T) {
	t.Logf("Starting load test with gateway URL: %s", gatewayURL)

	// Define test cases by service
	testCases := []struct {
		name      string
		serviceID string
		methods   []Method
	}{
		{
			name:      "F00C (Ethereum) service",
			serviceID: "F00C",
			methods: []Method{
				ethBlockNumber,
			},
		},
	}

	// Run tests for each service
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing service: %s", tc.serviceID)

			// Create results map
			results := make(map[Method]*MethodMetrics)

			// Run attack for each method
			for _, method := range tc.methods {
				methodDef, ok := methodDefinitions[method]
				if !ok {
					t.Fatalf("Method %s not found in methodDefinitions", method)
				}

				// Log method configuration
				t.Logf("Method: %s, Total Requests: %d, RPS: %d",
					method, methodDef.TotalRequests, methodDef.RPS)

				// Run the attack in a goroutine
				results[method] = runAttack(gatewayURL, tc.serviceID, method, methodDef)
			}

			// Validate results for each method
			for _, method := range tc.methods {
				validateResults(t, results[method], methodDefinitions[method])
			}
		})
	}
}

// createRPCTarget creates a vegeta.Targeter for the specified RPC method
func createRPCTarget(gatewayURL, serviceID string, method Method, params ...any) vegeta.Targeter {
	return func(tgt *vegeta.Target) error {
		req := RPCRequest{
			JSONRPC: "2.0",
			ID:      1,
			Method:  method,
		}
		// Initialize empty array for params if none provided
		if len(params) > 0 {
			req.Params = params
		}

		body, err := json.Marshal(req)
		if err != nil {
			return err
		}

		tgt.Method = "POST"
		tgt.URL = gatewayURL
		tgt.Body = body
		tgt.Header = http.Header{
			"Content-Type":      []string{"application/json"},
			"Target-Service-Id": []string{serviceID},
		}

		return nil
	}
}

// runAttack executes a load test for the given method
func runAttack(
	gatewayURL string,
	serviceID string,
	method Method,
	methodDef MethodDefinition,
	params ...any,
) *MethodMetrics {
	// Initialize metrics for the method
	metrics := &MethodMetrics{
		Method:      method,
		StatusCodes: make(map[int]int),
		Errors:      make(map[string]int),
		Results:     make([]*vegeta.Result, 0, methodDef.TotalRequests),
	}

	// Create target for the method
	target := createRPCTarget(gatewayURL, serviceID, method, params...)

	// Create a progress bar that shows N/total requests
	tmpl := `{{ string . "method" | blue}} {{ countf . "/" }} {{ bar . "[" "=" ">" "_" "]"}} {{percent . }} {{speed . "req/s" | rndcolor}} {{rtime . "ETA %s"}} `
	bar := pb.ProgressBarTemplate(tmpl).New(methodDef.TotalRequests)
	bar.Set("method", method)
	bar.Start()
	defer bar.Finish()

	// Create a rate with configured RPS
	rate := vegeta.Rate{
		Freq: methodDef.RPS,
		Per:  time.Second,
	}

	// Calculate max duration as a safety to prevent infinite runs
	// Allow 2x the theoretical time needed plus a 5 second buffer
	maxDuration := time.Duration(2*methodDef.TotalRequests/methodDef.RPS)*time.Second + 5*time.Second

	// Create an attacker
	attacker := vegeta.NewAttacker(
		vegeta.Timeout(5*time.Second),
		vegeta.KeepAlive(true),
		vegeta.Workers(methodDef.Workers),
	)

	// Use channels to control exactly how many requests are processed
	resultsChan := make(chan *vegeta.Result, methodDef.TotalRequests)

	// Create a channel with exactly the number of requests we want to process
	requestSlots := make(chan struct{}, methodDef.TotalRequests)
	for i := 0; i < methodDef.TotalRequests; i++ {
		requestSlots <- struct{}{}
	}
	close(requestSlots)

	// Start collection goroutine
	var resultsWg sync.WaitGroup
	resultsWg.Add(1)
	go func() {
		defer resultsWg.Done()
		for res := range resultsChan {
			processResult(metrics, res)
			bar.Increment()
		}
	}()

	// Use a targeting function that takes from our fixed pool of request slots
	targeter := func(tgt *vegeta.Target) error {
		select {
		case _, ok := <-requestSlots:
			if !ok {
				return vegeta.ErrNoTargets
			}
			return target(tgt)
		default:
			return vegeta.ErrNoTargets
		}
	}

	// Run the attack until we hit the total number of requests
	for res := range attacker.Attack(targeter, rate, maxDuration, string(method)) {
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
	m.Results = append(m.Results, result)

	// Process result
	if result.Code >= 200 && result.Code < 300 && result.Error == "" {
		m.Success++
	} else {
		m.Failed++
	}

	// Update status code counts
	m.StatusCodes[int(result.Code)]++

	// Update error counts if there's an error
	if result.Error != "" {
		m.Errors[result.Error]++
	}
}

// calculateSuccessRate computes the success rate
func calculateSuccessRate(m *MethodMetrics) {
	m.RequestCount = m.Success + m.Failed
	if m.RequestCount > 0 {
		m.SuccessRate = float64(m.Success) / float64(m.RequestCount)
	}
}

// calculatePercentiles computes P50, P95, and P99 latency percentiles
func calculatePercentiles(m *MethodMetrics) {
	if len(m.Results) == 0 {
		return
	}

	// Extract latencies
	latencies := make([]time.Duration, 0, len(m.Results))
	for _, res := range m.Results {
		latencies = append(latencies, res.Latency)
	}

	// Sort latencies
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	// Calculate percentiles
	m.P50 = percentile(latencies, 50)
	m.P95 = percentile(latencies, 95)
	m.P99 = percentile(latencies, 99)
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
func validateResults(t *testing.T, m *MethodMetrics, methodDef MethodDefinition) {
	c := require.New(t)

	// Print metrics
	t.Logf("========= Test Results for %s =========", m.Method)
	t.Logf("Success Rate: %.2f%% (%d/%d requests)", m.SuccessRate*100, m.Success, m.RequestCount)
	t.Logf("Latency P50: %s, P95: %s, P99: %s", m.P50, m.P95, m.P99)

	// Log status codes
	if len(m.StatusCodes) > 0 {
		statusText := "Status Codes: "
		for code, count := range m.StatusCodes {
			statusText += fmt.Sprintf("%d:%d ", code, count)
		}
		t.Log(statusText)
	}

	// Log top errors (limit to 5)
	if len(m.Errors) > 0 {
		t.Logf("Top Errors:")
		count := 0
		for err, errCount := range m.Errors {
			if count < 5 {
				t.Logf("  %s: %d", err, errCount)
				count++
			}
		}
		if len(m.Errors) > 5 {
			t.Logf("  ... and %d more error types", len(m.Errors)-5)
		}
	}

	// Perform assertions
	assertSuccessRate(c, m, methodDef.SuccessRate)
	assertLatency(c, m, methodDef)
}

// assertSuccessRate checks if the success rate meets requirements
func assertSuccessRate(c *require.Assertions, m *MethodMetrics, requiredRate float64) {
	msg := fmt.Sprintf("Method %s success rate %.2f%% should be >= %.2f%%",
		m.Method, m.SuccessRate*100, requiredRate*100)
	c.GreaterOrEqual(m.SuccessRate, requiredRate, msg)
}

// assertLatency checks if the latency meets requirements
func assertLatency(c *require.Assertions, m *MethodMetrics, methodDef MethodDefinition) {
	// P50 latency check
	msg := fmt.Sprintf("Method %s P50 latency %s should be <= %s",
		m.Method, m.P50, methodDef.MaxP50Latency)
	c.LessOrEqual(m.P50, methodDef.MaxP50Latency, msg)

	// P95 latency check
	msg = fmt.Sprintf("Method %s P95 latency %s should be <= %s",
		m.Method, m.P95, methodDef.MaxP95Latency)
	c.LessOrEqual(m.P95, methodDef.MaxP95Latency, msg)

	// P99 latency check
	msg = fmt.Sprintf("Method %s P99 latency %s should be <= %s",
		m.Method, m.P99, methodDef.MaxP99Latency)
	c.LessOrEqual(m.P99, methodDef.MaxP99Latency, msg)
}
