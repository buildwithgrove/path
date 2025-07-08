package gateway

import (
	"context"
	"fmt"
	"strings"
	"time"

	shannonmetrics "github.com/buildwithgrove/path/metrics/protocol/shannon"
	"github.com/buildwithgrove/path/protocol"
)

// HandleRelayRequest sends a relay from the perspective of a gateway.
// It performs the following steps:
//  1. Selects endpoints using the QoS context.
//  2. Sends the relay to multiple selected endpoints in parallel, using the protocol contexts.
//  3. Processes the first successful endpoint's response using the QoS context.
//
// HandleRelayRequest is written as a template method to allow the customization of key steps,
// e.g. endpoint selection and protocol-specific details of sending a relay.
// See the following link for more details:
// https://en.wikipedia.org/wiki/Template_method_pattern
func (rc *requestContext) HandleRelayRequest() error {
	// Track whether this is a parallel or single request
	isParallel := len(rc.protocolContexts) > 1

	// Log request type for monitoring
	rc.logger.Debug().
		Bool("is_parallel", isParallel).
		Int("protocol_contexts", len(rc.protocolContexts)).
		Str("service_id", string(rc.serviceID)).
		Msg("Handling relay request")

	// If we have multiple protocol contexts, send parallel requests
	if isParallel {
		return rc.handleParallelRelayRequests()
	}

	// Fallback to single request for backward compatibility
	return rc.handleSingleRelayRequest()
}

// handleSingleRelayRequest handles a single relay request (original behavior)
func (rc *requestContext) handleSingleRelayRequest() error {
	// Send the service request payload, through the protocol context, to the selected endpoint.
	// In this code path, we are always guaranteed to have exactly one protocol context.
	endpointResponse, err := rc.protocolContexts[0].HandleServiceRequest(rc.qosCtx.GetServicePayload())

	if err != nil {
		rc.logger.Warn().Err(err).Msg("Failed to send a relay request.")
		return err
	}

	rc.qosCtx.UpdateWithResponse(endpointResponse.EndpointAddr, endpointResponse.Bytes)
	return nil
}

// handleParallelRelayRequests sends relay requests to multiple endpoints in parallel
// and returns the first successful response
func (rc *requestContext) handleParallelRelayRequests() error {
	logger := rc.logger.With("service_id", rc.serviceID).With("method", "handleParallelRelayRequests")

	// Log comprehensive request information
	logger.Info().
		Int("endpoint_count", len(rc.protocolContexts)).
		Str("service_id", string(rc.serviceID)).
		Msg("Starting parallel relay requests")

	// Create a channel to receive the first successful response
	type relayResult struct {
		response  protocol.Response
		err       error
		index     int
		duration  time.Duration
		startTime time.Time
	}

	relayResultChan := make(chan relayResult, len(rc.protocolContexts))

	// Create context with timeout for parallel requests
	ctx, cancel := context.WithTimeout(rc.context, parallelRequestTimeout)
	defer cancel()

	overallStartTime := time.Now()

	// Launch parallel requests
	for i, protocolCtx := range rc.protocolContexts {
		go func(index int, pCtx ProtocolRequestContext) {
			startTime := time.Now()
			response, err := pCtx.HandleServiceRequest(rc.qosCtx.GetServicePayload())
			duration := time.Since(startTime)

			select {
			case relayResultChan <- relayResult{
				response:  response,
				err:       err,
				index:     index,
				duration:  duration,
				startTime: startTime,
			}:
			case <-ctx.Done():
				// Request was canceled, don't send result
				logger.Debug().Msgf("Request to endpoint %d canceled after %dms", index, duration.Milliseconds())
			}
		}(i, protocolCtx)
	}

	// Wait for the first successful response
	var lastErr error
	successfulResponses := 0
	totalRequests := len(rc.protocolContexts)
	var responseTimings []string

	for successfulResponses < totalRequests {
		select {
		case result := <-relayResultChan:
			successfulResponses++
			timingLog := fmt.Sprintf("endpoint_%d=%dms", result.index, result.duration.Milliseconds())
			responseTimings = append(responseTimings, timingLog)

			if result.err == nil {
				// First successful response - cancel other requests and return
				overallDuration := time.Since(overallStartTime)

				// Extract endpoint TLD for logging
				endpointTLD := ""
				if tlds := shannonmetrics.GetEndpointTLDs(protocol.EndpointAddrList{result.response.EndpointAddr}); len(tlds) > 0 {
					endpointTLD = tlds[result.response.EndpointAddr]
				}

				logger.Info().
					Int("winning_endpoint_index", result.index).
					Int64("winning_request_duration_ms", result.duration.Milliseconds()).
					Int64("overall_duration_ms", overallDuration.Milliseconds()).
					Str("winning_endpoint_addr", string(result.response.EndpointAddr)).
					Str("winning_endpoint_tld", endpointTLD).
					Int("completed_requests", successfulResponses).
					Int("total_requests", totalRequests).
					Str("all_response_timings", strings.Join(responseTimings, ", ")).
					Msg("Parallel relay race completed - first successful response received")

				cancel()
				rc.qosCtx.UpdateWithResponse(result.response.EndpointAddr, result.response.Bytes)
				return nil
			}
			// Log the error but continue waiting for other responses
			logger.Warn().Err(result.err).Msgf("Request to endpoint %d failed after %dms", result.index, result.duration.Milliseconds())
			lastErr = result.err
		case <-ctx.Done():
			// Context was canceled or timed out
			totalParallelRelayDuration := time.Since(overallStartTime).Milliseconds()
			if ctx.Err() == context.DeadlineExceeded {
				logger.Error().
					Int64("total_duration_ms", totalParallelRelayDuration).
					Int64("timeout_ms", parallelRequestTimeout.Milliseconds()).
					Int("completed_requests", successfulResponses).
					Int("total_requests", totalRequests).
					Str("response_timings", strings.Join(responseTimings, ", ")).
					Msg("Parallel relay requests timed out")
				return fmt.Errorf("parallel relay requests timed out after %v, last error: %w", parallelRequestTimeout, lastErr)
			}
			logger.Debug().
				Int64("total_duration_ms", totalParallelRelayDuration).
				Int("completed_requests", successfulResponses).
				Int("total_requests", totalRequests).
				Msg("Parallel relay requests canceled")
			return fmt.Errorf("parallel relay requests canceled, last error: %w", lastErr)
		}
	}

	// All requests failed
	totalParallelRelayDuration := time.Since(overallStartTime).Milliseconds()
	individualRequestDurationsStr := strings.Join(responseTimings, ", ")
	logger.Error().
		Int64("total_duration_ms", totalParallelRelayDuration).
		Int("failed_requests", totalRequests).
		Str("all_response_timings", individualRequestDurationsStr).
		Msg("All parallel relay requests failed")

	// Return the last error
	return fmt.Errorf("all parallel relay requests failed, last error: %w", lastErr)
}

