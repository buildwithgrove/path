package gateway

import (
	"github.com/buildwithgrove/path/observation"
)

// RequestResponseReporter defines the interface for reporting all the details
// regarding a request and its corresponding response and set of events to any interested entity.
// Examples of reporters includes:
// - The MetricsReporter component: to export metrics based on the observations.
// - The DataReporter: to export the observations to an external component: eg. a Messaging system or a Database.
type RequestResponseReporter interface {
	// Publish exports the details of the service request and response(s) to the external component used by the corresponding implementation.
	Publish(observation.RequestResponseDetails) error
}
