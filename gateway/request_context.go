package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/buildwithgrove/path/config/relay"
	shannonmetrics "github.com/buildwithgrove/path/metrics/protocol/shannon"
	"github.com/buildwithgrove/path/observation"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
)

var (
	errHTTPRequestRejectedByParser   = errors.New("HTTP request rejected by the HTTP parser")
	errHTTPRequestRejectedByQoS      = errors.New("HTTP request rejected by service QoS instance")
	errHTTPRequestRejectedByProtocol = errors.New("HTTP request rejected by protocol instance")
	errWebsocketRequestRejectedByQoS = errors.New("websocket request rejected by service QoS instance")
)

const (
	// Session state constants
	sessionStateNormal = "normal"
	sessionStateError  = "error"
	sessionStateActive = "active"
)

// Note: maxParallelRequests is now obtained from gateway configuration

// Gateway requestContext is responsible for performing the steps necessary to complete a service request.
//
// It contains two main contexts:
//
//  1. Protocol context
//     - Supplies the list of available endpoints for the requested service to the QoS ctx.
//     - Builds the Protocol ctx for the selected endpoint once it has been selected.
//     - Sends the relay request to the selected endpoint using the protocol-specific implementation.
//
//  2. QoS context
//     - Receives the list of available endpoints for the requested service from the Protocol instance.
//     - Selects a valid endpoint from among them based on the service-specific QoS implementation.
//     - Updates its internal store based on observations made during the handling of the request.
//
// As of PR #72, it is limited in scope to HTTP service requests.
type requestContext struct {
	logger polylog.Logger

	// httpRequestParser is used by the request context to interpret an HTTP request as a pair of:
	// 	1. service ID
	// 	2. The service ID's corresponding QoS instance.
	httpRequestParser HTTPRequestParser

	// metricsReporter is used to export metrics based on observations made in handling service requests.
	metricsReporter RequestResponseReporter

	// dataReporter is used to export, to the data pipeline, observations made in handling service requests.
	// It is declared separately from the `metricsReporter` to be consistent with the gateway package's role
	// of explicitly defining PATH gateway's components and their interactions.
	dataReporter RequestResponseReporter

	// QoS related request context
	serviceID  protocol.ServiceID
	serviceQoS QoSService
	qosCtx     RequestQoSContext

	// Protocol related request context
	protocol         Protocol
	protocolContexts []ProtocolRequestContext // Multiplicity of protocol contexts for parallel requests

	// presetFailureHTTPResponse, if set, is used to return a preconstructed error response to the user.
	// For example, this is used to return an error if the specified target service ID is invalid.
	presetFailureHTTPResponse HTTPResponse

	// httpObservations stores the observations related to the HTTP request.
	httpObservations observation.HTTPRequestObservations
	// gatewayObservations stores gateway related observations.
	gatewayObservations *observation.GatewayObservations
	// Tracks protocol observations.
	protocolObservations *protocolobservations.Observations

	// Enforces request completion deadline.
	// Passed to potentially long-running operations like protocol interactions.
	// Prevents HTTP handler timeouts that would return empty responses to clients.
	context context.Context

	// TODO_TECHDEBT(@adshmh): refactor the interfaces and interactions with Protocol and QoS, to remove the need for this field.
	// Tracks whether the request was rejected by the QoS.
	// This is needed for handling the observations: there will be no protocol context/observations in this case.
	requestRejectedByQoS bool

	// Timing fields for relay latency tracking
	relayStartTime time.Time

	// Gateway configuration for relay handling
	gatewayConfig relay.Config
}

// InitFromHTTPRequest builds the required context for serving an HTTP request.
// e.g.:
//   - The target service ID
//   - The Service QoS instance
func (rc *requestContext) InitFromHTTPRequest(httpReq *http.Request) error {
	rc.logger = rc.getHTTPRequestLogger(httpReq)

	// TODO_MVP(@adshmh): The HTTPRequestParser should return a context, similar to QoS, which is then used to get a QoS instance and the observation set.
	// Extract the service ID and find the target service's corresponding QoS instance.
	serviceID, serviceQoS, err := rc.httpRequestParser.GetQoSService(rc.context, httpReq)
	rc.serviceID = serviceID
	if err != nil {
		// TODO_MVP(@adshmh): consolidate gateway-level observations in one location.
		// Update gateway observations
		rc.updateGatewayObservations(err)

		// set an error response
		rc.presetFailureHTTPResponse = rc.httpRequestParser.GetHTTPErrorResponse(rc.context, err)

		// log the error
		rc.logger.Error().Err(err).Msg(errHTTPRequestRejectedByParser.Error())
		return errHTTPRequestRejectedByParser
	}

	rc.serviceQoS = serviceQoS
	return nil
}

