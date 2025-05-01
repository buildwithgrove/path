package framework

import (
	observations "github.com/buildwithgrove/path/observation/qos/framework"
)

// TODO_IN_THIS_PR: change all `*Kind*` enum names to `*Type*`.

	Type       SanctionType
	Reason     string
	ExpiryTime time.Time // Zero time means permanent

func (s *Sanction) buildObservation() *observations.Sanction {
	return &observations.Sanction{
		Type: translateToObservationSanctionType(s.Type),
		Reason: s.Reason,
		ExpiryTimestamp: timestampProto(s.ExpiryTime),
	}


}

func buildRequestErrorFromObservation(obs *observations.RequestError) *requestError {
	return &requestErro {
		errorKind: translateFromObservationRequestErrorKind(obs.ErrorKind()),
		errorDetails: obs.GetErrorDetails(),
		jsonrpcErrorResponse: buildJSONRPCResponseFromObservation(obs.GetJsonRpcResponse()),
	}
}


