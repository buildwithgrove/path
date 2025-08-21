// Package shannon provides functionality for exporting Shannon protocol metrics to Prometheus.
package shannon

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/prometheus/client_golang/prometheus"

	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
)

// TODO_TECHDEBT: Replace 'endpoint_domain' in the metrics to align with 'endpoint_url'
// used through the codebase or vice versa.
//
// TODO_METRICS: Add additional WebSocket-specific metrics
// - Message latency distribution (time between request and response for each message)
// - Connection duration histogram (time from connection establishment to termination)
// - Message size percentiles (distribution of message payload sizes)
// - Subscription event rates (frequency of subscription events per connection)

const (
	// The POSIX process that emits metrics
	pathProcess = "path"

	// HTTP relay metrics
	relaysTotalMetric       = "shannon_relays_total"
	relaysErrorsTotalMetric = "shannon_relay_errors_total"

	// WebSocket connection metrics
	websocketConnectionsTotalMetric = "shannon_websocket_connections_total"
	websocketConnectionErrorsMetric = "shannon_websocket_connection_errors_total"

	// WebSocket message metrics
	websocketMessagesTotalMetric = "shannon_websocket_messages_total"
	websocketMessageErrorsMetric = "shannon_websocket_message_errors_total"

	// Sanctions metrics (shared across HTTP and WebSocket)
	sanctionsByDomainMetric = "shannon_sanctions_by_domain"

	// Latency metrics (currently HTTP only)
	endpointLatencyMetric       = "shannon_endpoint_latency_seconds"
	relayMinerErrorsTotalMetric = "shannon_relay_miner_errors_total"
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

func init() {
	// HTTP relay metrics
	prometheus.MustRegister(relaysTotal)
	prometheus.MustRegister(relaysErrorsTotal)

	// WebSocket metrics
	prometheus.MustRegister(websocketConnectionsTotal)
	prometheus.MustRegister(websocketConnectionErrors)
	prometheus.MustRegister(websocketMessagesTotal)
	prometheus.MustRegister(websocketMessageErrors)

	// Sanctions metrics (shared across HTTP and WebSocket)
	prometheus.MustRegister(sanctionsByDomain)

	// Latency metrics
	prometheus.MustRegister(endpointLatency)
	prometheus.MustRegister(endpointResponseSize)
	prometheus.MustRegister(relayMinerErrorsTotal)
}

var (
	// relaysTotal tracks the total Shannon relay requests processed.
	// Labels:
	//   - service_id: Target service identifier (i.e. chain id in Shannon)
	//   - success: Whether the relay was successful (true if at least one endpoint had no error)
	//   - error_type: type of error encountered processing the request
	//   - used_fallback: Whether the request was served using a fallback endpoint.
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
	//   - Distribution of traffic between protocol and fallback endpoints.
	relaysTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      relaysTotalMetric,
			Help:      "Total number of relays processed by Shannon protocol instance(s)",
		},
		[]string{"service_id", "success", "error_type", "used_fallback"},
	)

	// relaysErrorsTotal tracks relay errors from Shannon protocol
	// Labels:
	//   - service_id: Target service identifier
	//   - endpoint_domain: Effective TLD+1 domain extracted from endpoint URL
	//   - error_type: Type of error encountered (based on trusted classification)
	//   - sanction_type: Type of sanction recommended (based on trusted classification)
	//
	// Exemplars:
	//   - endpoint_url: URL of the endpoint
	//
	// Use to analyze:
	//   - Shannon protocol errors by service and type
	//   - Sanctions recommended by the protocol
	//
	// TODO_TECHDEBT(@adshmh): Check whether merging SanctionsByDomain and relayErrorsTotal makes sense.
	relaysErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      relaysErrorsTotalMetric,
			Help:      "Total relay errors by service, endpoint domain, error type, and sanction type",
		},
		[]string{"service_id", "endpoint_domain", "error_type", "sanction_type"},
	)

	// websocketConnectionsTotal tracks the total WebSocket connection attempts processed.
	// Labels:
	//   - service_id: Target service identifier (i.e. chain id in Shannon)
	//   - success: Whether the connection was successful (true if no connection error)
	//   - error_type: type of error encountered during connection setup
	//   - used_fallback: Whether the connection used a fallback endpoint.
	//
	// Exemplars:
	//   - endpoint_url: URL of the WebSocket endpoint
	//
	// Use to analyze:
	//   - WebSocket connection volume by service
	//   - Connection success rates by service
	//   - Distribution between protocol and fallback endpoints for WebSocket connections
	websocketConnectionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      websocketConnectionsTotalMetric,
			Help:      "Total number of WebSocket connections processed by Shannon protocol instance(s)",
		},
		[]string{"service_id", "success", "error_type", "used_fallback"},
	)

	// websocketConnectionErrors tracks WebSocket connection establishment errors
	// Labels:
	//   - service_id: Target service identifier
	//   - endpoint_domain: Effective TLD+1 domain extracted from endpoint URL
	//   - error_type: Type of connection error encountered (based on trusted classification)
	//   - sanction_type: Type of sanction recommended (based on trusted classification)
	//
	// Exemplars:
	//   - endpoint_url: URL of the WebSocket endpoint
	//
	// Use to analyze:
	//   - WebSocket connection errors by service and type
	//   - Sanctions recommended for connection failures
	websocketConnectionErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      websocketConnectionErrorsMetric,
			Help:      "Total WebSocket connection errors by service, endpoint domain, error type, and sanction type",
		},
		[]string{"service_id", "endpoint_domain", "error_type", "sanction_type"},
	)

	// websocketMessagesTotal tracks the total WebSocket messages processed.
	// Labels:
	//   - service_id: Target service identifier (i.e. chain id in Shannon)
	//   - success: Whether the message was processed successfully (true if no message error)
	//   - error_type: type of error encountered processing the message
	//   - used_fallback: Whether the message was processed using a fallback endpoint.
	//
	// Exemplars:
	//   - endpoint_url: URL of the WebSocket endpoint
	//
	// Use to analyze:
	//   - WebSocket message volume by service
	//   - Message processing success rates by service
	//   - Distribution between protocol and fallback endpoints for WebSocket messages
	websocketMessagesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      websocketMessagesTotalMetric,
			Help:      "Total number of WebSocket messages processed by Shannon protocol instance(s)",
		},
		[]string{"service_id", "success", "error_type", "used_fallback"},
	)

	// websocketMessageErrors tracks WebSocket message processing errors
	// Labels:
	//   - service_id: Target service identifier
	//   - endpoint_domain: Effective TLD+1 domain extracted from endpoint URL
	//   - error_type: Type of message error encountered (based on trusted classification)
	//   - sanction_type: Type of sanction recommended (based on trusted classification)
	//
	// Exemplars:
	//   - endpoint_url: URL of the WebSocket endpoint
	//
	// Use to analyze:
	//   - WebSocket message errors by service and type
	//   - Sanctions recommended for message processing failures
	websocketMessageErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      websocketMessageErrorsMetric,
			Help:      "Total WebSocket message errors by service, endpoint domain, error type, and sanction type",
		},
		[]string{"service_id", "endpoint_domain", "error_type", "sanction_type"},
	)

	// sanctionsByDomain tracks sanctions applied by domain.
	// Labels:
	//   - service_id: Target service identifier
	//   - endpoint_domain: Effective TLD+1 domain extracted from endpoint URL
	//   - sanction_type: Type of sanction (based on trusted classification)
	//   - sanction_reason: The endpoint error type that caused the sanction (trusted)
	sanctionsByDomain = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      sanctionsByDomainMetric,
			Help:      "Total sanctions by service, endpoint domain (TLD+1), sanction type, and reason",
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

	// endpointResponseSize tracks the distribution of response payload sizes
	// Labels:
	//   - service_id: Target service identifier
	//   - endpoint_domain: Effective TLD+1 domain extracted from endpoint URL
	//   - success: Whether the request was successful (true if at least one endpoint had no error)
	//
	// Use to analyze:
	//   - Response size distribution patterns
	//   - Bandwidth usage across services and endpoints
	//   - Payload size percentiles
	endpointResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: pathProcess,
			Name:      "endpoint_response_size_bytes",
			Help:      "Histogram of endpoint response payload sizes in bytes",
			Buckets: []float64{
				1_024,      // 1KB
				10_240,     // 10KB
				51_200,     // 50KB
				102_400,    // 100KB
				512_000,    // 500KB
				1_048_576,  // 1MB
				5_242_880,  // 5MB
				10_485_760, // 10MB
			},
		},
		[]string{"service_id", "endpoint_domain", "success"},
	)

	// relayMinerErrorsTotal tracks RelayMinerError occurrences separately from Shannon protocol errors
	// This metric allows analysis of RelayMinerError patterns independently while including
	// endpoint error type for cross-referencing with Shannon protocol errors.
	// Labels:
	//   - service_id: Target service identifier
	//   - endpoint_domain: Effective TLD+1 domain extracted from endpoint URL
	//   - endpoint_error_type: Shannon endpoint error type for cross-referencing (empty if no endpoint error)
	//   - relay_miner_codespace: Codespace from RelayMinerError
	//   - relay_miner_code: Code from RelayMinerError
	//
	// Use to analyze:
	//   - RelayMinerError patterns by codespace and code
	//   - Correlation between endpoint errors and RelayMinerError occurrences
	//   - RelayMinerError distribution across services and endpoint domains
	relayMinerErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      relayMinerErrorsTotalMetric,
			Help:      "Total RelayMinerError occurrences by service, endpoint domain, endpoint error type, and relay miner details",
		},
		[]string{"service_id", "endpoint_domain", "endpoint_error_type", "relay_miner_codespace", "relay_miner_code"},
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
		// Handle different types of observations based on the oneof field
		switch obsData := observationSet.GetObservationData().(type) {
		case *protocolobservations.ShannonRequestObservations_HttpObservations:
			// HTTP observations - existing metrics processing
			httpObservations := obsData.HttpObservations
			if httpObservations == nil {
				logger.Warn().Msg("❌ SHOULD NEVER HAPPEN: skipping processing: received empty HTTP observations")
				continue
			}

			// Record the relay total with success/failure status
			recordRelayTotal(logger, observationSet)

			// Process endpoint errors
			processEndpointErrors(logger, observationSet.GetServiceId(), httpObservations.GetEndpointObservations())

			// Process sanctions by domain
			processSanctionsByDomain(logger, observationSet.GetServiceId(), httpObservations.GetEndpointObservations())

			// Process endpoint latency metrics
			processEndpointLatency(logger, observationSet.GetServiceId(), httpObservations.GetEndpointObservations())

			// Process RelayMinerError occurrences separately
			processRelayMinerErrors(logger, observationSet.GetServiceId(), httpObservations.GetEndpointObservations())

		case *protocolobservations.ShannonRequestObservations_WebsocketConnectionObservation:
			// WebSocket connection observation - new metrics processing
			wsConnectionObs := obsData.WebsocketConnectionObservation
			if wsConnectionObs == nil {
				logger.Warn().Msg("❌ SHOULD NEVER HAPPEN: skipping processing: received empty WebSocket connection observation")
				continue
			}

			// Record WebSocket connection metrics
			recordWebsocketConnectionTotal(logger, observationSet)
			processWebsocketConnectionErrors(logger, observationSet.GetServiceId(), wsConnectionObs)

		case *protocolobservations.ShannonRequestObservations_WebsocketMessageObservation:
			// WebSocket message observation - new metrics processing
			wsMessageObs := obsData.WebsocketMessageObservation
			if wsMessageObs == nil {
				logger.Warn().Msg("❌ SHOULD NEVER HAPPEN: skipping processing: received empty WebSocket message observation")
				continue
			}

			// Record WebSocket message metrics
			recordWebsocketMessageTotal(logger, observationSet)
			processWebsocketMessageErrors(logger, observationSet.GetServiceId(), wsMessageObs)

		default:
			logger.Warn().Msg("❌ SHOULD NEVER HAPPEN: received unknown observation type")
		}
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
				// Relay request failed before reaching out to any endpoints so no fallback was used.
				// Must be set to avoid inconsistent label cardinality error
				"used_fallback": "false",
			},
		).Inc()

		// Request has an error: no endpoint observations to process.
		return
	}

	// Extract endpoint observations and metrics data based on observation type
	var endpointURL string
	var success bool
	var usedFallbackEndpoint bool

	switch obsData := observations.GetObservationData().(type) {
	case *protocolobservations.ShannonRequestObservations_HttpObservations:
		endpointObservations := obsData.HttpObservations.GetEndpointObservations()
		// Skip if there are no endpoint observations
		if len(endpointObservations) == 0 {
			hydratedLogger.Warn().Msg("Request has no errors and no endpoint observations: endpoint selection has failed.")
			return
		}

		// Get the last observation for endpoint address
		lastObs := endpointObservations[len(endpointObservations)-1]
		endpointURL = lastObs.GetEndpointUrl()

		// Determine if any of the observations were successful.
		success = isAnyObservationSuccessful(endpointObservations)

		// Determine if any of the endpoints was a fallback
		usedFallbackEndpoint = isFallbackEndpointUsed(endpointObservations)

	case *protocolobservations.ShannonRequestObservations_WebsocketConnectionObservation:
		wsConnectionObs := obsData.WebsocketConnectionObservation
		endpointURL = wsConnectionObs.GetEndpointUrl()
		success = wsConnectionObs.GetErrorType() == protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_UNSPECIFIED
		usedFallbackEndpoint = wsConnectionObs.GetIsFallbackEndpoint()

	case *protocolobservations.ShannonRequestObservations_WebsocketMessageObservation:
		wsMessageObs := obsData.WebsocketMessageObservation
		endpointURL = wsMessageObs.GetEndpointUrl()
		success = wsMessageObs.GetErrorType() == protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_UNSPECIFIED
		usedFallbackEndpoint = wsMessageObs.GetIsFallbackEndpoint()

	default:
		hydratedLogger.Warn().Msg("Unknown observation type in recordRelayTotal")
		return
	}

	// Create exemplar with high-cardinality data
	// Truncate to 128 runes (Prometheus exemplar limit)
	// See `ExemplarMaxRunes` below:
	// https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#pkg-constants
	exLabels := prometheus.Labels{
		"endpoint_url": endpointURL[:min(len(endpointURL), 128)],
	}

	// Increment the relay total counter with exemplars
	relaysTotal.With(
		prometheus.Labels{
			"service_id":    serviceID,
			"success":       fmt.Sprintf("%t", success),
			"error_type":    "",
			"used_fallback": fmt.Sprintf("%t", usedFallbackEndpoint),
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

// isFallbackEndpointUsed returns true if any endpoint was a fallback endpoint.
func isFallbackEndpointUsed(observations []*protocolobservations.ShannonEndpointObservation) bool {
	for _, obs := range observations {
		if obs.GetIsFallbackEndpoint() {
			return true
		}
	}
	return false
}

// processEndpointErrors records error metrics with exemplars for high-cardinality data
func processEndpointErrors(
	logger polylog.Logger,
	serviceID string,
	observations []*protocolobservations.ShannonEndpointObservation,
) {
	for _, endpointObs := range observations {
		// Skip if there's no error
		if endpointObs.ErrorType == nil {
			continue
		}

		// Extract effective TLD+1 from endpoint URL
		// This function handles edge cases like IP addresses, localhost, invalid URLs
		endpointDomain, err := ExtractDomainOrHost(endpointObs.GetEndpointUrl())
		if err != nil {
			logger.With(
				"endpoint_url", endpointObs.EndpointUrl,
			).
				ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).
				Err(err).Msg("SHOULD NEVER HAPPEN: Could not extract domain from Shannon endpoint URL for relay errors metric")
			continue
		}

		// Extract low-cardinality labels (based on trusted error classification)
		errorType := endpointObs.ErrorType.String()

		// Extract sanction type (based on trusted error classification)
		var sanctionType string
		if endpointObs.RecommendedSanction != nil {
			sanctionType = endpointObs.RecommendedSanction.String()
		}

		// Extract high-cardinality values for exemplars
		endpointURL := endpointObs.EndpointUrl

		// Create exemplar with high-cardinality data
		// Truncate to 128 runes (Prometheus exemplar limit)
		exLabels := prometheus.Labels{
			"endpoint_url": endpointURL[:min(len(endpointURL), 128)],
		}

		// Record relay error
		relaysErrorsTotal.With(
			prometheus.Labels{
				"service_id":      serviceID,
				"endpoint_domain": endpointDomain,
				"error_type":      errorType,
				"sanction_type":   sanctionType,
			},
		).(prometheus.ExemplarAdder).AddWithExemplar(float64(1), exLabels)
	}
}

// processSanctionsByDomain records sanctions without RelayMinerError context
func processSanctionsByDomain(
	logger polylog.Logger,
	serviceID string,
	observations []*protocolobservations.ShannonEndpointObservation,
) {
	for _, endpointObs := range observations {
		// Skip if there's no recommended sanction (based on trusted error classification)
		if endpointObs.RecommendedSanction == nil {
			continue
		}

		// Extract effective domain from endpoint URL
		endpointDomain, err := ExtractDomainOrHost(endpointObs.GetEndpointUrl())
		// error extracting TLD+1, skip.
		if err != nil {
			logger.With(
				"endpoint_url", endpointObs.GetEndpointUrl(),
			).
				ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).
				Err(err).Msg("SHOULD NEVER HAPPEN: Could not extract domain from Shannon endpoint URL")
			continue
		}

		// Extract the sanction reason from the endpoint error type (trusted classification)
		var sanctionReason string
		if endpointObs.ErrorType != nil {
			sanctionReason = endpointObs.GetErrorType().String()
		}

		// Increment the sanctions counter without RelayMinerError context
		sanctionsByDomain.With(
			prometheus.Labels{
				"service_id":      serviceID,
				"endpoint_domain": endpointDomain,
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

		// Extract effective domain from endpoint URL
		endpointDomain, err := ExtractDomainOrHost(endpointObs.GetEndpointUrl())
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
				"endpoint_domain": endpointDomain,
				"success":         fmt.Sprintf("%t", success),
			}).Observe(latencySeconds)

		// Record response size
		responseSize := float64(endpointObs.GetEndpointBackendServiceHttpResponsePayloadSize())
		endpointResponseSize.With(
			prometheus.Labels{
				"service_id":      serviceID,
				"endpoint_domain": endpointDomain,
				"success":         fmt.Sprintf("%t", success),
			}).Observe(responseSize)
	}
}

