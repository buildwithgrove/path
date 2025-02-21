package evm

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/buildwithgrove/path/observation/qos"
)

const (
	// The POSIX process that emits metrics
	pathProcess = "path"

	// The list of metrics being tracked for EVM QoS
	requestsTotalMetric                 = "evm_requests_total"
	requestsValidationErrorsTotalMetric = "evm_request_validation_errors_total"
)

func init() {
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestValidationErrorsTotal)
}

var (
	// TODO_MVP(@adshmh): Track endpoint responses separately from requests if/when retries are implemented,
	// since a single request may generate multiple responses due to retry attempts.
	//
	// requestsTotal tracks the total EVM requests processed.
	// Labels:
	//   - chain_id: Target EVM chain identifier
	//   - valid_request: Whether request parsing succeeded
	//   - request_method: JSON-RPC method name
	//   - success: Whether a valid response was received
	//   - invalid_response_reason: the reason why an endpoint response failed QoS validation.
	//
	// Use to analyze:
	//   - Request volume by chain and method
	//   - Success rates across different PATH deployment regions
	//   - Method usage patterns across chains
	//   - End-to-end request success rates
	//   - Response validation errors by JSON-RPC method and chain
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      requestsTotalMetric,
			Help:      "Total number of requests processed by EVM QoS instance(s)",
		},
		[]string{"chain_id", "valid_request", "request_method", "success", "invalid_response_reason"},
	)

	// requestValidationErrorsTotal tracks validation errors of incoming EVM requests.
	// Labels:
	//   - chain_id: Target EVM chain identifier
	//   - validation_error_kind: Validation error kind
	//
	// Use to analyze:
	//   - Common request validation issues
	//   - Per-chain validation error patterns
	requestValidationErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      requestsValidationErrorsTotalMetric,
			Help:      "Total requests that failed validation BEFORE being sent to any endpoints; request was terminated in PATH. E.g. malformed JSON-RPC or parse errors",
		},
		[]string{"chain_id", "validation_error_kind"},
	)
)

// PublishMetrics exports all EVM-related Prometheus metrics using observations reported by EVM QoS service.
func PublishMetrics(
	observations *qos.EVMRequestObservations,
) {
	isRequestValid, requestValidationError := extractRequestValidationStatus(observations)

	// Increment request counters with all corresponding labels
	requestsTotal.With(
		prometheus.Labels{
			"chain_id":                observations.GetChainId(),
			"valid_request":           fmt.Sprintf("%t", isRequestValid),
			"request_method":          observations.GetJsonrpcRequest().GetMethod(),
			"success":                 fmt.Sprintf("%t", getRequestSuccess(observations)),
			"invalid_response_reason": getEndpointResponseValidationFailureReason(observations),
		},
	).Inc()

	// Only export validation error metrics for invalid requests
	if isRequestValid {
		return
	}

	// Increment the request validation failure counter.
	requestValidationErrorsTotal.With(
		prometheus.Labels{
			"chain_id":              observations.GetChainId(),
			"validation_error_kind": requestValidationError,
		},
	).Inc()
}

// getRequestSuccess checks if any endpoint provided a valid response.
// Alternatively, It can be thought of "isAnyResponseSuccessful".
func getRequestSuccess(
	observations *qos.EVMRequestObservations,
) bool {
	for _, observation := range observations.GetEndpointObservations() {
		if response := extractEndpointResponseFromObservation(observation); response != nil && response.GetValid() {
			return true
		}
	}

	return false
}

// TODO_MVP(@adshmh): When retry functionality is added, refactor to evaluate QoS based on a single endpoint response rather than
// aggregated observations.
//
// getEndpointResponseValidationFailureReason returns why the endpoint response failed QoS validation.
func getEndpointResponseValidationFailureReason(
	observations *qos.EVMRequestObservations,
) string {
	for _, observation := range observations.GetEndpointObservations() {
		if response := extractEndpointResponseFromObservation(observation); response != nil {
			return qos.EVMResponseValidationError_name[int32(response.GetResponseValidationError())]
		}
	}

	return ""
}

// extractRequestValidationStatus interprets validation results from the request observations.
// Returns (true, "") if valid, or (false, failureReason) if invalid.
func extractRequestValidationStatus(observations *qos.EVMRequestObservations) (bool, string) {
	reasonEnum := observations.GetRequestValidationError()

	// Valid request
	if reasonEnum == qos.EVMRequestValidationError_EVM_REQUEST_VALIDATION_ERROR_UNSPECIFIED {
		return true, ""
	}

	return false, qos.EVMRequestValidationError_name[int32(reasonEnum)]
}
