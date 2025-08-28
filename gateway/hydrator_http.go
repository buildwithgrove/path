package gateway

import (
	"context"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
)

// runHTTPQualityChecks performs HTTP-based quality checks for a specific endpoint.
func (eph *EndpointHydrator) runHTTPQualityChecks(
	endpointLogger polylog.Logger,
	serviceID protocol.ServiceID,
	serviceQoS QoSService,
	endpointAddr protocol.EndpointAddr,
) {
	// Retrieve all the required QoS checks for the endpoint.
	requiredQoSChecks := serviceQoS.GetRequiredQualityChecks(endpointAddr)
	if len(requiredQoSChecks) == 0 {
		endpointLogger.Warn().Msg("No required QoS checks for endpoint and service. Skipping checks...")
		return
	}

	// Iterate over every required QoS check for the endpoint and service.
	for _, serviceRequestCtx := range requiredQoSChecks {
		eph.performSingleQoSCheck(
			endpointLogger,
			serviceID,
			serviceQoS,
			endpointAddr,
			serviceRequestCtx,
		)
	}
}

// performSingleQoSCheck performs a single QoS check by sending a synthetic request to the endpoint.
func (eph *EndpointHydrator) performSingleQoSCheck(
	endpointLogger polylog.Logger,
	serviceID protocol.ServiceID,
	serviceQoS QoSService,
	endpointAddr protocol.EndpointAddr,
	serviceRequestCtx RequestQoSContext,
) {
	// Create a new protocol request context with a pre-selected endpoint for each request.
	// IMPORTANT: A new request context MUST be created on each iteration of the loop to
	// avoid race conditions related to concurrent access issues when running concurrent QoS checks.

	// Passing a nil as the HTTP request, because we assume the Centralized Operation Mode being used by the hydrator,
	// which means there is no need for specifying a specific app.
	// TODO_FUTURE(@adshmh): support specifying the app(s) used for sending/signing synthetic relay requests by the hydrator.
	// TODO_FUTURE(@adshmh): consider publishing observations here.
	hydratorRequestCtx, _, err := eph.BuildHTTPRequestContextForEndpoint(context.TODO(), serviceID, endpointAddr, nil)
	if err != nil {
		endpointLogger.Error().Err(err).Msg("Failed to build a protocol request context for the endpoint")
		return
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
		protocolContexts:    []ProtocolRequestContext{hydratorRequestCtx},
		// metrics reporter for exporting metrics on hydrator service requests.
		metricsReporter: eph.MetricsReporter,
		// data reporter for exporting data on hydrator service requests to the data pipeline.
		dataReporter: eph.DataReporter,
	}

	err = gatewayRequestCtx.handleRelayRequest()
	if err != nil {
		// TODO_FUTURE: consider skipping the rest of the checks based on the error.
		// e.g. if the endpoint is refusing connections it may be reasonable to skip it
		// in this iteration of QoS checks.
		//
		// TODO_FUTURE: consider retrying failed service requests
		// as the failure may not be related to the quality of the endpoint.
		endpointLogger.Warn().Err(err).Msg("Failed to send a relay. Only protocol-level observations will be applied.")
	}

	// publish all observations gathered through sending the synthetic service requests.
	// e.g. protocol-level, qos-level observations.
	gatewayRequestCtx.BroadcastAllObservations()
}
