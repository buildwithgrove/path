package data

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	"github.com/buildwithgrove/path/observation"
	protocolobservation "github.com/buildwithgrove/path/observation/protocol"
	qosobservation "github.com/buildwithgrove/path/observation/qos"
)

// TODO_MVP(@adshmh): Remove once the data pipeline has been updated.
//

// legacyRecord contains all the fields required by the legact data pipeline.
type legacyRecord struct {
	RequestID              string  `json:"request_id"`            // Request's ID
	Region                 string  `json:"region"`                // Region where the gateway serving the request is located (Grove legacy metadata)
	PortalAccountID        string  `json:"portal_account_id"`     // Portal account ID (Grove legacy metadata)
	PortalAppID            string  `json:"portal_application_id"` // Portal application ID (Grove legacy metadata)
	ChainID                string  `json:"chain_id"`              // The ID of the service/blockchain
	ChainMethod            string  `json:"chain_method"`          // The method of the JSONRPC request: applicable only to JSONRPC-based services.
	ProtocolAppPublicKey   string  `json:"protocol_application_public_key"`
	RelayType              string  `json:"relay_type"` // Type of request: User / EndpointQualityCheck (i.e. from the EndpointHydrator)
	IsError                bool    `json:"is_error"`
	ErrorType              string  `json:"error_type"`           // Type of error encountered: user, protocol, internal.
	ErrorMessage           string  `json:"error_message"`        // Details of the error, if available.
	ErrorSource            string  `json:"error_source"`         // Hardcoded to "PATH" to inform the legacy data pipeline.
	NodeQueryTimestamp     string  `json:"node_send_ts"`         // TIMESTAMP of when the request was sent to the endpoint.
	NodeReceiveTimestamp   string  `json:"node_receive_ts"`      // TIMESTAMP of when the endpoint's response was received.
	RequestStartTimestamp  string  `json:"relay_start_ts"`       // TIMESTAMP of when the request was received.
	RequestReturnTimestamp string  `json:"relay_return_ts"`      // TIMESTAMP of when the a response was returned to the client.
	RequestRoundTripTime   float64 `json:"relay_roundtrip_time"` // Request processing time, in seconds.
	PortalTripTime         float64 `json:"portal_trip_time"`     // In seconds: Total request processing time - time spent waiting for the endpoint.
	NodeTripTime           float64 `json:"node_trip_time"`       // In seconds: Total time spent waiting for the endpoint to respond.
	RequestDataSize        float64 `json:"request_data_size"`    // In bytes: the length of the request.
	RequestDate            string  `json:"date"`                 // BigQuery DATE type, format: "2025-04-11"
	RequestTimestamp       string  `json:"ts"`                   // BigQuery TIMESTAMP type, format: "2025-04-11T14:30:00.000Z"
	NodeAddress            string  `json:"pokt_node_address"`    // Address of the endpoint that served the request.
	NodeDomain             string  `json:"pokt_node_domain"`     // URL domain of the endpoint that served the request.

	// internal fields used for tracking protocol and QoS data.
	endpointTripTime float64 // endpint response timestamp - endpoint query time, in seconds.

}

