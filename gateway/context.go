package gateway

import (
	"context"
	"errors"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/observation"
	"github.com/buildwithgrove/path/protocol"
)

var (
	errHTTPRequestRejectedByParser   = errors.New("HTTP request rejected by the HTTP parser.")
	errHTTPRequestRejectedByQoS      = errors.New("HTTP request rejected by service QoS instance.")
	errHTTPRequestRejectedByProtocol = errors.New("HTTP request rejected by protocol instance.")
)

// requestContext is responsible for performing the steps necessary to complete a service request.
// As of PR #72, it is limited in scope to HTTP service requests.
type requestContext struct {
	httpRequestParser HTTPRequestParser

	// metricsReporter and dataReporter are intentionally declared separately, rather than using a slice of the same interface, to be consistent
	// with the gateway package's role of explicitly defining PATH gateway's components and their interactions.
	metricsReporter RequestResponseReporter
	dataReporter    RequestResponseReporter

	serviceID  protocol.ServiceID
	serviceQoS QoSService
	qosCtx     RequestQoSContext

	protocol    Protocol
	protocolCtx ProtocolRequestContext

	logger polylog.Logger
	// presetFailureHTTPResponse, if set, is used to return a preconstructed error response to the user.
	// For example, this is used to return an error if the specified target service ID is invalid.
	presetFailureHTTPResponse HTTPResponse

	httpObservations    observation.HTTPRequestObservations
	gatewayObservations observation.GatewayObservations
}

// InitFromHTTPRequest builds the required context for serving an HTTP request.
// e.g.:
//   - The target service ID
//   - The Service QoS instance
func (rc *requestContext) InitFromHTTPRequest(httpReq *http.Request) error {
	rc.logger = rc.getHTTPRequestLogger(httpReq)

	// TODO_MVP(@adshmh): The HTTPRequestParser should return a context, similar to QoS, which is then used to get a QoS instance and the observation set.
	// Extract the service ID and find the target service's corresponding QoS instance.
	serviceID, serviceQoS, err := rc.httpRequestParser.GetQoSService(context.TODO(), httpReq)
	if err != nil {
		rc.presetFailureHTTPResponse = rc.httpRequestParser.GetHTTPErrorResponse(context.TODO(), err)
		rc.logger.Info().Err(err).Msg(errHTTPRequestRejectedByParser.Error())
		return errHTTPRequestRejectedByParser
	}

	rc.serviceID = serviceID
	rc.serviceQoS = serviceQoS
	return nil
}

func (rc *requestContext) BuildQoSContextFromHTTP(ctx context.Context, httpReq *http.Request) error {
	// Build the payload for the requested service using the incoming HTTP request.
	// This poyload will be sent to an endpoint matching the requested service.
	qosCtx, isValid := rc.serviceQoS.ParseHTTPRequest(ctx, httpReq)
	if !isValid {
		rc.logger.Info().Msg(errHTTPRequestRejectedByQoS.Error())
		return errHTTPRequestRejectedByQoS
	}

	rc.qosCtx = qosCtx
	return nil
}

func (rc *requestContext) BuildProtocolContextFromHTTP(httpReq *http.Request) error {
	protocolCtx, err := rc.protocol.BuildRequestContext(rc.serviceID, httpReq)
	if err != nil {
		// TODO_UPNEXT(@adshmh): Add a unique identifier to each request to be used in generic user-facing error responses.
		// This will enable debugging of any potential issues.
		rc.logger.Info().Err(err).Msg(errHTTPRequestRejectedByProtocol.Error())
		return errHTTPRequestRejectedByProtocol
	}

	rc.protocolCtx = protocolCtx
	return nil
}

