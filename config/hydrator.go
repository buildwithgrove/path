package config

import (
	"time"

	"github.com/buildwithgrove/path/protocol"
)

/* --------------------------------- Hydrator Config Defaults -------------------------------- */

var (
	// endpointHydratorRunInterval specifies the run interval of an endpoint hydrator in milliseconds.
	defaultEndpointHydratorRunInterval = 10_000 * time.Millisecond
	// defaultMaxEndpointCheckWorkers specifies the maximum number of workers that will be used to concurrently check endpoints.
	defaultMaxEndpointCheckWorkers = 100
)

/* --------------------------------- Hydrator Config Struct -------------------------------- */

// EndpointHydratorConfig stores all the configuration
// settings required to run an instance of the
// Endpoint Hydrator.
// The EndpointHydrator will not be started if no
// service IDs are specified.
type EndpointHydratorConfig struct {
	// ServiceIDs is the list of IDs of services to be handled by the Endpoint Hydrator.
	ServiceIDs []protocol.ServiceID `yaml:"service_ids"`
	// RunInterval is the interval at which the Endpoint Hydrator will run.
	RunInterval time.Duration `yaml:"run_interval_ms"`
	// MaxEndpointCheckWorkers is the maximum number of
	// workers that will be used to concurrently check endpoints.
	MaxEndpointCheckWorkers int `yaml:"max_endpoint_check_workers"`
}

/* --------------------------------- Hydrator Config Private Helpers -------------------------------- */

// hydrateHydratorDefaults assigns default values to HydratorConfig fields if they are not set.
func (c *EndpointHydratorConfig) hydrateHydratorDefaults() {
	if c.RunInterval == 0 {
		c.RunInterval = defaultEndpointHydratorRunInterval
	}
	if c.MaxEndpointCheckWorkers == 0 {
		c.MaxEndpointCheckWorkers = defaultMaxEndpointCheckWorkers
	}
}
