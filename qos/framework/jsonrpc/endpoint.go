package jsonrpc

// TODO_TECHDEBT(@adshmh): Persist this state (which may include sanctions) across restarts to maintain endpoint exclusions.
// TODO_MVP(@adshmh): add an ExpiryTime field and the required support for removing expired items.
//
// Endpoint represents a service endpoint with its associated attributes.
// - Read-only for client code
// - All attributes are set internally by the framework
type Endpoint struct {
	attributes map[string]EndpointAttribute
}

// GetStringAttribute retrieves an attribute's string value, using its key.
func (e *Endpoint) GetStringAttribute(attrName string) (string, bool) {
	attr, exists := e.attributes[attrName]
	if !exists {
		return "", false
	}

	return attr.GetStringValue()
}

// GetIntAttribute retrieves an attribute's integer value, using its key.
func (e *Endpoint) GetIntAttribute(attrName string) (int, bool) {
	attr, exists := e.attributes[attrName]
	if !exists {
		return 0, false
	}

	return attr.GetIntValue()
}

// ApplyQueryResult updates the endpoint's attributes with attributes from the query result.
// It merges the EndpointAttributes from the query result into the endpoint's attributes map.
func (e *Endpoint) ApplyQueryResult(queryResult *EndpointQueryResult) {
	// Initialize attributes map if it doesn't exist
	if e.attributes == nil {
		e.attributes = make(map[string]EndpointAttribute)
	}

	// Add or update attributes from the query result
	for key, attr := range queryResult.EndpointAttributes {
		e.attributes[key] = attr
	}
}

// TODO_IN_THIS_PR: implement.
func (e *Endpoint) HasActiveSanction() (Sanction, bool) {

}