// processRelayMinerErrors records RelayMinerError occurrences separately from Shannon protocol errors
func processRelayMinerErrors(
	logger polylog.Logger,
	serviceID string,
	observations []*protocolobservations.ShannonEndpointObservation,
) {
	for _, endpointObs := range observations {
		// Skip if there's no RelayMinerError
		if endpointObs.RelayMinerError == nil {
			continue
		}

		// Extract effective domain from endpoint URL
		endpointDomain, err := ExtractDomainOrHost(endpointObs.GetEndpointUrl())
		if err != nil {
			logger.With(
				"endpoint_url", endpointObs.GetEndpointUrl(),
			).
				ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).
				Err(err).Msg("SHOULD NEVER HAPPEN: Could not extract domain from Shannon endpoint URL for RelayMinerError metric")
			continue
		}

		// Extract RelayMinerError details
		relayMinerCodespace := endpointObs.RelayMinerError.GetCodespace()
		relayMinerCode := fmt.Sprintf("%d", endpointObs.RelayMinerError.GetCode())

		// Extract endpoint error type for cross-referencing (empty if no endpoint error)
		var endpointErrorType string
		if endpointObs.ErrorType != nil {
			endpointErrorType = endpointObs.GetErrorType().String()
		}

		// Record RelayMinerError occurrence
		relayMinerErrorsTotal.With(
			prometheus.Labels{
				"service_id":            serviceID,
				"endpoint_domain":       endpointDomain,
				"endpoint_error_type":   endpointErrorType,
				"relay_miner_codespace": relayMinerCodespace,
				"relay_miner_code":      relayMinerCode,
			},
		).Inc()
	}
}

