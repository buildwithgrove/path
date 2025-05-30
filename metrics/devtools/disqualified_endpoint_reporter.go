package devtools

import (
	"context"
	"net/http"

	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
)

type (
	ProtocolDataReporter interface {
		// AvailableEndpoints returns the list of available endpoints matching both the service ID
		AvailableEndpoints(
			context.Context,
			protocol.ServiceID,
			*http.Request,
		) (protocol.EndpointAddrList, protocolobservations.Observations, error)

		HydrateDisqualifiedEndpointsResponse(protocol.ServiceID, *DisqualifiedEndpointResponse)
	}
	QoSDataReporter interface {
		HydrateDisqualifiedEndpointsResponse(protocol.ServiceID, *DisqualifiedEndpointResponse)
	}
)

type DisqualifiedEndpointReporter struct {
	ProtocolLevelReporter ProtocolDataReporter
	QoSLevelReporters     map[protocol.ServiceID]QoSDataReporter
}

func (r *DisqualifiedEndpointReporter) Report(serviceID protocol.ServiceID) (DisqualifiedEndpointResponse, error) {
	availableEndpoints, _, err := r.ProtocolLevelReporter.AvailableEndpoints(context.Background(), serviceID, nil)
	if err != nil {
		return DisqualifiedEndpointResponse{}, err
	}

	disqualifiedEndpointDetails := DisqualifiedEndpointResponse{
		AvailableEndpointsCount: len(availableEndpoints),
	}

	r.ProtocolLevelReporter.HydrateDisqualifiedEndpointsResponse(serviceID, &disqualifiedEndpointDetails)

	for qosServiceID, qoSLevelReporter := range r.QoSLevelReporters {
		if serviceID != "" && qosServiceID != serviceID {
			continue
		}

		qoSLevelReporter.HydrateDisqualifiedEndpointsResponse(qosServiceID, &disqualifiedEndpointDetails)
	}

	disqualifiedEndpointDetails.InvalidEndpointsCount = disqualifiedEndpointDetails.GetDisqualifiedEndpointsCount()
	disqualifiedEndpointDetails.ValidEndpointsCount = disqualifiedEndpointDetails.AvailableEndpointsCount - disqualifiedEndpointDetails.InvalidEndpointsCount

	return disqualifiedEndpointDetails, nil
}
