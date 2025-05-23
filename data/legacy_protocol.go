package data

import (
	"fmt"
	"net/url"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"golang.org/x/net/publicsuffix"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	protocolobservation "github.com/buildwithgrove/path/observation/protocol"
)

// TODO_IN_THIS_PR: implement the following function
// setLegacyFieldsFromShannonProtocolObservations

// setLegacyFieldsFromMorseProtocolObservations populates legacy record with Morse protocol data.
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
// - observationList: list of Morse protocol observations
// Returns: the populated legacy record
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

	// Update error fields if an endpoint error has occurred.
	legacyRecord = setLegacyErrFieldsFromMorseEndpointError(legacyRecord, endpointObservation)

	legacyRecord.ProtocolAppPublicKey = endpointObservation.AppPublicKey

	// Set endpoint query/response timestamps
	legacyRecord.NodeQueryTimestamp = formatTimestampPbForBigQueryJSON(endpointObservation.EndpointQueryTimestamp)
	legacyRecord.NodeReceiveTimestamp = formatTimestampPbForBigQueryJSON(endpointObservation.EndpointResponseTimestamp)

	// track time spent waiting for the endpoint: required for calculating the `PortalTripTime` legacy field.
	legacyRecord.endpointTripTime = endpointObservation.EndpointResponseTimestamp.AsTime().Sub(endpointObservation.EndpointQueryTimestamp.AsTime()).Seconds()

	// Set endpoint address
	legacyRecord.NodeAddress = endpointObservation.EndpointAddr

	// Extract the endpoint's domain from its URL.
	endpointDomain, err := extractEffectiveTLDPlusOne(endpointObservation.EndpointUrl)
	// Error extracting the endpoint domain: log the error.
	if err != nil {
		logger.With("endpoint_url", endpointObservation.EndpointUrl).Warn().Err(err).Msg("Could not extract domain from Morse endpoint URL")
		return legacyRecord
	}

	// Set the endpoint domain field: empty value if parsing the URL above failed.
	legacyRecord.NodeDomain = endpointDomain

	return legacyRecord
}

// setLegacyErrFieldsFromMorseEndpointError populates error fields in legacy record from endpoint error data.
// It handles:
// - Error type mapping
// - Error message construction
// - Sanction details when present
//
// Parameters:
// - legacyRecord: the record to update
// - endpointObservation: endpoint observation containing error data
// Returns: the updated legacy record
func setLegacyErrFieldsFromMorseEndpointError(
	legacyRecord *legacyRecord,
	endpointObservation *protocolobservation.MorseEndpointObservation,
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

// extractEffectiveTLDPlusOne extracts the "effective TLD+1" (eTLD+1) from a given URL.
// Example: "https://blog.example.co.uk" → "example.co.uk"
// - Parses the URL and validates the host.
// - Uses publicsuffix package to determine the registrable domain.
// - Returns an error if input is malformed or domain is not derivable.
func extractEffectiveTLDPlusOne(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err // malformed URL
	}

	host := parsedURL.Hostname()
	if host == "" {
		return "", fmt.Errorf("empty host") // no host in URL
	}

	etld, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return "", err // domain may not be derivable (e.g., IP, localhost)
	}
	return etld, nil
}

// formatTimestampPbForBigQueryJSON formats a protobuf Timestamp for BigQuery JSON inserts.
// BigQuery expects timestamps in RFC 3339 format: YYYY-MM-DDTHH:MM:SS[.SSSSSS]Z
func formatTimestampPbForBigQueryJSON(pbTimestamp *timestamppb.Timestamp) string {
	// Convert the protobuf timestamp to Go time.Time
	goTime := pbTimestamp.AsTime()

	// Format in RFC 3339 format which BigQuery expects
	return goTime.Format(time.RFC3339Nano)
}
