package evm

// endpoint captures the details required to validate an EVM endpoint.
// It contains all checks that should be run for the endpoint to validate
// it is providing a valid response to service requests.
type endpoint struct {
	// Whether the endpoint has returned an empty response
	hasReturnedEmptyResponse bool

	// Endpoint QoS checks - see `check_*.go` files
	// for implementation details
	checkBlockNumber endpointCheckBlockNumber
	checkChainID     endpointCheckChainID
	checkArchival    endpointCheckArchival

	// Latency tracking field - using exponential moving average
	// Stored directly in milliseconds
	averageLatencyMs float64
}

// Alpha defines the weight of the most recent observation in the exponential moving average.
// A higher value gives more weight to recent observations (faster adaptation to change).
// Common values are between 0.1 (more stable) and 0.3 (more responsive to recent changes).
const alphaLatency = 0.2

// recordLatency updates the exponential moving average of latency for this endpoint
// latencyNanos should be provided in nanoseconds (as would come from time.Duration.Nanoseconds())
func (e *endpoint) recordLatency(latencyNanos int64) {
	// Step 1: Convert from nanoseconds to milliseconds
	// 1 millisecond = 1,000,000 nanoseconds
	latencyMs := float64(latencyNanos) / 1_000_000.0

	// Step 2: Handle the first observation specially
	if e.averageLatencyMs == 0 {
		// For the first reading, simply use the value directly
		e.averageLatencyMs = latencyMs
		return
	}

	// Step 3: Apply the exponential moving average formula
	// Formula: newAvg = (1-α) * oldAvg + α * newValue
	// 	- α (alpha) controls how much weight is given to recent observations
	// 	- Higher α values (closer to 1) = faster adaptation to changes
	// 	- Lower α values (closer to 0) = more stable average over time
	oldAverage := e.averageLatencyMs
	newSample := latencyMs
	newAverage := (1-alphaLatency)*oldAverage + alphaLatency*newSample

	// Step 4: Store the updated average
	e.averageLatencyMs = newAverage
}
