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
	"strings"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
)

// HTTPHeaderTargetServiceID is the key used to lookup the HTTP header specifying the target service's ID.
// Please see the following link for more details on not including `X-` prefix in the HTTP header parameter names.
// https://www.rfc-editor.org/rfc/rfc6648#section-3
const HTTPHeaderTargetServiceID = "target-service-id"

type (
	Parser struct {
		Backend     Backend
		QoSServices map[protocol.ServiceID]gateway.QoSService
		Logger      polylog.Logger
	}
	Backend interface {
		GetServiceIDFromAlias(string) (protocol.ServiceID, bool)
	}
)

func NewParser(backend Backend, enabledServices map[protocol.ServiceID]gateway.QoSService, logger polylog.Logger) (*Parser, error) {
	return &Parser{
		Backend:     backend,
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

// getServiceID extracts the target service ID from the supplied HTTP request.
// As of now, it supports two options for specifying the target service ID, in the order of priority:
// 1. The value of the HTTP Header target-service-id, if defined.
// e.g. `target-service-id: eth` is interpreted as `eth` target service ID.
// 2. The subdomain of the HTTP request's Host field.
// eg. host = "eth.gateway.pokt.network" -> serviceID = "eth"
func (p *Parser) getServiceID(req *http.Request) (protocol.ServiceID, error) {
	// Prefer the custom HTTP Header for specification of the Target Service ID
	serviceID := req.Header.Get(HTTPHeaderTargetServiceID)
	if serviceID != "" {
		return p.getServiceIDFromAlias(serviceID), nil
	}

	// Fallback to using the HTTP request's host field's domain if the custom HTTP header is not set.
	hostParts := strings.Split(req.Host, ".")
	if len(hostParts) < 2 {
		return "", errNoServiceIDProvided
	}

	subdomain := hostParts[0]
	return p.getServiceIDFromAlias(subdomain), nil
}

// TODO_TECHDEBT(@adshmh): consider removing the alias concept altogether: it look like a DNS/Load Balancer level concept rather than a gateway feature.
// getServiceIDFromAlias returns the service ID for the supplied alias. The serviceAlias is returned as-is if no matching service IDs are found.
func (p *Parser) getServiceIDFromAlias(serviceAlias string) protocol.ServiceID {
	if serviceIDFromAlias, ok := p.Backend.GetServiceIDFromAlias(serviceAlias); ok {
		return serviceIDFromAlias
	}

	return protocol.ServiceID(serviceAlias)
}
