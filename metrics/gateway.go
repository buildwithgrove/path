package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/buildwithgrove/path/observation"
)

// See the metrics initialization below for details.
const (
	// The POSIX process that emits metrics
	pathProcess = "path"

	// The list of metrics being tracked for gateway-level observations
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
	// Increment on each service request with labels:
	//   - service_id: Identifies the service
	//   - request_type: "organic" or "synthetic"
	//   - request_error_kind: request error kind, if any.
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
		[]string{"service_id", "request_type", "request_error_kind"},
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
			// This means we need a higher resolution (smaller buckets and more granularity) for the lower values,
			// and less resolution (big buckets and low granularity) for the higher values because it'll have less
			// data points.
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
	relayResponseSizeBytes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: pathProcess,
			Name:      responseSizeBytes,
			Help:      "Histogram of response sizes in bytes for performance analysis.",
			// TODO_IMPROVE: Consider configuring bucket sizes externally for flexible adjustments
			// in response to different data patterns or deployment scenarios.
			Buckets: []float64{100, 500, 1000, 5000, 10000, 50000},
		},
		[]string{"service_id"},
	)

	// TODO_MVP(@adshmh): Add a serviceRequestSize metric once the `request` package is refactored to
	// fully encapsulate the task of dealing with the HTTP request, including:
	//	1. Reading of all HTTP headers: Target-Service-Id, etc.
	//	2. Reading of the HTTP request's body
	//	3. Building an HTTP observation using the extracted data.
	// This will also involve a small refactor on protocol and qos packages to accept a custom struct
	// rather than an HTTP request.
)

// publishGatewayMetrics publishes all metrics related to gateway-level observations.
// Returns:
// - true if the request was valid.
// - false otherwise.
func publishGatewayMetrics(gatewayObservations *observation.GatewayObservations) bool {
	serviceID := gatewayObservations.GetServiceId()

	var requestErrorKind string
	requestErr := gatewayObservations.GetRequestError()
	if requestErr != nil {
		requestErrorKind = requestErr.GetErrorKind().String()
	}

	// Increment on each service request with labels:
	//   - service_id: Identifies the service
	//   - request_type: "organic" or "synthetic"
	//   - request_error_kind: any gateway-level request errors: e.g. no service ID specified in request's HTTP headers.
	relaysTotal.With(
		prometheus.Labels{
			"service_id":         serviceID,
			"request_type":       observation.RequestType_name[int32(gatewayObservations.GetRequestType())],
			"request_error_kind": requestErrorKind,
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

	// Return the validity status of the request.
	return requestErr == nil
}
