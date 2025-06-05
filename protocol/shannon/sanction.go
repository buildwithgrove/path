package shannon

import (
	"time"

	"github.com/buildwithgrove/path/metrics/devtools"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
)

// TODO_FUTURE:
// - Consider expanding sanctions to apply across PATH instances and persist across gateway restarts.
// - The prior version of the Gateway worked this way, requiring a shared pubsub queue or database.
// - Only design and implement this after all basic quality-of-service checks are done.
//
// sanction represents a penalty applied to an endpoint based on observed behavior.
// Sanctions can be temporary (session-based) or permanent (gateway restart),
// depending on the severity of the observed issue.
type sanction struct {
	// reason: human-readable explanation for this sanction.
	reason string

	// errorType: the ErrorType that triggered the sanction.
	errorType protocolobservations.ShannonEndpointErrorType

	// createdAt: timestamp when the sanction was created.
	createdAt time.Time

	// sessionServiceID, sessionStartHeight: onchain session info at sanction creation (if available).
	sessionServiceID   string
	sessionStartHeight int64
}

// buildSanctionFromObservation creates a sanction struct from an endpoint observation.
func buildSanctionFromObservation(observation *protocolobservations.ShannonEndpointObservation) sanction {
	return sanction{
		reason:             observation.GetErrorDetails(),
		errorType:          observation.GetErrorType(),
		createdAt:          time.Now(),
		sessionServiceID:   observation.GetSessionServiceId(),
		sessionStartHeight: observation.GetSessionStartHeight(),
	}
}

// toSanctionDetails converts a sanction to a devtools.SanctionedEndpoint struct.
func (s sanction) toSanctionDetails(
	supplierAddress string,
	endpointURL string,
	appAddr string,
	sessionID string,
	sanctionType protocolobservations.ShannonSanctionType,
) devtools.SanctionedEndpoint {
	return devtools.SanctionedEndpoint{
		SupplierAddress: supplierAddress,
		EndpointURL:     endpointURL,
		AppAddr:         appAddr,
		SessionID:       sessionID,
		ServiceID:       protocol.ServiceID(s.sessionServiceID),
		Reason:          s.reason,
		SanctionType:    protocolobservations.ShannonSanctionType_name[int32(sanctionType)],
		ErrorType:       protocolobservations.ShannonEndpointErrorType_name[int32(s.errorType)],
		SessionHeight:   s.sessionStartHeight,
		CreatedAt:       s.createdAt,
	}
}
