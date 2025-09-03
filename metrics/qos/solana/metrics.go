package solana

import (
	"fmt"
	"strings"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/prometheus/client_golang/prometheus"

	metricshttp "github.com/buildwithgrove/path/metrics/http"
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
	// TODO_MVP(@adshmh):
	// - Add 'errorSubType' label for more granular error categorization
	// - Use 'errorType' for broad error categories (e.g., request validation, protocol error)
	// - Use 'errorSubType' for specifics (e.g., endpoint maxed out, timed out)
	// - Remove 'success' label (success = absence of errorType)
	// - Update EVM observations proto files and add interpreter support
	//
	// TODO_MVP(@adshmh):
	// - Track endpoint responses separately from requests if/when retries are implemented
	//   (A single request may generate multiple responses due to retries)
	//
	// requestsTotal tracks total Solana requests processed
	//
	// - Labels:
	//   - chain_id: Target Solana chain identifier
	//   - service_id: Service ID of the Solana QoS instance
	//   - request_origin: origin of the request: User or Hydrator.
	//   - request_method: JSON-RPC method name
	//   - success: Whether a valid response was received
	//   - error_type: Type of error if request failed (empty for success)
	//   - http_status_code: HTTP status code returned to user
	//   - endpoint_domain: Effective TLD+1 domain of the endpoint that served the request
	//
	// - Use cases:
	//   - Analyze request volume by chain and method
	//   - Track success rates across PATH deployment regions
	//   - Identify method usage patterns per chain
	//   - Measure end-to-end request success rates
	//   - Review error types by method and chain
	//   - Examine HTTP status code distribution
	//   - Performance and reliability by endpoint domain
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: pathProcess,
			Name:      requestsTotalMetric,
			Help:      "Total number of requests processed by Solana QoS instance(s)",
		},
		[]string{"chain_id", "service_id", "request_origin", "request_method", "success", "error_type", "http_status_code", "endpoint_domain"},
	)
)

// PublishMetrics:
// - Exports all Solana-related Prometheus metrics using observations from Solana QoS service
// - Logs errors for unexpected (should-never-happen) conditions
func PublishMetrics(logger polylog.Logger, observations *qos.SolanaRequestObservations) {
	logger = logger.With("method", "PublishMetricsSolana")

	// Skip if observations is nil.
	// This should never happen as PublishQoSMetrics uses nil checks to identify which QoS service produced the observations.
	if observations == nil {
		logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Unable to publish Solana metrics: received nil observations.")
		return
	}

	// Create an interpreter for the observations
	interpreter := &qos.SolanaObservationInterpreter{
		Logger:       logger,
		Observations: observations,
	}

	// Extract endpoint domain
	endpointDomain := extractEndpointDomain(logger, interpreter)

	// Increment request counters with all corresponding labels
	requestsTotal.With(
		prometheus.Labels{
			"chain_id":         interpreter.GetChainID(),
			"service_id":       interpreter.GetServiceID(),
			"request_origin":   observations.GetRequestOrigin().String(),
			"request_method":   interpreter.GetRequestMethod(),
			"success":          fmt.Sprintf("%t", interpreter.IsRequestSuccessful()),
			"error_type":       interpreter.GetRequestErrorType(),
			"http_status_code": fmt.Sprintf("%d", interpreter.GetRequestHTTPStatus()),
			"endpoint_domain":  endpointDomain,
		},
	).Inc()
}

// extractEndpointDomain extracts the endpoint domain from the selected endpoint in observations.
// Returns "unknown" if domain cannot be determined.
func extractEndpointDomain(logger polylog.Logger, interpreter *qos.SolanaObservationInterpreter) string {
	// Get endpoint observations and extract domain from the last one used
	endpointObservations := interpreter.Observations.GetEndpointObservations()
	if len(endpointObservations) == 0 {
		return "unknown"
	}

	// Use the last endpoint observation (most recent endpoint used, similar to Shannon metrics pattern)
	lastObs := endpointObservations[len(endpointObservations)-1]
	return extractDomainFromEndpointAddr(logger, lastObs.GetEndpointAddr())
}

// extractDomainFromEndpointAddr extracts the eTLD+1 domain from an endpoint address.
// Handles the format: "pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq-https://relayminer.shannon-mainnet.eu.nodefleet.net"
// Returns "unknown" if domain cannot be extracted.
func extractDomainFromEndpointAddr(logger polylog.Logger, endpointAddr string) string {
	// Split by dash to separate the address part from the URL part
	parts := strings.Split(endpointAddr, "-")
	if len(parts) < 2 {
		// No dash found, try to extract domain directly from the entire string
		if domain, err := metricshttp.ExtractEffectiveTLDPlusOne(endpointAddr); err == nil {
			return domain
		}
		logger.Debug().Str("endpoint_addr", endpointAddr).Msg("Could not extract domain from endpoint address - no dash separator found")
		return "unknown"
	}

	// Take everything after the first dash as the URL
	urlPart := strings.Join(parts[1:], "-")

	// Try to extract domain from the URL part
	if domain, err := metricshttp.ExtractEffectiveTLDPlusOne(urlPart); err == nil {
		return domain
	}

	logger.Debug().Str("endpoint_addr", endpointAddr).Str("url_part", urlPart).Msg("Could not extract eTLD+1 from URL part")

	// If domain extraction failed, return unknown
	return "unknown"
}
