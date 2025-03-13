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
	//   - http_status_code: The HTTP status code returned to the user
	//
	// Use to analyze:
	//   - Request volume by chain and method
	//   - Success rates across different PATH deployment regions
	//   - Method usage patterns across chains
	//   - End-to-end request success rates
	//   - Response validation errors by JSON-RPC method and chain
	//   - HTTP status code distribution
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      requestsTotalMetric,
			Help:      "Total number of requests processed by EVM QoS instance(s)",
		},
		[]string{"chain_id", "valid_request", "request_method", "success", "invalid_response_reason", "http_status_code"},
	)

	// requestValidationErrorsTotal tracks validation errors of incoming EVM requests.
	// Labels:
	//   - chain_id: Target EVM chain identifier
	//   - validation_error_kind: Validation error kind
	//   - http_status_code: The HTTP status code returned to the user
	//
	// Use to analyze:
	//   - Common request validation issues
	//   - Per-chain validation error patterns
	//   - HTTP status code distribution for validation errors
	requestValidationErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      requestsValidationErrorsTotalMetric,
			Help:      "Total requests that failed validation BEFORE being sent to any endpoints; request was terminated in PATH. E.g. malformed JSON-RPC or parse errors",
		},
		[]string{"chain_id", "validation_error_kind", "http_status_code"},
	)
)

// PublishMetrics exports all EVM-related Prometheus metrics using observations reported by EVM QoS service.
func PublishMetrics(observations *qos.EVMRequestObservations) {
	if observations == nil {
		return
	}

	req := newRequestAdapter(observations)
	isRequestValid := req.GetRequestValidationError() == nil

	// Get request method - handle the case where jsonrpc_request might be nil for invalid requests
	var requestMethod string
	if jsonReq := observations.GetJsonrpcRequest(); jsonReq != nil {
		requestMethod = jsonReq.GetMethod()
	}

	// Get HTTP status code that would be returned to the user
	httpStatusCode := getHTTPStatusCodeFromObservations(observations)
	httpStatusCodeStr := fmt.Sprintf("%d", httpStatusCode)

	// Increment request counters with all corresponding labels
	requestsTotal.With(
		prometheus.Labels{
			"chain_id":                observations.GetChainId(),
			"valid_request":           fmt.Sprintf("%t", isRequestValid),
			"request_method":          requestMethod,
			"success":                 fmt.Sprintf("%t", req.IsSuccessful()),
			"invalid_response_reason": getEndpointResponseValidationFailureReason(observations),
			"http_status_code":        httpStatusCodeStr,
		},
	).Inc()

	// Only export validation error metrics for invalid requests
	if isRequestValid {
		return
	}

	// Get the validation error kind as a string
	var errorKind string
	if validationErr := req.GetRequestValidationError(); validationErr != nil {
		errorKind = validationErr.String()
	} else {
		errorKind = "UNKNOWN"
	}

	// Increment the request validation failure counter
	requestValidationErrorsTotal.With(
		prometheus.Labels{
			"chain_id":              observations.GetChainId(),
			"validation_error_kind": errorKind,
			"http_status_code":      httpStatusCodeStr,
		},
	).Inc()
}
