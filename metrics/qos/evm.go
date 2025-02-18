package qos

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/buildwithgrove/path/observation/qos"
)

const (
	// The POSIX process that emits metrics
	pathProcess = "path"

	// The list of metrics being tracked for EVM QoS
	evmRequestsTotalMetric                 = "evm_requests_total"
	evmRequestsValidationErrorsTotalMetric = "evm_request_validation_errors_total"
)

func init() {
	prometheus.MustRegister(evmRequestsTotal)
	prometheus.MustRegister(evmRequestValidationErrorsTotal)
}

var (
	// evmRequestsTotal tracks the total EVM requests processed.
	// Labels:
	//   - chain_id: Target EVM chain identifier
	//   - valid_request: Whether request parsing succeeded
	//   - request_method: JSON-RPC method name
	//   - success: Whether a valid response was received
	//
	// Use to analyze:
	//   - Request volume by chain and method
	//   - Success rates across different PATH deployment regions
	//   - Method usage patterns across chains
	//   - End-to-end request success rates
	evmRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      evmRequestsTotalMetric,
			Help:      "Total number of requests processed by EVM QoS instance(s)",
		},
		[]string{"chain_id", "valid_request", "request_method", "success"},
	)

	// evmRequestValidationFailuresTotal tracks validation failures of incoming EVM requests.
	// Labels:
	//   - chain_id: Target EVM chain identifier
	//   - validation_error_kind: Validation error kind
	//
	// Use to analyze:
	//   - Common request validation issues
	//   - Per-chain validation error patterns
	evmRequestValidationErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      evmRequestsValidationErrorsTotalMetric,
			Help:      "Total requests that failed validation before being sent to any endpoints, e.g. malformed JSON-RPC or parse errors",
		},
		[]string{"chain_id", "validation_error_kind"},
	)
)

// PublishEVMMetrics exports all EVM-related Prometheus metrics using observations reported by EVM QoS service.
func PublishEVMMetrics(evmObservations *qos.EVMRequestObservations) {
	isRequestValid, requestValidationErrorKind := extractEVMRequestValidationStatus(evmObservations)

	// Increment request counters with all corresponding labels
	evmRequestsTotal.With(
		prometheus.Labels{
			"chain_id":       evmObservations.GetChainId(),
			"valid_request":  fmt.Sprintf("%t", isRequestValid),
			"request_method": evmObservations.GetJsonrpcRequest().GetMethod(),
			"success":        fmt.Sprintf("%t", getEVMRequestSuccess(evmObservations)),
		},
	).Inc()

	// Increment the request validation failure counter.
	if !isRequestValid {
		evmRequestValidationErrorsTotal.With(
			prometheus.Labels{
				"chain_id":              evmObservations.GetChainId(),
				"validation_error_kind": requestValidationErrorKind,
			},
		).Inc()
	}
}

// getEVMRequestSuccess returns true if the request is assumed successful.
// The request is assumed successful if any endpoint response is marked as valid.
func getEVMRequestSuccess(evmObservations *qos.EVMRequestObservations) bool {
	for _, observation := range evmObservations.GetEndpointObservations() {
		responses := []interface {
			GetValid() bool
		}{
			observation.GetChainIdResponse(),
			observation.GetBlockNumberResponse(),
			observation.GetUnrecognizedResponse(),
		}

		for _, response := range responses {
			if response != nil && response.GetValid() {
				return true
			}
		}
	}

	return false
}

// extractEVMRequestValidationStatus interprets validation results from the request observations.
// Returns (true, "") if valid, or (false, failureReason) if invalid.
func extractEVMRequestValidationStatus(evmObservations *qos.EVMRequestObservations) (bool, string) {
	reasonEnum := evmObservations.GetRequestValidationErrorKind()

	// Valid request
	if reasonEnum == qos.EVMRequestValidationErrorKind_EVM_REQUEST_VALIDATION_ERROR_KIND_UNSPECIFIED {
		return true, ""
	}

	return false, qos.EVMRequestValidationErrorKind_name[int32(reasonEnum)]
}
