package toolkit

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/buildwithgrove/path/qos/judge"
)

type QoSSpec struct {
	ServiceName string
	ServiceProbes map[jsonrpc.Method]ServiceProbe
}

func (qs *QoSSpec) NewQoSService(logger polylog.Logger) *judge.QoS {
	// build judge.QoSDefinition using QoSSpec
	qosDefinition := judge.QoSDefinition{
		Logger: logger,
		ServiceName: qs.ServiceName,

		// Build the QoS definition components using service probes.
		EndpointQualityChecksBuilder: qs.getEndpointQualityChecksBuilder(),
		ResultBuilders: qs.getEndpointQueryResultBuilders(),
		StateUpdater: qs.getStateUpdater(),
		EndpointSelector: qs.getEndpointSelector(),
	}

	// Instantiate a new QoS service using the constructed QoS Definition.
	return qosDefinition.NewQoSService()
}

// TODO_IN_THIS_PR: implement.
func (qs *QoSSpec) getEndpointQueryResultBuilders() map[jsonrpc.Method]judge.EndpointQueryResultBuilder {
	return nil
}

func (qs *QoSSpec) getEndpointQualityChecksBuilder() judge.EndpointQualityChecksBuilder {
	return nil
}

func (qs *QoSSpec) getStateUpdater() judge.StateUpdater {
	return nil
}

func (qs *QoSSpec) getEndpointSelector() judge.EndpointSelector {
	return nil
}
