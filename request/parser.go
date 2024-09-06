// Package request provides a struct for setting and retrieving relay
// request details from the context during the relay request lifecycle.
package request

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/relayer"
	"github.com/pokt-network/poktroll/pkg/polylog"
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

// TODO_INCOMPLETE: implement this
func (p *Parser) GetHTTPErrorResponse(ctx context.Context, err error) gateway.HTTPResponse {
	return nil
}

// getServiceID gets the service ID from the request host
// eg. host = "eth-mainnet.gateway.pokt.network" -> serviceID = "eth-mainnet"
func (p *Parser) getServiceID(host string) (relayer.ServiceID, error) {
	hostParts := strings.Split(host, ".")
	if len(hostParts) < 2 {
		return "", fmt.Errorf("no service ID provided")
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
