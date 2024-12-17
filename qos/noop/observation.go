package noop

import (
	"github.com/buildwithgrove/path/message"
)

var _ message.ObservationSet = observationSet{}

type observationSet struct{}

func (o observationSet) MarshalJSON() ([]byte, error) {
	return nil, nil
}

func (o observationSet) Broadcast() error {
	return nil
}
