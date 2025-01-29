package jsonrpc

import (
	"github.com/buildwithgrove/path/observation/qos"
)

// GetObservation builds and returns an `observation/qos` package's JsonRpcRequest struct that can be used by
// any QoS service to fill the corresponding observation field.
func (r Request) GetObservation() *qos.JsonRpcRequest {
	return &qos.JsonRpcRequest{
		Id:     r.ID.String(),
		Method: string(r.Method),
	}
}
