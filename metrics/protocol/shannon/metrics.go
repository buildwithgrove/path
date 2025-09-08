// Package shannon provides functionality for exporting Shannon protocol metrics to Prometheus.
package shannon

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/prometheus/client_golang/prometheus"

	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
)

// TODO_METRICS(@commoddity): Add additional WebSocket-specific metrics
// - Message latency distribution (time between request and response for each message)
// - Connection duration histogram (time from connection establishment to termination)
// - Message size percentiles (distribution of message payload sizes)
// - Subscription event rates (frequency of subscription events per connection)

const (
	// The POSIX process that emits metrics
	pathProcess = "path"

	// HTTP relay metrics
	relaysTotalMetric          = "shannon_relays_total"
	relaysErrorsTotalMetric    = "shannon_relay_errors_total"
	relaysActiveRequestsMetric = "shannon_relays_active"

	// WebSocket connection metrics
	websocketConnectionsTotalMetric  = "shannon_websocket_connections_total"
	websocketConnectionErrorsMetric  = "shannon_websocket_connection_errors_total"
	websocketConnectionsActiveMetric = "shannon_websocket_connections_active"

	// WebSocket message metrics
	websocketMessagesTotalMetric = "shannon_websocket_messages_total"
	websocketMessageErrorsMetric = "shannon_websocket_message_errors_total"

	// Sanctions metrics (shared across HTTP and WebSocket)
	sanctionsByDomainMetric = "shannon_sanctions_by_domain"

	// Latency metrics (currently HTTP only)
	endpointLatencyMetric       = "shannon_endpoint_latency_seconds"
	relayMinerErrorsTotalMetric = "shannon_relay_miner_errors_total"

	// The default value for a domain if it cannot be extracted from an endpoint URL
	errDomain = "error_extracting_domain"
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
	prometheus.MustRegister(activeRelays)

	// WebSocket metrics
	prometheus.MustRegister(websocketConnectionsTotal)
	prometheus.MustRegister(websocketConnectionErrors)
	prometheus.MustRegister(websocketMessagesTotal)
	prometheus.MustRegister(websocketMessageErrors)
	prometheus.MustRegister(activeWebsocketConnections)

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
	//   - endpoint_domain: Effective TLD+1 domain extracted from endpoint URL
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
		[]string{"service_id", "success", "error_type", "used_fallback", "endpoint_domain"},
	)

	// TODO_IMPROVE(@adshmh): This should be called endpointErrorsTotal
	//
	// relaysErrorsTotal tracks relay errors from Shannon protocol
	// Labels:
	//   - service_id: Target service identifier
	//   - error_type: Type of error encountered (based on trusted classification)
	//   - sanction_type: Type of sanction recommended (based on trusted classification)
	//   - endpoint_domain: Effective TLD+1 domain extracted from endpoint URL
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
		[]string{"service_id", "error_type", "sanction_type", "endpoint_domain"},
	)

	// activeRelays tracks the current number of active Shannon HTTP requests.
	// This gauge metric shows the real-time concurrency level for monitoring
	// request load and identifying potential bottlenecks.
	//
	// Use to analyze:
	//   - Current request concurrency levels
	//   - Request load patterns over time
	//   - Capacity planning and resource utilization
	//   - Identifying request spikes and bottlenecks
	activeRelays = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: pathProcess,
			Name:      relaysActiveRequestsMetric,
			Help:      "Current number of active Shannon requests being processed",
		},
		[]string{"request_type"},
	)

	// websocketConnectionsTotal tracks the total WebSocket connection attempts processed.
	// Labels:
	//   - service_id: Target service identifier (i.e. chain id in Shannon)
	//   - success: Whether the connection was successful (true if no connection error)
	//   - error_type: type of error encountered during connection setup
	//   - used_fallback: Whether the connection used a fallback endpoint.
	//   - endpoint_domain: Effective TLD+1 domain extracted from endpoint URL
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
		[]string{"service_id", "success", "error_type", "used_fallback", "endpoint_domain"},
	)

	// websocketConnectionErrors tracks WebSocket connection establishment errors
	// Labels:
	//   - service_id: Target service identifier
	//   - error_type: Type of connection error encountered (based on trusted classification)
	//   - sanction_type: Type of sanction recommended (based on trusted classification)
	//   - endpoint_domain: Effective TLD+1 domain extracted from endpoint URL

	// Use to analyze:
	//   - WebSocket connection errors by service and type
	//   - Sanctions recommended for connection failures
	websocketConnectionErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      websocketConnectionErrorsMetric,
			Help:      "Total WebSocket connection errors by service, endpoint domain, error type, and sanction type",
		},
		[]string{"service_id", "error_type", "sanction_type", "endpoint_domain"},
	)

	// websocketMessagesTotal tracks the total WebSocket messages processed.
	// Labels:
	//   - service_id: Target service identifier (i.e. chain id in Shannon)
	//   - success: Whether the message was processed successfully (true if no message error)
	//   - error_type: type of error encountered processing the message
	//   - used_fallback: Whether the message was processed using a fallback endpoint.
	//   - endpoint_domain: Effective TLD+1 domain extracted from endpoint URL
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
		[]string{"service_id", "success", "error_type", "used_fallback", "endpoint_domain"},
	)

	// websocketMessageErrors tracks WebSocket message processing errors
	// Labels:
	//   - service_id: Target service identifier
	//   - error_type: Type of message error encountered (based on trusted classification)
	//   - sanction_type: Type of sanction recommended (based on trusted classification)
	//   - endpoint_domain: Effective TLD+1 domain extracted from endpoint URL
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
		[]string{"service_id", "error_type", "sanction_type", "used_fallback", "endpoint_domain"},
	)

	// activeWebsocketConnections tracks the current number of active WebSocket connections.
	// This gauge metric shows the real-time WebSocket connection count for monitoring
	// persistent connection load and identifying potential bottlenecks.
	//
	// Use to analyze:
	//   - Current WebSocket connection counts
	//   - Connection load patterns over time
	//   - Capacity planning for persistent connections
	//   - Identifying connection spikes and bottlenecks
	activeWebsocketConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Subsystem: pathProcess,
			Name:      websocketConnectionsActiveMetric,
			Help:      "Current number of active Shannon WebSocket connections",
		},
	)

	// sanctionsByDomain tracks sanctions applied by domain.
	// Labels:
	//   - service_id: Target service identifier
	//   - sanction_type: Type of sanction (based on trusted classification)
	//   - sanction_reason: The endpoint error type that caused the sanction (trusted)
	//   - endpoint_domain: Effective TLD+1 domain extracted from endpoint URL
	sanctionsByDomain = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      sanctionsByDomainMetric,
			Help:      "Total sanctions by service, endpoint domain (TLD+1), sanction type, and reason",
		},
		[]string{"service_id", "sanction_type", "sanction_reason", "endpoint_domain"},
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

		// Check for request processing errors.
		// e.g. error fetching a session for the target service.
		if observationSet.GetRequestError() != nil {
			// Record the relay total with success/failure status
			recordRelayTotal(logger, observationSet)

			// Request processing encountered error.
			// Skip endpoint observations.
			continue
		}

		// TODO_IMPROVE(@adshmh): Replace dynamic type casts with nil checks.
		//
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
// Success determination varies by observation type:
// - HTTP observations: Success if ANY endpoint observation has ErrorType = UNSPECIFIED (supports parallel requests)
// - WebSocket connection observations: Success if ErrorType = UNSPECIFIED (single connection establishment)
// - WebSocket message observations: Success if ErrorType = UNSPECIFIED (individual message processing)
func recordRelayTotal(
	logger polylog.Logger,
	observations *protocolobservations.ShannonRequestObservations,
) {
	hydratedLogger := logger.With("method", "recordRelaysTotal")

	serviceID := observations.GetServiceId()

	// === FAILED RELAY ===
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

	// === SUCCESSFUL RELAY ===

	// Extract endpoint observations and metrics data based on observation type
	var endpointURL string
	var success bool
	var usedFallbackEndpoint bool

	switch obsData := observations.GetObservationData().(type) {
	case *protocolobservations.ShannonRequestObservations_HttpObservations:
		// HTTP observations can contain multiple endpoint attempts due to parallel requests or retries.
		// Success is determined by checking if ANY endpoint observation succeeded (ErrorType = UNSPECIFIED).
		// This supports the HTTP protocol's ability to try multiple endpoints for a single request.
		endpointObservations := obsData.HttpObservations.GetEndpointObservations()
		// Skip if there are no endpoint observations
		if len(endpointObservations) == 0 {
			hydratedLogger.Warn().Msg("Request has no errors and no endpoint observations: endpoint selection has failed.")
			return
		}

		// Get the last observation for endpoint address
		lastObs := endpointObservations[len(endpointObservations)-1]
		endpointURL = lastObs.GetEndpointUrl()

		// Determine if any of the observations were successful using explicit helper function
		success = isAnyObservationSuccessful(endpointObservations)

		// Determine if any of the endpoints was a fallback
		usedFallbackEndpoint = isFallbackEndpointUsed(endpointObservations)

	case *protocolobservations.ShannonRequestObservations_WebsocketConnectionObservation:
		// WebSocket connection observations track the establishment/termination of a single WebSocket connection.
		// Success is determined by whether the connection was established successfully (no ErrorType set).
		// This represents the initial handshake and connection setup phase, not individual message processing.
		wsConnectionObs := obsData.WebsocketConnectionObservation
		endpointURL = wsConnectionObs.GetEndpointUrl()
		success = isWebsocketConnectionSuccessful(wsConnectionObs)
		usedFallbackEndpoint = wsConnectionObs.GetIsFallbackEndpoint()

	case *protocolobservations.ShannonRequestObservations_WebsocketMessageObservation:
		// WebSocket message observations track individual message processing within an established connection.
		// Success is determined by whether the specific message was processed without errors.
		// This represents the processing of a single request/response or subscription event within the connection.
		wsMessageObs := obsData.WebsocketMessageObservation
		endpointURL = wsMessageObs.GetEndpointUrl()
		success = isWebsocketMessageSuccessful(wsMessageObs)
		usedFallbackEndpoint = wsMessageObs.GetIsFallbackEndpoint()

	default:
		hydratedLogger.Warn().Msg("Unknown observation type in recordRelayTotal")
		return
	}

	// Extract effective TLD+1 from endpoint URL
	// This function handles edge cases like IP addresses, localhost, invalid URLs
	endpointDomain, err := ExtractDomainOrHost(endpointURL)
	if err != nil {
		logger.Error().Err(err).Msgf("Could not extract domain from Shannon endpoint URL %s for relay errors metric", endpointURL)
		endpointDomain = errDomain
	}

	// Increment the relay total counter with exemplars
	relaysTotal.With(
		prometheus.Labels{
			"service_id":      serviceID,
			"success":         fmt.Sprintf("%t", success),
			"error_type":      "",
			"used_fallback":   fmt.Sprintf("%t", usedFallbackEndpoint),
			"endpoint_domain": endpointDomain,
		},
	).Add(1)
}