// recordWebsocketConnectionTotal tracks WebSocket connection counts.
func recordWebsocketConnectionTotal(
	logger polylog.Logger,
	observations *protocolobservations.ShannonRequestObservations,
) {
	hydratedLogger := logger.With("method", "recordWebsocketConnectionTotal")

	serviceID := observations.GetServiceId()

	// Check for request-level errors first
	if requestHasErr, requestErrorType := extractRequestError(observations); requestHasErr {
		websocketConnectionsTotal.With(
			prometheus.Labels{
				"service_id":    serviceID,
				"success":       "false",
				"error_type":    requestErrorType,
				"used_fallback": "false",
			},
		).Inc()
		return
	}

	wsConnectionObs := observations.GetWebsocketConnectionObservation()
	if wsConnectionObs == nil {
		hydratedLogger.Warn().Msg("WebSocket connection observation is nil")
		return
	}

	// Determine success based on error type
	success := wsConnectionObs.GetErrorType() == protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_UNSPECIFIED
	usedFallbackEndpoint := wsConnectionObs.GetIsFallbackEndpoint()

	// Extract endpoint URL for exemplars
	endpointURL := wsConnectionObs.GetEndpointUrl()
	exLabels := prometheus.Labels{
		"endpoint_url": endpointURL[:min(len(endpointURL), 128)],
	}

	// Record WebSocket connection total
	websocketConnectionsTotal.With(
		prometheus.Labels{
			"service_id":    serviceID,
			"success":       fmt.Sprintf("%t", success),
			"error_type":    "",
			"used_fallback": fmt.Sprintf("%t", usedFallbackEndpoint),
		},
	).(prometheus.ExemplarAdder).AddWithExemplar(1, exLabels)
}

