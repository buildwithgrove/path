package framework

import (
	observations "github.com/buildwithgrove/path/observation/qos/framework"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

func buildJSONRPCRequestObservation(jsonrpcReq *jsonrpc.Request) *observations.JsonRpcRequest {
	if jsonrpcReq == nil {
		return nil
	}

	return &observations.JsonRpcRequest {
		Id: jsonrpcReq.ID.String(),
		Method: jsonrpcReq.Method,
	}
}

// TODO_IN_THIS_PR: implement.
func buildJSONRPCResponseObservation(jsonrpcResp jsonrpc.Response) *observations.JsonRpcResponse {
	return nil
}

func extractJSONRPCRequestFromObservation(
	jsonrpcRequestObs *observations.JsonRpcRequest,
) *jsonrpc.Request {
	if jsonrpcRequestObs == nil {
		return nil
	}

	// The only field required in applying the observations is the request's method.
	return &jsonrpc.Request{
		Method: jsonrpcRequestObs.GetMethod(),
	}
}

func extractJSONRPCResponseFromObservation(
	observation *observations.JsonRpcResponse,
) *jsonrpc.Response {
	if observation == nil {
		return nil
	}

	jsonrpcResp := &jsonrpc.Response{
		ID: jsonrpc.IDFromStr(observation.GetId()),
		// TODO_MVP(@adshmh): consider capturing the result.
	}

	if jsonrpcErr := observation.GetErr(); jsonrpcErr != nil {
		jsonrpcResp.Error = &jsonrpc.ResponseError{
			Code: jsonrpcErr.GetCode(),
			Message: jsonrpcErr.GetMessage(),
		}
	}

	return jsonrpcResp
}
