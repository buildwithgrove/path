package protocol

// TODO_TECHDEBT(@adshmh): use an interface here that returns the serialized form of the request, with the following requirements:
// 1. Payload should return the serialized form of the request to be delivered to the backend service,
// i.e. the onchain service to which the protocol endpoint proxies relay requests.
// 2. Use an enum to represent the underlying spec/standard, e.g. REST/JSONRPC/gRPC/etc.
//
// Payload currently only supports HTTP(s) requests to an EVM blockchain (through its Data/Method/Path fields)
// TODO_DOCUMENT(@adshmh): add more examples, e.g. for RESTful services, as support for more types of services
// is added.
type Payload struct {
	Data            string
	Method          string
	Path            string
	Headers         map[string]string
	TimeoutMillisec int
}
