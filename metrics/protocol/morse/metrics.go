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
	//   - success: Whether the relay was successful (true if at least one endpoint had no error)
	//
	// Exemplars:
	//   - app_address: Application address that signed the relay
	//   - endpoint_addr: Address of the endpoint (from the last entry in observations list)
	//   - session_height: Height of the session (from the last entry in observations list)
	//
	// Low-cardinality labels are used for core metrics while high-cardinality data is
	// moved to exemplars to reduce Prometheus storage and query overhead while still
	// preserving detailed information for troubleshooting.
	//
	// Use to analyze:
	//   - Request volume by service
	//   - Success rates by service
	//   - Detailed endpoint and app data available via exemplars when needed
	relaysTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      relaysTotalMetric,
			Help:      "Total number of relays processed by Morse protocol instance(s)",
		},
		[]string{"service_id", "success"},
	)

	// relaysErrorsTotal tracks relay errors by error type.
	// Labels:
	//   - service_id: Target service identifier
	//   - error_type: Type of error encountered (connection, timeout, etc.)
	//   - sanction_type: Type of sanction recommended for this error (if any)
	//
	// Exemplars:
	//   - app_address: Application address that signed the relay
	//   - endpoint_addr: Address of the endpoint that returned an error
	//   - session_height: Height of the session when the error occurred
	//
	// Use to analyze:
	//   - Error patterns by service
	//   - Sanction distribution for different error types
	//   - Detailed endpoint and app data available via exemplars when needed
	relaysErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      relaysErrorsTotalMetric,
			Help:      "Total relay errors by type, service and sanction type",
		},
		[]string{"service_id", "error_type", "sanction_type"},
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

// recordRelayTotal tracks relay counts with exemplars for high-cardinality data.
func recordRelayTotal(serviceID string, observations []*protocol.MorseEndpointObservation) {
	// Skip if there are no observations
	if len(observations) == 0 {
		return
	}

	// Determine if any of the observations were successful (no error)
	success := isAnyObservationSuccessful(observations)

	// Get the last observation for endpoint address and session height
	lastObs := observations[len(observations)-1]

	// Extract values for core labels (low cardinality)
	successLabel := fmt.Sprintf("%t", success)

	// Extract high-cardinality values for exemplars
	appAddress := lastObs.GetAppAddress()
	endpointAddr := lastObs.GetEndpointAddr()
	sessionHeight := fmt.Sprintf("%d", lastObs.GetSessionHeight())

	// Create exemplar with high-cardinality data
	exLabels := prometheus.Labels{
		"app_address":    appAddress,
		"endpoint_addr":  endpointAddr,
		"session_height": sessionHeight,
	}

	// Increment the relay total counter with exemplars
	relaysTotal.With(
		prometheus.Labels{
			"service_id": serviceID,
			"success":    successLabel,
		},
	// This dynamic type cast is safe:
	// https://pkg.go.dev/github.com/prometheus/client_golang@v1.22.0/prometheus#NewCounter
	).(prometheus.ExemplarAdder).AddWithExemplar(float64(1), exLabels)
}

// isAnyObservationSuccessful returns true if any observation succeeded (no error)
func isAnyObservationSuccessful(observations []*protocol.MorseEndpointObservation) bool {
	for _, obs := range observations {
		if obs.GetErrorType() != protocol.MorseEndpointErrorType_MORSE_ENDPOINT_ERROR_UNSPECIFIED {
			return true
		}
	}
	return false
}

// processEndpointErrors records error metrics with exemplars for high-cardinality data
func processEndpointErrors(serviceID string, observations []*protocol.MorseEndpointObservation) {
	for _, endpointObs := range observations {
		// Skip if there's no error
		if endpointObs.ErrorType == nil {
			continue
		}

		// Extract low-cardinality labels
		errorType := endpointObs.GetErrorType().String()

		// Extract sanction type (if any)
		var sanctionType string
		if endpointObs.RecommendedSanction != nil {
			sanctionType = endpointObs.GetRecommendedSanction().String()
		}

		// Extract high-cardinality values for exemplars
		appAddress := endpointObs.GetAppAddress()
		endpointAddr := endpointObs.GetEndpointAddr()
		sessionHeight := fmt.Sprintf("%d", endpointObs.GetSessionHeight())

		// Create exemplar with high-cardinality data
		exLabels := prometheus.Labels{
			"app_address":    appAddress,
			"endpoint_addr":  endpointAddr,
			"session_height": sessionHeight,
		}

		// Record relay error with exemplars
		relaysErrorsTotal.With(
			prometheus.Labels{
				"service_id":    serviceID,
				"error_type":    errorType,
				"sanction_type": sanctionType,
			},
		// This dynamic type cast is safe:
		// https://pkg.go.dev/github.com/prometheus/client_golang@v1.22.0/prometheus#NewCounter
		).(prometheus.ExemplarAdder).AddWithExemplar(float64(1), exLabels)
	}
}
