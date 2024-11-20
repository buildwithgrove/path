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
	"github.com/buildwithgrove/path/relayer"
)

type (
	Parser struct {
		Backend     Backend
		QoSServices map[relayer.ServiceID]gateway.QoSService
		Logger      polylog.Logger
	}
	Backend interface {
		GetServiceIDFromAlias(string) (relayer.ServiceID, bool)
	}
)

func NewParser(backend Backend, enabledServices map[relayer.ServiceID]gateway.QoSService, logger polylog.Logger) (*Parser, error) {
	return &Parser{
		Backend:     backend,
		QoSServices: enabledServices,
		Logger:      logger,
	}, nil
}

/* --------------------------------- HTTP Request Parsing -------------------------------- */

func (p *Parser) GetQoSService(ctx context.Context, req *http.Request) (relayer.ServiceID, gateway.QoSService, error) {

	serviceID, err := p.getServiceID(req.Host)
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

// getServiceID gets the service ID from the request host
// eg. host = "eth.gateway.pokt.network" -> serviceID = "eth"
func (p *Parser) getServiceID(host string) (relayer.ServiceID, error) {
	hostParts := strings.Split(host, ".")
	if len(hostParts) < 2 {
		return "", errNoServiceIDProvided
	}

	subdomain := hostParts[0]

	var serviceID relayer.ServiceID
	if serviceIDFromAlias, ok := p.Backend.GetServiceIDFromAlias(subdomain); ok {
		serviceID = serviceIDFromAlias
	} else {
		serviceID = relayer.ServiceID(subdomain)
	}

	return serviceID, nil
}
