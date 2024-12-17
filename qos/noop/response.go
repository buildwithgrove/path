package noop

import (
	"github.com/buildwithgrove/path/protocol"
)

type endpointResponse struct {
	EndpointAddr  protocol.EndpointAddr
	ResponseBytes []byte
}
