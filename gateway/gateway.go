// Package gateway implements components for operating a gateway service.
//
// Protocols (Morse, Shannon):
// - Provide available endpoints for a service
// - Send relays to specific endpoints
//
// Gateways:
// - Select endpoints for relay transmission
//
// QoS Services:
// - Interpret user requests into endpoint payloads
// - Select optimal endpoints for request handling
//
// TODO_MVP(@adshmh): add a README with a diagram of all the above.
// TODO_MVP(@adshmh): add a section for the following packages once they are added: Metrics, Message.
package gateway

import (
	"context"
	"net/http"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	shannonmetrics "github.com/buildwithgrove/path/metrics/protocol/shannon"
)

// Gateway handles end-to-end service requests via HandleHTTPServiceRequest:
// - Receives user request
// - Processes request
// - Returns response
//
// TODO_FUTURE: Current HTTP-only format supports JSONRPC, REST, Websockets
// and gRPC. May expand to other formats in future.
type Gateway struct {
	Logger polylog.Logger

	// HTTPRequestParser is used by the gateway instance to
	// interpret an HTTP request as a pair of service ID and
	// its corresponding QoS instance.
	HTTPRequestParser

	// The Protocol instance is used to fulfill the
	// service requests received by the gateway through
	// sending the service payload to an endpoint.
	Protocol

	// MetricsReporter is used to export metrics based on observations made in handling service requests.
	MetricsReporter RequestResponseReporter

	// DataReporter is used to export, to the data pipeline, observations made in handling service requests.
	// It is declared separately from the `MetricsReporter` to be consistent with the gateway package's role
	// of explicitly defining PATH gateway's components and their interactions.
	DataReporter RequestResponseReporter
}

// HandleHTTPServiceRequest implements PATH gateway's HTTP request processing:
//
// This is written as a template method to allow customization of steps.
// Template pattern allows customization of service steps:
// - Establishing QoS context
// - Sending payload via relay protocols
// Reference: https://en.wikipedia.org/wiki/Template_method_pattern
//
// TODO_FUTURE: Refactor when adding other protocols (e.g. gRPC):
// - Extract generic processing into common method
// - Keep HTTP-specific details separate
func (g Gateway) HandleServiceRequest(ctx context.Context, httpReq *http.Request, w http.ResponseWriter) {
	// build a gatewayRequestContext with components necessary to process requests.
	gatewayRequestCtx := &requestContext{
		logger:              g.Logger,
		gatewayObservations: getUserRequestGatewayObservations(httpReq),
		protocol:            g.Protocol,
		httpRequestParser:   g.HTTPRequestParser,
		metricsReporter:     g.MetricsReporter,
		dataReporter:        g.DataReporter,
		context:             ctx,
	}

	defer func() {
		// Broadcast all observations, e.g. protocol-level, QoS-level, etc. contained in the gateway request context.
		gatewayRequestCtx.BroadcastAllObservations()
	}()

	// Initialize the GatewayRequestContext struct using the HTTP request.
	// e.g. extract the target service ID from the HTTP request.
	err := gatewayRequestCtx.InitFromHTTPRequest(httpReq)
	if err != nil {
		return
	}

	// Determine the type of service request and handle it accordingly.
	switch determineServiceRequestType(httpReq) {
	case websocketServiceRequest:
		g.handleWebSocketRequest(ctx, httpReq, gatewayRequestCtx, w)
	default:
		g.handleHTTPServiceRequest(ctx, httpReq, gatewayRequestCtx, w)
	}
}

