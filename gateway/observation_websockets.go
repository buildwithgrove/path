package gateway

import (
	"fmt"

	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
)

// buildConnectionEstablishmentFailureObservation creates a connection establishment failure observation
// when the protocol context build and bridge start fails.
func buildConnectionEstablishmentFailureObservation(
	serviceID protocol.ServiceID,
	selectedEndpoint protocol.EndpointAddr,
	err error,
) *protocolobservations.Observations {
	// Create a failure observation similar to what the protocol layer would create
	return &protocolobservations.Observations{
		Shannon: &protocolobservations.ShannonObservationsList{
			Observations: []*protocolobservations.ShannonRequestObservations{
				{
					ServiceId: string(serviceID),
					RequestError: &protocolobservations.ShannonRequestError{
						ErrorType:    protocolobservations.ShannonRequestErrorType_SHANNON_REQUEST_ERROR_INTERNAL,
						ErrorDetails: fmt.Sprintf("failed to build protocol context and start bridge: %v", err),
					},
					ObservationData: &protocolobservations.ShannonRequestObservations_WebsocketConnectionObservation{
						WebsocketConnectionObservation: &protocolobservations.ShannonWebsocketConnectionObservation{
							Supplier:     "", // Unknown at this point
							EndpointUrl:  string(selectedEndpoint),
							EventType:    protocolobservations.ShannonWebsocketConnectionObservation_CONNECTION_ESTABLISHMENT_FAILED,
							ErrorType:    &[]protocolobservations.ShannonEndpointErrorType{protocolobservations.ShannonEndpointErrorType_SHANNON_ENDPOINT_ERROR_UNKNOWN}[0],
							ErrorDetails: &[]string{fmt.Sprintf("failed to build protocol context and start bridge: %v", err)}[0],
						},
					},
				},
			},
		},
	}
}
