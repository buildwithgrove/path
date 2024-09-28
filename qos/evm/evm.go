package evm

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/relayer"
)

// QoS struct performs the functionality defined by
// gateway package's ServiceQoS, which consists of:
// A) a QoSRequestParser which builds EVM-specific
// ServiceRequestContext objects by parsing user HTTP
// requests.
// B) an EndpointSelector, which selects an endpoint
// for performing a service request.
var _ gateway.ServiceQoS = &QoS{}

var ( // compile-time checks to ensure EVMServiceQoS implements the required interfaces
	_ gateway.QoSService         = &EVMServiceQoS{}
	_ gateway.QoSResponseBuilder = &EVMResponseBuilder{}
	_ gateway.QoSRequestParser   = &EVMRequestParser{}
	_ gateway.HTTPResponse       = &EVMHTTPResponse{}
)

// QoS is the ServiceQoS implementations for EVM-based chains.
// It contains logic specific to EVM-based chains,
// including request parsing, response building,
// and endpoint validation/selection.
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
