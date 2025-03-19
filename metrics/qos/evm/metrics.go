package evm

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/buildwithgrove/path/observation/qos"
	"github.com/pokt-network/poktroll/pkg/polylog"
)

const (
	// The POSIX process that emits metrics
	pathProcess = "path"

	// The list of metrics being tracked for EVM QoS
	requestsTotalMetric = "evm_requests_total"
)

func init() {
	prometheus.MustRegister(requestsTotal)
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
	//   - request_method: JSON-RPC method name
	//   - success: Whether a valid response was received
	//   - error_type: Type of error if request failed (or "" for successful requests)
	//   - http_status_code: The HTTP status code returned to the user
	//
	// Use to analyze:
	//   - Request volume by chain and method
	//   - Success rates across different PATH deployment regions
	//   - Method usage patterns across chains
	//   - End-to-end request success rates
	//   - Error types by JSON-RPC method and chain
	//   - HTTP status code distribution
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      requestsTotalMetric,
			Help:      "Total number of requests processed by EVM QoS instance(s)",
		},
		[]string{"chain_id", "request_method", "success", "error_type", "http_status_code"},
	)
)

// PublishMetrics exports all EVM-related Prometheus metrics using observations reported by EVM QoS service.
// It logs errors for unexpected conditions that should never occur in normal operation.
func PublishMetrics(logger polylog.Logger, observations *qos.EVMRequestObservations) {
	// Skip if observations is nil.
	// This should never happen as PublishQoSMetrics uses nil checks to identify which QoS service produced the observations.
	if observations == nil {
		logger.Error().Msg("Unable to publish EVM metrics: received nil observations - this should never happen")
		return
	}

	// Create an interpreter for the observations
	interpreter := &qos.EVMObservationInterpreter{
		Observations: observations,
	}

	// Extract chain ID
	chainID := extractChainID(logger, interpreter)

	// Extract request method
	method := extractRequestMethod(logger, interpreter)

	// Get request status
	statusCode, requestError, err := interpreter.GetRequestStatus()

	// If we couldn't get status info due to missing observations, skip metrics
	// This should never happen if the observations are properly initialized
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get request status for EVM metrics - this indicates a programming error")
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
			"chain_id":         chainID,
			"request_method":   method,
			"success":          fmt.Sprintf("%t", requestError == nil),
			"error_type":       errorType,
			"http_status_code": fmt.Sprintf("%d", statusCode),
		},
	).Inc()
}

// extractChainID extracts the chain ID from the interpreter
// Returns empty string if chain ID cannot be determined
func extractChainID(logger polylog.Logger, interpreter *qos.EVMObservationInterpreter) string {
	chainID, chainIDFound := interpreter.GetChainID()
	if !chainIDFound {
		// For clarity in metrics, use empty string as the default value when chain ID can't be determined
		chainID = ""
		// This should rarely happen with properly configured EVM observations
		logger.Warn().Msg("Unable to determine chain ID for EVM metrics")
	}
	return chainID
}

// extractRequestMethod extracts the request method from the interpreter
// Returns empty string if method cannot be determined
func extractRequestMethod(logger polylog.Logger, interpreter *qos.EVMObservationInterpreter) string {
	method, methodFound := interpreter.GetRequestMethod()
	if !methodFound {
		// For clarity in metrics, use empty string as the default value when method can't be determined
		method = ""
		// This can happen for invalid requests, but we should still log it
		logger.Debug().Msg("Unable to determine request method for EVM metrics")
	}
	return method
}
