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
	relaysTotalMetric              = "shannon_relays_total"
	relaysErrorsTotalMetric        = "shannon_relay_errors_total"
	sanctionsByDomainMetric        = "shannon_sanctions_by_domain"
	sessionTransitionMetric        = "shannon_session_transitions_total"
	sessionCacheOperationsMetric   = "shannon_session_cache_operations_total"
	sessionGracePeriodMetric       = "shannon_session_grace_period_usage_total"
	sessionOperationDurationMetric = "shannon_session_operation_duration_seconds"
	relayLatencyMetric             = "shannon_relay_latency_seconds"
	backendServiceLatencyMetric    = "shannon_backend_service_latency_seconds"
)

func init() {
	prometheus.MustRegister(relaysTotal)
	prometheus.MustRegister(relaysErrorsTotal)
	prometheus.MustRegister(sanctionsByDomain)
	prometheus.MustRegister(sessionTransitions)
	prometheus.MustRegister(sessionCacheOperations)
	prometheus.MustRegister(sessionGracePeriodUsage)
	prometheus.MustRegister(sessionOperationDuration)
	prometheus.MustRegister(relayLatency)
	prometheus.MustRegister(backendServiceLatency)
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

	// sessionTransitions tracks session transitions and rollover events.
	// Labels:
	//   - service_id: Target service identifier
	//   - app_addr: Application address (truncated for cardinality)
	//   - transition_type: Type of transition (new_session, rollover, grace_period)
	//   - cache_hit: Whether the session was found in cache
	//
	// Use to analyze:
	//   - Session rollover frequency patterns
	//   - Cache effectiveness during transitions
	//   - Identify services with high session turnover
	sessionTransitions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      sessionTransitionMetric,
			Help:      "Total session transitions by service, transition type and cache performance",
		},
		[]string{"service_id", "app_addr_prefix", "transition_type", "cache_hit"},
	)

	// sessionCacheOperations tracks cache operations for session-related data.
	// Labels:
	//   - service_id: Target service identifier
	//   - operation: Type of operation (get, fetch, evict, refresh)
	//   - cache_type: Type of cache (session, shared_params, block_height)
	//   - result: Result of operation (hit, miss, error)
	//
	// Use to analyze:
	//   - Cache hit rates during session rollovers
	//   - Cache refresh patterns
	//   - Performance bottlenecks in cache operations
	sessionCacheOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      sessionCacheOperationsMetric,
			Help:      "Total cache operations for session-related data",
		},
		[]string{"service_id", "operation", "cache_type", "result"},
	)

	// sessionGracePeriodUsage tracks grace period usage patterns.
	// Labels:
	//   - service_id: Target service identifier
	//   - usage_type: Type of grace period usage (within_grace, outside_grace, scaled_grace)
	//   - session_decision: Which session was selected (current, previous)
	//
	// Use to analyze:
	//   - Grace period effectiveness
	//   - Session selection patterns during transitions
	//   - Impact of grace period scaling factor
	sessionGracePeriodUsage = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      sessionGracePeriodMetric,
			Help:      "Total grace period usage patterns by service and decision type",
		},
		[]string{"service_id", "usage_type", "session_decision"},
	)

	// sessionOperationDuration tracks latency of session-related operations.
	// Labels:
	//   - service_id: Target service identifier
	//   - operation: Type of operation (get_session, get_session_with_grace, get_active_sessions, cache_fetch, etc.)
	//   - cache_result: Result of cache operation (hit, miss, fetch_success, fetch_error)
	//   - grace_period_active: Whether grace period logic was applied (true/false)
	//
	// Buckets optimized for session operations (1ms to 30s):
	//   - Cache hits: < 1ms
	//   - Cache misses with network fetch: 10ms - 1s
	//   - Session rollover scenarios: 100ms - 5s
	//   - Network issues/timeouts: 5s - 30s
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
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0},
		},
		[]string{"service_id", "operation", "cache_result", "grace_period_active"},
	)

	// relayLatency tracks end-to-end relay request latency.
	// Labels:
	//   - service_id: Target service identifier
	//   - session_state: State of session during request (current, grace_period, rollover)
	//   - cache_effectiveness: Overall cache performance (all_hits, some_misses, all_misses)
	//
	// Buckets optimized for relay latency (1ms to 60s):
	//   - Fast cached responses: < 10ms
	//   - Normal responses: 10ms - 1s
	//   - Session rollover impact: 100ms - 10s
	//   - Slow/failed responses: 10s - 60s
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
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0},
		},
		[]string{"service_id", "session_state", "cache_effectiveness"},
	)

	// backendServiceLatency tracks the time spent waiting for backend service responses.
	// Labels:
	//   - service_id: Target service identifier
	//   - endpoint_domain: Backend service domain (TLD+1 for cardinality control)
	//   - http_status: HTTP response status (2xx, 4xx, 5xx, timeout)
	//   - request_size_bucket: Request size category (small, medium, large)
	//
	// Buckets optimized for backend service response times (1ms to 30s):
	//   - Fast responses: < 50ms (cache hits, simple queries)
	//   - Normal responses: 50ms - 2s (typical blockchain RPC calls)
	//   - Slow responses: 2s - 10s (complex queries, archival data)
	//   - Timeout/error responses: 10s - 30s (network issues, overloaded backends)
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
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0},
		},
		[]string{"service_id", "endpoint_domain", "http_status", "request_size_bucket"},
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

