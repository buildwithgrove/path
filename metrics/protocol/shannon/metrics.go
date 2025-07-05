// Package shannon provides functionality for exporting Shannon protocol metrics to Prometheus.
package shannon

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/prometheus/client_golang/prometheus"

	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	protocol "github.com/buildwithgrove/path/protocol"
)

var (
	defaultBuckets = []float64{
		// Sub-50ms (cache hits, internal optimization, fast responses, potential internal errors, etc.)
		0.01, 0.025, 0.05,
		// Primary range: 50ms to 1s (majority of traffic, normal responses, etc...)
		0.075, 0.1, 0.15, 0.2, 0.25, 0.3, 0.35, 0.4, 0.45, 0.5, 0.55, 0.6, 0.7, 0.8, 0.9, 1.0,
		// Long tail: > 1s (slow queries, rollovers, cold state, failed, etc.)
		1.5, 2.0, 3.0, 5.0, 10.0, 30.0,
	}
)

const (
	// The POSIX process that emits metrics
	pathProcess = "path"

	// The list of metrics being tracked for Shannon protocol
	relaysTotalMetric       = "shannon_relays_total"
	relaysErrorsTotalMetric = "shannon_relay_errors_total"
	sanctionsByDomainMetric = "shannon_sanctions_by_domain"
	endpointLatencyMetric   = "shannon_endpoint_latency_seconds"

	sessionTransitionMetric        = "shannon_session_transitions_total"
	sessionCacheOperationsMetric   = "shannon_session_cache_operations_total"
	sessionGracePeriodMetric       = "shannon_session_grace_period_usage_total"
	sessionOperationDurationMetric = "shannon_session_operation_duration_seconds"
	relayLatencyMetric             = "shannon_relay_latency_seconds"
	backendServiceLatencyMetric    = "shannon_backend_service_latency_seconds"
	requestSetupLatencyMetric      = "shannon_request_setup_latency_seconds"
)

func init() {
	prometheus.MustRegister(relaysTotal)
	prometheus.MustRegister(relaysErrorsTotal)
	prometheus.MustRegister(sanctionsByDomain)
	prometheus.MustRegister(endpointLatency)
	prometheus.MustRegister(sessionExtendedUsage)
	prometheus.MustRegister(sessionOperationDuration)
	prometheus.MustRegister(relayLatency)
	prometheus.MustRegister(backendServiceLatency)
	prometheus.MustRegister(requestSetupLatency)
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

	// endpointLatency tracks the latency distribution of endpoint responses.
	// Labels:
	//   - service_id: Target service identifier
	//   - endpoint_domain: Effective TLD+1 domain extracted from endpoint URL
	//   - success: Whether the request was successful (true if at least one endpoint had no error)
	//
	// This histogram measures the time between sending a request to an endpoint
	// and receiving its response. Only recorded for endpoints that actually respond
	// (excludes timeouts where no response timestamp is available).
	// A request with error not related to an endpoint will not have an endpoint query time set.
	//
	// Use to analyze:
	//   - Response time percentiles by service and domain
	//   - Performance comparison across different endpoint domains
	//   - Latency trends over time
	//   - Impact of errors on response times
	endpointLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: pathProcess,
			Name:      endpointLatencyMetric,
			Help:      "Histogram of endpoint response latencies in seconds",
			Buckets:   defaultBuckets,
		},
		[]string{"service_id", "endpoint_domain", "success"},
	)

	// sessionExtendedUsage tracks extended session usage patterns.
	// Labels:
	//   - service_id: Target service identifier
	//   - usage_type: Type of extended session usage (active, extended)
	//   - session_decision: Which session was selected (current, previous)
	//
	// Use to analyze:
	//   - Extended session effectiveness
	//   - Session selection patterns during transitions
	//   - Impact of grace period scaling factor
	sessionExtendedUsage = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      sessionExtendedUsageMetric,
			Help:      "Total extended session usage patterns by service and decision type",
		},
		[]string{"service_id", "usage_type", "session_decision"},
	)

	// sessionOperationDuration tracks latency of session-related operations.
	//
	// Labels:
	//   - service_id: Target service identifier
	//   - operation: Type of operation (get_session, get_session_with_grace, get_active_sessions, cache_fetch, etc.)
	//   - cache_result: Result of cache operation (hit, miss, fetch_success, fetch_error)
	//   - grace_period_active: Whether grace period logic was applied (true/false)
	//
	// Use to analyze:
	//   - Latency distribution during session rollovers
	//   - Cache vs network fetch performance
	//   - Grace period overhead
	//   - P95/P99 latency trends during rollover windows
	sessionOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: pathProcess,
			Name:      sessionOperationDurationMetric,
			Help:      "Duration of session-related operations in seconds",
			Buckets:   defaultBuckets,
		},
		[]string{"service_id", "operation", "cache_result", "grace_period_active"},
	)

	// relayLatency tracks end-to-end relay request latency.
	//
	// Labels:
	//   - service_id: Target service identifier
	//   - session_state: State of session during request (current, grace_period, rollover)
	//   - cache_effectiveness: Overall cache performance (all_hits, some_misses, all_misses)
	//
	// Use to analyze:
	//   - Impact of session rollovers on end-user latency
	//   - Correlation between cache misses and request latency
	//   - Session transition impact on user experience
	relayLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: pathProcess,
			Name:      relayLatencyMetric,
			Help:      "End-to-end relay request latency in seconds",
			Buckets:   defaultBuckets,
		},
		[]string{"service_id", "session_state", "cache_effectiveness"},
	)

	// backendServiceLatency tracks the time spent waiting for backend service responses.
	//
	// Labels:
	//   - service_id: Target service identifier
	//   - endpoint_domain: Backend service domain (TLD+1 for cardinality control)
	//   - http_status: HTTP response status (2xx, 4xx, 5xx, timeout)
	//   - request_size_bucket: Request size category (small, medium, large)
	//
	// Use to analyze:
	//   - Pure backend service performance (excluding PATH overhead)
	//   - Backend service degradation patterns
	//   - Correlation between backend latency and total request latency
	//   - Impact of request size on backend response time
	backendServiceLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: pathProcess,
			Name:      backendServiceLatencyMetric,
			Help:      "Backend service response latency in seconds",
			Buckets:   defaultBuckets,
		},
		[]string{"service_id", "endpoint_domain", "http_status", "request_size_bucket"},
	)

	// requestSetupLatency tracks the time spent in PATH's request setup phase before sending the relay.
	// Labels:
	//   - service_id: Target service identifier
	//   - setup_stage: Which setup stage completed successfully (qos_context, protocol_context, complete)
	//   - cache_performance: Overall cache hit rate during setup (all_hits, some_misses, all_misses)
	//
	// Use to analyze:
	//   - Setup overhead vs backend service latency
	//   - Impact of session rollovers on request preparation time
	//   - Cache effectiveness during session transitions
	//   - Bottlenecks in QoS vs Protocol context setup
	requestSetupLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: pathProcess,
			Name:      requestSetupLatencyMetric,
			Help:      "Request setup latency before relay transmission in seconds",
			Buckets:   defaultBuckets,
		},
		[]string{"service_id", "setup_stage", "cache_performance"},
	)
)

