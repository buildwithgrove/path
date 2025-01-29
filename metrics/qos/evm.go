package qos

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/buildwithgrove/path/observation/qos"
)

const (
	pathProcess            = "path"
	evmRequestsTotalMetric = "evm_requests_total"
)

func init() {
	prometheus.MustRegister(evmRequestsTotal)
}

var (
	// evmRequestsTotal is a Counter metric for requests processed by an EVM QoS instance.
	// It is incremented to track service requests and is labeled by:
	// 	- `chain_id`: the ID of the chain using EVM as its QoS
	//	- `request_method`: JSONRPC request's method field)
	///     - `success`: whether the request was considered successfull, i.e. received a valid response.
	// This is essential for monitoring EVM requests on different PATH instances.
	//
	// Usage:
	// - Monitor EVM requests load.
	// - Compare requests across different JSONRPC methods or chain IDs (i.e. different chains which use EVM as their QoS)
	evmRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      evmRequestsTotalMetric,
			Help:      "Total number of requests processed by EVM QoS instance(s), labeled by chain ID and request method.",
		},
		[]string{"chain_id", "request_method", "success"},
	)
)

// PublishEVMMetrics exports all EVM-related Prometheus metrics using observations reported by EVM QoS service.
func PublishEVMMetrics(evmObservations *qos.EVMRequestObservations) {
	// Update request counters with request_method and chain_id labels.
	evmRequestsTotal.With(
		prometheus.Labels{
			"success":        fmt.Sprintf("%t", getEVMRequestSuccess(evmObservations)),
			"chain_id":       evmObservations.GetChainId(),
			"request_method": evmObservations.GetJsonrpcRequest().GetMethod(),
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
