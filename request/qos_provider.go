package request

import (
	"errors"
	"fmt"
	"time"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/qos/evm"
	"github.com/buildwithgrove/path/relayer"
	"github.com/pokt-network/poktroll/pkg/polylog"
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

var (
	ErrServiceIDNotEnabled     = errors.New("service ID %s not enabled")
	ErrServiceNameNotSupported = errors.New("service name %s not supported")
)

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

	for serviceID, serviceConfig := range backend.GetEnabledServiceConfigs() {

		serviceName, ok := supportedServicesToQoSServiceName[serviceID]
		if !ok {
			return nil, fmt.Errorf(ErrServiceIDNotEnabled.Error(), serviceID)
		}

		switch serviceName {
		case ServiceNameEVM:
			qosServices[serviceID] = evm.NewEVMServiceQoS(serviceConfig.RequestTimeout, logger)
		case ServiceNameSolana:
			// TODO_TECHDEBT: add solana qos service here
		case ServiceNamePOKT:
			// TODO_TECHDEBT: add pokt qos service here
		default:
			return nil, fmt.Errorf(ErrServiceNameNotSupported.Error(), serviceName)
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
		return nil, fmt.Errorf(ErrServiceIDNotEnabled.Error(), serviceID)
	}
	return qosService, nil
}
