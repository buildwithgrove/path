package jsonrpc

import (
	"errors"
	"time"
)

// TODO_IMPROVE(@adshmh): Enhance EndpointAttribute to support data types commonly used for endpoint attributes.
// TODO_MVP(@adshmh): Add and support an ExpiryTime field.
//
// EndpointAttribute represents a single data item extracted from an endpoint response.
// - Stores either a string or integer value (not both)
// - Contains error/sanction information if associated with a failure
type EndpointAttribute struct {
	// Separate typed fields for different attribute types
	stringValue *string
	intValue    *int

	// Error information if this attribute is associated with a failure.
	// To be set and handled by framework only.
	// QoS services should use helper functions to set it, e.g. from the EndpointAttributeContext struct.
	err *EndpointAttributeError
}

// GetStringValue returns the attribute value as a string
// Returns false if the attribute does not have a string value set:
// It is either an integer attribute, or has an error value set.
func (a EndpointAttribute) GetStringValue() (string, bool) {
	if a.stringValue == nil {
		return "", false
	}

	return *a.stringValue, true
}

// GetIntValue returns the attribute value as an int
// Returns false if the attribute does not have an integer value set:
// It is either an string attribute, or has an error value set.
func (a EndpointAttribute) GetIntValue() (int, bool) {
	if a.intValue == nil {
		return 0, false
	}

	return *a.intValue, true
}