// hydrateGatewayObservations
// - updates the gateway-level observations in the request context with other metadata in the request context.
// - sets the gateway observation error with the one provided, if not already set
func (rc *requestContext) updateGatewayObservations(err error) {
	// set the service ID on the gateway observations.
	rc.gatewayObservations.ServiceId = string(rc.serviceID)

	// Update the request completion time on the gateway observation
	rc.gatewayObservations.CompletedTime = timestamppb.Now()

	// No errors: skip.
	if err == nil {
		return
	}

	// Request error already set: skip.
	if rc.gatewayObservations.GetRequestError() != nil {
		return
	}

	switch {
	// Service ID not specified
	case errors.Is(err, ErrNoServiceIDProvided):
		rc.logger.Error().Err(err).Msg("No service ID specified in the HTTP headers. Request will fail.")
		rc.gatewayObservations.RequestError = &observation.GatewayRequestError{
			// Set the error kind
			ErrorKind: observation.GatewayRequestErrorKind_GATEWAY_REQUEST_ERROR_KIND_MISSING_SERVICE_ID,
			// Use the error message as error details.
			Details: err.Error(),
		}

	// Request was rejected by the QoS instance.
	// e.g. HTTP payload could not be unmarshaled into a JSONRPC request.
	case errors.Is(err, ErrRejectedByQoS):
		rc.logger.Error().Err(err).Msg("QoS instance rejected the request. Request will fail.")
		rc.gatewayObservations.RequestError = &observation.GatewayRequestError{
			// Set the error kind
			ErrorKind: observation.GatewayRequestErrorKind_GATEWAY_REQUEST_ERROR_KIND_REJECTED_BY_QOS,
			// Use the error message as error details.
			Details: err.Error(),
		}

	default:
		rc.logger.Warn().Err(err).Msg("SHOULD NEVER HAPPEN: unrecognized gateway-level request error.")
		// Set a generic request error observation
		rc.gatewayObservations.RequestError = &observation.GatewayRequestError{
			// unspecified error kind: this should not happen
			ErrorKind: observation.GatewayRequestErrorKind_GATEWAY_REQUEST_ERROR_KIND_UNSPECIFIED,
			// Use the error message as error details.
			Details: err.Error(),
		}
	}
}

// BuildQoSContextFromHTTP builds the QoS context instance using the supplied HTTP request's payload.
func (rc *requestContext) BuildQoSContextFromHTTP(httpReq *http.Request) error {
	// TODO_MVP(@adshmh): Add an HTTP request size metric/observation at the gateway/http (L7) level.
	// Required steps:
	//  	1. Update QoSService interface to parse custom struct with []byte payload
	//  	2. Read HTTP request body in `request` package and return struct for QoS Service
	//  	3. Export HTTP observations from `request` package when reading body

	// Build the payload for the requested service using the incoming HTTP request.
	// This payload will be sent to an endpoint matching the requested service.
	qosCtx, isValid := rc.serviceQoS.ParseHTTPRequest(rc.context, httpReq)
	rc.qosCtx = qosCtx

	if !isValid {
		// mark the request was rejected by the QoS
		rc.requestRejectedByQoS = true

		// Update gateway observations
		rc.updateGatewayObservations(ErrRejectedByQoS)
		rc.logger.Info().Msg(errHTTPRequestRejectedByQoS.Error())
		return errHTTPRequestRejectedByQoS
	}

	return nil
}

// BuildQoSContextFromWebsocket builds the QoS context instance using the supplied WebSocket request.
// This method does not need to parse the HTTP request's payload as the WebSocket request does not have a body,
// so it will only return an error if called for a service that does not support WebSocket connections.
func (rc *requestContext) BuildQoSContextFromWebsocket(wsReq *http.Request) error {
	// Create the QoS request context using the WebSocket request.
	// This method will reject the request if it is for a service that does not support WebSocket connections.
	qosCtx, isValid := rc.serviceQoS.ParseWebsocketRequest(rc.context)
	rc.qosCtx = qosCtx

	// Only reject the request if the service QoS does not support WebSocket connections.
	// All other WebSocket requests will have `isValid` set to true.
	if !isValid {
		rc.logger.Info().Msg(errWebsocketRequestRejectedByQoS.Error())
		return errWebsocketRequestRejectedByQoS
	}

	return nil
}

