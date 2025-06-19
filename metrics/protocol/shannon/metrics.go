// Package shannon provides functionality for exporting Shannon protocol metrics to Prometheus.
package shannon

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/buildwithgrove/path/observation/protocol"
)

const (
	// The POSIX process that emits metrics
	pathProcess = "path"

	// The list of metrics being tracked for Shannon protocol
	relaysTotalMetric       = "shannon_relays_total"
	relaysErrorsTotalMetric = "shannon_relay_errors_total"
	sanctionsByDomainMetric = "shannon_sanctions_by_domain"
)

func init() {
	prometheus.MustRegister(relaysTotal)
	prometheus.MustRegister(relaysErrorsTotal)
	prometheus.MustRegister(sanctionsByDomain)
}

var (
	// relaysTotal tracks the total Shannon relay requests processed.
	// Labels:
	//   - service_id: Target service identifier (i.e. chain id in Shannon)
	//   - success: Whether the relay was successful (true if at least one endpoint had no error)
	//   - error_type: type of error encountered processing the request
	//
	// Exemplars:
	//   - endpoint_url: URL of the endpoint (from the last entry in observations list)
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
			Help:      "Total number of relays processed by Shannon protocol instance(s)",
		},
		[]string{"service_id", "success", "error_type"},
	)

	// relaysErrorsTotal tracks relay errors by error type.
	// Labels:
	//   - service_id: Target service identifier
	//   - error_type: Type of error encountered (connection, timeout, etc.)
	//   - sanction_type: Type of sanction recommended for this error (if any)
	//
	// Exemplars:
	//   - endpoint_url: URL of the endpoint (from the last entry in observations list)
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

	// sanctionsByDomain tracks sanctions applied by domain.
	// Labels:
	//   - service_id: Target service identifier
	//   - endpoint_domain: Effective TLD+1 domain extracted from endpoint URL
	//   - sanction_type: Type of sanction (session, permanent)
	//   - sanction_reason: The endpoint error type that caused the sanction
	//
	// This counter is incremented each time a sanction is applied to an endpoint.
	// Provides insight into sanction patterns by domain without high-cardinality supplier addresses.
	// Use Grafana time series functions (rate, increase) to analyze sanction trends.
	//
	// Use to analyze:
	//   - Sanction rate by endpoint domain and service
	//   - Endpoint domain-level reliability trends
	//   - Provider performance analysis over time
	//   - Root cause analysis of sanctions by error type
	sanctionsByDomain = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      sanctionsByDomainMetric,
			Help:      "Total sanctions by service, endpoint domain (TLD+1), sanction type and reason",
		},
		[]string{"service_id", "endpoint_domain", "sanction_type", "sanction_reason"},
	)
)

// PublishMetrics exports all Shannon-related Prometheus metrics using observations
// reported by the Shannon protocol.
func PublishMetrics(
	logger polylog.Logger,
	observations *protocol.ShannonObservationsList,
) {

	shannonObservations := observations.GetObservations()
	if len(shannonObservations) == 0 {
		logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Unable to publish Shannon metrics: received nil observations.")
		return
	}

	// Process each observation for metrics
	for _, observationSet := range shannonObservations {
		// Record the relay total with success/failure status
		recordRelayTotal(logger, observationSet)

		// Process each endpoint observation for errors and sanctions
		processEndpointErrors(observationSet.GetServiceId(), observationSet.GetEndpointObservations())

		// Process sanctions by domain
		processSanctionsByDomain(logger, observationSet.GetServiceId(), observationSet.GetEndpointObservations())
	}
}

