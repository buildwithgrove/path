package framework

import (
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// EndpointQuery represents a raw communication attempt with an endpoint.
// Instantiated by: RequestQoSContext.
// Used in EndpointQueryResultContext.
type endpointQuery struct {
	// request is the JSONRPC request that was sent.
	request *jsonrpc.Request

	// TODO_IN_THIS_PR: REMOVE this field to be consistent with the proto files.
	//
	// endpointAddr identifies the endpoint
	endpointAddr protocol.EndpointAddr

	// receivedData is the raw response data received from the endpoint (may be nil)
	receivedData []byte

	// JSONRPC response, parsed from the data received from the endpoint.
	// Only set if the data received from the endpoint could be parsed into a JSONRPC response.
	parsedResponse *jsonrpc.Response
}
