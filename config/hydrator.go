package config

import (
	"github.com/buildwithgrove/path/protocol"
)

// EndpointHydratorConfig stores all the configuration
// settings required to run an instance of the
// Endpoint Hydrator.
// The EndpointHydrator will not be started if no
// service IDs are specified.
type EndpointHydratorConfig struct {
	// ServiceIDs is the list of IDs of services to be handled by
	// the Endpoint Hydrator.
	ServiceIDs []protocol.ServiceID `yaml:"service_ids"`
}
