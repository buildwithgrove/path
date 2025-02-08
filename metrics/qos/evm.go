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
	evmRequestsTotalMetric = "evm_requests_total"
)

func init() {
	prometheus.MustRegister(evmRequestsTotal)
}

var (
	// TODO_MVP(@adshmh): add a `validation` object field to indicate whether
	// the user's request was valid, with two fields:
	//	1. Valid: whether the user's request was valid.
	//	2. Reason: The reason the request is considered invalid, if applicable.
	//
	// evmRequestsTotal counts EVM QoS processed requests with labels:
	//   - chain_id: Chain identifier using EVM QoS
	//   - request_method: JSONRPC method name
	//   - success: Whether request received a valid response
	//
	// Usage:
	// - Monitor EVM requests load across chains and methods
	// - Monitor EVM requests across different PATH instances
	// - Compare requests across different JSONRPC methods or chain IDs (i.e. different chains which use EVM as their QoS)
	evmRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      evmRequestsTotalMetric,
			Help:      "Total number of requests processed by EVM QoS instance(s)",
		},
		[]string{"chain_id", "request_method", "success"},
	)
)

// PublishEVMMetrics exports all EVM-related Prometheus metrics using observations reported by EVM QoS service.
func PublishEVMMetrics(evmObservations *qos.EVMRequestObservations) {
	// Increment request counters with all corresponding labels
	evmRequestsTotal.With(
		prometheus.Labels{
			"chain_id":       evmObservations.GetChainId(),
			"request_method": evmObservations.GetJsonrpcRequest().GetMethod(),
			"success":        fmt.Sprintf("%t", getEVMRequestSuccess(evmObservations)),
		},
	).Inc()
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