// converts a data record to legacy format, for compatibility with the existing data pipeline.
func buildLegacyDataRecord(
	logger polylog.Logger,
	observations *observation.RequestResponseObservations,
) *legacyRecord {
	// initialize the legacy-compatible data record.
	legacyRecord := &legacyRecord{}

	// Update the legacy data record from Gateway observations.
	if gatewayObservations := observations.GetGateway(); gatewayObservations != nil {
		legacyRecord = setLegacyFieldsFromGatewayObservations(logger, legacyRecord, gatewayObservations)
	}

	// TODO_MVP(@adshmh): Set legacy fields from Shannon observations.
	//
	// Extract protocol observations
	protocolObservations := observations.GetProtocol()
	// Update the data record from Morse protocol data
	morseObservations := protocolObservations.GetMorse()
	if morseObservations != nil {
		legacyRecord = setLegacyFieldsFromMorseProtocolObservations(logger, legacyRecord, morseObservations)
	}

	// Update the legacy data record from QoS observations.
	if qosObservations := observations.GetQos(); qosObservations != nil {
		legacyRecord = setLegacyFieldsFromQoSObservations(logger, legacyRecord, qosObservations)
	}

	// Set constant/calculated/inferred fields' values.
	//
	if legacyRecord.ErrorType != "" {
		// Redundant value, set to comply with the legacy data pipeline.
		legacyRecord.IsError = true
	}

	// Hardcoded to "PATH" to inform the legacy data pipeline.
	legacyRecord.ErrorSource = "PATH"

	// Time spent waiting for the endpoint's response, in seconds.
	legacyRecord.NodeTripTime = legacyRecord.endpointTripTime

	// Total request processing time - time spent waiting for the endpoint, measured in seconds.
	legacyRecord.PortalTripTime = legacyRecord.RequestRoundTripTime - legacyRecord.endpointTripTime

	return legacyRecord
}

func setLegacyFieldsFromGatewayAuthData(
	legacyRecord *legacyRecord,
	authObservations *observation.RequestAuth,
) *legacyRecord {
	legacyRecord.RequestID = authObservations.RequestId
	legacyRecord.Region = authObservations.Region

	legacyRecord.PortalAccountID = authObservations.PortalAccountId
	legacyRecord.PortalAppID = authObservations.PortalApplicationId

	return legacyRecord
}

func setLegacyFieldsFromGatewayObservations(
	logger polylog.Logger,
	legacyRecord *legacyRecord,
	observations *observation.GatewayObservations,
) *legacyRecord {
	legacyRecord = setLegacyFieldsFromGatewayAuthData(legacyRecord, observations.GetRequestAuth())

	// Track organic (i.e. from the user) and synthetic (i.e. from the endpoint hydrator) requests.
	legacyRecord.RelayType = observations.RequestType.String()

	// Update request reception and completion timestamps.
	legacyRecord.RequestStartTimestamp = formatTimestampPbForBigQueryJSON(observations.ReceivedTime)
	legacyRecord.RequestReturnTimestamp = formatTimestampPbForBigQueryJSON(observations.CompletedTime)

	// BigQuery DATE type, format: "2025-04-11"
	legacyRecord.RequestDate = observations.ReceivedTime.AsTime().Format("2006-01-02")

	// BigQuery TIMESTAMP type, format: "2025-04-11T14:30:00.000Z"
	legacyRecord.RequestTimestamp = formatTimestampPbForBigQueryJSON(observations.ReceivedTime)

	// Request processing time, in seconds.
	legacyRecord.RequestRoundTripTime = observations.CompletedTime.AsTime().Sub(observations.ReceivedTime.AsTime()).Seconds()

	return legacyRecord
}

// TODO_MVP(@adshmh): handle QoS observations for:
// - Solana
// - CometBFT
func setLegacyFieldsFromQoSObservations(
	logger polylog.Logger,
	legacyRecord *legacyRecord,
	observations *qosobservation.Observations,
) *legacyRecord {
	if evmObservations := observations.GetEvm(); evmObservations != nil {
		return setLegacyFieldsFromQoSEVMObservations(logger, legacyRecord, evmObservations)
	}

	return legacyRecord
}

const qosEVMErrorTypeStr = "QOS_EVM_"

