package devtools

import (
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
)

type (
	ProtocolDataReporter interface {
		GetTotalServiceEndpointsCount(protocol.ServiceID, *http.Request) (int, error)

		HydrateDisqualifiedEndpointsResponse(protocol.ServiceID, *DisqualifiedEndpointResponse)
	}
	QoSDataReporter interface {
		HydrateDisqualifiedEndpointsResponse(protocol.ServiceID, *DisqualifiedEndpointResponse)
	}
)

type DisqualifiedEndpointReporter struct {
	Logger                polylog.Logger
	ProtocolLevelReporter ProtocolDataReporter
	QoSLevelReporters     map[protocol.ServiceID]QoSDataReporter
}

func (r *DisqualifiedEndpointReporter) Report(serviceID protocol.ServiceID, httpReq *http.Request) (DisqualifiedEndpointResponse, error) {
	r.Logger.Info().Msgf("Reporting disqualified endpoints for service ID: %s", serviceID)

	var serviceEndpointsCount int
	serviceEndpointsCount, err := r.ProtocolLevelReporter.GetTotalServiceEndpointsCount(serviceID, httpReq)
	if err != nil {
		return DisqualifiedEndpointResponse{}, err
	}

	r.Logger.Info().Msgf("DisqualifiedEndpointReporter.Report: Successfully got available endpoints for service ID: %s", serviceID)

	details := DisqualifiedEndpointResponse{
		TotalServiceEndpointsCount: serviceEndpointsCount,
	}

	// Get Protocol-level sanctioned endpoints
	r.ProtocolLevelReporter.HydrateDisqualifiedEndpointsResponse(serviceID, &details)

	// Get QoS-level sanctioned endpoints
	for qosServiceID, qoSLevelReporter := range r.QoSLevelReporters {
		if serviceID != "" && qosServiceID != serviceID {
			continue
		}

		qoSLevelReporter.HydrateDisqualifiedEndpointsResponse(qosServiceID, &details)
	}

	r.Logger.Info().Msgf("DisqualifiedEndpointReporter.Report: Successfully hydrated disqualified endpoint details for service ID: %s", serviceID)

	details.InvalidServiceEndpointsCount = details.GetDisqualifiedEndpointsCount()
	details.ValidServiceEndpointsCount = details.GetValidServiceEndpointsCount()

	return details, nil
}
