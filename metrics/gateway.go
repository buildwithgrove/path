package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/buildwithgrove/path/observation"
)

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
	// relaysTotal is a Counter metric for requests processed by a PATH instance.
	// It increments to track service requests and is labeled by 'service_id' and 'request_type' (organic or synthetic),
	// essential for monitoring load and traffic on different PATH instances and services.
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

	// relaysDurationSeconds observes request durations in the gateway.
	// This histogram, labeled by 'service_id', measures response times,
	// vital for performance analysis under different loads.
	//
	// Buckets:
	// - 0.1s to 15s range, capturing response times from very fast to upper limit.
	//
	// Usage:
	// - Analyze typical response times and long-tail latency issues.
	// - Compare performance across services.
	relaysDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: pathProcess,
			Name:      relayDurationSeconds,
			Help:      "Histogram of request durations for performance analysis.",
			Buckets:   []float64{0.1, 0.5, 1, 2, 5, 15},
		},
		[]string{"service_id"},
	)

	// relayResponseSizeBytes is a histogram metric for observing response size distribution.
	// It counts responses in bytes, with buckets:
	// - 100 bytes to 50,000 bytes, capturing a range from small to large responses.
	// This data helps in accurately representing response size distribution and is vital
	// for performance tuning.
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
	// Update request counters with request_type and service ID labels.
	relaysTotal.With(
		prometheus.Labels{
			"service_id":   serviceID,
			"request_type": observation.RequestType_name[int32(gatewayObservations.GetRequestType())],
		},
	).Inc()

	// Publish duration of the request
	duration := gatewayObservations.GetCompletedTime().AsTime().Sub(gatewayObservations.GetReceivedTime().AsTime()).Seconds()
	relaysDurationSeconds.With(
		prometheus.Labels{
			"service_id": serviceID,
		},
	).Observe(duration)

	// Publish response size
	relayResponseSizeBytes.With(
		prometheus.Labels{
			"service_id": serviceID,
		},
	).Observe(float64(gatewayObservations.GetResponseSize()))
}