func (rc *requestContext) SendRelay() error {
	// Send the service request payload, to a service provider endpoint.
	endpointResponse, err := SendRelay(
		rc.protocolCtx,
		rc.qosCtx.GetServicePayload(),
		rc.qosCtx.GetEndpointSelector(),
	)

	// Ignore any errors returned from the SendRelay call above.
	// These would be protocol-level errors, which are the responsibility
	// of the specific protocol instance used in serving the request.
	// e.g. the Protocol instance should drop an endpoint that is
	// temporarily/permanently unavailable from the set returned by
	// the AvailableEndpoints() method.
	//
	// There is no action required from the QoS perspective, if no
	// responses were received from an endpoint.
	if err != nil {
		rc.logger.Warn().Err(err).Msg("Failed to send a relay request.")
		// TODO_TECHDEBT: the correct reaction to a failure in sending the relay to an endpoint and getting
		// a response could be retrying with another endpoint, depending on the error.
		// This should be revisited once a retry mechanism for failed relays is within scope.
		//
		// TODO_TECHDEBT(@adshmh): use the relay error in the response returned to the user.
		return err
	}

	// TODO_TECHDEBT: implement a service-specific retry mechanism based on the protocol's response/error:
	// This would need to distinguish between:
	// a) protocol errors, e.g. when an endpoint is maxed out for a service+app combination,
	// b) endpoint errors, e.g. when an endpoint is (temporarily) unreachable due to some network issue,
	// c) request errors: these do not result in an error from SendRelay, but the payload from the endpoint indicates
	// an error, e.g. an insufficinet funds response to a transaction: note that such validation issues on requests
	// can only be identified onchain, i.e. the requests will pass the validation by the OffchainServicesSpecsEnforcer.
	//
	// TODO_FUTURE: Support multiple concurrent relays to multiple
	// endpoints for a single user request.
	// e.g. for handling JSONRPC batch requests.
	rc.qosCtx.UpdateWithResponse(endpointResponse.EndpointAddr, endpointResponse.Bytes)

	// TODO_TECHDEBT: Enhance the returned qosCtx so it can be optionally queried on both:
	// a) whether the endpoint failed to provide a valid response, and
	// b) whether a retry with another endpoint makes sense, if a failure occurred.
	return nil
}

// WriteHTTPUserResponse uses the data contained in the gateway request context to write the user-facing HTTP response.
func (rc *requestContext) WriteHTTPUserResponse(w http.ResponseWriter) {
	// If the HTTP request was invalid, write a generic response.
	// e.g. if the specified target service ID was invalid.
	if rc.presetFailureHTTPResponse != nil {
		rc.writeHTTPResponse(rc.presetFailureHTTPResponse, w)
		return
	}

	// Processing a request only gets to this point if a QoS instance was matched to the request.
	// Use the QoS context to obtain an HTTP response.
	// There are 3 possible scenarios:
	// 	1. The QoS instance rejected the request, e.g. a non-JSONRPC payload for an EVM service:
	//		QoS returns a properly formatted error response.
	// 	2. Protocol relay failed for any reason:
	//		QoS returns a generic, properly formatted response: e.g. a JSONRPC error response.
	//	3. Protocol relay was sent successfully:
	//		QoS returns the endpoint's response: e.g. the chain ID for a `eth_chainId` request.
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

	w.WriteHeader(statusCode)
	numWrittenBz, writeErr := w.Write(responsePayload)
	if writeErr != nil {
		logger.With("http_response_bytes_writte", numWrittenBz).Warn().Err(writeErr).Msg("Error writing the HTTP response.")
		return
	}

	logger.Info().Msg("Completed processing the HTTP request and returned an HTTP response.")
}

// BroadcastAllObservations delivers the collected details regarding all aspects of the service request to all the interested parties.
// For example:
//   - QoS-level observations, e.g. endpoint validation results
//   - Protocol-level observations, e.g. "maxed-out" endpoints.
//   - Gateway-level observations, e.g. the request ID.
func (rc *requestContext) BroadcastAllObservations() {
	// observation-related tasks are called in Goroutines to avoid potentially blocking the HTTP handler.
	go func() {
		protocolObservations := rc.protocolCtx.GetObservations()
		rc.protocol.ApplyObservations(protocolObservations)

		// The service request context contains all the details the QoS needs to update its internal metrics about endpoint(s), which it should use to build
		// the observation.QoSObservations struct.
		// This ensures that separate PATH instances can communicate and share their QoS observations.
		qosObservations := rc.qosCtx.GetObservations()
		rc.serviceQoS.ApplyObservations(qosObservations)

		observations := observation.RequestResponseObservations{
			HttpRequest: &rc.httpObservations,
			Gateway:     &rc.gatewayObservations,
			Protocol:    &protocolObservations,
			Qos:         &qosObservations,
		}

		rc.metricsReporter.Publish(observations)
		rc.dataReporter.Publish(observations)
	}()
}

// getHTTPRequestLogger returns a logger with attributes set using the supplied HTTP request.
func (rc requestContext) getHTTPRequestLogger(httpReq *http.Request) polylog.Logger {
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