// BuildProtocolContextsFromHTTPRequest builds multiple Protocol contexts using the supplied HTTP request.
//
// Steps:
// 1. Get available endpoints for the requested service from the Protocol instance
// 2. Select multiple endpoints for parallel relay attempts
// 3. Build Protocol contexts for each selected endpoint
//
// The constructed Protocol instances will be used for:
//   - Sending parallel relay requests to multiple endpoints
//   - Getting the list of protocol-level observations
//
// TODO_TECHDEBT: Either rename to `PrepareProtocol` or return the built protocol context.
func (rc *requestContext) BuildProtocolContextsFromHTTPRequest(httpReq *http.Request) error {
	// Retrieve the list of available endpoints for the requested service.
	availableEndpoints, endpointLookupObs, err := rc.protocol.AvailableEndpoints(rc.context, rc.serviceID, httpReq)
	if err != nil {
		// error encountered: use the supplied observations as protocol observations.
		rc.updateProtocolObservations(&endpointLookupObs)
		rc.logger.
			With("service_id", rc.serviceID).
			ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).
			Err(err).Msg("error getting available endpoints for the service request. Request will fail.")
		return fmt.Errorf("BuildProtocolContextsFromHTTPRequest: error getting available endpoints for service %s: %w", rc.serviceID, err)
	}

	// Select multiple endpoints for parallel relay attempts  
	maxParallelRequests := rc.gatewayConfig.MaxParallelRequests
	selectedEndpoints := rc.selectMultipleEndpoints(availableEndpoints, maxParallelRequests)
	if len(selectedEndpoints) == 0 {
		// no protocol context will be built: use the endpointLookup observation.
		rc.updateProtocolObservations(&endpointLookupObs)
		rc.logger.
			With("service_id", rc.serviceID).
			ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).
			Msg("error selecting endpoints for the service request. Request will fail.")
		return fmt.Errorf("BuildProtocolContextsFromHTTPRequest: no endpoints could be selected")
	}

	rc.logger.Info().Msgf("Selected %d endpoints for parallel relay requests", len(selectedEndpoints))

	// Prepare Protocol contexts for all selected endpoints
	rc.protocolContexts = make([]ProtocolRequestContext, 0, len(selectedEndpoints))
	var lastProtocolCtxSetupErrObs *protocolobservations.Observations

	for i, endpointAddr := range selectedEndpoints {
		rc.logger.Debug().Msgf("Building protocol context for endpoint %d/%d: %s", i+1, len(selectedEndpoints), endpointAddr)
		protocolCtx, protocolCtxSetupErrObs, err := rc.protocol.BuildRequestContextForEndpoint(rc.context, rc.serviceID, endpointAddr, httpReq)
		if err != nil {
			lastProtocolCtxSetupErrObs = &protocolCtxSetupErrObs
			rc.logger.Warn().
				Err(err).
				Str("endpoint_addr", string(endpointAddr)).
				Msgf("Failed to build protocol context for endpoint %d/%d, skipping", i+1, len(selectedEndpoints))
			// Continue with other endpoints rather than failing completely
			continue
		}
		rc.protocolContexts = append(rc.protocolContexts, protocolCtx)
		rc.logger.Debug().Msgf("Successfully built protocol context for endpoint %d/%d: %s", i+1, len(selectedEndpoints), endpointAddr)
	}

	if len(rc.protocolContexts) == 0 {
		// error encountered: use the supplied observations as protocol observations.
		rc.updateProtocolObservations(lastProtocolCtxSetupErrObs)
		rc.logger.
			With("service_id", rc.serviceID).
			ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).
			Msg(errHTTPRequestRejectedByProtocol.Error())
		return errHTTPRequestRejectedByProtocol
	}

	return nil
}

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
	// If we have multiple protocol contexts, send parallel requests
	if len(rc.protocolContexts) > 1 {
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
	logger.Info().Msgf("Sending parallel relay requests to %d endpoints", len(rc.protocolContexts))

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
	parallelTimeout := rc.gatewayConfig.ParallelRequestTimeout
	ctx, cancel := context.WithTimeout(rc.context, parallelTimeout)
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
				// Request was cancelled, don't send result
				logger.Debug().Msgf("Request to endpoint %d cancelled after %dms", index, duration.Milliseconds())
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
				logger.Info().Msgf("Received successful response from endpoint %d after %dms (overall: %dms), cancelling other requests. Timings: [%s]",
					result.index, result.duration.Milliseconds(), overallDuration.Milliseconds(), strings.Join(responseTimings, ", "))
				cancel()
				rc.qosCtx.UpdateWithResponse(result.response.EndpointAddr, result.response.Bytes)
				return nil
			}
			// Log the error but continue waiting for other responses
			logger.Warn().Err(result.err).Msgf("Request to endpoint %d failed after %dms", result.index, result.duration.Milliseconds())
			lastErr = result.err
		case <-ctx.Done():
			// Context was cancelled or timed out
			totalParallelRelayDuration := time.Since(overallStartTime).Milliseconds()
			if ctx.Err() == context.DeadlineExceeded {
				logger.Error().Msgf("Parallel relay requests timed out after %dms (timeout: %v), received %d/%d responses", 
					totalParallelRelayDuration, parallelTimeout, successfulResponses, totalRequests)
				return fmt.Errorf("parallel relay requests timed out after %v, last error: %w", parallelTimeout, lastErr)
			}
			logger.Debug().Msgf("Parallel relay requests cancelled after %dms", totalParallelRelayDuration)
			return fmt.Errorf("parallel relay requests cancelled, last error: %w", lastErr)
		}
	}

	// All requests failed
	totalParallelRelayDuration := time.Since(overallStartTime).Milliseconds()
	individualRequestDurationsStr := strings.Join(responseTimings, ", ")
	logger.Error().Msgf("All parallel relay requests failed after %dms. Timings: [%s]", totalParallelRelayDuration, individualRequestDurationsStr)

	// Return the last error
	return fmt.Errorf("all parallel relay requests failed, last error: %w", lastErr)
}

