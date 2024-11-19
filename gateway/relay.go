package gateway

import (
	"fmt"

	"github.com/buildwithgrove/path/protocol"
)

// SendRelay is sending a relay from the perspective of a gateway.
// It is responsible for calling Protocol.SendRelay to a specific endpoint for a specific application.
// It does so by calling the correct sequence of functions on the Relayer and the EndpointSelector.
//
// SendRelay is written as a template method to allow the customization of key steps,
// e.g. endpoint selection and protocol-specific details of sending a relay.
// See the following link for more details:
// https://en.wikipedia.org/wiki/Template_method_pattern
func SendRelay(
	protocolRequestCtx ProtocolRequestContext,
	payload protocol.Payload,
	endpointSelector protocol.EndpointSelector,
) (protocol.Response, error) {
	if err := protocolRequestCtx.SelectEndpoint(endpointSelector); err != nil {
		return protocol.Response{}, fmt.Errorf("SendRelay: error selecting an endpoint: %w", err)
	}

	// TODO_FUTURE: add a protocol publisher to enable sending feedback on the endpoint that served the request.
	// e.g. on Morse protocol, an endpoint that rejects a request due to being maxed out for the app+service
	// combination, should be dropped until the start of the next session.
	return protocolRequestCtx.HandleServiceRequest(payload)
}
