// gateway package defines the components and
// their interactions necessary for operating a gateway.
// It defines, in a template design pattern function, all
// the steps involved in handling a service request.
package gateway

import (
	"context"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/relayer"
)

// Gateway performs end-to-end handling of all service requests
// through a single function, i.e. HandleHTTPServiceRequest,
// which starts from the point of receiving a user request,
// and ends once a response has been returned to the user.
// TODO_FUTURE: Currently, the only supported format for both the
// request and the response is HTTP as it is sufficient for JSONRPC,
// REST, Websockets and gRPC but may expand in the future.
type Gateway struct {
	// HTTPRequestParser is used by the gateway instance to
	// interpret an HTTP request as a pair of service ID and
	// its corresponding QoS instance.
	HTTPRequestParser

	// The relayer.Relayer instance is used to fulfill the
	// service requests received by the gateway through
	// sending the service payload to an endpoint.
	*relayer.Relayer

	// QoSPublisher is used to publish QoS-related observations.
	// It can be "local" i.e. inform the local QoS
	// instance, or publisher that sends QoS observations over
	// a messaging platform to share among multiple PATH instances.
	QoSPublisher

	MetricsPublisher RequestResponseDetailsPublisher
	DataPublisher RequestResponseDetailsPublisher

	Logger polylog.Logger
}

// HandleHTTPServiceRequest defines the steps the PATH gateway takes to
// handle a service request. It is currently limited in scope to
// service requests received over HTTP, to avoid adding any abstraction
// layers that are not necessary yet.
// TODO_FUTURE: Once other service request protocols, e.g. GRPC, are
// within scope, the HandleHTTPServiceRequest needs to be
// refactored to keep HTTP-specific details and move the generic service
// request processing steps into a common method.
//
// HandleHTTPServiceRequest is written as a template method to allow the customization of steps
// invovled in serving a service request, e.g.:
// authenticating the request, parsing into a service payload,
// sending the service payload through a relayer, etc.
//
// See the following link for more details:
// https://en.wikipedia.org/wiki/Template_method_pattern
func (g Gateway) HandleHTTPServiceRequest(ctx context.Context, httpReq *http.Request, w http.ResponseWriter) {
	var httpRes HTTPResponse

	// TODO_INCOMPLETE: add request response observation and uncomment the following line when implemented.
	// defer g.RequestResponseObserver.ObserveReqRes(ctx, httpReq, httpRes)

	// TODO_TECHDEBT: add request authentication: e.g. using a portal app ID extracted from the HTTP request's path.
	// This is currently out of scope since the gateway MVP is to accept all incoming HTTP requests.

	// TODO_MVP(@adshmh): The HTTPRequestParser should return a context, similar to QoS, which is then used to get a QoS instance and the observation set.
	// Extract the service ID and find the target service's corresponding QoS instance.
	serviceID, serviceQoS, err := g.HTTPRequestParser.GetQoSService(ctx, httpReq)
	if err != nil {
		httpRes = g.HTTPRequestParser.GetHTTPErrorResponse(ctx, err)
		g.writeResponse(ctx, httpRes, w)
		return
	}

	// TODO_TECHDEBT: add request authorization, e.g. rate limiting would block an otherwise valid service request.
	// This is currently out fo scope since the gateway MVP is to accept and serve all incoming HTTP requests.

	// Build the payload for the requested service using the incoming HTTP request.
	// This poyload will be sent to an endpoint matching the requested service.
	serviceRequestCtx, isValid := serviceQoS.ParseHTTPRequest(ctx, httpReq)
	if !isValid {
		// Use the offchain service spec to decide the HTTP response returned to the user.
		// This is service-specific because we know which service the user is requesting.
		// e.g. for a JSONRPC service, the offchain spec enforcer can return a JSONRPC-formatted payload for the HTTP response returned to the user.
		g.writeResponse(ctx, serviceRequestCtx.GetHTTPResponse(), w)
		return
	}

	// TODO_IN_THIS_PR: Use a Protocol context struct, similar to QoS, to send relays and to get the protocol-level observation set, once 
	// the relayer package's refactor PR is merged.

	// Send the service request payload, through the relayer, to a service provider endpoint.
	endpointResponse, err := g.Relayer.SendRelay(
		ctx,
		serviceID,
		serviceRequestCtx.GetServicePayload(),
		serviceRequestCtx.GetEndpointSelector(),
	)
	if err != nil {
		// TODO_TECHDEBT: the correct reaction to a failure in sending the relay to an endpoint and getting
		// a response could be retrying with another endpoint, depending on the error.
		// This should be revisited once a retry mechanism for failed relays is within scope.
		g.writeResponse(ctx, serviceRequestCtx.GetHTTPResponse(), w)

		// The serviceQoS.Observe method call is intentionally skipped here,
		// because the relayer package is expected to handle protocol-specific errors.
		return
	}

	// TODO_TECHDEBT: implement a service-specific retry mechanism based on the relayer response/error:
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
	serviceRequestCtx.UpdateWithResponse(endpointResponse.EndpointAddr, endpointResponse.Bytes)

	// TODO_TECHDEBT: Enhance the returned serviceRequestCtx so it can be optionally queried on both:
	// a) whether the endpoint failed to provide a valid response, and
	// b) whether a retry with another endpoint makes sense, if a failure occurred.
	g.writeResponse(ctx, serviceRequestCtx.GetHTTPResponse(), w)

	// TODO_MVP(@adshmh): Apply Protocol-level observations once the relayer package's refactor PR is merged:
	// https://github.com/buildwithgrove/path/pull/61

	// The service request context contains all the details the QoS needs to update its internal metrics about endpoint(s), which it should use to build
	// the observation.QoSDetails struct.
	// This ensures that separate PATH instances can communicate and share their QoS observations.
	qosObservations := serviceRequestCtx.GetObservations()
	// observation-related tasks are called in Goroutines to avoid potentially blocking the HTTP handler.
	go func() {
		g.applyQoSObservations(serviceID, serviceQoS, qosObservations)
		g.applyProtocolObservations(serviceID, protocolObservations)

		g.publishRequestResponseDetails(
			serviceID,
			gatewayObservations,
			httpRequestObservations,
			protocolObservations,
			qosObservations, 
		)
	}()
}

