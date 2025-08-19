package data

import (
	"fmt"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	shannonmetrics "github.com/buildwithgrove/path/metrics/protocol/shannon"
	protocolobservation "github.com/buildwithgrove/path/observation/protocol"
)

// setLegacyFieldsFromShannonProtocolObservations populates legacy record with Shannon protocol data.
// It processes:
// - Service ID mapping to chain ID
// - Request errors
// - Endpoint observations and errors
// - Timestamps for queries and responses
// - Endpoint location information
//
// Parameters:
// - logger: logging interface
// - legacyRecord: the record to populate
// - observationList: list of Shannon protocol observations
// Returns: the populated legacy record
func setLegacyFieldsFromShannonProtocolObservations(
	logger polylog.Logger,
	legacyRecord *legacyRecord,
	observationList *protocolobservation.ShannonObservationsList,
) *legacyRecord {
	// TODO_MVP(@adshmh): Simplify this if ShannonObservationsList type is dropped in favor of using a single ShannonRequestObservation per service request.
	//
	requestObservations := observationList.GetObservations()
	if requestObservations == nil {
		return legacyRecord
	}
	// Pick the last entry: this can be dropped once the above TODO is completed.
	observations := requestObservations[len(requestObservations)-1]

	// Use the ServiceID as the legacy record's chain ID.
	legacyRecord.ChainID = observations.ServiceId

	// Request processing error: set the fields and skip further processing.
	if requestErr := observations.GetRequestError(); requestErr != nil {
		legacyRecord.ErrorType = requestErr.ErrorType.String()
		legacyRecord.ErrorMessage = requestErr.ErrorDetails

		// Request error: no more data to add.
		return legacyRecord
	}

	// Handle different observation types based on the oneof field
	switch obsData := observations.GetObservationData().(type) {
	case *protocolobservation.ShannonRequestObservations_HttpObservations:
		return setLegacyFieldsFromHTTPObservations(logger, legacyRecord, obsData.HttpObservations)
	case *protocolobservation.ShannonRequestObservations_WebsocketEndpointObservation:
		return setLegacyFieldsFromWebsocketEndpointObservation(logger, legacyRecord, obsData.WebsocketEndpointObservation)
	case *protocolobservation.ShannonRequestObservations_WebsocketMessageObservation:
		return setLegacyFieldsFromWebsocketMessageObservation(logger, legacyRecord, obsData.WebsocketMessageObservation)
	default:
		logger.Warn().Msg("Unknown observation type received for legacy record processing")
		return legacyRecord
	}
}

// setLegacyFieldsFromHTTPObservations populates legacy record with HTTP endpoint observation data.
// This handles the original HTTP relay processing logic.
func setLegacyFieldsFromHTTPObservations(
	logger polylog.Logger,
	legacyRecord *legacyRecord,
	httpObservations *protocolobservation.ShannonHTTPEndpointObservations,
) *legacyRecord {
	endpointObservations := httpObservations.GetEndpointObservations()
	// No endpoint observations: this should not happen as the request has not error set.
	// Log a warning entry.
	if len(endpointObservations) == 0 {
		logger.Warn().Err(fmt.Errorf("")).Msg("Received no Shannon endpoint observations for a valid request.")
		return legacyRecord
	}

	// TODO_FUTURE(@adshmh): Update the processing method if a retry mechanism is implemented:
	// Retries will result in multiple endpoint observations for a single request.
	//
	// Use the most recent entry in the endpoint observations.
	endpointObservation := endpointObservations[len(endpointObservations)-1]

	// Update error fields if an endpoint error has occurred.
	legacyRecord = setLegacyErrFieldsFromShannonEndpointError(legacyRecord, endpointObservation)

	// TODO_MVP(@adshmh): surface application public key if it is a must for the data pipeline.
	legacyRecord.ProtocolAppPublicKey = endpointObservation.GetEndpointAppAddress()

	// Set endpoint query/response timestamps
	legacyRecord.NodeQueryTimestamp = formatTimestampPbForBigQueryJSON(endpointObservation.EndpointQueryTimestamp)
	legacyRecord.NodeReceiveTimestamp = formatTimestampPbForBigQueryJSON(endpointObservation.EndpointResponseTimestamp)

	// track time spent waiting for the endpoint: required for calculating the `PortalTripTime` legacy field.
	legacyRecord.endpointTripTime = endpointObservation.EndpointResponseTimestamp.AsTime().Sub(endpointObservation.EndpointQueryTimestamp.AsTime()).Seconds()

	// Set endpoint address to the supplier address.
	// Will be "fallback" in the case of a request sent to a fallback endpoint.
	legacyRecord.NodeAddress = endpointObservation.GetSupplier()

	// Extract the endpoint's domain from its URL.
	endpointDomain, err := shannonmetrics.ExtractDomainOrHost(endpointObservation.GetEndpointUrl())
	if err != nil {
		logger.With("endpoint_url", endpointObservation.EndpointUrl).Warn().Err(err).Msg("Could not extract domain from Shannon endpoint URL")
		return legacyRecord
	}

	// Set the endpoint domain field: empty value if parsing the URL above failed.
	legacyRecord.NodeDomain = endpointDomain

	return legacyRecord
}