// PublishMetrics exports all Shannon-related Prometheus metrics using observations
// reported by the Shannon protocol.
func PublishMetrics(
	logger polylog.Logger,
	observations *protocolobservations.ShannonObservationsList,
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

		// Process endpoint latency metrics
		processEndpointLatency(logger, observationSet.GetServiceId(), observationSet.GetEndpointObservations())
	}
}

// recordRelayTotal tracks relay counts with exemplars for high-cardinality data.
func recordRelayTotal(
	logger polylog.Logger,
	observations *protocolobservations.ShannonRequestObservations,
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
		hydratedLogger.Warn().Msg("Request has no errors and no endpoint observations: endpoint selection has failed.")
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
func extractRequestError(observations *protocolobservations.ShannonRequestObservations) (bool, string) {
	requestErr := observations.GetRequestError()
	// No request errors.
	if requestErr == nil {
		return false, ""
	}

	return true, requestErr.GetErrorType().String()
}

// isAnyObservationSuccessful returns true if any endpoint observation indicates a success.
func isAnyObservationSuccessful(observations []*protocolobservations.ShannonEndpointObservation) bool {
	for _, obs := range observations {
		if obs.GetErrorType() == protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_UNSPECIFIED {
			return true
		}
	}
	return false
}

// processEndpointErrors records error metrics with exemplars for high-cardinality data
func processEndpointErrors(serviceID string, observations []*protocolobservations.ShannonEndpointObservation) {
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
	observations []*protocolobservations.ShannonEndpointObservation,
) {
	for _, endpointObs := range observations {
		// Skip if there's no recommended sanction
		if endpointObs.RecommendedSanction == nil {
			continue
		}

		// Extract effective TLD+1 from endpoint URL
		// This function handles edge cases like IP addresses, localhost, invalid URLs
		endpointTLDPlusOne, err := ExtractEffectiveTLDPlusOne(endpointObs.GetEndpointUrl())
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

// processEndpointLatency records endpoint response latency metrics.
// Only records latency for endpoints that actually responded (have both query and response timestamps).
// A request with error not related to an endpoint will not have an endpoint query time set.
func processEndpointLatency(
	logger polylog.Logger,
	serviceID string,
	observations []*protocolobservations.ShannonEndpointObservation,
) {
	// Calculate overall success status for the request
	success := isAnyObservationSuccessful(observations)

	for _, endpointObs := range observations {
		// Skip if we don't have both timestamps (e.g., timeouts)
		// These will be caught by other metrics indicating endpoint errors.
		queryTime := endpointObs.GetEndpointQueryTimestamp()
		responseTime := endpointObs.GetEndpointResponseTimestamp()

		if queryTime == nil || responseTime == nil {
			continue
		}

		// Calculate latency in seconds
		queryTimestamp := queryTime.AsTime()
		responseTimestamp := responseTime.AsTime()
		latencySeconds := responseTimestamp.Sub(queryTimestamp).Seconds()

		// Skip negative latencies (invalid timestamps)
		if latencySeconds < 0 {
			logger.With(
				"endpoint_url", endpointObs.GetEndpointUrl(),
				"latency_seconds", latencySeconds,
			).ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).
				Msg("SHOULD RARELY HAPPEN: Negative latency detected, skipping metric")
			continue
		}

		// Extract effective TLD+1 from endpoint URL
		endpointTLDPlusOne, err := ExtractEffectiveTLDPlusOne(endpointObs.GetEndpointUrl())
		if err != nil {
			logger.With(
				"endpoint_url", endpointObs.GetEndpointUrl(),
			).ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).
				Err(err).Msg("SHOULD NEVER HAPPEN: Could not extract domain from Shannon endpoint URL for latency metric")
			continue
		}

		// Record latency
		endpointLatency.With(
			prometheus.Labels{
				"service_id":      serviceID,
				"endpoint_domain": endpointTLDPlusOne,
				"success":         fmt.Sprintf("%t", success),
			},
		).Observe(latencySeconds)
	}
}

// RecordSessionTransition records a session transition event with cache performance data.
func RecordSessionTransition(
	serviceID protocol.ServiceID,
	appAddr, transitionType string,
	cacheHit bool,
) {
	// Truncate app address to first 8 characters to reduce cardinality while maintaining uniqueness
	appAddrPrefix := appAddr
	if len(appAddr) > 8 {
		appAddrPrefix = appAddr[:8]
	}

	sessionTransitions.With(prometheus.Labels{
		"service_id":      string(serviceID),
		"app_addr_prefix": appAddrPrefix,
		"transition_type": transitionType,
		"cache_hit":       fmt.Sprintf("%t", cacheHit),
	}).Inc()
}

// RecordSessionGracePeriodUsage records grace period usage patterns.
func RecordSessionGracePeriodUsage(
	serviceID protocol.ServiceID,
	usageType, sessionDecision string,
) {
	sessionExtendedUsage.With(prometheus.Labels{
		"service_id":       string(serviceID),
		"usage_type":       usageType,
		"session_decision": sessionDecision,
	}).Inc()
}

// RecordSessionOperationDuration records the duration of session-related operations.
func RecordSessionOperationDuration(
	serviceID protocol.ServiceID,
	operation, cacheResult string,
	isExtendedSession bool,
	duration float64,
) {
	sessionOperationDuration.With(prometheus.Labels{
		"service_id":          string(serviceID),
		"operation":           operation,
		"cache_result":        cacheResult,
		"is_extended_session": fmt.Sprintf("%t", isExtendedSession),
	}).Observe(duration)
}

// RecordRelayLatency records end-to-end relay request latency.
func RecordRelayLatency(
	serviceID protocol.ServiceID,
	sessionState, cacheEffectiveness string,
	duration float64,
) {
	relayLatency.With(prometheus.Labels{
		"service_id":          string(serviceID),
		"session_state":       sessionState,
		"cache_effectiveness": cacheEffectiveness,
	}).Observe(duration)
}

// RecordBackendServiceLatency records backend service response latency.
func RecordBackendServiceLatency(
	serviceID protocol.ServiceID,
	endpointDomain, httpStatus, requestSizeBucket string,
	duration float64,
) {
	backendServiceLatency.With(prometheus.Labels{
		"service_id":          string(serviceID),
		"endpoint_domain":     endpointDomain,
		"http_status":         httpStatus,
		"request_size_bucket": requestSizeBucket,
	}).Observe(duration)
}

// RecordRequestSetupLatency records the time spent in PATH's request setup phase.
func RecordRequestSetupLatency(
	serviceID protocol.ServiceID,
	setupStage, cachePerformance string,
	duration float64,
) {
	requestSetupLatency.With(prometheus.Labels{
		"service_id":        string(serviceID),
		"setup_stage":       setupStage,
		"cache_performance": cachePerformance,
	}).Observe(duration)
}
