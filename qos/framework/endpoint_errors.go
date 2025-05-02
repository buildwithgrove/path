package framework

type EndpointErrorKind int

const (
	EndpointErrKindUnspecified   EndpointErrorKind = iota // matches the "UNSPECIFIED" enum value in proto definitions.
	EndpointErrKindEmptyPayload                           // Empty payload from endpoint
	EndpointErrKindParseErr                               // Could not parse endpoint payload
	EndpointErrKindValidationErr                          // Parsed endpoint payload, in the form of JSONRPC response, failed validation.
	EndpointErrKindInvalidResult                          // Payload result doesn't match expected value: e.g. invalid chainID value
)

// TODO_FUTURE(@adshmh): Allow custom QoS implementations to provide a custom JSONRPC response:
// - Add a CustomJSONRPCResponse field to EndpointError struct.
// - Support setting the above by custom QoS implementations.
// - If set, the above should be returned to the client instead of the JSONRPC response parsed from endpoint's returned payload.
//
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
