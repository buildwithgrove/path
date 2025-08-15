//go:build e2e

package e2e

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
)

// ===== ANSI Color Constants (for log output) =====
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

// ===== Helper Functions for Colors and Formatting =====

// getSuccessRateColor returns color based on success rate
func getSuccessRateColor(rate float64) string {
	if rate >= 0.90 {
		return GREEN
	} else if rate >= 0.70 {
		return YELLOW
	}
	return RED
}

// getRateColor returns color for success rates compared to required rate
func getRateColor(rate, requiredRate float64) string {
	if rate >= requiredRate {
		return GREEN
	} else if rate >= requiredRate*0.50 {
		return YELLOW
	}
	return RED
}

// getRateEmoji returns emoji for success rates
func getRateEmoji(rate, requiredRate float64) string {
	if rate >= requiredRate {
		return "🟢"
	} else if rate >= requiredRate*0.50 {
		return "🟡"
	}
	return "🔴"
}

// getLatencyColor returns color for latency values
func getLatencyColor(actual, maxAllowed time.Duration) string {
	if float64(actual) <= float64(maxAllowed)*0.5 {
		return GREEN // Well under limit
	} else if float64(actual) <= float64(maxAllowed) {
		return YELLOW // Close to limit
	}
	return RED // Over limit
}

// getStatusCodeColor returns color based on HTTP status code
func getStatusCodeColor(code int) string {
	if code >= 200 && code < 300 {
		return GREEN // 2xx success
	} else if code >= 300 && code < 400 {
		return YELLOW // 3xx redirect
	}
	return RED // 4xx/5xx error
}

// formatLatency formats latency values to whole milliseconds
func formatLatency(d time.Duration) string {
	return fmt.Sprintf("%dms", d/time.Millisecond)
}

// ===== Progress Bar Management =====
//
// Progress bars provide visual feedback during HTTP load tests.
// They are automatically disabled in CI environments for clean log output.

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