// handleHTTPRequest handles a standard HTTP service request.
func (g Gateway) handleHTTPServiceRequest(_ context.Context, httpReq *http.Request, gatewayRequestCtx *requestContext, w http.ResponseWriter) {
	// Record the overall request start time for end-to-end latency tracking
	gatewayRequestCtx.relayStartTime = time.Now()

	defer func() {
		// Write the user-facing HTTP response. This is deliberately not called for websocket requests as they do not return an HTTP response.
		gatewayRequestCtx.WriteHTTPUserResponse(w)
	}()

	// Record setup phase timing - starts before QoS context building
	setupStartTime := time.Now()
	setupStage := "qos_context"   // Track which stage we reach
	cachePerformance := "unknown" // Will be determined based on timing

	// TODO_TECHDEBT(@adshmh): Pass the context with deadline to QoS once it can handle deadlines.
	// Build the QoS context for the target service ID using the HTTP request's payload.
	err := gatewayRequestCtx.BuildQoSContextFromHTTP(httpReq)
	if err != nil {
		// Record setup latency even for failed QoS context building
		setupDuration := time.Since(setupStartTime).Seconds()
		cachePerformance = categorizeSetupCachePerformance(setupDuration)
		shannonmetrics.RecordRequestSetupLatency(string(gatewayRequestCtx.serviceID), setupStage, cachePerformance, setupDuration)
		return
	}

	// Update setup stage - QoS context built successfully
	setupStage = "protocol_context"

	// TODO_MVP(@adshmh): Enhance the protocol interface used by the gateway to provide explicit error classification.
	// Implementation should:
	//   1. Differentiate between user errors (e.g., invalid Service ID in request) and system errors (e.g., endpoint timeout)
	//   2. Add error type field to protocol response structure
	//   3. Pass specific error codes from the protocol back to QoS service
	// This will allow the QoS service to return more helpful diagnostic messages and enable better metrics collection for different failure modes.
	//
	// Build the protocol context for the HTTP request.
	err = gatewayRequestCtx.BuildProtocolContextFromHTTP(httpReq)
	if err != nil {
		// Record setup latency for failed protocol context building
		setupDuration := time.Since(setupStartTime).Seconds()
		cachePerformance = categorizeSetupCachePerformance(setupDuration)
		shannonmetrics.RecordRequestSetupLatency(string(gatewayRequestCtx.serviceID), setupStage, cachePerformance, setupDuration)
		return
	}

	// Setup phase complete - record successful setup latency
	setupDuration := time.Since(setupStartTime).Seconds()
	cachePerformance = categorizeSetupCachePerformance(setupDuration)
	shannonmetrics.RecordRequestSetupLatency(string(gatewayRequestCtx.serviceID), "complete", cachePerformance, setupDuration)

	// Use the gateway request context to process the relay(s) corresponding to the HTTP request.
	// Any returned errors are ignored here and processed by the gateway context in the deferred calls.
	// See the `BroadcastAllObservations` method of `gateway.requestContext` struct for details.
	_ = gatewayRequestCtx.HandleRelayRequest()
}

// handleWebsocketRequest handles WebSocket connection requests
func (g Gateway) handleWebSocketRequest(_ context.Context, httpReq *http.Request, gatewayRequestCtx *requestContext, w http.ResponseWriter) {
	// Build the QoS context for the target service ID using the HTTP request's payload.
	err := gatewayRequestCtx.BuildQoSContextFromWebsocket(httpReq)
	if err != nil {
		return
	}

	// Build the protocol context for the HTTP request.
	err = gatewayRequestCtx.BuildProtocolContextFromHTTP(httpReq)
	if err != nil {
		return
	}

	// Use the gateway request context to process the websocket connection request.
	// Any returned errors are ignored here and processed by the gateway context in the deferred calls.
	// See the `BroadcastAllObservations` method of `gateway.requestContext` struct for details.
	_ = gatewayRequestCtx.HandleWebsocketRequest(httpReq, w)
}

// categorizeSetupCachePerformance categorizes setup performance based on timing patterns.
// This helps identify whether session cache hits/misses are affecting setup time.
func categorizeSetupCachePerformance(setupDurationSeconds float64) string {
	setupDurationMs := setupDurationSeconds * 1000

	switch {
	case setupDurationMs < 50: // Very fast setup, likely all cache hits
		return "all_hits"
	case setupDurationMs < 500: // Moderate setup time, some cache misses
		return "some_misses"
	default: // Slow setup, likely session rollover or cache failures
		return "all_misses"
	}
}
