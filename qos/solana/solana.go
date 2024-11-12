package solana

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
)

// QoS struct performs the functionality defined by gateway package's ServiceQoS,
// which consists of:
// A) a QoSRequestParser which builds Solana-specific RequestQoSContext objects,
// by parsing user HTTP requests.
// B) an EndpointSelector, which selects an endpoint for performing a service request.
var _ gateway.QoSService = &QoS{}

// QoS is the ServiceQoS implementations for the Solana blockchain.
// It contains logic specific to Solana, including request parsing,
// response building, and endpoint validation/selection.
type QoS struct {
	EndpointStore *EndpointStore
	ServiceState  *ServiceState
	Logger        polylog.Logger
}
