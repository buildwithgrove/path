package gateway

// Protocol specifies the set of behaviours and interactions needed
// by the EndpointDataAugmenter with the underlying protocol.
// It is defined separately, rather than reusing relayer.Protocol interface,
// to ensure only minimum necessary capabilities are available to the augmenter.
type Protocol interface {
	Endpoints(relayer.ServiceID) ([]relayer.Endpoint, error)
}

// EndpointDataAugmenter transforms endpoint data to enable real-time business-logic decisions.
// 
// For example, it can be used to process raw data into QoS data.
// This ensures that each service on each instance has the information
// needed to make real-time decisions to handle user requests.
// 
// An example QoS transformation workflow can be:
// 1. Consulting each service's QoS instance on the checks required to validate an endpoint.
// 2. Performing the required checks on the endpoint, in the form of a (synthetic) service request.
// 3. Reporting the results back to the service's QoS instance.
type EndpointDataAugmenter struct {
	Protocol
	*relayer.Relayer
	QoSPublisher
	Services map[relayer.ServiceID]QoSEndpointCheckGenerator
}

func (eda *EndpointDataAugmenter) Start() error {
	if eda.Protocol == nil {
		return errors.New("a Protocol instance must be proivded.")
	}

	if eda.Relayer == nil {
		return errors.New("a Relayer must be provided.")
	}

	if len(Services) == 0 {
		return errors.New("at-least one covered service must be specified")
	}

	go func() {
		// TODO_IMPROVE: support configuring a custom running interval.
		ticker := time.NewTicker(30 * time.Second)
		for {
			eda.run()
			<-ticker.C
		}
	}()

	return nil
}

func (eda *EndpointDataAugmenter) run() {
	for svcID, svcQoS := range s.services {
		go func(serviceID relayer.ServiceID, serviceQoS ServiceQoS) {
			eda.performChecks(serviceID, serviceQoS)
		}(svcID, svcQoS)
	}

	// TODO_IMPROVE: use waitgroups to wait for all goroutines to finish before returning.
}

func (eda *EndpointDataAugmenter) performChecks(serviceID relayer.ServiceID, serviceQoS QoSEndpointCheckGenerator) {
	// endpoints here is expected to be: []Endpoint (AppAddr and EndpointAddr should be properties of Endpoint interface)
	endpoints, err := eda.Protocol.AvailableEndpoints(serviceID)
	if err != nil {
		// TODO_IMPROVE: log the error
		return
	}

	// TODO_IMPROVE: use a single goroutine per endpoint
	for _, endpointAddr := range endpoints {
		endpointChecks := serviceQoS.GetRequiredQualityChecks(endpointAddr)
		if len(endpointChecks) == 0 {
			// TODO_FUTURE: Log an info-level message
			continue
		}

		singleEndpointSelector := singleEndpointSelector{EndpointAddr: endpointAddr}

		for _, serviceRequestCtx := range endpointChecks {
			// TODO_IMPROVE: Sending a request here should use some method shared with the user request handler.
			// This would ensure that both organic (i.e. user-generated) and synthetic (i.e. for QoS data) requests
			// take the same execution path.
			_, endpointAddr, endpointResponse, err := eda.Relayer.SendRelay(context.TODO(), serviceID, serviceReq.GetPayload(), singleEndpointSelector)

			// Protocol-level errors are the responsibility of the specific
			// protocol instance used in serving the request.
			// e.g. an endpoint that is temporarily/permanently unavailable
			// should not be returned by the AvailableEndpoints() method.
			if err != nil {
				// TODO_FUTURE: consider retrying failed service requests
				// as the failure may not be related to the quality of the endpoint.
				// TODO_IMPROVE: log the error
				continue
			}

			serviceRequestCtx.UpdateWithResponse(endpointAddr, endpointResponse)

			// TODO_FUTURE: consider supplying additional data to QoS.
			// e.g. data on the latency of an endpoint.
			if err := eda.QoSPublisher.Publish(serviceRequestCtx.GetObservationSet()); err != nil {
				// TODO_IMPROVE: log the error
			}
		}

	}

	// TODO_FUTURE: publish aggregated QoS reports (in addition to reports on endpoints of a specific service)
}
