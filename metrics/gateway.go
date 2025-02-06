package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/buildwithgrove/path/observation"
)

// See the metrics initialization below for details.
const (
	pathProcess = "path"

	requestsTotal        = "requests_total"
	responseSizeBytes    = "response_size_bytes"
	relayDurationSeconds = "relay_duration_seconds"
)

func init() {
	prometheus.MustRegister(relaysTotal)
	prometheus.MustRegister(relaysDurationSeconds)
	prometheus.MustRegister(relayResponseSizeBytes)
}

var (
	// relaysTotal is a counter tracking processed requests per PATH instance.
	// It increments on each service request with labels:
	//   - service_id: Identifies the service
	//   - request_type: "organic" or "synthetic"
	//
	// Usage:
	// - Monitor total request load.
	// - Compare requests across services or PATH instances.
	relaysTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      requestsTotal,
			Help:      "Total number of requests processed, labeled by service ID.",
		},
		[]string{"service_id", "request_type"},
	)

	// relaysDurationSeconds measures request processing duration with the service_id label.
	// Histogram buckets from 0.1s to 15s capture performance from fast responses to timeouts.
	//
	// Usage:
	// - Analyze typical response times and long-tail latency issues.
	// - Compare performance across services.
	// - Compare performance under different loads.
	relaysDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: pathProcess,
			Name:      relayDurationSeconds,
			Help:      "Histogram of request processing time (duration) in seconds",
			// Buckets are selected as: [0, 0.1), [0.1, 0.5), [0.5, 1), [1, 2), [2, 5), [5, 15)
			// This is because the request processing time is expected to be normally distributed.
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 15},
		},
		[]string{"service_id"},
	)

	// relayResponseSizeBytes tracks response payload sizes in bytes.
	// Histogram buckets from 100B to 50KB capture size distribution.
	//
	// Usage:
	// 	- Performance tuning to understand skew of data distribution
	//	- Visibility into small & large response size distribution
	//
	// TODO_TECHDEBT: Consider configuring bucket sizes externally for flexible adjustments
	// in response to different data patterns or deployment scenarios.
	relayResponseSizeBytes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: pathProcess,
			Name:      responseSizeBytes,
			Help:      "Histogram of response sizes in bytes for performance analysis.",
			Buckets:   []float64{100, 500, 1000, 5000, 10000, 50000},
		},
		[]string{"service_id"},
	)

	// TODO_MVP(@adshmh): Add a serviceRequestSize metric once the `request` package is refactored to
	// fully encapsulate the task of dealing with the HTTP request, including:
	//	!. Reading of all HTTP headers: target-service-id, etc.
	//	2. Reading of the HTTP request's body
	//	3. Building an HTTP observation using the extracted data.
	// This will also involve a small refactor on protocol and qos packages to accept a custom struct
	// rather than an HTTP request.
)

// publishGatewayMetrics publishes all metrics related to gateway-level observations.
func publishGatewayMetrics(gatewayObservations *observation.GatewayObservations) {
	serviceID := gatewayObservations.GetServiceId()

	// Increment request counter with the following labels:
	// 	- request_type
	// 	- service_id
	relaysTotal.With(
		prometheus.Labels{
			"service_id":   serviceID,
			"request_type": observation.RequestType_name[int32(gatewayObservations.GetRequestType())],
		},
	).Inc()

	// Publish request duration in seconds with the following labels
	// 	- service_id
	duration := gatewayObservations.GetCompletedTime().AsTime().Sub(gatewayObservations.GetReceivedTime().AsTime()).Seconds()
	relaysDurationSeconds.With(
		prometheus.Labels{
			"service_id": serviceID,
		},
	).Observe(duration)

	// Publish response_size in bytes with the following labels
	// 	- service_id
	relayResponseSizeBytes.With(
		prometheus.Labels{
			"service_id": serviceID,
		},
	).Observe(float64(gatewayObservations.GetResponseSize()))
}