// selectMultipleEndpoints selects up to maxNumEndpoints endpoints from the available endpoints
// with optional bias towards different TLDs for improved diversity and resilience
func (rc *requestContext) selectMultipleEndpoints(
	availableEndpoints protocol.EndpointAddrList,
	maxNumEndpoints int,
) protocol.EndpointAddrList {
	logger := rc.logger.With("method", "selectMultipleEndpoints")

	if len(availableEndpoints) == 0 {
		logger.Warn().
			Int("requested_endpoints", maxNumEndpoints).
			Str("service_id", string(rc.serviceID)).
			Msg("No endpoints available for selection")
		return nil
	}

	// Log available endpoints information
	availableTLDs := shannonmetrics.GetEndpointTLDs(availableEndpoints)
	uniqueTLDs := make(map[string]bool)
	for _, tld := range availableTLDs {
		if tld != "" {
			uniqueTLDs[tld] = true
		}
	}

	logger.Info().
		Int("available_endpoints", len(availableEndpoints)).
		Int("requested_endpoints", maxNumEndpoints).
		Int("unique_tlds_available", len(uniqueTLDs)).
		Str("service_id", string(rc.serviceID)).
		Msg("Selecting multiple endpoints for parallel requests")

	// Select multiple endpoints
	multipleSelectedEndpointAddr, err := rc.qosCtx.GetEndpointSelector().SelectMultiple(availableEndpoints, maxNumEndpoints)
	if err != nil {
		logger.Warn().
			Err(err).
			Int("available_endpoints", len(availableEndpoints)).
			Int("requested_endpoints", maxNumEndpoints).
			Msg("Failed to select multiple endpoints")
		return nil
	}

	logger.Info().
		Int("selected_endpoints", len(multipleSelectedEndpointAddr)).
		Int("requested_endpoints", maxNumEndpoints).
		Msg("Successfully selected endpoints")

	return multipleSelectedEndpointAddr

	// selectedEndpointAddr, err := rc.qosCtx.GetEndpointSelector().Select(availableEndpoints)
	// if err != nil {
	// 	rc.logger.Warn().Err(err).Msg("Failed to select endpoint")
	// 	return nil
	// }
	// return protocol.EndpointAddrList{selectedEndpointAddr}
}

// logEndpointTLDDiversity logs TLD diversity information for selected endpoints
func (rc *requestContext) logEndpointTLDDiversity(endpoints protocol.EndpointAddrList) {
	endpointTLDs := shannonmetrics.GetEndpointTLDs(endpoints)

	// Count unique TLDs
	tldCounts := make(map[string]int)
	for _, tld := range endpointTLDs {
		if tld != "" {
			tldCounts[tld]++
		}
	}

	// Log TLD distribution
	tldDistribution := make([]string, 0, len(tldCounts))
	for tld, count := range tldCounts {
		tldDistribution = append(tldDistribution, fmt.Sprintf("%s=%d", tld, count))
	}

	rc.logger.Info().
		Int("unique_tlds", len(tldCounts)).
		Int("endpoint_count", len(endpoints)).
		Str("tld_distribution", strings.Join(tldDistribution, ", ")).
		Msg("Endpoint TLD diversity for parallel requests")
}
