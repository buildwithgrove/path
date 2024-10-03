// TODO_UPNEXT(@adshmh): Add a mermaid diagram of the different structural
// (i.e. packages, types) components to help clarify the role of each.
package gateway

import (
	"context"
	"errors"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/relayer"
)

// endpointHydratorRunIntervalMillisec specifies the running
// interval of an endpoint hydrator.
const endpointHydratorRunIntervalMillisec = 30_000

// TODO_UPNEXT(@adshmh): Complete the following to remove the confusing Protocol interface below:
//
//	1- Split the relayer package's Protocol interface.
//	2- Import the appropriate interface here, e.g. a new `EndpointProvider` interface.
//	3- Update/remove the comment below.
//
// Protocol specifies the interactions of the EndpointHydrator with
// the underlying protocol.
// It is defined separately, rather than reusing relayer.Protocol interface,
// to ensure only minimum necessary capabilities are available to the augmenter.
type Protocol interface {
	Endpoints(relayer.ServiceID) (map[relayer.AppAddr][]relayer.Endpoint, error)
}

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
	*relayer.Relayer
	QoSPublisher
	ServiceQoSGenerators map[relayer.ServiceID]QoSEndpointCheckGenerator
	Logger               polylog.Logger
}

// Start should be called to signal this instance of the hydrator
// to start generating and sending out the endpoint check requests.
func (eph *EndpointHydrator) Start() error {
	if eph.Protocol == nil {
		return errors.New("a Protocol instance must be proivded.")
	}

	if eph.Relayer == nil {
		return errors.New("a Relayer must be provided.")
	}

	if eph.QoSPublisher == nil {
		return errors.New("a QoS Publisher must be provided.")
	}

	if len(eph.ServiceQoSGenerators) == 0 {
		return errors.New("at-least one covered service must be specified")
	}

	go func() {
		// TODO_IMPROVE: support configuring a custom running interval.
		ticker := time.NewTicker(endpointHydratorRunIntervalMillisec * time.Millisecond)
		for {
			eph.run()
			<-ticker.C
		}
	}()

	return nil
}

func (eph *EndpointHydrator) run() {
	for svcID, svcQoS := range eph.ServiceQoSGenerators {
		go func(serviceID relayer.ServiceID, serviceQoS QoSEndpointCheckGenerator) {
			eph.performChecks(serviceID, serviceQoS)
		}(svcID, svcQoS)
	}

	// TODO_IMPROVE: use waitgroups to wait for all goroutines to finish before returning.
}

func (eph *EndpointHydrator) performChecks(serviceID relayer.ServiceID, serviceQoS QoSEndpointCheckGenerator) {
	logger := eph.Logger.With(
		"service", string(serviceID),
	)

	allEndpoints, err := eph.Protocol.Endpoints(serviceID)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to get the list of available endpoints")
		return
	}

	// TODO_UPNEXT(@adshmh): remove this once the Protocol interface
	// is updated to directly return the set of unique endpoints.
	uniqueEndpoints := make(map[relayer.EndpointAddr]struct{})
	for _, endpoints := range allEndpoints {
		for _, endpoint := range endpoints {
			uniqueEndpoints[endpoint.Addr()] = struct{}{}
		}
	}

	// TODO_IMPROVE: use a single goroutine per endpoint
	for endpointAddr := range uniqueEndpoints {
		requiredChecks := serviceQoS.GetRequiredQualityChecks(endpointAddr)
		if len(requiredChecks) == 0 {
			logger.With("endpoint", string(endpointAddr)).Warn().Msg("service QoS returned 0 required checks")
			continue
		}

		for _, serviceRequestCtx := range requiredChecks {
			// TODO_IMPROVE: Sending a request here should use some method shared with
			// the user request (i.e. HTTP request) handler.
			// This would ensure that both organic, i.e. user-generated, and quality data augmenting service requests
			// take the same execution path.
			// TODO_UPNEXT(@adshmh): remove the context input argument once the Relayer interface's
			// SendRelay function is updated.
			endpointResponse, err := eph.Relayer.SendRelay(
				context.TODO(),
				serviceID,
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
}
