package framework

type EndpointErrorKind int

const (
	_                            EndpointErrorKind = iota // skip the 0 value: it matches the "UNSPECIFIED" enum value in proto definitions.
	EndpointErrKindEmptyPayload                           // Empty payload from endpoint
	EndpointErrKindParseErr                               // Could not parse endpoint payload
	EndpointErrKindInvalidResult                          // Payload result doesn't match expected value: e.g. invalid chainID value
)

// EndpointError contains error details for endpoint queries.
// An EndpointError is always associated with an Endpoint Attribute struct.
type EndpointError struct {
	// The category of endpoint error
	ErrorKind EndpointErrorKind

	// Description is set by the custom service implementation
	Description string

	// RecommendedSanction is set by the custom service implementation
	// It is under ResultError to clarify the reason a sanction was recommended.
	RecommendedSanction *Sanction
}
