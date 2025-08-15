//go:build e2e

package e2e

import (
	"fmt"
	"time"
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
