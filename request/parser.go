// The responsibility of the `request` package is to extract the service ID and find the target service's corresponding QoS instance.
// See: https://github.com/buildwithgrove/path/blob/e0067eb0f9ab0956127c952980b09909a795b300/gateway/gateway.go#L52C2-L52C45
//
// Request package decides how the requested service is referenced by the user (currently: subdomain of the HTTP request).
//
// Processing should fail here only if:
// A) No service is provided - Bad Request
// B) The provided service is not found/configured for the gateway instance - Not Found
package request

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/relayer"
)

type (
	Parser struct {
		Backend            Backend
		QoSServiceProvider provider
		Logger             polylog.Logger
	}
	Backend interface {
		GetEnabledServiceConfigs() map[relayer.ServiceID]QoSServiceConfig
		GetServiceIDFromAlias(string) (relayer.ServiceID, bool)
	}
)

type provider interface {
	GetQoSService(relayer.ServiceID) (gateway.QoSService, error)
}

func NewParser(backend Backend, logger polylog.Logger) (*Parser, error) {
	qosServiceProvider, err := newQoSServiceProvider(backend, logger)
	if err != nil {
		return nil, err
	}

	return &Parser{
		Backend:            backend,
		QoSServiceProvider: qosServiceProvider,
		Logger:             logger,
	}, nil
}

/* --------------------------------- HTTP Request Parsing -------------------------------- */

func (p *Parser) GetQoSService(ctx context.Context, req *http.Request) (relayer.ServiceID, gateway.QoSService, error) {

	serviceID, err := p.getServiceID(req.Host)
	if err != nil {
		return "", nil, err
	}

	qosService, err := p.QoSServiceProvider.GetQoSService(serviceID)
	if err != nil {
		return "", nil, err
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
// eg. host = "eth-mainnet.gateway.pokt.network" -> serviceID = "eth-mainnet"
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