// selectMultipleEndpoints selects up to maxCount endpoints from the available endpoints
// with optional bias towards different TLDs for improved diversity and resilience
//
// 🏗️ ARCHITECTURAL CONCERN: This TLD-based selection logic is tightly coupled to URL parsing
// and makes assumptions about endpoint address format. Consider abstracting endpoint diversity
// selection into a pluggable strategy pattern.
func (rc *requestContext) selectMultipleEndpoints(availableEndpoints protocol.EndpointAddrList, maxCount int) []protocol.EndpointAddr {
	if len(availableEndpoints) == 0 {
		return nil
	}

	// If diversity is disabled, use simple sequential selection
	if !rc.gatewayConfig.EnableEndpointDiversity {
		return rc.selectEndpointsSequentially(availableEndpoints, maxCount)
	}

	// Use diversity-aware selection
	return rc.selectEndpointsWithDiversity(availableEndpoints, maxCount)
}

// selectEndpointsSequentially selects endpoints without diversity considerations
func (rc *requestContext) selectEndpointsSequentially(availableEndpoints protocol.EndpointAddrList, maxCount int) []protocol.EndpointAddr {
	var selectedEndpoints []protocol.EndpointAddr
	remainingEndpoints := make(protocol.EndpointAddrList, len(availableEndpoints))
	copy(remainingEndpoints, availableEndpoints)

	for i := 0; i < maxCount && len(remainingEndpoints) > 0; i++ {
		selectedEndpoint, err := rc.qosCtx.GetEndpointSelector().Select(remainingEndpoints)
		if err != nil {
			rc.logger.Warn().Err(err).Msgf("Failed to select endpoint %d, stopping selection", i+1)
			break
		}

		selectedEndpoints = append(selectedEndpoints, selectedEndpoint)

		// Remove the selected endpoint from the remaining pool
		newRemainingEndpoints := make(protocol.EndpointAddrList, 0, len(remainingEndpoints)-1)
		for _, endpoint := range remainingEndpoints {
			if endpoint != selectedEndpoint {
				newRemainingEndpoints = append(newRemainingEndpoints, endpoint)
			}
		}
		remainingEndpoints = newRemainingEndpoints
	}

	rc.logger.Info().Msgf("Selected %d endpoints (diversity disabled)", len(selectedEndpoints))
	return selectedEndpoints
}