// extractRequestError  extracts from the observations the status (success/failure) and the first encountered error, if any.
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

// isAnyObservationSuccessful returns true if any HTTP endpoint observation indicates a success.
// Success is determined by checking if ErrorType is UNSPECIFIED (meaning no error occurred).
func isAnyObservationSuccessful(observations []*protocolobservations.ShannonEndpointObservation) bool {
	for _, obs := range observations {
		if obs.GetErrorType() == protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_UNSPECIFIED {
			return true
		}
	}
	return false
}

// isWebsocketConnectionSuccessful returns true if the WebSocket connection observation indicates success.
// For WebSocket connections, success is determined by checking if ErrorType is UNSPECIFIED.
// Unlike HTTP observations which can have multiple endpoint attempts, WebSocket connections
// use a single endpoint and have a single success/failure status.
func isWebsocketConnectionSuccessful(wsConnectionObs *protocolobservations.ShannonWebsocketConnectionObservation) bool {
	return wsConnectionObs.GetErrorType() == protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_UNSPECIFIED
}

// isWebsocketMessageSuccessful returns true if the WebSocket message observation indicates success.
// For WebSocket messages, success is determined by checking if ErrorType is UNSPECIFIED.
// Each WebSocket message is processed individually and has its own success/failure status.
func isWebsocketMessageSuccessful(wsMessageObs *protocolobservations.ShannonWebsocketMessageObservation) bool {
	return wsMessageObs.GetErrorType() == protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_UNSPECIFIED
}

