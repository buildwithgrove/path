package morse

import (
	"time"

	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
)

// sanction represents a penalty applied to an endpoint based on observed behavior.
// Sanctions can be temporary (e.g. session-based) or permanent (e.g. gateway restart)
// depending on the severity of the observed issue.
type sanction struct {
	// reason provides a human-readable explanation for this sanction
	reason string

	// ErrorType that triggered the sanction
	errorType protocolobservations.MorseEndpointErrorType

	// CreatedAt captures the timestamp when the sanction was created
	createdAt time.Time

	// Onchain session information when sanction was created if available
	sessionServiceID string
	sessionHeight    int
}

// buildSanctionFromObservation creates a sanction struct from an endpoint observation.
func buildSanctionFromObservation(observation *protocolobservations.MorseEndpointObservation) sanction {
	return sanction{
		// Type no longer stored in the sanction struct as it's implicitly known by
		// which store it's saved in (permanent vs session)
		reason:           observation.GetErrorDetails(),
		errorType:        observation.GetErrorType(),
		createdAt:        time.Now(),
		sessionServiceID: observation.GetSessionServiceId(),
		sessionHeight:    int(observation.GetSessionHeight()),
	}
}
