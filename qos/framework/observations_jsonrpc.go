package framework

func buildJSONRPCRequestObservation(jsonrpcReq jsonrpc.Request) *qosobservations.JsonRpcRequest {
	return &qosobservations.JsonRpcRequest {
		Id: jsonrpcReq.ID.String(),
		Method: jsonrpcReq.Method,
	}
}

// TODO_IN_THIS_PR: implement.
func buildJSONRPCResponseObservation(jsonrpcResp jsonrpc.Response) *qosobservations.JsonRpcResponse {
	return nil
}

func extractJSONRPCRequestFromObservation(
	observation *qosobservations.RequestObservation,
) *jsonrpc.Request {
	// Nil 
	if observation == nil {
		return nil 
	}

	jsonrpcRequestObs := observation.GetJsonRpcRequest()
	if jsonrpcRequestObs == nil {
		return nil
	}

	// The only field required in applying the observations is the request's method.
	return &jsonrpc.Request{
		Method: jsonrpcRequestObs.GetMethod(),
	}
}

func extractJSONRPCResponseFRomObservation(
	observation *qosobservations.
) *jsonrpc.Response {

}
