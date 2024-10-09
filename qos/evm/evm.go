package evm

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
)

// QoS struct performs the functionality defined by gateway package's ServiceQoS,
// which consists of:
// A) a QoSRequestParser which builds EVM-specific RequestQoSContext objects,
// by parsing user HTTP requests.
// B) an EndpointSelector, which selects an endpoint for performing a service request.
var _ gateway.QoSService = &QoS{}

// QoS is the ServiceQoS implementations for EVM-based chains.
// It contains logic specific to EVM-based chains, including request parsing,
// response building, and endpoint validation/selection.
type QoS struct {
	endpointStore *EndpointStore
	logger        polylog.Logger
}

func NewServiceQoS(endpointStore *EndpointStore, logger polylog.Logger) *QoS {
	return &QoS{
		endpointStore: endpointStore,
		logger:        logger,
	}
}
