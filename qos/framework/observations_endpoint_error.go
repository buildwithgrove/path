package framework

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	observations "github.com/buildwithgrove/path/observation/qos/framework"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// buildObservation converts an EndpointError to an observations.EndpointError
func (ee *EndpointError) buildObservation() *observations.EndpointError {
	if ee == nil {
		return nil
	}

	observationError := &observations.EndpointError{
		Description: ee.Description,
		ErrorKind:   translateToObservationErrorKind(ee.ErrorKind),
	}

	// Include sanction information if available
	if ee.RecommendedSanction != nil {
		observationError.Sanction = &observations.Sanction{
			Reason: ee.Description,
			Type:   observations.SanctionType_SANCTION_TYPE_TEMPORARY,
		}

		// Convert expiry timestamp if available
		if !ee.RecommendedSanction.Duration.IsZero() {
			// Convert Go time.Duration to proto timestamp
			observationError.Sanction.ExpiryTimestamp = timestampProto(time.Now().Add(ee.RecommendedSanction.Duration))
		}
	}

	return observationError
}

// extractEndpointErrorFromObservation extracts an EndpointError from an observations.EndpointError
func extractEndpointErrorFromObservation(obsError *observations.EndpointError) *EndpointError {
	if obsError == nil {
		return nil
	}

	err := &EndpointError{
		Description: obsError.Description,
		ErrorKind:   translateFromObservationErrorKind(obsError.ErrorKind),
	}

	// Include sanction information if available
	if obsError.Sanction != nil {
		err.RecommendedSanction = &Sanction{}

		// Convert sanction expiry timestamp to Duration
		if obsError.Sanction.ExpiryTimestamp != nil {
			sanctionExpiry := timeFromProto(obsError.Sanction.ExpiryTimestamp)
			err.RecommendedSanction.Duration = sanctionExpiry.Sub(time.Now())
		}
	}

	return err
}

// TODO_IN_THIS_PR: verify errorKind conversion to/from proto.
//
// DEV_NOTE: you MUST update this function when changing the set of endpoint error kinds.
func translateToObservationErrorKind(errKind EndpointErrorKind) observations.ErrorKind {
	switch errKind {
	case EndpointErrorKindResponseMalformed:
		return observations.ErrorKind_ERROR_KIND_RESPONSE_MALFORMED
	case EndpointErrorKindResponseUnexpectedError:
		return observations.ErrorKind_ERROR_KIND_RESPONSE_UNEXPECTED_ERROR
	case EndpointErrorKindResponseInvalidValue:
		return observations.ErrorKind_ERROR_KIND_RESPONSE_INVALID_VALUE
	case EndpointErrorKindRequestTimedOut:
		return observations.ErrorKind_ERROR_KIND_REQUEST_TIMED_OUT
	case EndpointErrorKindInsufficientFunds:
		return observations.ErrorKind_ERROR_KIND_INSUFFICIENT_FUNDS
	case EndpointErrorKindRateLimited:
		return observations.ErrorKind_ERROR_KIND_RATE_LIMITED
	case EndpointErrorKindInternalError:
		return observations.ErrorKind_ERROR_KIND_INTERNAL_ERROR
	default:
		return observations.ErrorKind_ERROR_KIND_UNSPECIFIED
	}
}

// DEV_NOTE: you MUST update this function when changing the set of endpoint error kinds.
func translateFromObservationErrorKind(errKind observations.ErrorKind) EndpointErrorKind {
	switch errKind {
	case observations.ErrorKind_ERROR_KIND_RESPONSE_MALFORMED:
		return EndpointErrorKindResponseMalformed
	case observations.ErrorKind_ERROR_KIND_RESPONSE_UNEXPECTED_ERROR:
		return EndpointErrorKindResponseUnexpectedError
	case observations.ErrorKind_ERROR_KIND_RESPONSE_INVALID_VALUE:
		return EndpointErrorKindResponseInvalidValue
	case observations.ErrorKind_ERROR_KIND_REQUEST_TIMED_OUT:
		return EndpointErrorKindRequestTimedOut
	case observations.ErrorKind_ERROR_KIND_INSUFFICIENT_FUNDS:
		return EndpointErrorKindInsufficientFunds
	case observations.ErrorKind_ERROR_KIND_RATE_LIMITED:
		return EndpointErrorKindRateLimited
	case observations.ErrorKind_ERROR_KIND_INTERNAL_ERROR:
		return EndpointErrorKindInternalError
	default:
		return EndpointErrorKindUnspecified
	}
}

// Helper functions for proto timestamp conversion
func timestampProto(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func timeFromProto(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}
