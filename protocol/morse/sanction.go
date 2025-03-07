package morse

import (
	"time"

	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
)

// sanction represents a penalty applied to an endpoint based on observed behavior.
// Sanctions can be temporary (e.g. session-based) or permanent (e.g. gateway restart)
// depending on the severity of the observed issue.
type sanction struct {
	// Type of sanction (session or permanent)
	Type protocolobservations.MorseSanctionType

	// Reason provides a human-readable explanation for this sanction
	Reason string

	// ErrorType that triggered the sanction
	ErrorType protocolobservations.MorseEndpointErrorType

	// CreatedAt captures the timestamp when the sanction was created
	CreatedAt time.Time

	// Onchain session information when sanction was created if available
	SessionChain  string
	SessionHeight int
}
