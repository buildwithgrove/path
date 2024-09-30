// package message defines the terminology required for sharing
// data between multiple PATH instances.
package message

// ObservationSet defines the functionality required for sharing
// data between PATH instances.
// e.g. QoS instances can share data by supplying an implementation
// of this interface.
type ObservationSet interface {
	// MarshalJSON returns the serialized form
	// of the set of observations, in JSON format.
	// This is required for sharing QoS data between
	// multiple PATH instances.
	MarshalJSON() ([]byte, error)

	// Broadcast is used to communicate the observations contained
	// in the set to the interested entities.
	// e.g. The observation set returned by a gateway.ServiceRequestContext can
	// guide the target service's QoS instance to update the quality data of
	// one or more endpoints.
	Broadcast() error
}

// TODO_UPNEXT(@adshmh): consider using protobuf.
// Unmarshaller builds an instance of the ObservationSet,
// matching a specific implementation, e.g. one provided by a service QoS.
// Each service's QoS instance provides its own unique unmarshaller.
// This allows sharing data between PATH instances.
type Unmarshaller interface {
	// UnmarshalJSONObservationSet constructs an observation set by parsing
	// its JSON-formatted serialized form.
	UnmarshalJSONObservationSet([]byte) (ObservationSet, error)
}