func setLegacyFieldsFromQoSEVMObservations(
	logger polylog.Logger,
	legacyRecord *legacyRecord,
	observations *qosobservation.EVMRequestObservations,
) *legacyRecord {
	// In bytes: the length of the request: float64 type is for compatibility with the legacy data pipeline.
	legacyRecord.RequestDataSize = float64(observations.RequestPayloadLength)

	evmInterpreter := &qosobservation.EVMObservationInterpreter{
		Observations: observations,
	}

	// Extract the JSONRPC request's method.
	jsonrpcRequestMethod, _ := evmInterpreter.GetRequestMethod()
	legacyRecord.ChainMethod = jsonrpcRequestMethod

	_, requestErr, err := evmInterpreter.GetRequestStatus()
	// Could not extract request error details, skip the rest of the updates.
	if err != nil || requestErr == nil {
		return legacyRecord
	}

	legacyRecord.ErrorMessage = requestErr.String()

	switch {
	case requestErr.IsRequestError():
		legacyRecord.ErrorType = fmt.Sprintf("%s_REQUEST_ERROR", qosEVMErrorTypeStr)
	case requestErr.IsResponseError():
		legacyRecord.ErrorType = fmt.Sprintf("%s_ENDPOINT_ERROR", qosEVMErrorTypeStr)
	default:
		legacyRecord.ErrorType = fmt.Sprintf("%s_UNKNOWN_ERROR", qosEVMErrorTypeStr)
	}

	return legacyRecord
}

func setLegacyFieldsFromMorseProtocolObservations(
	logger polylog.Logger,
	legacyRecord *legacyRecord,
	observationList *protocolobservation.MorseObservationsList,
) *legacyRecord {
	// TODO_MVP(@adshmh): Simplify this if MorseObservationsList type is dropped in favor of using a single MorseRequestObservation per service request.
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

	endpointObservations := observations.GetEndpointObservations()
	// No endpoint observations: this should not happen as the request has not error set.
	// Log a warning entry.
	if len(endpointObservations) == 0 {
		logger.Warn().Err(fmt.Errorf("")).Msg("Received no Morse endpoint observations for a valid request.")
		return legacyRecord
	}

	// TODO_FUTURE(@adshmh): Update the processing method if a retry mechanism is implemented:
	// Retries will result in multiple endpoint observations for a single request.
	//
	// Use the most recent entry in the endpoint observations.
	endpointObservation := endpointObservations[len(endpointObservations)-1]

	legacyRecord.ProtocolAppPublicKey = endpointObservation.AppPublicKey

	// Set endpoint query/response timestamps
	legacyRecord.NodeQueryTimestamp = formatTimestampPbForBigQueryJSON(endpointObservation.EndpointQueryTimestamp)
	legacyRecord.NodeReceiveTimestamp = formatTimestampPbForBigQueryJSON(endpointObservation.EndpointResponseTimestamp)

	// track time spent waiting for the endpoint: required for calculating the `PortalTripTime` legacy field.
	legacyRecord.endpointTripTime = endpointObservation.EndpointResponseTimestamp.AsTime().Sub(endpointObservation.EndpointQueryTimestamp.AsTime()).Seconds()

	// Set endpoint address
	legacyRecord.NodeAddress = endpointObservation.EndpointAddr

	// Extract the endpoint's domain from its URL.
	endpointDomain, err := extractEndpointDomainFromURL(endpointObservation.EndpointUrl)
	// Error extracting the endpoint domain: log the error.
	if err != nil {
		logger.Warn().Err(fmt.Errorf("")).Msg("Received no Morse endpoint observations for a valid request.")
	}
	// Set the endpoint domain field: empty value if parsing the URL above failed.
	legacyRecord.NodeDomain = endpointDomain

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

// extractEndpointDomainFromURL extracts the domain name from a URL string.
// It handles various URL formats including those with or without protocol,
// subdomains, ports, paths, queries, and fragments.
// Returns the domain without protocol, path, query parameters, or fragments.
// Returns an empty string and error if the URL is invalid.
func extractEndpointDomainFromURL(rawURL string) (string, error) {
	// If the URL doesn't have a scheme, add one to make it parseable
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Extract just the host part (domain + port if present)
	host := parsedURL.Host

	// Remove port number if present
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	return host, nil
}