// recordRelayTotal tracks relay counts with exemplars for high-cardinality data.
func recordRelayTotal(
	logger polylog.Logger,
	observations *protocol.ShannonRequestObservations,
) {
	hydratedLogger := logger.With("method", "recordRelaysTotal")

	serviceID := observations.GetServiceId()
	// Relay request failed before reaching out to any endpoints.
	// e.g. there were no available endpoints.
	// Skip processing endpoint observations.
	if requestHasErr, requestErrorType := extractRequestError(observations); requestHasErr {
		relaysTotal.With(
			prometheus.Labels{
				"service_id": serviceID,
				"success":    "false",
				"error_type": requestErrorType,
			},
		).Inc()

		// Request has an error: no endpoint observations to process.
		return
	}

	endpointObservations := observations.GetEndpointObservations()
	// Skip if there are no endpoint observations
	// This happens if endpoint selection logic failed to select an endpoint from the available endpoints list.
	if len(endpointObservations) == 0 {
		hydratedLogger.Info().Msg("Request has no errors and no endpoint observations: endpoint selection has failed.")
		return
	}

	// Get the last observation for endpoint address and session height
	lastObs := endpointObservations[len(endpointObservations)-1]

	// Extract high-cardinality values for exemplars
	endpointURL := lastObs.GetEndpointUrl()

	// Create exemplar with high-cardinality data
	// Truncate to 128 runes (Prometheus exemplar limit)
	// See `ExemplarMaxRunes` below:
	// https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#pkg-constants
	exLabels := prometheus.Labels{
		"endpoint_url": endpointURL[:min(len(endpointURL), 128)],
	}

	// Determine if any of the observations were successful.
	success := isAnyObservationSuccessful(endpointObservations)

	// Increment the relay total counter with exemplars
	relaysTotal.With(
		prometheus.Labels{
			"service_id": serviceID,
			"success":    fmt.Sprintf("%t", success),
			"error_type": "",
		},
	// This dynamic type cast is safe:
	// https://pkg.go.dev/github.com/prometheus/client_golang@v1.22.0/prometheus#NewCounter
	).(prometheus.ExemplarAdder).AddWithExemplar(float64(1), exLabels)
}

// extractRequestError  extracts from the observations the stauts (success/failure) and the first encountered error, if any.
// Returns:
// - false, "" if the relay was successful.
// - true, error_type if the relay failed.
func extractRequestError(observations *protocol.ShannonRequestObservations) (bool, string) {
	requestErr := observations.GetRequestError()
	// No request errors.
	if requestErr == nil {
		return false, ""
	}

	return true, requestErr.GetErrorType().String()
}

// isAnyObservationSuccessful returns true if any endpoint observation indicates a success.
func isAnyObservationSuccessful(observations []*protocol.ShannonEndpointObservation) bool {
	for _, obs := range observations {
		if obs.GetErrorType() == protocol.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_UNSPECIFIED {
			return true
		}
	}
	return false
}

// processEndpointErrors records error metrics with exemplars for high-cardinality data
func processEndpointErrors(serviceID string, observations []*protocol.ShannonEndpointObservation) {
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
		endpointURL := endpointObs.GetEndpointUrl()

		// Create exemplar with high-cardinality data
		// Truncate to 128 runes (Prometheus exemplar limit)
		// See `ExemplarMaxRunes` below:
		// https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#pkg-constants
		exLabels := prometheus.Labels{
			"endpoint_url": endpointURL[:min(len(endpointURL), 128)],
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

// processSanctionsByDomain records sanction events by domain using a counter.
// This function tracks sanctions at the domain level to provide operational visibility
// without the high cardinality of individual supplier addresses.
// Use Grafana time series functions to analyze trends and rates.
func processSanctionsByDomain(
	logger polylog.Logger,
	serviceID string,
	observations []*protocol.ShannonEndpointObservation,
) {
	for _, endpointObs := range observations {
		// Skip if there's no recommended sanction
		if endpointObs.RecommendedSanction == nil {
			continue
		}

		// Extract effective TLD+1 from endpoint URL
		// This function handles edge cases like IP addresses, localhost, invalid URLs
		endpointTLDPlusOne, err := extractEffectiveTLDPlusOne(endpointObs.GetEndpointUrl())
		// error extracting TLD+1, skip.
		if err != nil {
			logger.With(
				"endpoint_url", endpointObs.GetEndpointUrl(),
			).
				ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).
				Err(err).Msg("SHOULD NEVER HAPPEN: Could not extract domain from Shannon endpoint URL")

			continue
		}

		// Extract the sanction reason from the endpoint error type
		var sanctionReason string
		if endpointObs.ErrorType != nil {
			sanctionReason = endpointObs.GetErrorType().String()
		}

		// Increment the sanctions counter for this domain
		sanctionsByDomain.With(
			prometheus.Labels{
				"service_id":      serviceID,
				"endpoint_domain": endpointTLDPlusOne,
				"sanction_type":   endpointObs.GetRecommendedSanction().String(),
				"sanction_reason": sanctionReason,
			},
		).Inc()
	}
}
