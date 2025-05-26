package devtools

import "github.com/buildwithgrove/path/protocol"

type DataReporter interface {
	GetInvalidEndpointResponses(protocol.ServiceID, *InvalidEndpointResponses)
}

type InvalidEndpointsReporter struct {
	ProtocolLevelReporter DataReporter
	QoSLevelReporters     map[protocol.ServiceID]DataReporter
}

func (r *InvalidEndpointsReporter) Report(serviceID protocol.ServiceID) InvalidEndpointResponses {
	invalidEndpointResponses := InvalidEndpointResponses{}

	r.ProtocolLevelReporter.GetInvalidEndpointResponses(serviceID, &invalidEndpointResponses)

	for qosServiceID, qoSLevelReporter := range r.QoSLevelReporters {
		if serviceID != "" && qosServiceID != serviceID {
			continue
		}

		qoSLevelReporter.GetInvalidEndpointResponses(qosServiceID, &invalidEndpointResponses)
	}

	return invalidEndpointResponses
}
