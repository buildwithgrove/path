// Request package is responsible for parsing and forwarding user requests.
// It is not responsible for QoS, authorization, etc...
// For example, Processing should fail here only if no authoritative service ID is provided - Bad Request
//
// The responsibility of the `request` package is to extract the authoritative service ID and return the target service's corresponding QoS instance.
// See: https://github.com/buildwithgrove/path/blob/e0067eb0f9ab0956127c952980b09909a795b300/gateway/gateway.go#L52C2-L52C45
package request

import (
	"context"
	"errors"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/noop"
)

// HTTPHeaderTargetServiceID is the key used to lookup the HTTP header specifying the target
// service's ID. Please see the following link on the deprecation of X- prefix in HTTP header
// parameter names and why it wasn't used: https://www.rfc-editor.org/rfc/rfc6648#section-3
const HTTPHeaderTargetServiceID = "target-service-id"

// The Parser struct is responsible for parsing the authoritative service ID from the request's
// 'target-service-id' header and returning the corresponding QoS service implementation.
type Parser struct {
	// QoSServices is the set of QoS services to which the request parser should map requests based on the extracted service ID.
	QoSServices map[protocol.ServiceID]gateway.QoSService
	Logger      polylog.Logger
}

/* --------------------------------- HTTP Request Parsing -------------------------------- */

// GetQoSService returns the QoS service implementation for the given request, as well as the authoritative service ID.
// If the service ID does not have a corresponding QoS implementation, the NoOp QoS service is returned.
func (p *Parser) GetQoSService(ctx context.Context, req *http.Request) (protocol.ServiceID, gateway.QoSService, error) {
	// Get the authoritative service ID from the request's header.
	serviceID, err := p.getServiceID(req)
	if err != nil {
		return "", nil, err
	}

	// Return the QoS service implementation for the request's service ID if it exists.
	if qosService, ok := p.QoSServices[serviceID]; ok {
		return serviceID, qosService, nil
	}

	// If the service does not have a corresponding QoS implementation,
	// return the NoOp QoS service, which will select a random endpoint.
	return serviceID, noop.NoOpQoS{}, nil
}

// getServiceID extracts the authoritative service ID from the HTTP request's `target-service-id` header.
func (p *Parser) getServiceID(req *http.Request) (protocol.ServiceID, error) {
	if serviceID := req.Header.Get(HTTPHeaderTargetServiceID); serviceID != "" {
		return protocol.ServiceID(serviceID), nil
	}
	return "", errNoServiceIDProvided
}

/* --------------------------------- HTTP Error Response -------------------------------- */

// GetHTTPErrorResponse returns an HTTP response with the appropriate status code and
// error message, which ensures the error response is returned in a valid JSON format.
func (p *Parser) GetHTTPErrorResponse(ctx context.Context, err error) gateway.HTTPResponse {
	if errors.Is(err, errNoServiceIDProvided) {
		return &parserErrorResponse{err: err.Error(), code: http.StatusBadRequest}
	}
	return &parserErrorResponse{err: err.Error(), code: http.StatusNotFound}
}
