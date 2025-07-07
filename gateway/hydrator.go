// TODO_MVP(@adshmh): Add a mermaid diagram of the different structural
// (i.e. packages, types) components to help clarify the role of each.
package gateway

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/health"
	"github.com/buildwithgrove/path/protocol"
)

// EndpointHydrator provides the functionality required for health check.
var _ health.Check = &EndpointHydrator{}

// componentNameHydrator is the name used when reporting the status of the endpoint hydrator
const componentNameHydrator = "endpoint-hydrator"

// Please see the following link for details on the use of `Hydrator` word in the name.
// https://stackoverflow.com/questions/6991135/what-does-it-mean-to-hydrate-an-object
//
// EndpointHydrator augments the available dataset on quality of endpoints.
// For example, it can be used to process raw data into QoS data.
// This ensures that each service on each instance has the information
// needed to make real-time decisions to handle user requests.
//
// An example QoS transformation workflow can be:
// 1. Consulting each service's QoS instance on the checks required to validate an endpoint.
// 2. Performing the required checks on the endpoint, in the form of a (synthetic) service request.
// 3. Reporting the results back to the service's QoS instance.
type EndpointHydrator struct {
	Logger polylog.Logger

	// Protocol instance to be used by the hydrator when listing endpoints and sending relays.
	Protocol

	// ActiveQoSServices provides the hydrator with the QoS instances
	// it needs to invoke for generating synthetic service requests.
	// IMPORTANT: ActiveQoSServices should not be modified after the hydrator is started.
	ActiveQoSServices map[protocol.ServiceID]QoSService

	// MetricsReporter is used to export metrics based on observations made in handling service requests.
	MetricsReporter RequestResponseReporter

	// DataReporter is used to export, to the data pipeline, observations made in handling service requests.
	// It is declared separately from the `MetricsReporter` to be consistent with the gateway package's role
	// of explicitly defining PATH gateway's components and their interactions.
	DataReporter RequestResponseReporter

	// RunInterval is the interval at which the Endpoint Hydrator will run in milliseconds.
	RunInterval time.Duration
	// MaxEndpointCheckWorkers is the maximum number of workers that will be used to concurrently check endpoints.
	MaxEndpointCheckWorkers int

	// TODO_FUTURE: a more sophisticated health status indicator
	// may eventually be needed, e.g. one that checks whether any
	// of the attempted service requests returned a response.
	//
	// isHealthy indicates whether the hydrator's
	// most recent iteration has been successful
	// i.e. it has successfully run checks against
	// every configured service.
	isHealthy         bool
	healthStatusMutex sync.RWMutex
}

// Start should be called to signal this instance of the hydrator
// to start generating and sending endpoint check requests.
func (eph *EndpointHydrator) Start() error {
	if eph.Protocol == nil {
		return errors.New("an instance of Protocol must be provided")
	}

	if len(eph.ActiveQoSServices) == 0 {
		return errors.New("at least one QoS instance must be provided to the endpoint hydrator to start sending check requests")
	}

	go func() {
		ticker := time.NewTicker(eph.RunInterval)
		for {
			eph.run()
			<-ticker.C
		}
	}()

	return nil
}

func (eph *EndpointHydrator) run() {
	logger := eph.Logger.With("services_count", len(eph.ActiveQoSServices))
	logger.Info().Msg("Running Endpoint Hydrator")

	// TODO_TECHDEBT: ensure every outgoing request (or the goroutine checking a service ID)
	// has a timeout set.
	var wg sync.WaitGroup
	// A sync.Map is optimized for the use case here,
	// i.e. each map entry is written only once.
	var successfulServiceChecks sync.Map

	for svcID, svcQoS := range eph.ActiveQoSServices {
		wg.Add(1)
		go func(serviceID protocol.ServiceID, serviceQoS QoSService) {
			defer wg.Done()

			logger := eph.Logger.With("serviceID", serviceID)

			err := eph.performChecks(serviceID, serviceQoS)
			if err != nil {
				logger.Warn().Err(err).Msg("failed to run QoS checks for service")
				return
			}

			successfulServiceChecks.Store(svcID, true)
			logger.Info().Msg("successfully completed QoS checks for service")
		}(svcID, svcQoS)
	}
	wg.Wait()

	eph.healthStatusMutex.Lock()
	defer eph.healthStatusMutex.Unlock()

	eph.isHealthy = eph.getHealthStatus(&successfulServiceChecks)
}