// recordWebsocketMessageTotal tracks WebSocket message counts.
func recordWebsocketMessageTotal(
	logger polylog.Logger,
	observations *protocolobservations.ShannonRequestObservations,
) {
	hydratedLogger := logger.With("method", "recordWebsocketMessageTotal")

	serviceID := observations.GetServiceId()

	// Check for request-level errors first
	if requestHasErr, requestErrorType := extractRequestError(observations); requestHasErr {
		websocketMessagesTotal.With(
			prometheus.Labels{
				"service_id":    serviceID,
				"success":       "false",
				"error_type":    requestErrorType,
				"used_fallback": "false",
			},
		).Inc()
		return
	}

	wsMessageObs := observations.GetWebsocketMessageObservation()
	if wsMessageObs == nil {
		hydratedLogger.Warn().Msg("WebSocket message observation is nil")
		return
	}

	// Determine success based on error type
	success := wsMessageObs.GetErrorType() == protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_UNSPECIFIED
	usedFallbackEndpoint := wsMessageObs.GetIsFallbackEndpoint()

	// Extract endpoint URL for exemplars
	endpointURL := wsMessageObs.GetEndpointUrl()
	exLabels := prometheus.Labels{
		"endpoint_url": endpointURL[:min(len(endpointURL), 128)],
	}

	// Record WebSocket message total
	websocketMessagesTotal.With(
		prometheus.Labels{
			"service_id":    serviceID,
			"success":       fmt.Sprintf("%t", success),
			"error_type":    "",
			"used_fallback": fmt.Sprintf("%t", usedFallbackEndpoint),
		},
	).(prometheus.ExemplarAdder).AddWithExemplar(1, exLabels)
}

