package shannon

import (
	"time"

	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
)

// TODO_FUTURE: Consider expanding sanctions to apply across PATH instances and persist across gateway restarts.
// This is (in part) how the prior version of the Gateway worked and would require a shared pubsub queue or a shared database.
// If you're an engineer reading this, this likely sounds like a cool and fun task, but it SHOLD ONLY be designed and implemented
// AFTER all the low hanging (less sexy) quality-of-service checks have been implemented.
//
// sanction represents a penalty applied to an endpoint based on observed behavior.
// Sanctions can be temporary (e.g. session-based) or permanent (e.g. gateway restart)
// depending on the severity of the observed issue.
type sanction struct {
	// reason provides a human-readable explanation for this sanction
	reason string

	// ErrorType that triggered the sanction
	errorType protocolobservations.ShannonEndpointErrorType

	// CreatedAt captures the timestamp when the sanction was created
	createdAt time.Time

	// Onchain session information when sanction was created if available
	sessionServiceID   string
	sessionStartHeight int64
}

// buildSanctionFromObservation creates a sanction struct from an endpoint observation.
func buildSanctionFromObservation(observation *protocolobservations.ShannonEndpointObservation) sanction {
	return sanction{
		// Type no longer stored in the sanction struct as it's implicitly known by
		// which store it's saved in (permanent vs session)
		reason:             observation.GetErrorDetails(),
		errorType:          observation.GetErrorType(),
		createdAt:          time.Now(),
		sessionServiceID:   observation.GetSessionServiceId(),
		sessionStartHeight: observation.GetSessionStartHeight(),
	}
}