// setLegacyFieldsFromWebsocketEndpointObservation populates legacy record with WebSocket connection observation data.
// This handles WebSocket connection establishment (not individual messages).
func setLegacyFieldsFromWebsocketEndpointObservation(
	logger polylog.Logger,
	legacyRecord *legacyRecord,
	wsEndpointObs *protocolobservation.ShannonWebsocketEndpointObservation,
) *legacyRecord {
	// Update error fields if a connection error has occurred.
	legacyRecord = setLegacyErrFieldsFromWebsocketEndpointError(legacyRecord, wsEndpointObs)

	// Set application address
	legacyRecord.ProtocolAppPublicKey = wsEndpointObs.GetEndpointAppAddress()

	// Set WebSocket connection timestamps
	legacyRecord.NodeQueryTimestamp = formatTimestampPbForBigQueryJSON(wsEndpointObs.ConnectionAttemptTimestamp)

	// Use connection established timestamp as "receive" time if available
	if wsEndpointObs.ConnectionEstablishedTimestamp != nil {
		legacyRecord.NodeReceiveTimestamp = formatTimestampPbForBigQueryJSON(wsEndpointObs.ConnectionEstablishedTimestamp)
		// Calculate connection establishment time
		legacyRecord.endpointTripTime = wsEndpointObs.ConnectionEstablishedTimestamp.AsTime().Sub(wsEndpointObs.ConnectionAttemptTimestamp.AsTime()).Seconds()
	} else {
		// TODO_INSPECT(@fredteumer): Failed connections have no response timestamp - using attempt timestamp maintains consistency with immediate HTTP failures
		legacyRecord.NodeReceiveTimestamp = formatTimestampPbForBigQueryJSON(wsEndpointObs.ConnectionAttemptTimestamp)
		legacyRecord.endpointTripTime = 0
	}

	// Set endpoint address to the supplier address.
	legacyRecord.NodeAddress = wsEndpointObs.GetSupplier()

	// Extract the endpoint's domain from its URL.
	endpointDomain, err := shannonmetrics.ExtractDomainOrHost(wsEndpointObs.GetEndpointUrl())
	if err != nil {
		logger.With("endpoint_url", wsEndpointObs.EndpointUrl).Warn().Err(err).Msg("Could not extract domain from WebSocket endpoint URL")
		return legacyRecord
	}

	// Set the endpoint domain field
	legacyRecord.NodeDomain = endpointDomain

	// TODO_INSPECT(@fredteumer): WebSocket connections lack HTTP-style methods - using identifier for analytics differentiation
	legacyRecord.ChainMethod = "websocket_connection"

	return legacyRecord
}

// setLegacyFieldsFromWebsocketMessageObservation populates legacy record with WebSocket message observation data.
// This handles individual WebSocket messages sent over an established connection.
func setLegacyFieldsFromWebsocketMessageObservation(
	logger polylog.Logger,
	legacyRecord *legacyRecord,
	wsMessageObs *protocolobservation.ShannonWebsocketMessageObservation,
) *legacyRecord {
	// Update error fields if a message error has occurred.
	legacyRecord = setLegacyErrFieldsFromWebsocketMessageError(legacyRecord, wsMessageObs)

	// Set application address
	legacyRecord.ProtocolAppPublicKey = wsMessageObs.GetEndpointAppAddress()

	// Set WebSocket message timestamp as both query and receive time
	// TODO_INSPECT(@fredteumer): WebSocket messages lack separate request/response cycles - MessageTimestamp represents processing time
	legacyRecord.NodeQueryTimestamp = formatTimestampPbForBigQueryJSON(wsMessageObs.MessageTimestamp)
	legacyRecord.NodeReceiveTimestamp = formatTimestampPbForBigQueryJSON(wsMessageObs.MessageTimestamp)

	// TODO_INSPECT(@fredteumer): WebSocket messages have no request/response latency - using connection duration if available for analytics value
	if wsMessageObs.ConnectionDurationMs != nil {
		legacyRecord.endpointTripTime = float64(*wsMessageObs.ConnectionDurationMs) / 1000.0 // Convert ms to seconds
	} else {
		legacyRecord.endpointTripTime = 0
	}

	// Set endpoint address to the supplier address.
	legacyRecord.NodeAddress = wsMessageObs.GetSupplier()

	// Extract the endpoint's domain from its URL.
	endpointDomain, err := shannonmetrics.ExtractDomainOrHost(wsMessageObs.GetEndpointUrl())
	if err != nil {
		logger.With("endpoint_url", wsMessageObs.EndpointUrl).Warn().Err(err).Msg("Could not extract domain from WebSocket message endpoint URL")
		return legacyRecord
	}

	// Set the endpoint domain field
	legacyRecord.NodeDomain = endpointDomain

	// TODO_INSPECT(@fredteumer): WebSocket messages lack HTTP-style methods and JSON-RPC extraction is QoS-level - using identifier for analytics
	legacyRecord.ChainMethod = "websocket_message"

	// TODO_INSPECT(@fredteumer): Using MessagePayloadSize as closest equivalent to HTTP request size for bandwidth analytics
	legacyRecord.RequestDataSize = float64(wsMessageObs.GetMessagePayloadSize())

	return legacyRecord
}

