package gateway

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	shannonmetrics "github.com/buildwithgrove/path/metrics/protocol/shannon"
	"github.com/buildwithgrove/path/protocol"
)

// parallelRelayResult is used to track the result of a parallel relay request.
// It is intended for internal use by the requestContext.
type parallelRelayResult struct {
	response  protocol.Response
	err       error
	index     int
	duration  time.Duration
	startTime time.Time
}

// parallelRequestMetrics tracks metrics for parallel requests
type parallelRequestMetrics struct {
	numRequestsToAttempt     int
	numCompletedSuccessfully int
	numFailedOrErrored       int
	overallStartTime         time.Time
}

// HandleRelayRequest sends a relay from the perspective of a gateway.
//
// It performs the following steps:
//  1. Selects endpoints using the QoS context
//  2. Sends the relay to multiple selected endpoints in parallel, using the protocol contexts
//  3. Processes the first successful endpoint's response using the QoS context
//
// HandleRelayRequest is written as a template method to allow the customization of key steps,
// e.g. endpoint selection and protocol-specific details of sending a relay.
// See the following link for more details:
// https://en.wikipedia.org/wiki/Template_method_pattern
func (rc *requestContext) HandleRelayRequest() error {
	logger := rc.logger.
		With("service_id", rc.serviceID).
		With("method", "HandleRelayRequest").
		With("num_protocol_contexts", len(rc.protocolContexts))

	// Track whether this is a parallel or single request
	isParallel := len(rc.protocolContexts) > 1

	// If we have multiple protocol contexts, send parallel requests
	if isParallel {
		logger.Debug().Msgf("Handling %d parallel relay requests", len(rc.protocolContexts))
		return rc.handleParallelRelayRequests()
	}

	// Fallback to single request for backward compatibility
	logger.Debug().Msg("Handling single relay request")
	return rc.handleSingleRelayRequest()
}

// handleSingleRelayRequest handles a single relay request (original behavior)
func (rc *requestContext) handleSingleRelayRequest() error {
	// Send the service request payload, through the protocol context, to the selected endpoint.
	// In this code path, we are always guaranteed to have exactly one protocol context.
	endpointResponse, err := rc.protocolContexts[0].HandleServiceRequest(rc.qosCtx.GetServicePayload())
	if err != nil {
		rc.logger.Warn().Err(err).Msg("Failed to send a single relay request.")
		return err
	}

	rc.qosCtx.UpdateWithResponse(endpointResponse.EndpointAddr, endpointResponse.Bytes)
	return nil
}

// handleParallelRelayRequests orchestrates parallel relay requests and returns the first successful response.
func (rc *requestContext) handleParallelRelayRequests() error {
	metrics := &parallelRequestMetrics{
		numRequestsToAttempt: len(rc.protocolContexts),
		overallStartTime:     time.Now(),
	}
	defer rc.updateParallelRequestMetrics(metrics)

	logger := rc.logger.
		With("method", "handleParallelRelayRequests").
		With("num_protocol_contexts", len(rc.protocolContexts)).
		With("service_id", rc.serviceID)
	logger.Debug().Msg("Starting parallel relay race")

	ctx, cancel := context.WithTimeout(rc.context, parallelRequestTimeout)
	defer cancel()

	resultChan := rc.launchParallelRequests(ctx, logger)

	return rc.waitForFirstSuccessfulResponse(ctx, logger, resultChan, metrics)
}

// updateParallelRequestMetrics updates gateway observations with parallel request metrics
func (rc *requestContext) updateParallelRequestMetrics(metrics *parallelRequestMetrics) {
	numCanceledByContext := metrics.numRequestsToAttempt - metrics.numCompletedSuccessfully - metrics.numFailedOrErrored
	rc.updateGatewayObservationsWithParallelRequests(
		metrics.numRequestsToAttempt,
		metrics.numCompletedSuccessfully,
		metrics.numFailedOrErrored,
		numCanceledByContext,
	)
}

// launchParallelRequests starts all parallel relay requests and returns a result channel
func (rc *requestContext) launchParallelRequests(ctx context.Context, logger polylog.Logger) <-chan parallelRelayResult {
	resultChan := make(chan parallelRelayResult, len(rc.protocolContexts))

	for protocolCtxIdx, protocolCtx := range rc.protocolContexts {
		go rc.executeRelayRequest(ctx, logger, protocolCtx, protocolCtxIdx, resultChan)
	}

	return resultChan
}

