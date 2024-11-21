package protocol

// TODO_TECHDEBT: use an interace here that returns the serialized form the request:
// Payload should return the serialized form of the request to be delivered to the backend service,
// i.e. the service to which the protocol endpoint proxies relay requests.
//
// Payload currently only supports HTTP requests to an EVM blockchain (through its Data/Method/Path fields)
// TODO_DOCUMENT: add more examples, e.g. for RESTful services, as support for more types of services
// is added.
type Payload struct {
	Data            string
	Method          string
	Path            string
	TimeoutMillisec int
}