// RecordSessionTransition records a session transition event with cache performance data.
func RecordSessionTransition(serviceID, appAddr, transitionType string, cacheHit bool) {
	// Truncate app address to first 8 characters to reduce cardinality while maintaining uniqueness
	appAddrPrefix := appAddr
	if len(appAddr) > 8 {
		appAddrPrefix = appAddr[:8]
	}

	sessionTransitions.With(prometheus.Labels{
		"service_id":      serviceID,
		"app_addr_prefix": appAddrPrefix,
		"transition_type": transitionType,
		"cache_hit":       fmt.Sprintf("%t", cacheHit),
	}).Inc()
}

// RecordSessionCacheOperation records cache operations for session-related data.
func RecordSessionCacheOperation(serviceID, operation, cacheType, result string) {
	sessionCacheOperations.With(prometheus.Labels{
		"service_id": serviceID,
		"operation":  operation,
		"cache_type": cacheType,
		"result":     result,
	}).Inc()
}

// RecordSessionGracePeriodUsage records grace period usage patterns.
func RecordSessionGracePeriodUsage(serviceID, usageType, sessionDecision string) {
	sessionGracePeriodUsage.With(prometheus.Labels{
		"service_id":       serviceID,
		"usage_type":       usageType,
		"session_decision": sessionDecision,
	}).Inc()
}

// RecordSessionOperationDuration records the duration of session-related operations.
func RecordSessionOperationDuration(serviceID, operation, cacheResult string, gracePeriodActive bool, duration float64) {
	sessionOperationDuration.With(prometheus.Labels{
		"service_id":          serviceID,
		"operation":           operation,
		"cache_result":        cacheResult,
		"grace_period_active": fmt.Sprintf("%t", gracePeriodActive),
	}).Observe(duration)
}

// RecordRelayLatency records end-to-end relay request latency.
func RecordRelayLatency(serviceID, sessionState, cacheEffectiveness string, duration float64) {
	relayLatency.With(prometheus.Labels{
		"service_id":          serviceID,
		"session_state":       sessionState,
		"cache_effectiveness": cacheEffectiveness,
	}).Observe(duration)
}

// RecordBackendServiceLatency records backend service response latency.
func RecordBackendServiceLatency(serviceID, endpointDomain, httpStatus, requestSizeBucket string, duration float64) {
	backendServiceLatency.With(prometheus.Labels{
		"service_id":          serviceID,
		"endpoint_domain":     endpointDomain,
		"http_status":         httpStatus,
		"request_size_bucket": requestSizeBucket,
	}).Observe(duration)
}
