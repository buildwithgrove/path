package evm

import (
	"fmt"
	"strings"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/prometheus/client_golang/prometheus"

	metricshttp "github.com/buildwithgrove/path/metrics/http"
	"github.com/buildwithgrove/path/observation/qos"
)

const (
	// The POSIX process that emits metrics
	pathProcess = "path"

	// The list of metrics being tracked for EVM QoS
	requestsTotalMetric            = "evm_requests_total"
	availableEndpointsMetric       = "evm_available_endpoints"
	validEndpointsMetric           = "evm_valid_endpoints"
	endpointValidationsTotalMetric = "evm_endpoint_validations_total"
)

func init() {
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(availableEndpoints)
	prometheus.MustRegister(validEndpoints)
	prometheus.MustRegister(endpointValidationsTotal)
}

var (
	// TODO_MVP(@adshmh): Update requestsTotal metric labels:
	// - Add 'errorSubType' field to further categorize errors
	// - Use errorType for broad categories (request validation, protocol error)
	// - Use errorSubType for specifics (endpoint maxed out, endpoint timed out)
	// - Remove 'success' field (success indicated by absence of errorType)
	// - Update EVM observations proto files and add observation interpreter support
	//
	// TODO_MVP(@adshmh): Track endpoint responses separately from requests if/when retries are implemented,
	// since a single request may generate multiple responses due to retry attempts.
	//
	// requestsTotal tracks the total EVM requests processed.
	// Labels:
	//   - chain_id: Target EVM chain identifier
	//   - service_id: Service ID of the EVM QoS instance
	//   - request_origin: origin of the request: Organic (i.e. user) or Synthetic (i.e. hydrator)
	//   - request_method: JSON-RPC method name
	//   - success: Whether a valid response was received
	//   - error_type: Type of error if request failed (or "" for successful requests)
	//   - http_status_code: The HTTP status code returned to the user
	//   - random_endpoint_fallback: Random endpoint selected when all failed validation
	//
	// Use to analyze:
	//   - Request volume by chain and method
	//   - Success rates across different PATH deployment regions
	//   - Method usage patterns across chains
	//   - End-to-end request success rates
	//   - Error types by JSON-RPC method and chain
	//   - HTTP status code distribution
	//   - Service degradation when random endpoint fallback is used
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      requestsTotalMetric,
			Help:      "Total number of requests processed by EVM QoS instance(s)",
		},
		[]string{"chain_id", "service_id", "request_origin", "request_method", "success", "error_type", "http_status_code", "random_endpoint_fallback"},
	)

	// availableEndpoints tracks the number of available endpoints per service.
	// Labels:
	//   - chain_id: Target EVM chain identifier
	//   - service_id: Service ID of the EVM QoS instance
	//
	// Use to analyze:
	//   - Endpoint pool size per service
	//   - Service capacity trends
	availableEndpoints = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: pathProcess,
			Name:      availableEndpointsMetric,
			Help:      "Number of available endpoints for EVM QoS instance(s)",
		},
		[]string{"chain_id", "service_id"},
	)

	// validEndpoints tracks the number of valid endpoints per service after filtering.
	// Labels:
	//   - chain_id: Target EVM chain identifier
	//   - service_id: Service ID of the EVM QoS instance
	//
	// Use to analyze:
	//   - Endpoint health per service
	//   - Validation failure rates
	//   - Service quality trends
	validEndpoints = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: pathProcess,
			Name:      validEndpointsMetric,
			Help:      "Number of valid endpoints for EVM QoS instance(s) after validation",
		},
		[]string{"chain_id", "service_id"},
	)

	// endpointValidationsTotal tracks all endpoint validation attempts with detailed results.
	// This metric provides comprehensive validation tracking for calculating success rates and analyzing failure patterns.
	//
	// Note: Multiple endpoint validations occur during each service request processing:
	// - All available endpoints are validated before selection for the service request
	// - Failed endpoints are filtered out with specific failure reasons captured
	// - Valid endpoints are identified and one is selected to handle the request
	// - This metric captures ALL validation attempts (both successful and failed) that occurred during endpoint selection
	//
	// Labels:
	//   - chain_id: Target EVM chain identifier
	//   - service_id: Service ID of the EVM QoS instance
	//   - domain: eTLD+1 of endpoint URL for provider analysis (extracted from endpoint_addr)
	//   - success: "true" for successful validations, "false" for failed validations
	//   - validation_failure_reason: Specific failure reason for failed validations (empty for successful ones)
	//
	// Validation failure reasons include:
	//   - "ENDPOINT_VALIDATION_FAILURE_REASON_EMPTY_RESPONSE_HISTORY"
	//   - "ENDPOINT_VALIDATION_FAILURE_REASON_RECENT_INVALID_RESPONSE"
	//   - "ENDPOINT_VALIDATION_FAILURE_REASON_BLOCK_NUMBER_BEHIND"
	//   - "ENDPOINT_VALIDATION_FAILURE_REASON_CHAIN_ID_MISMATCH"
	//   - "ENDPOINT_VALIDATION_FAILURE_REASON_NO_BLOCK_NUMBER_OBSERVATION"
	//   - "ENDPOINT_VALIDATION_FAILURE_REASON_NO_CHAIN_ID_OBSERVATION"
	//   - "ENDPOINT_VALIDATION_FAILURE_REASON_ARCHIVAL_CHECK_FAILED"
	//   - "ENDPOINT_VALIDATION_FAILURE_REASON_ENDPOINT_NOT_FOUND"
	//   - "ENDPOINT_VALIDATION_FAILURE_REASON_UNKNOWN"
	//
	// Use to analyze:
	//   - Validation success rate: sum(success="true") / sum(all) by domain
	//   - Validation failure rate: sum(success="false") / sum(all) by domain
	//   - Most common failure types: sum by (validation_failure_reason) where success="false"
	//   - Provider reliability comparison across domains
	//   - Service capacity utilization per provider
	//   - Trends in endpoint quality over time by domain
	endpointValidationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      endpointValidationsTotalMetric,
			Help:      "Total endpoint validation attempts with success status and failure reasons at EVM QoS level",
		},
		[]string{"chain_id", "service_id", "domain", "success", "validation_failure_reason"},
	)
)

