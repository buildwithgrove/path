// Request package is responsible for parsing and forwarding user requests.
// It is not responsible for QoS, authorization, etc...
//
// For example, Processing should fail here only if:
// A) No service is provided - Bad Request
// B) The provided service is not found/configured for the gateway instance - Not Found
//
// The responsibility of the `request` package is to extract the service ID and find the target service's corresponding QoS instance.
// See: https://github.com/buildwithgrove/path/blob/e0067eb0f9ab0956127c952980b09909a795b300/gateway/gateway.go#L52C2-L52C45
package request

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
)

// HTTPHeaderTargetServiceID is the key used to lookup the HTTP header specifying the target service's ID.
// Please see the following link on the deprecation of X- prefix in the HTTP header parameter names and why it wasn't used.
// https://www.rfc-editor.org/rfc/rfc6648#section-3
const HTTPHeaderTargetServiceID = "target-service-id"

type (
	Parser struct {
		QoSServices map[protocol.ServiceID]gateway.QoSService
		Logger      polylog.Logger
	}
)

func NewParser(enabledServices map[protocol.ServiceID]gateway.QoSService, logger polylog.Logger) (*Parser, error) {
	return &Parser{
		QoSServices: enabledServices,
		Logger:      logger,
	}, nil
}

/* --------------------------------- HTTP Request Parsing -------------------------------- */

func (p *Parser) GetQoSService(ctx context.Context, req *http.Request) (protocol.ServiceID, gateway.QoSService, error) {

	serviceID, err := p.getServiceID(req)
	if err != nil {
		return "", nil, err
	}

	qosService, ok := p.QoSServices[serviceID]
	if !ok {
		return serviceID, nil, fmt.Errorf("service ID %q not supported", serviceID)
	}

	return serviceID, qosService, nil
}

func (p *Parser) GetHTTPErrorResponse(ctx context.Context, err error) gateway.HTTPResponse {
	if errors.Is(err, errNoServiceIDProvided) {
		return &parserErrorResponse{err: err.Error(), code: http.StatusBadRequest}
	}
	return &parserErrorResponse{err: err.Error(), code: http.StatusNotFound}
}

// getServiceID extracts the target service ID from the HTTP request's headers.
func (p *Parser) getServiceID(req *http.Request) (protocol.ServiceID, error) {
	// Prefer the custom HTTP Header for specification of the Target Service ID
	if serviceID := req.Header.Get(HTTPHeaderTargetServiceID); serviceID != "" {
		return protocol.ServiceID(serviceID), nil
	}

	return "", errNoServiceIDProvided
}