// selectEndpointsWithDiversity selects endpoints with TLD diversity preference
func (rc *requestContext) selectEndpointsWithDiversity(availableEndpoints protocol.EndpointAddrList, maxCount int) []protocol.EndpointAddr {
	// Get endpoint URLs to extract TLD information
	endpointTLDs := rc.getEndpointTLDs(availableEndpoints)

	// Count unique TLDs for logging
	uniqueTLDs := make(map[string]bool)
	for _, tld := range endpointTLDs {
		if tld != "" {
			uniqueTLDs[tld] = true
		}
	}

	rc.logger.Debug().Msgf("Endpoint selection: %d available endpoints across %d unique TLDs, selecting up to %d endpoints",
		len(availableEndpoints), len(uniqueTLDs), maxCount)

	var selectedEndpoints []protocol.EndpointAddr
	usedTLDs := make(map[string]bool)
	remainingEndpoints := make(protocol.EndpointAddrList, len(availableEndpoints))
	copy(remainingEndpoints, availableEndpoints)

	// First pass: Try to select endpoints with different TLDs
	for i := 0; i < maxCount && len(remainingEndpoints) > 0; i++ {
		var selectedEndpoint protocol.EndpointAddr
		var err error

		// Try to find an endpoint with a different TLD
		if i > 0 && len(usedTLDs) > 0 {
			selectedEndpoint, err = rc.selectEndpointWithDifferentTLD(remainingEndpoints, endpointTLDs, usedTLDs)
			if err != nil {
				// Fallback to standard selection if no different TLD found
				selectedEndpoint, err = rc.qosCtx.GetEndpointSelector().Select(remainingEndpoints)
			}
		} else {
			// First endpoint: use standard selection
			selectedEndpoint, err = rc.qosCtx.GetEndpointSelector().Select(remainingEndpoints)
		}

		if err != nil {
			rc.logger.Warn().Err(err).Msgf("Failed to select endpoint %d, stopping selection", i+1)
			break
		}

		selectedEndpoints = append(selectedEndpoints, selectedEndpoint)

		// Track the TLD of the selected endpoint
		if tld, exists := endpointTLDs[selectedEndpoint]; exists {
			usedTLDs[tld] = true
			rc.logger.Debug().Msgf("Selected endpoint with TLD: %s (endpoint: %s)", tld, selectedEndpoint)
		}

		// Remove the selected endpoint from the remaining pool
		newRemainingEndpoints := make(protocol.EndpointAddrList, 0, len(remainingEndpoints)-1)
		for _, endpoint := range remainingEndpoints {
			if endpoint != selectedEndpoint {
				newRemainingEndpoints = append(newRemainingEndpoints, endpoint)
			}
		}
		remainingEndpoints = newRemainingEndpoints
	}

	// Count fallback selections (endpoints without TLD diversity)
	fallbackSelections := 0
	for _, endpoint := range selectedEndpoints {
		if tld, exists := endpointTLDs[endpoint]; exists && tld != "" {
			// Count how many endpoints use this TLD
			tldCount := 0
			for _, otherEndpoint := range selectedEndpoints {
				if otherTLD, exists := endpointTLDs[otherEndpoint]; exists && otherTLD == tld {
					tldCount++
				}
			}
			if tldCount > 1 {
				fallbackSelections++
			}
		}
	}

	rc.logger.Info().Msgf("Selected %d endpoints across %d different TLDs (diversity: %.1f%%, fallback selections: %d)",
		len(selectedEndpoints), len(usedTLDs),
		float64(len(usedTLDs))/float64(len(selectedEndpoints))*100, fallbackSelections)
	return selectedEndpoints
}

// getEndpointTLDs extracts TLD information from endpoint addresses
// Returns a map of endpoint address to TLD for efficient lookup
func (rc *requestContext) getEndpointTLDs(endpoints protocol.EndpointAddrList) map[protocol.EndpointAddr]string {
	endpointTLDs := make(map[protocol.EndpointAddr]string)

	for _, endpointAddr := range endpoints {
		tld := rc.extractTLDFromEndpointAddr(endpointAddr)
		if tld != "" {
			endpointTLDs[endpointAddr] = tld
		}
	}

	return endpointTLDs
}

// extractTLDFromEndpointAddr extracts the TLD from an endpoint address
// Shannon endpoints are formatted as "supplier-url", so we need to extract the URL part
func (rc *requestContext) extractTLDFromEndpointAddr(endpointAddr protocol.EndpointAddr) string {
	addrStr := string(endpointAddr)

	// Find the first occurrence of "http" to locate the URL part
	httpIndex := strings.Index(addrStr, "http")
	if httpIndex == -1 {
		// No http found, try to find domain-like patterns
		// Look for first part that contains a dot (likely a domain)
		parts := strings.Split(addrStr, "-")
		for i, part := range parts {
			if strings.Contains(part, ".") {
				// Reconstruct potential URL from this point
				urlPart := strings.Join(parts[i:], "-")
				return rc.extractTLDFromURL(urlPart, true)
			}
		}
		return ""
	}

	// Extract URL part starting from "http"
	// 🐛 POTENTIAL BUG: This URL extraction is fragile and depends on specific formatting
	urlPart := addrStr[httpIndex:]
	return rc.extractTLDFromURL(urlPart, false)
}

