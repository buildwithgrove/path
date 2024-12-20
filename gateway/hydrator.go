// TODO_UPNEXT(@adshmh): Add a mermaid diagram of the different structural
// (i.e. packages, types) components to help clarify the role of each.
package gateway

import (
	"errors"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/health"
	"github.com/buildwithgrove/path/protocol"
)

// EndpointHydrator provides the functionality required for health check.
var _ health.Check = &EndpointHydrator{}

const (
	// componentNameHydrator is the name used when reporting the status of the endpoint hydrator
	componentNameHydrator = "endpoint-hydrator"
)

// endpointHydratorRunInterval specifies the running
// interval of an endpoint hydrator.
var endpointHydratorRunInterval = 10_000 * time.Millisecond

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
	Protocol
	QoSPublisher

	// ServiceQoSGenerators provides the hydrator with the EndpointCheckGenerator
	// it needs to invoke for a service ID.
	// ServiceQoSGenerators should not be modified after the hydrator is started.
	ServiceQoSGenerators map[protocol.ServiceID]QoSEndpointCheckGenerator
	Logger               polylog.Logger

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
// to start generating and sending out the endpoint check requests.
func (eph *EndpointHydrator) Start() error {
	if eph.Protocol == nil {
		return errors.New("an instance of Protocol must be proivded.")
	}

	if eph.QoSPublisher == nil {
		return errors.New("a QoS Publisher must be provided.")
	}

	if len(eph.ServiceQoSGenerators) == 0 {
		return errors.New("at-least one covered service must be specified")
	}

	go func() {
		// TODO_IMPROVE: support configuring a custom running interval.
		ticker := time.NewTicker(endpointHydratorRunInterval)
		for {
			eph.run()
			<-ticker.C
		}
	}()

	return nil
}

func (eph *EndpointHydrator) run() {
	eph.Logger.With("services count", len(eph.ServiceQoSGenerators)).Info().Msg("Running Hydrator")

	// TODO_TECHDEBT: ensure every outgoing request (or the goroutine checking a service ID)
	// has a timeout set.
	var wg sync.WaitGroup
	// A sync.Map is optimized for the use case here,
	// i.e. each map entry is written only once.
	var successfulServiceChecks sync.Map

	for svcID, svcQoS := range eph.ServiceQoSGenerators {
		wg.Add(1)
		go func(serviceID protocol.ServiceID, serviceQoS QoSEndpointCheckGenerator) {
			defer wg.Done()

			logger := eph.Logger.With("serviceID", serviceID)

			err := eph.performChecks(serviceID, serviceQoS)
			if err != nil {
				logger.Warn().Err(err).Msg("failed to run checks for service")
				return
			}

			successfulServiceChecks.Store(svcID, true)
			logger.Info().Msg("successfully completed checks for service")
		}(svcID, svcQoS)
	}
	wg.Wait()

	eph.healthStatusMutex.Lock()
	defer eph.healthStatusMutex.Unlock()
	eph.isHealthy = eph.getHealthStatus(&successfulServiceChecks)
}

func (eph *EndpointHydrator) performChecks(serviceID protocol.ServiceID, serviceQoS QoSEndpointCheckGenerator) error {
	logger := eph.Logger.With(
		"service", string(serviceID),
	)

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

	logger = logger.With("number of endpoints", len(uniqueEndpoints))
	// TODO_IMPROVE: use a single goroutine per endpoint
	for _, endpoint := range uniqueEndpoints {
		logger.With("endpoint", endpoint.Addr()).Info().Msg("running checks against the endpoint")

		requiredChecks := serviceQoS.GetRequiredQualityChecks(endpoint.Addr())
		if len(requiredChecks) == 0 {
			logger.With("endpoint", string(endpoint.Addr())).Warn().Msg("service QoS returned 0 required checks")
			continue
		}

		for _, serviceRequestCtx := range requiredChecks {
			// TODO_IMPROVE: Sending a request here should use some method shared with
			// the user request (i.e. HTTP request) handler.
			// This would ensure that both organic, i.e. user-generated, and quality data augmenting service requests
			// take the same execution path.
			endpointResponse, err := SendRelay(
				protocolRequestCtx,
				serviceRequestCtx.GetServicePayload(),
				serviceRequestCtx.GetEndpointSelector(),
			)

			// Ignore any errors returned from the SendRelay call above.
			// These would be protocol-level errors, which are the responsibility
			// of the specific protocol instance used in serving the request.
			// e.g. the Protocol instance should drop an endpoint that is
			// temporarily/permanently unavailable from the set returned by
			// the Endpoints() method.
			//
			// There is no action required from the QoS perspective, if no
			// responses were received from an endpoint.
			if err != nil {
				// TODO_FUTURE: consider skipping the rest of the checks based on the error.
				// e.g. if the endpoint is refusing connections it may be reasonable to skip it
				// in this iteration of QoS checks.
				//
				// TODO_FUTURE: consider retrying failed service requests
				// as the failure may not be related to the quality of the endpoint.
				logger.Warn().Err(err).Msg("Failed to send relay.")
				continue
			}

			serviceRequestCtx.UpdateWithResponse(endpointResponse.EndpointAddr, endpointResponse.Bytes)

			// TODO_FUTURE: consider supplying additional data to QoS.
			// e.g. data on the latency of an endpoint.
			if err := eph.QoSPublisher.Publish(serviceRequestCtx.GetObservationSet()); err != nil {
				logger.Warn().Err(err).Msg("Failed to publish QoS observations.")
			}
		}
	}

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
	for svcID := range eph.ServiceQoSGenerators {
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