func (eph *EndpointHydrator) performChecks(serviceID protocol.ServiceID, serviceQoS QoSService) error {
	logger := eph.Logger.With(
		"method", "performChecks",
		"service_id", string(serviceID),
	)

	// Passing a nil as the HTTP request, because we assume the hydrator uses "Centralized Operation Mode".
	// This implies there is no need to specify a specific app.
	// TODO_TECHDEBT(@adshmh): support specifying the app(s) used for sending/signing synthetic relay requests by the hydrator.
	// TODO_FUTURE(@adshmh): consider publishing observations if endpoint lookup fails.
	availableEndpoints, _, err := eph.AvailableEndpoints(context.TODO(), serviceID, nil)
	if err != nil || len(availableEndpoints) == 0 {
		// No session found or no endpoints available for service: skip.
		logger.Warn().Msg("no session found or no endpoints available for service when running hydrator checks.")
		// do NOT return an error: hydrator and PATH should not report unhealthy status if a single service is unavailable.
		return nil
	}

	logger = logger.With("number_of_endpoints", len(availableEndpoints))

	// Prepare a channel that will keep track of all the parallel async job to perform QoS checks on every endpoint.
	endpointCheckChan := make(chan protocol.EndpointAddr, len(availableEndpoints))

	var wgEndpoints sync.WaitGroup
	for range eph.MaxEndpointCheckWorkers {
		wgEndpoints.Add(1)

		go func() {
			defer wgEndpoints.Done()

			for endpointAddr := range endpointCheckChan {
				// Creating a new locally scoped logger
				endpointLogger := logger.With("endpoint_addr", string(endpointAddr))
				endpointLogger.Info().Msg("About to run QoS checks against the current endpoint and service")

				// Retrieve all the required QoS checks for the endpoint.
				requiredQoSChecks := serviceQoS.GetRequiredQualityChecks(endpointAddr)
				if len(requiredQoSChecks) == 0 {
					endpointLogger.Warn().Msg("No required QoS checks for endpoint and service. Skipping checks...")
					continue
				}

				// Iterate over every required QoS check for the endpoint and service.
				for _, serviceRequestCtx := range requiredQoSChecks {
					// Create a new protocol request context with a pre-selected endpoint for each request.
					// IMPORTANT: A new request context MUST be created on each iteration of the loop to
					// avoid race conditions related to concurrent access issues when running concurrent QoS checks.

					// Passing a nil as the HTTP request, because we assume the Centralized Operation Mode being used by the hydrator,
					// which means there is no need for specifying a specific app.
					// TODO_FUTURE(@adshmh): support specifying the app(s) used for sending/signing synthetic relay requests by the hydrator.
					// TODO_FUTURE(@adshmh): consider publishing observations here.
					hydratorRequestCtx, _, err := eph.BuildRequestContextForEndpoint(context.TODO(), serviceID, endpointAddr, nil)
					if err != nil {
						logger.Error().Err(err).Msg("Failed to build a protocol request context for the endpoint")
						continue
					}

					// Prepare a request context to submit a synthetic relay request to the endpoint on behalf of the gateway for QoS purposes.
					gatewayRequestCtx := requestContext{
						logger:  endpointLogger,
						context: context.TODO(),
						// TODO_MVP(@adshmh): populate the fields of gatewayObservations struct.
						// Mark the request as Synthetic using the following steps:
						// 	1. Define a `gatewayObserver` function as a field in the `requestContext` struct.
						//	2. Define a `hydratorObserver` function in this file: it should at-least set the request type as `Synthetic`
						//	3. Set the `hydratorObserver` function in the `gatewayRequestContext` below.
						gatewayObservations: getSyntheticRequestGatewayObservations(),
						serviceID:           serviceID,
						serviceQoS:          serviceQoS,
						qosCtx:              serviceRequestCtx,
						protocol:            eph.Protocol,
						protocolCtx:         hydratorRequestCtx,
						// metrics reporter for exporting metrics on hydrator service requests.
						metricsReporter: eph.MetricsReporter,
						// data reporter for exporting data on hydrator service requests to the data pipeline.
						dataReporter: eph.DataReporter,
					}

					err = gatewayRequestCtx.HandleRelayRequest()
					if err != nil {
						// TODO_FUTURE: consider skipping the rest of the checks based on the error.
						// e.g. if the endpoint is refusing connections it may be reasonable to skip it
						// in this iteration of QoS checks.
						//
						// TODO_FUTURE: consider retrying failed service requests
						// as the failure may not be related to the quality of the endpoint.
						logger.Warn().Err(err).Msg("Failed to send a relay. Only protocol-level observations will be applied.")
					}

					// publish all observations gathered through sending the synthetic service requests.
					// e.g. protocol-level, qos-level observations.
					gatewayRequestCtx.BroadcastAllObservations()
				}
			}
		}()
	}

	// Kick off the workers above for every unique endpoint.
	for _, endpointAddr := range availableEndpoints {
		endpointCheckChan <- endpointAddr
	}

	close(endpointCheckChan)

	// Wait for all workers to finish processing the endpoints.
	wgEndpoints.Wait()

	// TODO_FUTURE: publish aggregated QoS reports (in addition to reports on endpoints of a specific service)
	return nil
}

// Name is used when checking the status/health of the hydrator.
func (eph *EndpointHydrator) Name() string {
	return componentNameHydrator
}

// IsAlive returns true if the hydrator has completed 1 iteration.
// It is used to check the status/health of the hydrator
func (eph *EndpointHydrator) IsAlive() bool {
	eph.healthStatusMutex.RLock()
	defer eph.healthStatusMutex.RUnlock()

	return eph.isHealthy
}

// getHealthStatus returns the health status of the hydrator
// based on the results of the most recently completed iteration
// of running checks against service endpoints.
func (eph *EndpointHydrator) getHealthStatus(successfulServiceChecks *sync.Map) bool {
	// TODO_FUTURE: allow reporting unhealthy status if
	// certain services could not be processed.
	for svcID := range eph.ActiveQoSServices {
		value, found := successfulServiceChecks.Load(svcID)
		if !found {
			return false
		}

		successful, ok := value.(bool)
		if !ok || !successful {
			return false
		}
	}

	return true
}