// TODO_INCOMPLETE: writeResponse should use the context to write the user-facing
// HTTP Response.
func (g Gateway) writeResponse(ctx context.Context, response HTTPResponse, w http.ResponseWriter) {
	for key, value := range response.GetHTTPHeaders() {
		w.Header().Set(key, value)
	}

	statusCode := response.GetHTTPStatusCode()
	// TODO_IMPROVE: introduce handling for cases where the status code is not set.
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	w.WriteHeader(statusCode)

	// TODO_TECHDEBT: add logging in case the payload is not written correctly;
	// this could be a silent failure. Gateway currently has no logger.
	_, _ = w.Write(response.GetPayload())
}

// applyQoSObservations calls the supplied QoS instance to apply the supplied observations.
func (g Gateway) applyQoSObservations(serviceID relayer.ServiceID, serviceQoS QoSService, qosObservations *observation.QoSDetails) {
	logger := g.Logger.With(
		"service", string(serviceID),
		"method": "publishQoSObservations",
	)

	if qosObservations == nil {
		logger.Info().Msg("QoS observation set is nil; skipping.")
		return
	}

	err := qos.ApplyObservations(qosObservations)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to apply QoS observations")
	}
}

// publishRequestResponseDetails delivers the collected details regarding all aspects of the service request to all the interested parties, e.g. the QoS service instance.
func (g Gateway) publishRequestResponseDetails(
	serviceID relayer.ServiceID,
	gatewayObservations *observation.Gateway,
	httpObservations *observation.HTTPRequest,
	protocolObservations *observation.Protocol,
	qosObservations *observation.Qos,
) {
	logger := g.Logger.With(
		"service", string(serviceID),
		"method": "publishRequestResponseDetails",
	)

	requestResponseDetails := observation.RequestResponseDetails {
		HTTPRequest: httpObservations,
		Gateway: gatewayObservations,
		Protocol: protocolObservations,
		QoS: qosObservations,
	}

	err = g.MetricsPublisher.Publish(requestResponseDetails)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to publish metrics")
	}

	err = g.DataPublisher.Publish(requestResponseDetails)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to publish data")
	}
}