// PublishMetrics exports all EVM-related Prometheus metrics using observations reported by EVM QoS service.
// It logs errors for unexpected conditions that should never occur in normal operation.
func PublishMetrics(logger polylog.Logger, observations *qos.EVMRequestObservations) {
	logger = logger.With("method", "PublishMetricsEVM")

	// Skip if observations is nil.
	// This should never happen as PublishQoSMetrics uses nil checks to identify which QoS service produced the observations.
	if observations == nil {
		logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Unable to publish EVM metrics: received nil observations.")
		return
	}

	// Create an interpreter for the observations
	interpreter := &qos.EVMObservationInterpreter{
		Logger:       logger,
		Observations: observations,
	}

	// Extract chain ID
	chainID := extractChainID(logger, interpreter)

	// Extract service ID
	serviceID := extractServiceID(logger, interpreter)

	// Extract request method
	method := extractRequestMethod(logger, interpreter)

	// Extract endpoint selection metadata
	endpointSelectionMetadata := extractEndpointSelectionMetadata(interpreter)

	// Get request status
	statusCode, requestError, err := interpreter.GetRequestStatus()
	// If we couldn't get status info due to missing observations, skip metrics.
	// This should never happen if the observations are properly initialized.
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get request status for EVM metrics - this indicates a programming/implementation error")
		return
	}

	// Determine error type
	var errorType string // Default to empty string for successful requests
	if requestError != nil {
		// Use the String() method on the RequestError to get the string representation
		errorType = requestError.String()
	}

	// Increment request counters with all corresponding labels
	requestsTotal.With(
		prometheus.Labels{
			"chain_id":                 chainID,
			"service_id":               serviceID,
			"request_origin":           interpreter.GetRequestOrigin(),
			"request_method":           method,
			"success":                  fmt.Sprintf("%t", requestError == nil),
			"error_type":               errorType,
			"http_status_code":         fmt.Sprintf("%d", statusCode),
			"random_endpoint_fallback": fmt.Sprintf("%t", endpointSelectionMetadata.RandomEndpointFallback),
		},
	).Inc()

	// Update endpoint count gauges (calculated from validation results)
	availableCount := calculateAvailableEndpointsCount(endpointSelectionMetadata)
	validCount := calculateValidEndpointsCount(endpointSelectionMetadata)

	availableEndpoints.With(
		prometheus.Labels{
			"chain_id":   chainID,
			"service_id": serviceID,
		},
	).Set(float64(availableCount))

	validEndpoints.With(
		prometheus.Labels{
			"chain_id":   chainID,
			"service_id": serviceID,
		},
	).Set(float64(validCount))

	// Publish validation failure metrics using the structured data from metadata
	publishValidationMetricsFromMetadata(logger, chainID, serviceID, endpointSelectionMetadata)
}

// publishValidationMetricsFromMetadata publishes validation metrics (both failures and successes)
// using the structured data from endpoint selection metadata.
// Extracts domain information from endpoint addresses at metrics time.
func publishValidationMetricsFromMetadata(logger polylog.Logger, chainID, serviceID string, metadata *qos.EndpointSelectionMetadata) {
	if metadata == nil {
		return
	}

	// Process all validation results in a single loop
	for _, result := range metadata.ValidationResults {
		domain := extractDomainFromEndpointAddr(logger, result.EndpointAddr)

		// Determine failure reason for failed validations
		failureReason := ""
		if !result.Success && result.FailureReason != nil {
			failureReason = result.FailureReason.String()
		}

		// Track validation result
		endpointValidationsTotal.With(
			prometheus.Labels{
				"chain_id":                  chainID,
				"service_id":                serviceID,
				"domain":                    domain,
				"success":                   fmt.Sprintf("%t", result.Success),
				"validation_failure_reason": failureReason,
			},
		).Inc()
	}
}

