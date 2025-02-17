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
	requestsTotalMetric = "evm_requests_total"
)

func init() {
	prometheus.MustRegister(requestsTotal)
}

var (
	// TODO_MVP(@adshmh): add a `validation` object field to indicate whether
	// the user's request was valid, with two fields:
	//	1. Valid: whether the user's request was valid.
	//	2. Reason: The reason the request is considered invalid, if applicable.
	//
	// TODO_MVP(@adshmh): Track endpoint responses separately from requests if/when retries are implemented,
	// since a single request may generate multiple responses due to retry attempts.
	//
	// requestsTotal counts EVM QoS processed requests with labels:
	//   - chain_id: Chain identifier using EVM QoS
	//   - request_method: JSONRPC method name
	//   - success: Whether request received a valid response
	//   - invalid_response_reason: the reason why an endpoint response failed QoS validation.
	//
	// Usage:
	// - Monitor EVM requests load across chains and methods
	// - Monitor EVM requests across different PATH instances
	// - Compare requests across different JSONRPC methods or chain IDs (i.e. different chains which use EVM as their QoS)
	// - Compare endpoint response validation failures across JSONRPC methods or chain IDs.
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      requestsTotalMetric,
			Help:      "Total number of requests processed by EVM QoS instance(s)",
		},
		[]string{"chain_id", "request_method", "success", "invalid_response_reason"},
	)
)

// PublishMetrics exports all EVM-related Prometheus metrics using observations reported by EVM QoS service.
func PublishMetrics(
	observations *qos.EVMRequestObservations,
) {
	// Increment request counters with all corresponding labels
	requestsTotal.With(
		prometheus.Labels{
			"chain_id":                observations.GetChainId(),
			"request_method":          observations.GetJsonrpcRequest().GetMethod(),
			"success":                 fmt.Sprintf("%t", getRequestSuccess(observations)),
			"invalid_response_reason": getEndpointResponseValidationFailureReason(observations),
		},
	).Inc()
}

// getRequestSuccess returns true if the request is assumed successful.
// The request is assumed successful if any endpoint response is marked as valid.
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

// TODO_MVP(@adshmh): Add Grafana panel(s) to visualize the validation failure reason label in Local development mode.
//
// TODO_MVP(@adshmh): When retry functionality is added, refactor to evaluate QoS based on a single endpoint response rather than
// aggregated observations.
//
// getEndpointResponseValidationFailureReason returns why the endpoint response failed QoS validation.
func getEndpointResponseValidationFailureReason(
	observations *qos.EVMRequestObservations,
) string {
	for _, observation := range observations.GetEndpointObservations() {
		if response := extractEndpointResponseFromObservation(observation); response != nil {
			return qos.EVMResponseInvalidReason_name[int32(response.GetInvalidReason())]
		}
	}

	return ""
}