// extractTLDFromURL extracts TLD from a URL string
func (rc *requestContext) extractTLDFromURL(urlStr string, addScheme bool) string {
	// Add scheme if needed for proper URL parsing
	if addScheme && !strings.HasPrefix(urlStr, "http") {
		urlStr = "https://" + urlStr
	}

	// Clean up any URL encoding issues
	urlStr = strings.ReplaceAll(urlStr, "%3A", ":")
	urlStr = strings.ReplaceAll(urlStr, "%2F", "/")

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		rc.logger.Debug().Err(err).Msgf("Failed to parse URL: %s", urlStr)
		// Fallback: try to extract hostname manually
		return rc.extractTLDManually(urlStr)
	}

	hostname := parsedURL.Hostname()
	if hostname == "" {
		// Fallback: try to extract hostname manually
		return rc.extractTLDManually(urlStr)
	}

	// Extract TLD from hostname
	domainParts := strings.Split(hostname, ".")
	if len(domainParts) < 2 {
		return ""
	}

	// Return the TLD (last part of the domain)
	tld := domainParts[len(domainParts)-1]
	return tld
}

// extractTLDManually attempts to extract TLD when URL parsing fails
func (rc *requestContext) extractTLDManually(urlStr string) string {
	// Remove protocol if present
	urlStr = strings.TrimPrefix(urlStr, "http://")
	urlStr = strings.TrimPrefix(urlStr, "https://")

	// Take only the host part (before any path, query, or fragment)
	if idx := strings.Index(urlStr, "/"); idx != -1 {
		urlStr = urlStr[:idx]
	}
	if idx := strings.Index(urlStr, "?"); idx != -1 {
		urlStr = urlStr[:idx]
	}
	if idx := strings.Index(urlStr, "#"); idx != -1 {
		urlStr = urlStr[:idx]
	}

	// Remove port if present
	if idx := strings.LastIndex(urlStr, ":"); idx != -1 {
		// Check if this is actually a port (numeric after the colon)
		portPart := urlStr[idx+1:]
		isPort := true
		for _, char := range portPart {
			if char < '0' || char > '9' {
				isPort = false
				break
			}
		}
		if isPort {
			urlStr = urlStr[:idx]
		}
	}

	// Split by dots and get the last part
	domainParts := strings.Split(urlStr, ".")
	if len(domainParts) < 2 {
		return ""
	}

	return domainParts[len(domainParts)-1]
}

// selectEndpointWithDifferentTLD attempts to select an endpoint with a TLD that hasn't been used yet
func (rc *requestContext) selectEndpointWithDifferentTLD(
	availableEndpoints protocol.EndpointAddrList,
	endpointTLDs map[protocol.EndpointAddr]string,
	usedTLDs map[string]bool,
) (protocol.EndpointAddr, error) {
	// Filter endpoints to only those with different TLDs
	var endpointsWithDifferentTLDs protocol.EndpointAddrList

	for _, endpoint := range availableEndpoints {
		if tld, exists := endpointTLDs[endpoint]; exists {
			if !usedTLDs[tld] {
				endpointsWithDifferentTLDs = append(endpointsWithDifferentTLDs, endpoint)
			}
		} else {
			// If we can't determine TLD, include it anyway
			endpointsWithDifferentTLDs = append(endpointsWithDifferentTLDs, endpoint)
		}
	}

	if len(endpointsWithDifferentTLDs) == 0 {
		return "", fmt.Errorf("no endpoints with different TLDs available")
	}

	// Use the QoS selector on the filtered list
	return rc.qosCtx.GetEndpointSelector().Select(endpointsWithDifferentTLDs)
}

