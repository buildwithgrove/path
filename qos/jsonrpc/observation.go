package jsonrpc

import (
	"github.com/buildwithgrove/path/observation/qos"
)

// GetObservation returns a qos.JsonRpcRequest struct that can be used by QoS services
// to populate observation fields.
func (r Request) GetObservation() *qos.JsonRpcRequest {
	return &qos.JsonRpcRequest{
		Id:     r.ID.String(),
		Method: string(r.Method),
	}
}