// processWebsocketConnectionErrors records WebSocket connection error metrics.
func processWebsocketConnectionErrors(
	logger polylog.Logger,
	serviceID string,
	wsConnectionObs *protocolobservations.ShannonWebsocketConnectionObservation,
) {
	// Skip if there's no error
	if wsConnectionObs.ErrorType == nil {
		return
	}

	// Extract effective TLD+1 from endpoint URL
	endpointDomain, err := ExtractDomainOrHost(wsConnectionObs.GetEndpointUrl())
	if err != nil {
		logger.With(
			"endpoint_url", wsConnectionObs.EndpointUrl,
		).
			ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).
			Err(err).Msg("Could not extract domain from WebSocket endpoint URL for connection errors metric")
		return
	}

	// Extract error information
	errorType := wsConnectionObs.ErrorType.String()
	var sanctionType string
	if wsConnectionObs.RecommendedSanction != nil {
		sanctionType = wsConnectionObs.RecommendedSanction.String()
	}

	// Extract endpoint URL for exemplars
	endpointURL := wsConnectionObs.EndpointUrl
	exLabels := prometheus.Labels{
		"endpoint_url": endpointURL[:min(len(endpointURL), 128)],
	}

	// Record WebSocket connection error
	websocketConnectionErrors.With(
		prometheus.Labels{
			"service_id":      serviceID,
			"endpoint_domain": endpointDomain,
			"error_type":      errorType,
			"sanction_type":   sanctionType,
		},
	).(prometheus.ExemplarAdder).AddWithExemplar(float64(1), exLabels)

	// Record sanction if recommended
	if wsConnectionObs.RecommendedSanction != nil {
		sanctionsByDomain.With(
			prometheus.Labels{
				"service_id":      serviceID,
				"endpoint_domain": endpointDomain,
				"sanction_type":   sanctionType,
				"sanction_reason": errorType,
			},
		).Inc()
	}

	// Record RelayMinerError if present
	if wsConnectionObs.RelayMinerError != nil {
		relayMinerCodespace := wsConnectionObs.RelayMinerError.GetCodespace()
		relayMinerCode := fmt.Sprintf("%d", wsConnectionObs.RelayMinerError.GetCode())

		relayMinerErrorsTotal.With(
			prometheus.Labels{
				"service_id":            serviceID,
				"endpoint_domain":       endpointDomain,
				"endpoint_error_type":   errorType,
				"relay_miner_codespace": relayMinerCodespace,
				"relay_miner_code":      relayMinerCode,
			},
		).Inc()
	}
}

