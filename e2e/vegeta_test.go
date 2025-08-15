//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cheggaaa/pb/v3"
	vegeta "github.com/tsenart/vegeta/lib"

	"github.com/buildwithgrove/path/protocol"
)

// ===== Type Aliases for Vegeta =====
// These aliases allow us to use Vegeta types in our exported structs
// while maintaining clean separation between packages
type VegetaResult = vegeta.Result

// ===== Metrics Types (Exported for use across test files) =====

// MethodMetrics stores metrics for each method
// Tracks HTTP and JSON-RPC results and derived rates
// Used for assertion and reporting
type MethodMetrics struct {
	Method       string          // RPC method name
	Success      int             // Number of successful requests
	Failed       int             // Number of failed requests
	StatusCodes  map[int]int     // Count of each status code
	Errors       map[string]int  // Count of each error type
	Results      []*VegetaResult // All raw results for this method
	RequestCount int             // Total number of requests
	SuccessRate  float64         // Success rate as a ratio (0-1)
	P50          time.Duration   // 50th percentile latency
	P95          time.Duration   // 95th percentile latency
	P99          time.Duration   // 99th percentile latency

	// JSON-RPC specific validation metrics
	JSONRPCResponses       int // Count of responses we could unmarshal as JSON-RPC
	JSONRPCUnmarshalErrors int // Count of responses we couldn't unmarshal
	JSONRPCErrorField      int // Count of responses with non-nil Error field
	JSONRPCNilResult       int // Count of responses with nil Result field
	JSONRPCValidateErrors  int // Count of responses that fail validation

	// Error tracking with response previews
	JSONRPCParseErrors      map[string]int // Parse errors with response previews
	JSONRPCValidationErrors map[string]int // Validation errors with response previews

	// Success rates for specific checks
	JSONRPCSuccessRate    float64 // Success rate for JSON-RPC unmarshaling
	JSONRPCErrorFieldRate float64 // Error field absent rate (success = no error)
	JSONRPCResultRate     float64 // Non-nil result rate
	JSONRPCValidateRate   float64 // Validation success rate
}

// serviceSummary holds aggregated metrics for a service
// Used for service-level reporting
type serviceSummary struct {
	ServiceID protocol.ServiceID

	AvgP50Latency  time.Duration
	AvgP90Latency  time.Duration
	AvgLatency     time.Duration
	AvgSuccessRate float64

	TotalRequests int
	TotalSuccess  int
	TotalFailure  int

	ServiceConfig ServiceConfig
	MethodErrors  map[string]map[string]int
	MethodCount   int
	TotalErrors   int
}

// NewServiceSummary creates a new service summary
func newServiceSummary(serviceID protocol.ServiceID, serviceConfig ServiceConfig, testMethodsMap map[string]testMethodConfig) *serviceSummary {
	return &serviceSummary{
		ServiceID:     serviceID,
		ServiceConfig: serviceConfig,
		MethodErrors:  make(map[string]map[string]int),
		MethodCount:   len(testMethodsMap),
	}
}

// ===== Vegeta Helper Functions =====

// runServiceTest runs the E2E test for a single EVM service in a test case.
func runServiceTest(t *testing.T, ctx context.Context, ts *TestService) (serviceTestFailed bool) {
	results := make(map[string]*MethodMetrics)
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
func runMethodAttack(ctx context.Context, method string, ts *TestService, progBar *pb.ProgressBar) *MethodMetrics {
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
func runAttack(ctx context.Context, method string, ts *TestService, progressBar *pb.ProgressBar) *MethodMetrics {
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

	calculateAllSuccessRates(metrics)
	calculatePercentiles(metrics)
	return metrics
}

// initMethodMetrics
// • Initializes MethodMetrics struct for a method
func initMethodMetrics(method string, totalRequests int) *MethodMetrics {
	return &MethodMetrics{
		Method:      method,
		StatusCodes: make(map[int]int),
		Errors:      make(map[string]int),
		Results:     make([]*vegeta.Result, 0, totalRequests),
		// Initialize the error tracking fields
		JSONRPCParseErrors:      make(map[string]int),
		JSONRPCValidationErrors: make(map[string]int),
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
	ts *TestService,
	method string,
	methodConfig testMethodConfig,
	metrics *MethodMetrics,
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
				processResult(metrics, res, ts.serviceType, methodConfig.target.Body)
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

// processResult
// • Updates metrics based on a single result
func processResult(
	m *MethodMetrics,
	result *vegeta.Result,
	serviceType serviceType,
	httpRequestBody []byte,
) {
	// Skip "no targets to attack" errors (not actual requests)
	if result.Error == "no targets to attack" {
		return
	}
	// Store the raw result
	m.Results = append(m.Results, result)
	// Process HTTP result
	if result.Code >= 200 && result.Code < 300 && result.Error == "" {
		m.Success++
	} else {
		m.Failed++
	}
	// Update status code counts
	m.StatusCodes[int(result.Code)]++

	// If the request body contains "jsonrpc", it's a JSON-RPC request,
	// and we should process the result as a JSON-RPC response.
	if strings.Contains(string(httpRequestBody), "jsonrpc") {
		// Use the decoupled JSON-RPC validation function
		_ = validateJSONRPCResponse(result.Body, getExpectedID(serviceType), m)
	}
}

// calculateServiceSummary validates method results, aggregates summary metrics, and updates the service summary.
func calculateServiceSummary(
	t *testing.T,
	ts *TestService,
	results map[string]*MethodMetrics,
) bool {
	var serviceTestFailed bool = false

	methodConfigs := ts.testMethodsMap
	summary := ts.summary
	serviceId := ts.ServiceID

	// Validate results for each method
	for method := range methodConfigs {
		metrics := results[method]

		// Skip methods with no data
		if metrics == nil || len(metrics.Results) == 0 {
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

		// Use the decoupled validation function
		if !validateMethodResults(t, serviceId, metrics, methodTestConfig) {
			serviceTestFailed = true
		}
	}

	// Calculate service averages using the decoupled calculation functions
	calculateServiceAverages(summary, results)
	collectServiceErrors(summary, results)

	return serviceTestFailed
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
