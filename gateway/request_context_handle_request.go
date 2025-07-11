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
	// Track whether this is a parallel or single request
	isParallel := len(rc.protocolContexts) > 1

	logger := rc.logger.
		With("service_id", rc.serviceID).
		With("method", "HandleRelayRequest").
		With("num_protocol_contexts", len(rc.protocolContexts))

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

// handleParallelRelayRequests sends relay requests to multiple endpoints in parallel
// and returns the first successful response.
func (rc *requestContext) handleParallelRelayRequests() error {
	logger := rc.logger.
		With("method", "handleParallelRelayRequests").
		With("num_protocol_contexts", len(rc.protocolContexts)).
		With("service_id", rc.serviceID)

	// Log comprehensive request information
	logger.Debug().Msg("Starting parallel relay race")

	// Create a channel to receive the first successful response
	type relayResult struct {
		response  protocol.Response
		err       error
		index     int
		duration  time.Duration
		startTime time.Time
	}

	// Channel to capture successful responses
	relayResultChan := make(chan relayResult, len(rc.protocolContexts))

	// Create context with timeout for parallel requests
	ctx, cancel := context.WithTimeout(rc.context, parallelRequestTimeout)
	defer cancel()

	// Track overall request start time
	overallStartTime := time.Now()

	// Launch parallel requests
	for protocolCtxIdx, protocolCtx := range rc.protocolContexts {
		go func(index int, pCtx ProtocolRequestContext) {
			specificRelayStartTime := time.Now()
			relayResponse, err := pCtx.HandleServiceRequest(rc.qosCtx.GetServicePayload())
			duration := time.Since(specificRelayStartTime)

			select {
			case relayResultChan <- relayResult{
				response:  relayResponse,
				err:       err,
				index:     index,
				duration:  duration,
				startTime: specificRelayStartTime,
			}:
			case <-ctx.Done():
				// Request was canceled, don't send result
				logger.Debug().Msgf("Request to endpoint %d canceled after %dms", index, duration.Milliseconds())
			}
		}(protocolCtxIdx, protocolCtx)
	}

	// Wait for the first successful response
	var lastErr error
	successfulResponses := 0
	totalRequests := len(rc.protocolContexts)
	var responseTimings []string

	// TODO_IN_THIS_PR: Add metrics for parallel request success rates and timing distributions

	for successfulResponses < totalRequests {
		select {
		case result := <-relayResultChan:
			successfulResponses++
			timingLog := fmt.Sprintf("endpoint_%d=%dms", result.index, result.duration.Milliseconds())
			responseTimings = append(responseTimings, timingLog)

			// First successful response - cancel other requests and return
			if result.err == nil {
				// Extract and log results
				overallDurationToFirstSuccess := time.Since(overallStartTime)
				endpointDomain := shannonmetrics.ExtractTLDFromEndpointAddr(string(result.response.EndpointAddr))
				logger.Info().
					Str("endpoint_domain", endpointDomain).
					Msgf("Parallel request success: endpoint %d/%d responded in %dms", result.index+1, totalRequests, overallDurationToFirstSuccess.Milliseconds())

				// Update QoS context and return
				rc.qosCtx.UpdateWithResponse(result.response.EndpointAddr, result.response.Bytes)
				return nil
			}
			// Log the error but continue waiting for other responses
			logger.Warn().Err(result.err).Msgf("[Parallel Requests] Request to endpoint %d failed after %dms", result.index, result.duration.Milliseconds())
			lastErr = result.err
		case <-ctx.Done():
			// Context was canceled or timed out
			totalParallelRelayDuration := time.Since(overallStartTime).Milliseconds()
			if ctx.Err() == context.DeadlineExceeded {
				logger.Error().Msgf("Parallel requests timed out after %dms and %d completed requests", totalParallelRelayDuration, successfulResponses)
				return fmt.Errorf("parallel relay requests timed out after %dms and %d completed requests, last error: %w", totalParallelRelayDuration, successfulResponses, lastErr)
			}
			logger.Debug().Msg("Parallel requests canceled")
			return fmt.Errorf("parallel relay requests canceled after %dms and %d completed requests, last error: %w", totalParallelRelayDuration, successfulResponses, lastErr)
		}
	}

	// All requests failed
	totalParallelRelayDuration := time.Since(overallStartTime).Milliseconds()
	individualRequestDurationsStr := strings.Join(responseTimings, ", ")
	logger.Error().Msgf("All %d parallel requests failed after %dms with individual request durations: %s", totalRequests, totalParallelRelayDuration, individualRequestDurationsStr)

	// Return the last error
	return fmt.Errorf("all parallel relay requests failed, last error: %w", lastErr)
}

// logEndpointTLDDiversity logs TLD diversity information for selected endpoints.
func (rc *requestContext) logEndpointTLDDiversity(endpoints protocol.EndpointAddrList) {
	logger := rc.logger.
		With("method", "logEndpointTLDDiversity").
		With("service_id", rc.serviceID).
		With("num_endpoints", len(endpoints))

	// Count unique TLDs
	endpointTLDs := shannonmetrics.GetEndpointTLDs(endpoints)
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
	logger.Info().Msgf("Endpoint TLD diversity: %s", strings.Join(tldDistribution, ", "))
}