// isFallbackEndpointUsed returns true if any HTTP endpoint observation indicates a fallback endpoint was used.
// This function is specific to HTTP observations which can have multiple endpoint attempts.
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

		// Extract effective TLD+1 from endpoint URL.
		endpointDomain, err := ExtractDomainOrHost(endpointObs.GetEndpointUrl())
		if err != nil {
			logger.Error().Err(err).Msgf("Could not extract domain from Shannon endpoint URL %s for relay errors metric", endpointObs.GetEndpointUrl())
			endpointDomain = errDomain
		}

		// Extract low-cardinality labels (based on trusted error classification)
		errorType := endpointObs.ErrorType.String()

		// Extract sanction type (based on trusted error classification)
		var sanctionType string
		if endpointObs.RecommendedSanction != nil {
			sanctionType = endpointObs.RecommendedSanction.String()
		}

		// Record relay error
		relaysErrorsTotal.With(
			prometheus.Labels{
				"service_id":      serviceID,
				"error_type":      errorType,
				"sanction_type":   sanctionType,
				"endpoint_domain": endpointDomain,
			},
		).Inc()
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

		// Extract effective TLD+1 from endpoint URL.
		endpointDomain, err := ExtractDomainOrHost(endpointObs.GetEndpointUrl())
		if err != nil {
			logger.Error().Err(err).Msgf("Could not extract domain from endpoint URL %s.", endpointObs.GetEndpointUrl())
			endpointDomain = errDomain
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
				"sanction_type":   endpointObs.GetRecommendedSanction().String(),
				"sanction_reason": sanctionReason,
				"endpoint_domain": endpointDomain,
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

		// Extract effective TLD+1 from endpoint URL.
		endpointUrl := endpointObs.GetEndpointUrl()
		endpointDomain, err := ExtractDomainOrHost(endpointUrl)
		if err != nil {
			logger.Error().Err(err).Msgf("Could not extract domain from endpoint URL %s.", endpointUrl)
			endpointDomain = errDomain
		}

		// Calculate latency in seconds
		queryTimestamp := queryTime.AsTime()
		responseTimestamp := responseTime.AsTime()
		latencySeconds := responseTimestamp.Sub(queryTimestamp).Seconds()

		// Skip negative latencies (invalid timestamps)
		if latencySeconds < 0 {
			logger.Error().Err(fmt.Errorf("negative latency detected")).Msgf("SHOULD NEVER HAPPEN: Negative latency (%f) detected, skipping metric for endpoint %s", latencySeconds, endpointUrl)
			continue
		}

		// Record latency
		endpointLatency.With(
			prometheus.Labels{
				"service_id":      serviceID,
				"success":         fmt.Sprintf("%t", success),
				"endpoint_domain": endpointDomain,
			}).Observe(latencySeconds)

		// Record response size
		responseSize := float64(endpointObs.GetEndpointBackendServiceHttpResponsePayloadSize())
		endpointResponseSize.With(
			prometheus.Labels{
				"service_id":      serviceID,
				"success":         fmt.Sprintf("%t", success),
				"endpoint_domain": endpointDomain,
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
		endpointUrl := endpointObs.GetEndpointUrl()
		endpointDomain, err := ExtractDomainOrHost(endpointUrl)
		if err != nil {
			logger.Error().Err(err).Msgf("Could not extract domain from endpoint URL %s.", endpointUrl)
			endpointDomain = errDomain
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
				"endpoint_error_type":   endpointErrorType,
				"relay_miner_codespace": relayMinerCodespace,
				"relay_miner_code":      relayMinerCode,
				"endpoint_domain":       endpointDomain,
			},
		).Inc()
	}
}