// setLegacyErrFieldsFromWebsocketEndpointError populates error fields in legacy record from WebSocket endpoint error data.
func setLegacyErrFieldsFromWebsocketEndpointError(
	legacyRecord *legacyRecord,
	wsEndpointObs *protocolobservation.ShannonWebsocketEndpointObservation,
) *legacyRecord {
	endpointErr := wsEndpointObs.ErrorType
	// No endpoint error has occurred: no error processing required.
	if endpointErr == nil {
		return legacyRecord
	}

	// Update ErrorType using the observed endpoint error.
	legacyRecord.ErrorType = endpointErr.String()

	// Build the endpoint error details, including any sanctions.
	var errMsg string
	if errDetails := wsEndpointObs.GetErrorDetails(); errDetails != "" {
		errMsg = fmt.Sprintf("error details: %s", errDetails)
	}

	// Add the sanction details to the error message.
	if endpointSanction := wsEndpointObs.RecommendedSanction; endpointSanction != nil {
		errMsg = fmt.Sprintf("%s, sanction: %s", errMsg, endpointSanction.String())
	}

	legacyRecord.ErrorMessage = errMsg

	return legacyRecord
}

// setLegacyErrFieldsFromWebsocketMessageError populates error fields in legacy record from WebSocket message error data.
func setLegacyErrFieldsFromWebsocketMessageError(
	legacyRecord *legacyRecord,
	wsMessageObs *protocolobservation.ShannonWebsocketMessageObservation,
) *legacyRecord {
	endpointErr := wsMessageObs.ErrorType
	// No endpoint error has occurred: no error processing required.
	if endpointErr == nil {
		return legacyRecord
	}

	// Update ErrorType using the observed endpoint error.
	legacyRecord.ErrorType = endpointErr.String()

	// Build the endpoint error details, including any sanctions.
	var errMsg string
	if errDetails := wsMessageObs.GetErrorDetails(); errDetails != "" {
		errMsg = fmt.Sprintf("error details: %s", errDetails)
	}

	// Add the sanction details to the error message.
	if endpointSanction := wsMessageObs.RecommendedSanction; endpointSanction != nil {
		errMsg = fmt.Sprintf("%s, sanction: %s", errMsg, endpointSanction.String())
	}

	legacyRecord.ErrorMessage = errMsg

	return legacyRecord
}

// setLegacyErrFieldsFromShannonEndpointError populates error fields in legacy record from endpoint error data.
// It handles:
// - Error type mapping
// - Error message construction
// - Sanction details when present
//
// Parameters:
// - legacyRecord: the record to update
// - endpointObservation: endpoint observation containing error data
// Returns: the updated legacy record
func setLegacyErrFieldsFromShannonEndpointError(
	legacyRecord *legacyRecord,
	endpointObservation *protocolobservation.ShannonEndpointObservation,
) *legacyRecord {

	endpointErr := endpointObservation.ErrorType
	// No endpoint error has occurred: no error processing required.
	if endpointErr == nil {
		return legacyRecord
	}

	// Update ErrorType using the observed endpoint error.
	legacyRecord.ErrorType = endpointErr.String()

	// Build the endpoint error details, including any sanctions.
	var errMsg string
	if errDetails := endpointObservation.GetErrorDetails(); errDetails != "" {
		errMsg = fmt.Sprintf("error details: %s", errDetails)
	}

	// Add the sanction details to the error message.
	if endpointSanction := endpointObservation.RecommendedSanction; endpointSanction != nil {
		errMsg = fmt.Sprintf("%s, sanction: %s", errMsg, endpointSanction.String())
	}

	legacyRecord.ErrorMessage = errMsg

	return legacyRecord
}

// formatTimestampPbForBigQueryJSON formats a protobuf Timestamp for BigQuery JSON inserts.
// BigQuery expects timestamps in RFC 3339 format: YYYY-MM-DDTHH:MM:SS[.SSSSSS]Z
func formatTimestampPbForBigQueryJSON(pbTimestamp *timestamppb.Timestamp) string {
	// Convert the protobuf timestamp to Go time.Time
	goTime := pbTimestamp.AsTime()

	// Format in RFC 3339 format which BigQuery expects
	return goTime.Format(time.RFC3339Nano)
}
