// TODO_MVP(@adshmh): Add a mermaid diagram of the different structural
// (i.e. packages, types) components to help clarify the role of each.
package gateway

import (
	"errors"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/buildwithgrove/path/health"
	"github.com/buildwithgrove/path/observation"
	"github.com/buildwithgrove/path/protocol"
)

// EndpointHydrator provides the functionality required for health check.
var _ health.Check = &EndpointHydrator{}

const (
	// componentNameHydrator is the name used when reporting the status of the endpoint hydrator
	componentNameHydrator = "endpoint-hydrator"
)

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
			if !eph.Protocol.IsAlive() {
				eph.Logger.Warn().Msg("Protocol is not alive, skipping endpoint hydrator")
				continue
			}

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
	logger := eph.Logger.With("service_id", string(serviceID))

	// TODO_FUTURE(@adshmh): support specifying the app(s) used for sending/signing synthetic relay requests by the hydrator.
	// Passing a nil as the HTTP request, because we assume the Centralized Operation Mode being used by the hydrator, which means there is
	// no need for specifying a specific app.
	protocolRequestCtx, err := eph.Protocol.BuildRequestContext(serviceID, nil)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to build a protocol request context")
		return err
	}

	uniqueEndpoints, err := protocolRequestCtx.AvailableEndpoints()
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to get the list of available endpoints")
		return err
	}

	var checksToRun []QualityCheck
	for _, endpoint := range uniqueEndpoints {
		for _, check := range serviceQoS.GetRequiredQualityChecks(endpoint.Addr()) {
			if check.ExpiresAt().IsZero() || check.ExpiresAt().After(time.Now()) {
				checksToRun = append(checksToRun, check)
			}
		}
	}

	logger = logger.With("number_of_checks", len(checksToRun))

	jobs := make(chan QualityCheck, len(checksToRun))

	var wgChecks sync.WaitGroup
	for i := 0; i < eph.MaxEndpointCheckWorkers; i++ {
		wgChecks.Add(1)

		go func() {
			defer wgChecks.Done()

			for check := range jobs {
				// Creating a new locally scoped logger
				endpointLogger := logger.With(
					"endpoint", string(check.EndpointAddr()),
					"check", check.CheckName(),
					"expires_at", check.ExpiresAt().Format(time.RFC3339),
				)
				endpointLogger.Info().Msg("running checks against the endpoint")

				// TODO_MVP(@adshmh): populate the fields of gatewayObservations struct.
				// Mark the request as Synthetic using the following steps:
				// 	1. Define a `gatewayObserver` function as a field in the `requestContext` struct.
				//	2. Define a `hydratorObserver` function in this file: it should at-least set the request type as `Synthetic`
				//	3. Set the `hydratorObserver` function in the `gatewayRequestContext` below.
				gatewayRequestCtx := requestContext{
					logger: logger,

					gatewayObservations: getSyntheticRequestGatewayObservations(),
					serviceID:           serviceID,
					serviceQoS:          serviceQoS,
					qosCtx:              check.GetRequestContext(),
					protocol:            eph.Protocol,
					protocolCtx:         protocolRequestCtx,
				}

				err := gatewayRequestCtx.HandleRelayRequest()
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
		}()
	}

	// Kick off the workers above for every unique endpoint.
	for _, check := range checksToRun {
		jobs <- check
	}

	close(jobs)

	// Wait for all workers to finish processing the endpoints.
	wgChecks.Wait()

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

// getSyntheticRequestGatewayObservations returns the gateway-level observations for a synthetic request.
// Example: request originated from the hydrator.
func getSyntheticRequestGatewayObservations() observation.GatewayObservations {
	return observation.GatewayObservations{
		RequestType:  observation.RequestType_REQUEST_TYPE_SYNTHETIC,
		ReceivedTime: timestamppb.Now(),
	}
}