// calculateAvailableEndpointsCount returns the total number of endpoints that were validated.
func calculateAvailableEndpointsCount(metadata *qos.EndpointSelectionMetadata) int {
	if metadata == nil {
		return 0
	}
	return len(metadata.ValidationResults)
}

// calculateValidEndpointsCount returns the number of endpoints that passed validation.
func calculateValidEndpointsCount(metadata *qos.EndpointSelectionMetadata) int {
	if metadata == nil {
		return 0
	}

	validCount := 0
	for _, result := range metadata.ValidationResults {
		if result.Success {
			validCount++
		}
	}
	return validCount
}

// publishSuccessfulValidations is now deprecated since we track actual validation results.
// Keeping for backward compatibility but it's no longer used.
func publishSuccessfulValidations(logger polylog.Logger, chainID, serviceID string, totalSuccessful int) {
	// This function is deprecated - we now track actual validation results
	// from the ValidationResults field in metadata
	logger.Debug().Msg("publishSuccessfulValidations is deprecated - using actual ValidationResults from metadata")
}

// extractDomainFromEndpointAddr extracts the eTLD+1 domain from an endpoint address.
// Handles the format: "pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq-https://relayminer.shannon-mainnet.eu.nodefleet.net"
// Returns "unknown" if domain cannot be extracted.
func extractDomainFromEndpointAddr(logger polylog.Logger, endpointAddr string) string {
	// Split by dash to separate the address part from the URL part
	parts := strings.Split(endpointAddr, "-")
	if len(parts) < 2 {
		// No dash found, try to extract domain directly from the entire string
		if domain, err := metricshttp.ExtractEffectiveTLDPlusOne(endpointAddr); err == nil {
			return domain
		}
		logger.Debug().Str("endpoint_addr", endpointAddr).Msg("Could not extract domain from endpoint address - no dash separator found")
		return "unknown"
	}

	// Take everything after the first dash as the URL
	urlPart := strings.Join(parts[1:], "-")

	// Try to extract domain from the URL part
	if domain, err := metricshttp.ExtractEffectiveTLDPlusOne(urlPart); err == nil {
		return domain
	}

	logger.Debug().Str("endpoint_addr", endpointAddr).Str("url_part", urlPart).Msg("Could not extract eTLD+1 from URL part")

	// If domain extraction failed, return unknown
	return "unknown"
}

// extractChainID extracts the chain ID from the interpreter.
// Returns empty string if chain ID cannot be determined.
func extractChainID(logger polylog.Logger, interpreter *qos.EVMObservationInterpreter) string {
	chainID, chainIDFound := interpreter.GetChainID()
	if !chainIDFound {
		// For clarity in metrics, use empty string as the default value when chain ID can't be determined
		chainID = ""
		// This should rarely happen with properly configured EVM observations
		logger.Warn().Msgf("Should happen very rarely: Unable to determine chain ID for EVM metrics: %+v", interpreter)
	}
	return chainID
}

// extractServiceID extracts the service ID from the interpreter.
// Returns empty string if service ID cannot be determined.
func extractServiceID(logger polylog.Logger, interpreter *qos.EVMObservationInterpreter) string {
	serviceID, serviceIDFound := interpreter.GetServiceID()
	if !serviceIDFound {
		// For clarity in metrics, use empty string as the default value when service ID can't be determined
		serviceID = ""
		// This should rarely happen with properly configured EVM observations
		logger.Warn().Msgf("Should happen very rarely: Unable to determine service ID for EVM metrics: %+v", interpreter)
	}
	return serviceID
}

// extractRequestMethod extracts the request method from the interpreter.
// Returns empty string if method cannot be determined.
func extractRequestMethod(logger polylog.Logger, interpreter *qos.EVMObservationInterpreter) string {
	method, methodFound := interpreter.GetRequestMethod()
	if !methodFound {
		// For clarity in metrics, use empty string as the default value when method can't be determined
		method = ""
		// This can happen for invalid requests, but we should still log it
		logger.Debug().Msgf("Should happen very rarely: Unable to determine request method for EVM metrics: %+v", interpreter)
	}
	return method
}

// extractEndpointSelectionMetadata extracts endpoint selection metadata from observations.
// Returns metadata about the endpoint selection process including counts and fallback status.
func extractEndpointSelectionMetadata(interpreter *qos.EVMObservationInterpreter) *qos.EndpointSelectionMetadata {
	// Return the endpoint selection metadata directly from observations
	if interpreter.Observations.EndpointSelectionMetadata != nil {
		return interpreter.Observations.EndpointSelectionMetadata
	}
	// Return empty metadata if not set
	return &qos.EndpointSelectionMetadata{}
}