// SetActiveHTTPRelays updates the gauge metric with the current number of active HTTP relays.
// This should be called whenever the active HTTP relay count changes in the concurrency limiter.
// TODO_TECHDEBT: the metrics package should use the passed observation to report metrics: it should not expose any methods for external usage (other than PublishMetrics)
func SetActiveHTTPRelays(activeCount int64) {
	activeRelays.With(prometheus.Labels{
		"request_type": "http",
	}).Set(float64(activeCount))
}

// SetActiveWebsocketConnections updates the gauge metric with the current number of active WebSocket connections.
// This should be called whenever the WebSocket connection count changes.
func SetActiveWebsocketConnections(activeCount int64) {
	activeWebsocketConnections.Set(float64(activeCount))
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

	// Determine success based on error type using explicit helper function
	success := isWebsocketConnectionSuccessful(wsConnectionObs)
	usedFallbackEndpoint := wsConnectionObs.GetIsFallbackEndpoint()

	// Extract endpoint URL for exemplars
	endpointURL := wsConnectionObs.GetEndpointUrl()
	endpointDomain, err := ExtractDomainOrHost(endpointURL)
	if err != nil {
		logger.Error().Err(err).Msgf("Could not extract domain from Shannon endpoint URL %s for relay errors metric", endpointURL)
		endpointDomain = errDomain
	}

	// Record WebSocket connection total
	websocketConnectionsTotal.With(
		prometheus.Labels{
			"service_id":      serviceID,
			"success":         fmt.Sprintf("%t", success),
			"error_type":      "",
			"used_fallback":   fmt.Sprintf("%t", usedFallbackEndpoint),
			"endpoint_domain": endpointDomain,
		},
	).Add(1)
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

	// Determine success based on error type using explicit helper function
	success := isWebsocketMessageSuccessful(wsMessageObs)
	usedFallbackEndpoint := wsMessageObs.GetIsFallbackEndpoint()

	// Extract endpoint URL for exemplars
	endpointURL := wsMessageObs.GetEndpointUrl()
	endpointDomain, err := ExtractDomainOrHost(endpointURL)
	if err != nil {
		logger.Error().Err(err).Msgf("Could not extract domain from WebSocket endpoint URL %s for message errors metric", endpointURL)
		endpointDomain = errDomain
	}

	// Record WebSocket message total
	websocketMessagesTotal.With(
		prometheus.Labels{
			"service_id":      serviceID,
			"success":         fmt.Sprintf("%t", success),
			"error_type":      "",
			"used_fallback":   fmt.Sprintf("%t", usedFallbackEndpoint),
			"endpoint_domain": endpointDomain,
		}).Inc()
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

	// Extract effective TLD+1 from endpoint URL.
	endpointUrl := wsConnectionObs.GetEndpointUrl()
	endpointDomain, err := ExtractDomainOrHost(endpointUrl)
	if err != nil {
		logger.Error().Err(err).Msgf("Could not extract domain from endpoint URL %s.", endpointUrl)
		endpointDomain = errDomain
	}

	// Extract error information
	errorType := wsConnectionObs.ErrorType.String()
	var sanctionType string
	if wsConnectionObs.RecommendedSanction != nil {
		sanctionType = wsConnectionObs.RecommendedSanction.String()
	}

	// Record WebSocket connection error
	websocketConnectionErrors.With(
		prometheus.Labels{
			"service_id":      serviceID,
			"error_type":      errorType,
			"sanction_type":   sanctionType,
			"endpoint_domain": endpointDomain,
		}).Inc()

	// Record sanction if recommended
	if wsConnectionObs.RecommendedSanction != nil {
		sanctionsByDomain.With(
			prometheus.Labels{
				"service_id":      serviceID,
				"sanction_type":   sanctionType,
				"sanction_reason": errorType,
				"endpoint_domain": endpointDomain,
			}).Inc()
	}

	// Record RelayMinerError if present
	if wsConnectionObs.RelayMinerError != nil {
		relayMinerCodespace := wsConnectionObs.RelayMinerError.GetCodespace()
		relayMinerCode := fmt.Sprintf("%d", wsConnectionObs.RelayMinerError.GetCode())

		relayMinerErrorsTotal.With(
			prometheus.Labels{
				"service_id":            serviceID,
				"endpoint_error_type":   errorType,
				"relay_miner_codespace": relayMinerCodespace,
				"relay_miner_code":      relayMinerCode,
				"endpoint_domain":       endpointDomain,
			}).Inc()
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

	// Extract effective TLD+1 from endpoint URL.
	endpointDomain, err := ExtractDomainOrHost(wsMessageObs.GetEndpointUrl())
	if err != nil {
		logger.Error().Err(err).Msgf("Could not extract domain from endpoint URL %s.", wsMessageObs.GetEndpointUrl())
		endpointDomain = errDomain
	}

	// Extract error information
	errorType := wsMessageObs.ErrorType.String()
	var sanctionType string
	if wsMessageObs.RecommendedSanction != nil {
		sanctionType = wsMessageObs.RecommendedSanction.String()
	}

	// Record WebSocket message error
	websocketMessageErrors.With(
		prometheus.Labels{
			"service_id":      serviceID,
			"error_type":      errorType,
			"sanction_type":   sanctionType,
			"used_fallback":   fmt.Sprintf("%t", wsMessageObs.GetIsFallbackEndpoint()),
			"endpoint_domain": endpointDomain,
		}).Inc()

	// Record sanction if recommended
	if wsMessageObs.RecommendedSanction != nil {
		sanctionsByDomain.With(
			prometheus.Labels{
				"service_id":      serviceID,
				"sanction_type":   sanctionType,
				"sanction_reason": errorType,
				"endpoint_domain": endpointDomain,
			}).Inc()
	}

	// Record RelayMinerError if present
	if wsMessageObs.RelayMinerError != nil {
		relayMinerCodespace := wsMessageObs.RelayMinerError.GetCodespace()
		relayMinerCode := fmt.Sprintf("%d", wsMessageObs.RelayMinerError.GetCode())

		relayMinerErrorsTotal.With(
			prometheus.Labels{
				"service_id":            serviceID,
				"endpoint_error_type":   errorType,
				"relay_miner_codespace": relayMinerCodespace,
				"relay_miner_code":      relayMinerCode,
				"endpoint_domain":       endpointDomain,
			}).Inc()
	}
}
