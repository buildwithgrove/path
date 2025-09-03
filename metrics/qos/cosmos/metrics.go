package cosmos

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

	// The list of metrics being tracked for Cosmos SDK QoS
	requestsTotalMetric = "cosmos_sdk_requests_total"
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
	// requestsTotal tracks total Cosmos SDK requests processed
	//
	// - Labels:
	//   - cosmos_chain_id: Target Cosmos chain identifier
	//   - evm_chain_id: Target EVM chain identifier for Cosmos chains with native EVM support, e.g. XRPLEVM, evmos, etc...
	//   - service_id: Service ID of the Cosmos SDK QoS instance
	//   - request_origin: origin of the request: User or Hydrator.
	//   - rpc_type: Backend service type (JSONRPC, REST, COMETBFT)
	//   - request_method: Cosmos SDK RPC method name (e.g., health, status)
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
			Help:      "Total number of requests processed by Cosmos SDK QoS instance(s)",
		},
		[]string{"cosmos_chain_id", "evm_chain_id", "service_id", "request_origin", "rpc_type", "request_method", "success", "error_type", "http_status_code", "endpoint_domain"},
	)
)

// PublishMetrics:
// - Exports all Cosmos SDK-related Prometheus metrics using observations from Cosmos SDK QoS service
// - Logs errors for unexpected (should-never-happen) conditions
func PublishMetrics(logger polylog.Logger, observations *qos.CosmosRequestObservations) {
	logger = logger.With("method", "PublishMetricsCosmosSDK")

	// Skip if observations is nil.
	// This should never happen as PublishQoSMetrics uses nil checks to identify which QoS service produced the observations.
	if observations == nil {
		logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Unable to publish Cosmos SDK metrics: received nil observations.")
		return
	}

	// Create an interpreter for the observations
	interpreter := &qos.CosmosSDKObservationInterpreter{
		Logger:       logger,
		Observations: observations,
	}

	// Extract endpoint domain
	endpointDomain := extractEndpointDomain(logger, interpreter)

	// Increment request counters with all corresponding labels
	requestsTotal.With(
		prometheus.Labels{
			"cosmos_chain_id":  interpreter.GetCosmosChainID(),
			"evm_chain_id":     interpreter.GetEVMChainID(),
			"service_id":       interpreter.GetServiceID(),
			"request_origin":   observations.GetRequestOrigin().String(),
			"rpc_type":         interpreter.GetRPCType(),
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
func extractEndpointDomain(logger polylog.Logger, interpreter *qos.CosmosSDKObservationInterpreter) string {
	// Get endpoint observations and extract domain from the first available one
	endpointObservations := interpreter.Observations.GetEndpointObservations()
	if len(endpointObservations) == 0 {
		return "unknown"
	}

	// Use the last endpoint observation (most recent/final endpoint used)
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