// processWebsocketMessageErrors records WebSocket message error metrics.
func processWebsocketMessageErrors(
	logger polylog.Logger,
	serviceID string,
	wsMessageObs *protocolobservations.ShannonWebsocketMessageObservation,
) {
	// Skip if there's no error
	if wsMessageObs.ErrorType == nil {
		return
	}

	// Extract effective TLD+1 from endpoint URL
	endpointDomain, err := ExtractDomainOrHost(wsMessageObs.GetEndpointUrl())
	if err != nil {
		logger.With(
			"endpoint_url", wsMessageObs.EndpointUrl,
		).
			ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).
			Err(err).Msg("Could not extract domain from WebSocket endpoint URL for message errors metric")
		return
	}

	// Extract error information
	errorType := wsMessageObs.ErrorType.String()
	var sanctionType string
	if wsMessageObs.RecommendedSanction != nil {
		sanctionType = wsMessageObs.RecommendedSanction.String()
	}

	// Extract endpoint URL for exemplars
	endpointURL := wsMessageObs.EndpointUrl
	exLabels := prometheus.Labels{
		"endpoint_url": endpointURL[:min(len(endpointURL), 128)],
	}

	// Record WebSocket message error
	websocketMessageErrors.With(
		prometheus.Labels{
			"service_id":      serviceID,
			"endpoint_domain": endpointDomain,
			"error_type":      errorType,
			"sanction_type":   sanctionType,
		},
	).(prometheus.ExemplarAdder).AddWithExemplar(float64(1), exLabels)

	// Record sanction if recommended
	if wsMessageObs.RecommendedSanction != nil {
		sanctionsByDomain.With(
			prometheus.Labels{
				"service_id":      serviceID,
				"endpoint_domain": endpointDomain,
				"sanction_type":   sanctionType,
				"sanction_reason": errorType,
			},
		).Inc()
	}

	// Record RelayMinerError if present
	if wsMessageObs.RelayMinerError != nil {
		relayMinerCodespace := wsMessageObs.RelayMinerError.GetCodespace()
		relayMinerCode := fmt.Sprintf("%d", wsMessageObs.RelayMinerError.GetCode())

		relayMinerErrorsTotal.With(
			prometheus.Labels{
				"service_id":            serviceID,
				"endpoint_domain":       endpointDomain,
				"endpoint_error_type":   errorType,
				"relay_miner_codespace": relayMinerCodespace,
				"relay_miner_code":      relayMinerCode,
			},
		).Inc()
	}
}
