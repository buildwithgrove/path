package framework

import (
	"fmt"
	"time"
)

// TODO_FUTURE(@adshmh): Add capability to override default sanctions via the QoSDefinition struct.
// TODO_FUTURE(@adshmh): make these sanction durations/types configurable through service config,
const (
	// Default sanction duration for empty responses
	DefaultEmptyResponseSanctionDuration = 5 * time.Minute

	// Default sanction duration for parse errors
	DefaultParseErrorSanctionDuration = 5 * time.Minute

	// Default sanction duration for no responses
	DefaultNoResponseSanctionDuration = 5 * time.Minute
)

func getRecommendedSanction(endpointErrKind EndpointErrorKind, err error) *Sanction {
	switch endpointErrKind {
	case EndpointErrKindEmptyPayload:
		return newSanctionForEmptyResponse()
	case EndpointErrKindParseErr:
		return newSanctionForUnmarshalingError(err)
	default:
		return nil
	}
}

// newSanctionForEmptyResponse returns the default sanction for empty responses.
func newSanctionForEmptyResponse() *Sanction {
	return &Sanction{
		Type: SanctionTypeTemporary,              
		Reason: "Empty response from the endpoint",
		ExpiryTime: time.Now().Add(DefaultEmptyResponseSanctionDuration),
	}
}

// newSanctionForUnmarshalingError returns the default sanction for parse errors.
func newSanctionForUnmarshalingError(err error) *Sanction {
	return &Sanction{
		Type: SanctionTypeTemporary,              
		Reason: fmt.Sprintf("Endpoint payload failed to parse into JSONRPC response: %s", err.Error()),
		ExpiryTime: time.Now().Add(DefaultParseErrorSanctionDuration),
	}
}
