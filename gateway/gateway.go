// gateway package defines the components and
// their interactions necessary for operating a gateway.
// It defines, in a template design pattern function, all
// the steps involved in handling a service request.
package gateway

import (
	"context"
	"net/http"

	"github.com/buildwithgrove/path/relayer"
	reqCtx "github.com/buildwithgrove/path/request/context"
)

// Gateway performs end-to-end handling of all service requests
// through a single function, i.e. HandleHTTPServiceRequest,
// which starts from the point of receiving a user request,
// and ends once a response has been returned to the user.
// TODO_FUTURE: Currently, the only supported format for both the
// request and the response is HTTP as it is sufficient for JSONRPC,
// REST, Websockets and gRPC but may expand in the future.
type Gateway struct {
	HTTPRequestParser
	*relayer.Relayer
	RequestResponseObserver
	UserRequestAuthenticator
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

	// If the request ctx contains a userAppID, authenticate the request. This performs user data auth
	// and rate limiting auth. If the req fails authentication an HTTPResponse error is returned to the user.
	if appID := reqCtx.GetUserAppIDFromCtx(ctx); appID != "" {
		if authFailedResp := g.UserRequestAuthenticator.AuthenticateReq(ctx, httpReq, appID); authFailedResp != nil {
			g.writeResponse(ctx, authFailedResp, w)
			return
		}
	}

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
	servicePayload, err := serviceQoS.ParseHTTPRequest(ctx, httpReq)
	if err != nil {
		// Use the offchain service spec to decide the HTTP response returned to the user.
		// This is service-specific because we know which service the user is requesting.
		// e.g. for a JSONRPC service, the offchain spec enforcer can return a JSONRPC-formatted payload for the HTTP response returned to the user.
		httpRes = serviceQoS.GetHTTPErrorResponse(ctx, err)
		g.writeResponse(ctx, httpRes, w)
		return
	}

	// Send the service request payload, through the relayer, to a service provider endpoint.
	endpointResponse, err := g.Relayer.SendRelay(ctx, serviceID, servicePayload, serviceQoS)
	if err != nil {
		// TODO_TECHDEBT: the correct reaction to a failure in sending the relay to an endpoint and getting
		// a response could be retrying with another endpoint, depending on the error.
		// This should be revisited once a retry mechanism for failed relays is within scope.
		httpRes = serviceQoS.GetHTTPErrorResponse(ctx, err)
		g.writeResponse(ctx, httpRes, w)
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
	// TODO_INCOMPLETE: ParseResponse should use the supplied context of the service request to access any details about
	// the request that is rrquired to validate and parse the response.
	httpRes, err = serviceQoS.GetHTTPResponse(ctx, endpointResponse)
	if err != nil {
		httpRes = serviceQoS.GetHTTPErrorResponse(ctx, err)
	}

	g.writeResponse(ctx, httpRes, w)
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

	w.Write(response.GetPayload())
}
