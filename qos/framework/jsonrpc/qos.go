package jsonrpc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// QoS represents a service that processes JSONRPC requests and applies QoS policies based on data returned by endpoints.
type QoS struct {
	// Logger for diagnostics
	logger polylog.Logger

	// Service identification
	serviceID ServiceID
	
	endpointCallProcessor  endpointCallProcessor
	endpointSelector endpointSelector
	serviceState serviceState

	// TODO_MVP(@adshmh): Enable custom service QoS implementations to provide 
	// a list of allowed methods which the requestValidator needs to enforce.
}

// Stored all the fields required for identification of the service.
// Used when interpreting observations.
type ServiceID struct {
	ID string
	Description string
}

// ParseHTTPRequest handles parsing an HTTP request and validating its content
// It returns a RequestQoSContext and a boolean indicating if processing should continue
func (s *QoSService) ParseHTTPRequest(
	ctx context.Context,
	httpReq *http.Request,
) (*requestQoSContext, bool) {
	builder := requestContextBuilder{
		Logger: s.Logger,
		ServiceID: s.ServiceID,
	}
	
	// Parse the HTTP request
	builder.ParseHTTPRequest(httpReq)
	
	// Validate the JSONRPC request
	builder.ValidateRequest()
	
	// Build and return the final context
	reqContext, shouldContinue := builder.Build()
	return reqContext, shouldContinue
}

// TODO_IN_THIS_PR: implement this method
// func (qos *QoS) ParseWebsocketRequest(_ context.Context) (gateway.RequestQoSContext, bool)


func (s *QoSService) ApplyObservations(observations *qosobservations.Observations) error
) {
//	-> Framework updates the endpoints + state as part of ApplyObservations
//	-> custom ResultBuilders return the set of attributes for the endpoint.
//	--> + expiryTime to drop endpoint attributes after expiry.
	jsonrpcSvcObservations := observations.GetJsonrpc()
	endpointResults := extractEndpointResultsFromObservations(jsonrpcSvcObservations)
	return s.serviceState.UpdateFromEndpointResults(endpointResults)
}
