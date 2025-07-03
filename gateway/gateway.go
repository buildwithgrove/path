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

const (
	requestSetupStageQoSContext      = "qos_context"
	requestSetupStageProtocolContext = "protocol_context"
	requestSetupStageComplete        = "complete"
)

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

func (g Gateway) handleHTTPServiceRequest(
	_ context.Context,
	httpReq *http.Request,
	gatewayRequestCtx *requestContext,
	w http.ResponseWriter,
) {
	gatewayRequestCtx.relayStartTime = time.Now()

	// Track setup phase timing
	requestSetupStartTime := time.Now()
	requestSetupStage := requestSetupStageQoSContext
	var setupCompleted bool

	defer func() {
		// Record setup metrics if setup didn't complete successfully
		if !setupCompleted {
			requestSetupDuration := time.Since(requestSetupStartTime).Seconds()
			requestCachePerformance := categorizeRequestSetupCachePerformance(requestSetupDuration)
			shannonmetrics.RecordRequestSetupLatency(gatewayRequestCtx.serviceID, requestSetupStage, requestCachePerformance, requestSetupDuration)
		}

		// Write the HTTP response
		gatewayRequestCtx.WriteHTTPUserResponse(w)
	}()

	// Build QoS context
	err := gatewayRequestCtx.BuildQoSContextFromHTTP(httpReq)
	if err != nil {
		return
	}

	requestSetupStage = requestSetupStageProtocolContext

	// Build protocol context
	err = gatewayRequestCtx.BuildProtocolContextFromHTTP(httpReq)
	if err != nil {
		return
	}

	// Setup completed successfully - record metrics
	requestSetupDuration := time.Since(requestSetupStartTime).Seconds()
	requestCachePerformance := categorizeRequestSetupCachePerformance(requestSetupDuration)
	shannonmetrics.RecordRequestSetupLatency(gatewayRequestCtx.serviceID, requestSetupStageComplete, requestCachePerformance, requestSetupDuration)
	setupCompleted = true

	// Handle the actual relay request
	_ = gatewayRequestCtx.HandleRelayRequest()
}

// handleWebsocketRequest handles WebSocket connection requests
func (g Gateway) handleWebSocketRequest(
	_ context.Context,
	httpReq *http.Request,
	gatewayRequestCtx *requestContext,
	w http.ResponseWriter,
) {
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

// categorizeRequestSetupCachePerformance categorizes request setup performance based on timing patterns.
// This helps identify whether session cache hits/misses are affecting setup time.
func categorizeRequestSetupCachePerformance(setupDurationSeconds float64) string {
	setupDurationMs := setupDurationSeconds * 1000

	switch {
	// Very fast setup, likely all cache hits
	case setupDurationMs < 50:
		return "all_hits"

	// Moderate setup time, some cache misses
	case setupDurationMs < 500:
		return "some_misses"

	// Slow setup, likely session rollover or cache failures
	default:
		return "all_misses"
	}
}
