package request

import (
	"fmt"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/qos/evm"
	"github.com/buildwithgrove/path/relayer"
)

/* --------------------------------- QoS Service Name Enum -------------------------------- */

// ServiceName represents a distinct category of service. This distinction is for the purposes
// of handling different service-specific logic, such as request parsing, response building,
// and endpoint selection. For example: EVM, Solana, POKT, etc.
type ServiceName string

const (
	ServiceNameEVM    ServiceName = "evm"    // ServiceNameEVM represents the EVM service type, containing all EVM-based blockchains.
	ServiceNameSolana ServiceName = "solana" // ServiceNameSolana represents the Solana blockchain service type.
	ServiceNamePOKT   ServiceName = "pokt"   // ServiceNamePOKT represents the POKT blockchain service type.

	// TODO_IMPROVE: add other service types here
)

/* --------------------------------- QoS Service Provider -------------------------------- */

type (
	qosServiceProvider struct {
		qosServices map[relayer.ServiceID]gateway.QoSService
		logger      polylog.Logger
	}
	QoSServiceConfig struct {
		RequestTimeout time.Duration
	}
)

type qosBackend interface {
	GetEnabledServiceConfigs() map[relayer.ServiceID]QoSServiceConfig
}

// newQoSServiceProvider creates a new QoSServiceProvider with the given backend and logger.
// It initializes the QoSServiceProvider with the service IDs and types that are supported by the relayer.
func newQoSServiceProvider(backend qosBackend, logger polylog.Logger) (*qosServiceProvider, error) {

	qosServices := make(map[relayer.ServiceID]gateway.QoSService)

	// TODO_UPNEXT(@adshmh): Move config-related code/initialization to the config package.
	for serviceID := range backend.GetEnabledServiceConfigs() {

		serviceName, ok := supportedServicesToQoSServiceName[serviceID]
		if !ok {
			return nil, fmt.Errorf(errServiceIDNotEnabled.Error(), serviceID)
		}

		switch serviceName {
		case ServiceNameEVM:
			// TODO_UPNEXT(@adshmh): initialize the EVM Service QoS instance.
			// TODO_UPNEXT(@adshmh): move this initialization to the config package.
			qosServices[serviceID] = evm.NewServiceQoS(nil, logger)
		case ServiceNameSolana:
			// TODO_TECHDEBT: add solana qos service here
		case ServiceNamePOKT:
			// TODO_TECHDEBT: add pokt qos service here
		default:
			return nil, errServiceNameNotSupported
		}
	}

	return &qosServiceProvider{
		qosServices: qosServices,
		logger:      logger,
	}, nil
}

func (p *qosServiceProvider) GetQoSService(serviceID relayer.ServiceID) (gateway.QoSService, error) {
	qosService, ok := p.qosServices[serviceID]
	if !ok {
		return nil, errServiceIDNotEnabled
	}
	return qosService, nil
}
