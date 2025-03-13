// Package morse provides functionality for exporting Morse protocol metrics to Prometheus.
package morse

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/buildwithgrove/path/observation/protocol"
)

const (
	// The POSIX process that emits metrics
	pathProcess = "path"

	// The list of metrics being tracked for Morse protocol
	relaysTotalMetric       = "morse_relays_total"
	relaysErrorsTotalMetric = "morse_relay_errors_total"
)

func init() {
	prometheus.MustRegister(relaysTotal)
	prometheus.MustRegister(relaysErrorsTotal)
}

var (
	// relaysTotal tracks the total Morse relay requests processed.
	// Labels:
	//   - service_id: Target service identifier (i.e. chain id in Morse)
	//   - app_address: Application address that signed the relay
	//   - success: Whether the relay was successful (true if at least one endpoint had no error)
	//   - endpoint_addr: Address of the endpoint (from the last entry in observations list)
	//     NOTE: Currently only using a single endpoint (the last one) for this metric.
	//     TODO_FUTURE: This needs to be revisited when retry mechanism is implemented at protocol level.
	//   - session_height: Height of the session (from the last entry in observations list)
	//     NOTE: Using the same endpoint entry as endpoint_addr above.
	//     TODO_FUTURE: This needs to be revisited when retry mechanism is implemented at protocol level.
	//
	// Use to analyze:
	//   - Request volume by service
	//   - Request distribution across applications
	//   - Success rates by service and application
	//   - Endpoint participation
	//   - Session height distribution
	relaysTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      relaysTotalMetric,
			Help:      "Total number of relays processed by Morse protocol instance(s)",
		},
		[]string{"service_id", "app_address", "success", "endpoint_addr", "session_height"},
	)

	// relaysErrorsTotal tracks relay errors by error type.
	// Labels:
	//   - service_id: Target service identifier
	//   - app_address: Application address that signed the relay
	//   - error_type: Type of error encountered (connection, timeout, etc.)
	//   - endpoint_addr: Address of the endpoint that returned an error
	//   - session_height: Height of the session when the error occurred
	//   - sanction_type: Type of sanction recommended for this error (if any)
	//
	// Use to analyze:
	//   - Error patterns by service and endpoint
	//   - Error patterns by application
	//   - Session height correlation with errors
	//   - Sanction distribution for different error types
	relaysErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      relaysErrorsTotalMetric,
			Help:      "Total relay errors by type, service, endpoint, and session height",
		},
		[]string{"service_id", "app_address", "error_type", "endpoint_addr", "session_height", "sanction_type"},
	)
)

// PublishMetrics exports all Morse-related Prometheus metrics using observations
// reported by the Morse protocol.
func PublishMetrics(observations *protocol.MorseObservationsList) {
	if observations == nil {
		return
	}

	morseObservations := observations.GetObservations()
	if len(morseObservations) == 0 {
		return
	}

	// Process each observation for metrics
	for _, observationSet := range morseObservations {
		serviceID := observationSet.GetServiceId()
		endpointObservations := observationSet.GetEndpointObservations()

		// Skip if there are no endpoint observations
		if len(endpointObservations) == 0 {
			continue
		}

		// Record the relay total with success/failure status
		recordRelayTotal(serviceID, endpointObservations)

		// Process each endpoint observation for errors
		processEndpointErrors(serviceID, endpointObservations)
	}
}

// recordRelayTotal records the total relays metric collected.
// See relaysTotal above for details on additional metadata and use-cases.
func recordRelayTotal(serviceID string, observations []*protocol.MorseEndpointObservation) {
	// Skip if there are no observations
	if len(observations) == 0 {
		return
	}

	// Determine if any of the observations were successful (no error)
	success := isAnyObservationSuccessful(observations)

	// Get the last observation for endpoint address and session height
	// TODO_TECHDEBT(@adshmh): Currently using the last entry in observations list for endpoint_addr and session_height.
	// This is a simplification that should be revisited when implementing retry mechanisms.
	lastObs := observations[len(observations)-1]

	// Extract values for labels
	appAddress := lastObs.GetAppAddress()
	endpointAddr := lastObs.GetEndpointAddr()
	sessionHeight := fmt.Sprintf("%d", lastObs.GetSessionHeight())

	// Increment the relay total counter
	relaysTotal.With(
		prometheus.Labels{
			"service_id":     serviceID,
			"app_address":    appAddress,
			"success":        fmt.Sprintf("%t", success),
			"endpoint_addr":  endpointAddr,
			"session_height": sessionHeight,
		},
	).Inc()
}

// isAnyObservationSuccessful checks if at least one of the observations has no error
// (indicating success)
func isAnyObservationSuccessful(observations []*protocol.MorseEndpointObservation) bool {
	for _, obs := range observations {
		if obs.GetErrorType() != protocol.MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_UNSPECIFIED {
			return true
		}
	}
	return false
}

// processEndpointErrors iterates through endpoint observations to record error metrics
func processEndpointErrors(serviceID string, observations []*protocol.MorseEndpointObservation) {
	for _, endpointObs := range observations {
		// Skip if there's no error
		if endpointObs.ErrorType == nil {
			continue
		}

		// Extract base labels
		appAddress := endpointObs.GetAppAddress()
		endpointAddr := endpointObs.GetEndpointAddr()
		sessionHeight := fmt.Sprintf("%d", endpointObs.GetSessionHeight())
		errorType := endpointObs.GetErrorType().String()

		// Extract sanction type (if any)
		sanctionType := "none"
		if endpointObs.RecommendedSanction != nil {
			sanctionType = endpointObs.GetRecommendedSanction().String()
		}

		// Record relay error
		relaysErrorsTotal.With(
			prometheus.Labels{
				"service_id":     serviceID,
				"app_address":    appAddress,
				"error_type":     errorType,
				"endpoint_addr":  endpointAddr,
				"session_height": sessionHeight,
				"sanction_type":  sanctionType,
			},
		).Inc()
	}
}
