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

	// defaultBootstrapInitialQoSDataChecks specifies the number of rounds of checks to run immediately on PATH startup.
	defaultBootstrapInitialQoSDataChecks = 5
)

/* --------------------------------- Hydrator Config Struct -------------------------------- */

// EndpointHydratorConfig stores configuration settings for running an
// Endpoint Hydrator instance to collect observations about service endpoints.
// The hydrator will not start without specified service IDs.
type EndpointHydratorConfig struct {
	// List of service IDs to be handled for observation collection
	ServiceIDs []protocol.ServiceID `yaml:"service_ids"`

	// Interval between hydrator runs during which endpoint checks are performed
	RunInterval time.Duration `yaml:"run_interval_ms"`

	// Maximum number of concurrent endpoint check workers for performance tuning
	MaxEndpointCheckWorkers int `yaml:"max_endpoint_check_workers"`

	// BootstrapInitialQoSDataChecks specifies the number of rounds of checks to run immediately on PATH startup.
	// This helps to identify and filter out invalid endpoints as soon as possible.
	BootstrapInitialQoSDataChecks int `yaml:"bootstrap_initial_qos_data_checks"`
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
	if c.BootstrapInitialQoSDataChecks == 0 {
		c.BootstrapInitialQoSDataChecks = defaultBootstrapInitialQoSDataChecks
	}
}