// determineRelayMetricLabels determines the session state and cache effectiveness for relay metrics.
func (rc *requestContext) determineRelayMetricLabels() (sessionState, cacheEffectiveness string) {
	// Default values
	sessionState = sessionStateNormal

	// Check if we have Shannon observations to determine session state
	if rc.protocolObservations != nil && rc.protocolObservations.GetShannon() != nil {
		shannonObs := rc.protocolObservations.GetShannon().GetObservations()
		if len(shannonObs) > 0 {
			// Look for grace period or rollover patterns in the observations
			// This is a simplified heuristic - could be improved with explicit session state tracking
			for _, obs := range shannonObs {
				if obs.GetRequestError() != nil {
					sessionState = sessionStateError
					break
				}
				// If we have endpoint observations, we can infer some patterns
				if len(obs.GetEndpointObservations()) > 0 {
					sessionState = sessionStateActive
				}
			}
		}
	}

	// Determine cache effectiveness based on timing patterns
	// This is a simplified heuristic - in practice you might want more sophisticated logic
	relayDurationMs := time.Since(rc.relayStartTime).Milliseconds()
	cacheEffectiveness = categorizeRequestSetupCachePerformance(float64(relayDurationMs))

	return sessionState, cacheEffectiveness
}

// recordRelayLatencyMetrics records the end-to-end relay latency metrics.
func (rc *requestContext) recordRelayLatencyMetrics(duration float64, sessionState, cacheEffectiveness string) {
	// Only record metrics for Shannon protocol
	if rc.protocol != nil {
		// Check if this is a Shannon protocol (we could add a method to identify protocol type)
		// For now, we'll record metrics for all protocols but with service_id differentiation
		shannonmetrics.RecordRelayLatency(rc.serviceID, sessionState, cacheEffectiveness, duration)
	}
}

// HandleWebsocketRequest handles a websocket request.
func (rc *requestContext) HandleWebsocketRequest(request *http.Request, responseWriter http.ResponseWriter) error {
	// Establish a websocket connection with the selected endpoint and handle the request.
	// In this code path, we are always guaranteed to have exactly one protocol context.
	if err := rc.protocolContexts[0].HandleWebsocketRequest(rc.logger, request, responseWriter); err != nil {
		rc.logger.Warn().Err(err).Msg("Failed to establish a websocket connection.")
		return err
	}

	return nil
}

// WriteHTTPUserResponse uses the data contained in the gateway request context to write the user-facing HTTP response.
func (rc *requestContext) WriteHTTPUserResponse(w http.ResponseWriter) {
	// Always record relay latency metrics when writing the response
	defer func() {
		if rc.relayStartTime.IsZero() {
			// No start time recorded, skip metrics
			return
		}

		// Calculate end-to-end relay duration
		relayDuration := time.Since(rc.relayStartTime).Seconds()

		// Determine session state and cache effectiveness
		sessionState, cacheEffectiveness := rc.determineRelayMetricLabels()

		// Record the relay latency metrics
		rc.recordRelayLatencyMetrics(relayDuration, sessionState, cacheEffectiveness)
	}()

	// If the HTTP request was invalid, write a generic response.
	// e.g. if the specified target service ID was invalid.
	if rc.presetFailureHTTPResponse != nil {
		rc.writeHTTPResponse(rc.presetFailureHTTPResponse, w)
		return
	}

	// Processing a request only gets to this point if a QoS instance was matched to the request.
	// Use the QoS context to obtain an HTTP response.
	// There are 3 possible scenarios:
	// 	1. The QoS instance rejected the request:
	//		QoS returns a properly formatted error response.
	//               e.g. a non-JSONRPC payload for an EVM service.
	// 	2. Protocol relay failed for any reason:
	//		QoS returns a generic, properly formatted response.
	//.              e.g. a JSONRPC error response.
	//	3. Protocol relay was sent successfully:
	//		QoS returns the endpoint's response.
	//               e.g. the chain ID for a `eth_chainId` request.
	rc.writeHTTPResponse(rc.qosCtx.GetHTTPResponse(), w)
}

