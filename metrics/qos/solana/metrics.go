package solana

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/buildwithgrove/path/observation/qos"
)

const (
	// The POSIX process that emits metrics
	pathProcess = "path"

	// The list of metrics being tracked for Solana QoS
	requestsTotalMetric = "solana_requests_total"
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
	// requestsTotal tracks the total Solana requests processed.
	// Labels:
	//   - chain_id: Target Solana chain identifier
	//   - service_id: Service ID of the EVM QoS instance
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
			Help:      "Total number of requests processed by Solana QoS instance(s)",
		},
		[]string{"chain_id", "service_id", "request_method", "success", "error_type", "http_status_code"},
	)
)

// PublishMetrics exports all EVM-related Prometheus metrics using observations reported by EVM QoS service.
// It logs errors for unexpected conditions that should never occur in normal operation.
func PublishMetrics(logger polylog.Logger, observations *qos.SolanaRequestObservations) {
	logger = logger.With("method", "PublishSolanaMetrics")

	// Skip if observations is nil.
	// This should never happen as PublishQoSMetrics uses nil checks to identify which QoS service produced the observations.
	if observations == nil {
		logger.Error().Msg("SHOULD NEVER HAPPEN: Unable to publish Solana metrics: received nil observations.")
		return
	}

	// Create an interpreter for the observations
	interpreter := &qos.SolanaObservationInterpreter{
		Logger:       logger,
		Observations: observations,
	}

	// Increment request counters with all corresponding labels
	requestsTotal.With(
		prometheus.Labels{
			"chain_id":         interpreter.GetChainID(),
			"service_id":       interpreter.GetServiceID(),
			"request_method":   interpreter.GetRequestMethod(),
			"success":          fmt.Sprintf("%t", interpreter.IsRequestSuccessful()),
			"error_type":       interpreter.GetRequestErrorType(),
			"http_status_code": fmt.Sprintf("%d", interpreter.GetRequestHTTPStatus()),
		},
	).Inc()
}
