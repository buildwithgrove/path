package jsonrpc

// EndpointAttributeError contains error details for endpoint queries.
// An EndpointError is always associated with an Endpoint Attribute struct.
type EndpointAttributeError struct {
	// Description is set by the custom service implementation
	Description string

	// RecommendedSanction is set by the custom service implementation
	// It is under ResultError to clarify the reason a sanction was recommended.
	RecommendedSanction *Sanction

	// Internal fields used by the framework
	kind       endpointErrorKind
	rawPayload []byte
}
