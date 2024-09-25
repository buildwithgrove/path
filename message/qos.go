// package message provides the functionality required for sharing
// data between multiple PATH instances.
package message

// The topic used by QoS publishers and subscribers
// for individual service request contexts.
const qosServiceRequestTopic = "qos.service_request"

// QoSMessenger provides the functionality required by
// the gateway package for publishing Service QoS data,
// to be shared among multiple PATH instances.
var _ gateway.QoSPublisher = &QoSMessenger{}

// QoSObserver is used to communicate the details of a service
// request context to the corresponding service's QoS instance.
// It is used to notify the local QoS instance of the data shared
// by other PATH instances.
type QoSObserver interface {
	Observe(gateway.ServiceRequestContext) error
}

// TODO_FUTURE: consider using protobuf.
// QoSRequestContextUnmarshaller can build an instance of the
// ServiceRequestContext matching the specific service QoS instance.
// Each service's QoS instance provides its own unique unmarshaller.
// This is required to allow sharing service request data between PATH instances.
type QoSRequestContextUnmarshaller interface {
	// UnmarshalJSON constructs a service request context by parsing the
	// provided JSON-formatted serialization.
	UnmarshalJSON([]byte) (gateway.ServiceRequestContext, error)
}

type ServiceQoS interface {
	QoSObserver
	QoSRequestContextUnmarshaller
}

type QoSServiceRequestContextMessage struct {
	relayer.ServiceID `json:"service_id"`
	Payload           []byte
}

type MessagePlatform interface {
	Publish(topic string, data []byte) error
	Subscribe(topic string) <-chan []byte
}

type QoSMessenger struct {
	MessagePlatform
	Services map[relayer.ServiceID]ServiceQoS
}

func (qm *QoSMessenger) Publish(serviceRequestCtx gateway.ServiceRequestContext) error {
	// TODO_IMPROVE: there may be some performance advantage to directly
	// sending a ServiceRequestContext to the service's QoS instance,
	// over publishing it to the shared medium to be picked up by
	// the same PATH instance.
	bz, err := serviceRequestCtx.MarshalJSON()
	if err != nil {
		return fmt.Errorf("publish: error marshalling service request context: %w", err)
	}

	return qm.MessagePlatform.Publish(qosServiceRequestTopic, bz)
}

func (qm *QoSMessenger) Start() error {
	// TODO_INCOMPLETE: validate the struct.

	serviceRequestContextMsgCh := qm.MessagePlatform.Subscribe(qosServiceRequestTopic)

	go func() {
		qm.run(serviceRequestContextMsgCh)
	}()

	return nil
}

func (qm *QoSMessenger) run(messageCh <-chan []byte) {
	// TODO_INCOMPLETE: use multiple goroutines here.
	for bz := range messageCh {
		var qosMsg QoSServiceRequestContextMessage
		if err := json.Unmarshal(bz, &qosMsg); err != nil {
			// TODO_IMPROVE: log the error
			continue
		}

		serviceQoS, found := qm.Services[qosMsg.ServiceID]
		if !found {
			// TODO_IMPROVE: log the error
			continue
		}

		// TODO_FUTURE: find out if there is a meaningful performance difference
		// if the code is refactored to use a single Unmarshal method call.
		serviceRequestCtx, err := serviceQoS.UnmarshalJSON(qosMsg.Payload)
		if err != nil {
			// TODO_IMPROVE: log the error
			continue
		}

		serviceQoS.Observe(serviceRequestCtx)
	}
}
