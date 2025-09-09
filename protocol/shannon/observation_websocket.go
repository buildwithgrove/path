package shannon

import (
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"google.golang.org/protobuf/types/known/timestamppb"

	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
)

// buildWebsocketMessageSuccessObservation creates a Shannon websocket message observation for successful message processing.
// It includes endpoint details, session information, and message-specific data.
// Used when websocket message handling succeeds.
func buildWebsocketMessageSuccessObservation(
	_ polylog.Logger,
	endpoint endpoint,
	msgSize int64,
) *protocolobservations.ShannonWebsocketMessageObservation {
	session := *endpoint.Session()
	sessionHeader := session.GetHeader()

	return &protocolobservations.ShannonWebsocketMessageObservation{
		// Endpoint information
		Supplier:           endpoint.Supplier(),
		EndpointUrl:        endpoint.PublicURL(),
		EndpointAppAddress: sessionHeader.ApplicationAddress,
		IsFallbackEndpoint: endpoint.IsFallback(),

		// Session information
		SessionServiceId:   sessionHeader.ServiceId,
		SessionId:          sessionHeader.SessionId,
		SessionStartHeight: sessionHeader.SessionStartBlockHeight,
		SessionEndHeight:   sessionHeader.SessionEndBlockHeight,

		// Message information
		MessageTimestamp:   timestamppb.New(time.Now()),
		MessagePayloadSize: msgSize,
	}
}

// buildWebsocketMessageErrorObservation creates a Shannon websocket message observation for failed message processing.
// It includes endpoint details, session information, message data, and error details.
// Used when websocket message handling fails.
func buildWebsocketMessageErrorObservation(
	endpoint endpoint,
	msgSize int64,
	errorType protocolobservations.ShannonEndpointErrorType,
	errorDetails string,
	sanctionType protocolobservations.ShannonSanctionType,
) *protocolobservations.ShannonWebsocketMessageObservation {
	session := *endpoint.Session()
	sessionHeader := session.GetHeader()

	return &protocolobservations.ShannonWebsocketMessageObservation{
		// Endpoint information
		Supplier:           endpoint.Supplier(),
		EndpointUrl:        endpoint.PublicURL(),
		EndpointAppAddress: sessionHeader.ApplicationAddress,
		IsFallbackEndpoint: endpoint.IsFallback(),

		// Session information
		SessionServiceId:   sessionHeader.ServiceId,
		SessionId:          sessionHeader.SessionId,
		SessionStartHeight: sessionHeader.SessionStartBlockHeight,
		SessionEndHeight:   sessionHeader.SessionEndBlockHeight,

		// Message information
		MessageTimestamp:   timestamppb.New(time.Now()),
		MessagePayloadSize: msgSize,

		// Error information
		ErrorType:           &errorType,
		ErrorDetails:        &errorDetails,
		RecommendedSanction: &sanctionType,
	}
}

// buildWebsocketConnectionObservation creates a Shannon websocket connection observation for connection lifecycle events.
// It includes endpoint details and session information for connection-level tracking.
// Used when websocket connection setup succeeds or when connection closes.
func buildWebsocketConnectionObservation(
	_ polylog.Logger,
	endpoint endpoint,
	eventType protocolobservations.ShannonWebsocketConnectionObservation_ConnectionEventType,
) *protocolobservations.ShannonWebsocketConnectionObservation {
	session := *endpoint.Session()
	sessionHeader := session.GetHeader()

	return &protocolobservations.ShannonWebsocketConnectionObservation{
		// Endpoint information
		Supplier:           endpoint.Supplier(),
		EndpointUrl:        endpoint.PublicURL(),
		EndpointAppAddress: sessionHeader.ApplicationAddress,
		IsFallbackEndpoint: endpoint.IsFallback(),

		// Session information
		SessionServiceId:   sessionHeader.ServiceId,
		SessionId:          sessionHeader.SessionId,
		SessionStartHeight: sessionHeader.SessionStartBlockHeight,
		SessionEndHeight:   sessionHeader.SessionEndBlockHeight,

		// Connection lifecycle
		ConnectionEstablishedTimestamp: timestamppb.New(time.Now()),
		EventType:                      eventType,
	}
}

// buildWebsocketConnectionErrorObservation creates a Shannon websocket connection observation for failed connection events.
// It includes endpoint details, session information, and error details.
// Used when websocket connection setup fails or when connection closes with an error.
func buildWebsocketConnectionErrorObservation(
	_ polylog.Logger,
	endpoint endpoint,
	errorType protocolobservations.ShannonEndpointErrorType,
	errorDetails string,
	sanctionType protocolobservations.ShannonSanctionType,
	eventType protocolobservations.ShannonWebsocketConnectionObservation_ConnectionEventType,
) *protocolobservations.ShannonWebsocketConnectionObservation {
	return &protocolobservations.ShannonWebsocketConnectionObservation{
		// Endpoint information
		Supplier:           endpoint.Supplier(),
		EndpointUrl:        endpoint.PublicURL(),
		EndpointAppAddress: endpoint.Session().GetHeader().ApplicationAddress,
		IsFallbackEndpoint: endpoint.IsFallback(),

		// Session information
		SessionServiceId:   endpoint.Session().GetHeader().ServiceId,
		SessionId:          endpoint.Session().GetHeader().SessionId,
		SessionStartHeight: endpoint.Session().GetHeader().SessionStartBlockHeight,
		SessionEndHeight:   endpoint.Session().GetHeader().SessionEndBlockHeight,

		// Error information
		ErrorType:           &errorType,
		ErrorDetails:        &errorDetails,
		RecommendedSanction: &sanctionType,

		// Connection lifecycle
		ConnectionEstablishedTimestamp: timestamppb.New(time.Now()),
		EventType:                      eventType,
	}
}