// writeResponse uses the supplied http.ResponseWriter to write the supplied HTTP response.
func (rc *requestContext) writeHTTPResponse(response HTTPResponse, w http.ResponseWriter) {
	for key, value := range response.GetHTTPHeaders() {
		w.Header().Set(key, value)
	}

	statusCode := response.GetHTTPStatusCode()
	// TODO_IMPROVE: introduce handling for cases where the status code is not set.
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	responsePayload := response.GetPayload()
	logger := rc.logger.With(
		"http_response_payload_length", len(responsePayload),
		"http_response_status", statusCode,
	)

	// TODO_TECHDEBT(@adshmh): Refactor to consolidate all gateway observation updates in one function.
	// Required steps:
	// 	1. Update requestContext.WriteHTTPUserResponse to return response length
	// 	2. Update Gateway.HandleHTTPServiceRequest to use length for gateway observations
	//
	// Update response size observation
	rc.gatewayObservations.ResponseSize = uint64(len(responsePayload))

	w.WriteHeader(statusCode)

	numWrittenBz, writeErr := w.Write(responsePayload)
	if writeErr != nil {
		logger.With("http_response_bytes_written", numWrittenBz).Warn().Err(writeErr).Msg("Error writing the HTTP response.")
		return
	}

	logger.Info().Msg("Completed processing the HTTP request and returned an HTTP response.")
}

// BroadcastAllObservations delivers the collected details regarding all aspects
// of the service request to all the interested parties.
//
// For example:
//   - QoS-level observations; e.g. endpoint validation results
//   - Protocol-level observations; e.g. "maxed-out" endpoints.
//   - Gateway-level observations; e.g. the request ID.
func (rc *requestContext) BroadcastAllObservations() {
	// observation-related tasks are called in Goroutines to avoid potentially blocking the HTTP handler.
	go func() {
		var (
			qosObservations qosobservations.Observations
		)

		// update gateway-level observations: no request error encountered.
		rc.updateGatewayObservations(nil)

		// update protocol-level observations: no errors encountered setting up the protocol context.
		rc.updateProtocolObservations(nil)

		if rc.protocolObservations != nil {
			err := rc.protocol.ApplyObservations(rc.protocolObservations)
			if err != nil {
				rc.logger.Warn().Err(err).Msg("error applying protocol observations.")
			}
		}

		// The service request context contains all the details the QoS needs to update its internal metrics about endpoint(s), which it should use to build
		// the observation.QoSObservations struct.
		// This ensures that separate PATH instances can communicate and share their QoS observations.
		// The QoS context will be nil if the target service ID is not specified correctly by the request.
		if rc.qosCtx != nil {
			qosObservations = rc.qosCtx.GetObservations()
			if err := rc.serviceQoS.ApplyObservations(&qosObservations); err != nil {
				rc.logger.Warn().Err(err).Msg("error applying QoS observations.")
			}
		}

		observations := &observation.RequestResponseObservations{
			HttpRequest: &rc.httpObservations,
			Gateway:     rc.gatewayObservations,
			Protocol:    rc.protocolObservations,
			Qos:         &qosObservations,
		}

		if rc.metricsReporter != nil {
			rc.metricsReporter.Publish(observations)
		}

		if rc.dataReporter != nil {
			rc.dataReporter.Publish(observations)
		}
	}()
}

// getHTTPRequestLogger returns a logger with attributes set using the supplied HTTP request.
func (rc *requestContext) getHTTPRequestLogger(httpReq *http.Request) polylog.Logger {
	var urlStr string
	if httpReq.URL != nil {
		urlStr = httpReq.URL.String()
	}

	return rc.logger.With(
		"http_req_url", urlStr,
		"http_req_host", httpReq.Host,
		"http_req_remote_addr", httpReq.RemoteAddr,
		"http_req_content_length", httpReq.ContentLength,
	)
}

// updateProtocolObservations updates the stored protocol-level observations.
// It is called at:
// - Protocol context setup error.
// - When broadcasting observations.
func (rc *requestContext) updateProtocolObservations(protocolContextSetupErrorObservation *protocolobservations.Observations) {
	// protocol observation already set: skip.
	// This happens when a protocol context setup observation was reported earlier.
	if rc.protocolObservations != nil {
		return
	}

	// protocol context setup error observation is set: skip.
	if protocolContextSetupErrorObservation != nil {
		rc.protocolObservations = protocolContextSetupErrorObservation
		return
	}

	// Check if we have multiple protocol contexts and use the first successful one
	if len(rc.protocolContexts) > 0 {
		// TODO_IN_THIS_PR: Refactor to all observations.
		observations := rc.protocolContexts[0].GetObservations()
		rc.protocolObservations = &observations
		return
	}

	// QoS rejected the request: there is no protocol context/observation.
	if rc.requestRejectedByQoS {
		return
	}

	// This should never happen: either protocol context is setup, or an observation is reported to use directly for the request.
	rc.logger.
		With("service_id", rc.serviceID).
		ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).
		Msg("SHOULD NEVER HAPPEN: protocol context is nil, but no protocol setup observation have been reported.")
}
