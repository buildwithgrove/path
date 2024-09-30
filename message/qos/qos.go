// package qos provides the functionality required for
// messaging (seriliaizing, sharing, etc...) QoS data between
// multiple PATH instances.
package qos

// The topic used by QoS publishers and subscribers
// for individual service request contexts.
const observationSetTopic = "qos.observation_set"

// Messenger provides the functionality required by
// the gateway package for publishing QoS data,
// to be shared among multiple PATH instances.
var _ gateway.QoSPublisher = &Messenger{}

type ServiceQoS interface {
	message.Unmarshaller
}

// ObservationSetMessage is the expected format of QoS messages shared
// between multiple PATH instances, using the provided MessagePlatform
type ObservationSetMessage struct {
	relayer.ServiceID `json:"service_id"`
	Payload           []byte `json:"payload"`
}

// TODO_UPNEXT(@adshmh): implement the MessagePlatform interface in a separate package, using NATS or REDIS.
// MessagePlatform is used to:
// A) Publish QoS observation sets for sharing
// with other PATH instances, and
// B) Receive, through subscription to a topic, QoS observation
// sets shared by other PATH instances
type MessagePlatform interface {
	Publish(topic string, data []byte) error
	Subscribe(topic string) <-chan []byte
}

type QoSMessenger struct {
	MessagePlatform
	Services map[relayer.ServiceID]ServiceQoS
}

func (qm *QoSMessenger) Publish(observationSet message.ObservationSet) error {
	// TODO_IMPROVE: there may be some performance advantage to directly
	// sending a ServiceRequestContext to the service's QoS instance,
	// over publishing it to the shared medium to be picked up by
	// the same PATH instance.
	bz, err := observationSet.MarshalJSON()
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
		var qosMsg QoSObservationSetMessage
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
		observationSet, err := serviceQoS.UnmarshalJSONObservationSet(qosMsg.Payload)
		if err != nil {
			// TODO_IMPROVE: log the error
			continue
		}

		if err := observationSet.NotifyStakeHolders(); err != nil {
			// TODO_IMPROVE: log the error
		}
	}
}
