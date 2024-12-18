package noop

import (
	"github.com/buildwithgrove/path/message"
)

// observationSet provides all the functionality defined in the message.ObservationSet interface.
var _ message.ObservationSet = observationSet{}

// observationSet provides a noop versions of all functionality required for sharing and processing
// service request and the corresponding endpoint(s) response(s) observations.
// It is an empty struct because noop qos does not produce or use any observations.
type observationSet struct{}

// MarshalJSON is a noop implementation of Marshalling for the noop observation set.
// This method implements the message.ObservationSet interface.
func (o observationSet) MarshalJSON() ([]byte, error) {
	return nil, nil
}

// Broadcast is a noop implementation of Broadcast for the noop observation set.
// This method implements the message.ObservationSet interface.
func (o observationSet) Broadcast() error {
	return nil
}