// executeRelayRequest handles a single relay request in a goroutine
func (rc *requestContext) executeRelayRequest(
	ctx context.Context,
	logger polylog.Logger,
	protocolCtx ProtocolRequestContext,
	index int,
	resultChan chan<- parallelRelayResult,
) {
	startTime := time.Now()
	response, err := protocolCtx.HandleServiceRequest(rc.qosCtx.GetServicePayload())
	duration := time.Since(startTime)

	result := parallelRelayResult{
		response:  response,
		err:       err,
		index:     index,
		duration:  duration,
		startTime: startTime,
	}

	select {
	case resultChan <- result:
		// Result sent successfully
	case <-ctx.Done():
		logger.Debug().Msgf("Request to endpoint %d canceled after %dms", index, duration.Milliseconds())
	}
}

// waitForFirstSuccessfulResponse waits for the first successful response or handles all failures
func (rc *requestContext) waitForFirstSuccessfulResponse(
	ctx context.Context,
	logger polylog.Logger,
	resultChan <-chan parallelRelayResult,
	metrics *parallelRequestMetrics,
) error {
	var lastErr error
	var responseTimings []string

	for metrics.numCompletedSuccessfully < metrics.numRequestsToAttempt {
		select {
		case result := <-resultChan:
			responseTimings = append(responseTimings, rc.formatTimingLog(result))

			if result.err == nil {
				return rc.handleSuccessfulResponse(logger, result, metrics)
			} else {
				rc.handleFailedResponse(logger, result, metrics, &lastErr)
			}

		case <-ctx.Done():
			return rc.handleContextDone(ctx, logger, metrics, lastErr)
		}
	}

	return rc.handleAllRequestsFailed(logger, metrics, responseTimings, lastErr)
}

// handleSuccessfulResponse processes the first successful response
func (rc *requestContext) handleSuccessfulResponse(
	logger polylog.Logger,
	result parallelRelayResult,
	metrics *parallelRequestMetrics,
) error {
	metrics.numCompletedSuccessfully++
	overallDuration := time.Since(metrics.overallStartTime)
	endpointDomain := shannonmetrics.ExtractTLDFromEndpointAddr(string(result.response.EndpointAddr))

	logger.Info().
		Str("endpoint_domain", endpointDomain).
		Msgf("Parallel request success: endpoint %d/%d responded in %dms",
			result.index+1, metrics.numRequestsToAttempt, overallDuration.Milliseconds())

	rc.qosCtx.UpdateWithResponse(result.response.EndpointAddr, result.response.Bytes)
	return nil
}

// handleFailedResponse processes a failed response
func (rc *requestContext) handleFailedResponse(
	logger polylog.Logger,
	result parallelRelayResult,
	metrics *parallelRequestMetrics,
	lastErr *error,
) {
	metrics.numFailedOrErrored++
	logger.Warn().Err(result.err).
		Msgf("Request to endpoint %d failed after %dms", result.index, result.duration.Milliseconds())
	*lastErr = result.err
}

// handleContextDone processes context cancellation or timeout
func (rc *requestContext) handleContextDone(
	ctx context.Context,
	logger polylog.Logger,
	metrics *parallelRequestMetrics,
	lastErr error,
) error {
	totalDuration := time.Since(metrics.overallStartTime).Milliseconds()

	if ctx.Err() == context.DeadlineExceeded {
		logger.Error().Msgf("Parallel requests timed out after %dms and %d completed requests",
			totalDuration, metrics.numCompletedSuccessfully)
		return fmt.Errorf("parallel relay requests timed out after %dms and %d completed requests, last error: %w",
			totalDuration, metrics.numCompletedSuccessfully, lastErr)
	}

	logger.Debug().Msg("Parallel requests canceled")
	return fmt.Errorf("parallel relay requests canceled after %dms and %d completed requests, last error: %w",
		totalDuration, metrics.numCompletedSuccessfully, lastErr)
}

// handleAllRequestsFailed processes the case where all requests failed
func (rc *requestContext) handleAllRequestsFailed(
	logger polylog.Logger,
	metrics *parallelRequestMetrics,
	responseTimings []string,
	lastErr error,
) error {
	totalDuration := time.Since(metrics.overallStartTime).Milliseconds()
	timingsStr := strings.Join(responseTimings, ", ")

	logger.Error().Msgf("All %d parallel requests failed after %dms with individual request durations: %s",
		metrics.numRequestsToAttempt, totalDuration, timingsStr)

	return fmt.Errorf("all parallel relay requests failed, last error: %w", lastErr)
}

// formatTimingLog creates a timing log string for a relay result
func (rc *requestContext) formatTimingLog(result parallelRelayResult) string {
	return fmt.Sprintf("endpoint_%d=%dms", result.index, result.duration.Milliseconds())
}
