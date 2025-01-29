package gateway

import (
	"context"
	"errors"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/buildwithgrove/path/observation"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
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
	protocol    Protocol
	protocolCtx ProtocolRequestContext

	// presetFailureHTTPResponse, if set, is used to return a preconstructed error response to the user.
	// For example, this is used to return an error if the specified target service ID is invalid.
	presetFailureHTTPResponse HTTPResponse

	// httpObservations stores the observations related to the HTTP request.
	httpObservations observation.HTTPRequestObservations
	// gatewayObservations stores gateway related observations.
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
	rc.gatewayObservations.ServiceId = string(serviceID)
	rc.serviceQoS = serviceQoS
	return nil
}

// BuildQoSContextFromHTTP builds the QoS context instance using the supplied HTTP request's payload.
func (rc *requestContext) BuildQoSContextFromHTTP(ctx context.Context, httpReq *http.Request) error {
	// TODO_MVP(@adshmh): Add an HTTP request size metric/observation at the gateway/http level.
	// This needs the following steps:
	// 	1. Udate the QoSService interface to Parse a custom struct including a payload of type []byte.
	//	2. Read the HTTP request's body in the `request` package and return the struct required by the updated QoS Service interface.
	//	3. Export HTTP-related observations from the `request` package at the time of reading the HTTP request's body.
	//
	// Build the payload for the requested service using the incoming HTTP request.
	// This payload will be sent to an endpoint matching the requested service.
	qosCtx, isValid := rc.serviceQoS.ParseHTTPRequest(ctx, httpReq)
	if !isValid {
		rc.logger.Info().Msg(errHTTPRequestRejectedByQoS.Error())
		return errHTTPRequestRejectedByQoS
	}

	rc.qosCtx = qosCtx
	return nil
}

// BuildProtocolContextFromHTTP builds the Protocol context using the supplied HTTP request.
// The constructed Protocol instance will be used for:
//   - Sending relays to endpoint(s)
//   - Getting the list of protocol-level observations.
func (rc *requestContext) BuildProtocolContextFromHTTP(httpReq *http.Request) error {
	protocolCtx, err := rc.protocol.BuildRequestContext(rc.serviceID, httpReq)
	if err != nil {
		// TODO_MVP(@adshmh): Add a unique identifier to each request to be used in generic user-facing error responses.
		// This will enable debugging of any potential issues.
		rc.logger.Info().Err(err).Msg(errHTTPRequestRejectedByProtocol.Error())
		return errHTTPRequestRejectedByProtocol
	}

	rc.protocolCtx = protocolCtx
	return nil
}

// HandleRelayRequest sends a relay from the perspective of a gateway.
// It performs the following steps:
//  1. Selects an endpoint using the QoS context.
//  2. Sends the relay to the selected endpoint, using the protocol context.
//  3. Processes the endpoint's response using the QoS context.
//
// HandleRelayRequest is written as a template method to allow the customization of key steps,
// e.g. endpoint selection and protocol-specific details of sending a relay.
// See the following link for more details:
// https://en.wikipedia.org/wiki/Template_method_pattern
func (rc *requestContext) HandleRelayRequest() error {
	// Make an endpoint selection using the QoS context.
	if err := rc.protocolCtx.SelectEndpoint(rc.qosCtx.GetEndpointSelector()); err != nil {
		rc.logger.Warn().Err(err).Msg("SendRelay: error selecting an endpoint.")
		return err
	}

	// Send the service request payload, through the protocol context, to the selected endpoint.
	endpointResponse, err := rc.protocolCtx.HandleServiceRequest(rc.qosCtx.GetServicePayload())
	if err != nil {
		rc.logger.Warn().Err(err).Msg("Failed to send a relay request.")
		// TODO_TECHDEBT(@commoddity): the correct reaction to a failure in sending the relay to an endpoint and getting
		// a response could be retrying with another endpoint, depending on the error.
		// This should be revisited once a retry mechanism for failed relays is within scope.
		//
		// TODO_TECHDEBT(@adshmh): use the relay error in the response returned to the user.
		return err
	}

	// TODO_TECHDEBT(@commoddity): implement a service-specific retry mechanism based on the protocol's response/error:
	// This would need to distinguish between:
	// 1) protocol errors, e.g. when an endpoint is maxed out for a service+app combination,
	// 2) QoS errors, e.g.:
	// 	A. The request is invalid: e.g. a JSONRPC request with no specified method.
	//	B. An endpoint returns an invalid response.
	//
	// TODO_FUTURE: Support multiple concurrent relays to multiple endpoints for a single user request.
	// e.g. for handling JSONRPC batch requests.
	rc.qosCtx.UpdateWithResponse(endpointResponse.EndpointAddr, endpointResponse.Bytes)

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
	// This will require the following:
	// 	1. Update the WriteHTTPUserResponse method of `requestContext` struct to return the length of the response.
	//	2. Update the `HandleHTTPServiceRequest` method of `Gateway` struct to use the above for updating Gateway observations.
	//
	// Update response size observation
	rc.gatewayObservations.ResponseSize = uint64(len(responsePayload))

	w.WriteHeader(statusCode)

	numWrittenBz, writeErr := w.Write(responsePayload)
	if writeErr != nil {
		logger.With("http_response_bytes_writte", numWrittenBz).Warn().Err(writeErr).Msg("Error writing the HTTP response.")
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
	// Update the request completion time on the gateway observation
	rc.gatewayObservations.CompletedTime = timestamppb.Now()

	var (
		protocolObservations protocolobservations.Observations
		qosObservations      qosobservations.Observations
	)

	// observation-related tasks are called in Goroutines to avoid potentially blocking the HTTP handler.
	go func() {
		if rc.protocolCtx != nil {
			protocolObservations = rc.protocolCtx.GetObservations()
			if err := rc.protocol.ApplyObservations(&protocolObservations); err != nil {
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
			Gateway:     &rc.gatewayObservations,
			Protocol:    &protocolObservations,
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
