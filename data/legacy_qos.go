package data

// TODO_MVP(@adshmh): handle QoS observations for:
// - Solana
// - CometBFT

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservation "github.com/buildwithgrove/path/observation/qos"
)

// setLegacyFieldsFromQoSObservations populates legacy record with QoS observation data.
// Currently supports:
// - EVM observations
// Future support planned for Solana and CometBFT
//
// Parameters:
// - logger: logging interface
// - legacyRecord: the record to populate
// - observations: QoS observations data
// Returns: the populated legacy record
func setLegacyFieldsFromQoSObservations(
	logger polylog.Logger,
	legacyRecord *legacyRecord,
	observations *qosobservation.Observations,
) *legacyRecord {
	if evmObservations := observations.GetEvm(); evmObservations != nil {
		return setLegacyFieldsFromQoSEVMObservations(logger, legacyRecord, evmObservations)
	}

	return legacyRecord
}

// qosEVMErrorTypeStr defines the prefix for EVM QoS error types in legacy records
const qosEVMErrorTypeStr = "QOS_EVM"

// setLegacyFieldsFromQoSEVMObservations populates legacy record with EVM-specific QoS data.
// It captures:
// - Request payload size
// - JSONRPC method information
// - Error details (when applicable)
//
// Parameters:
// - logger: logging interface
// - legacyRecord: the record to populate
// - observations: EVM-specific QoS observations
// Returns: the populated legacy record
func setLegacyFieldsFromQoSEVMObservations(
	logger polylog.Logger,
	legacyRecord *legacyRecord,
	observations *qosobservation.EVMRequestObservations,
) *legacyRecord {
	// In bytes: the length of the request: float64 type is for compatibility with the legacy data pipeline.
	legacyRecord.RequestDataSize = float64(observations.RequestPayloadLength)

	evmInterpreter := &qosobservation.EVMObservationInterpreter{
		Observations: observations,
	}

	// Extract the JSONRPC request's method.
	jsonrpcRequestMethod, _ := evmInterpreter.GetRequestMethod()
	legacyRecord.ChainMethod = jsonrpcRequestMethod

	// ErrorType is already set at gateway or protocol level.
	// Skip updating the error fields to preserve the original error.
	if legacyRecord.ErrorType != "" {
		return legacyRecord
	}

	_, requestErr, err := evmInterpreter.GetRequestStatus()
	// Could not extract request error details, skip the rest of the updates.
	if err != nil || requestErr == nil {
		return legacyRecord
	}

	legacyRecord.ErrorMessage = requestErr.String()

	switch {
	case requestErr.IsRequestError():
		legacyRecord.ErrorType = fmt.Sprintf("%s_REQUEST_ERROR", qosEVMErrorTypeStr)
	case requestErr.IsResponseError():
		legacyRecord.ErrorType = fmt.Sprintf("%s_ENDPOINT_ERROR", qosEVMErrorTypeStr)
	default:
		legacyRecord.ErrorType = fmt.Sprintf("%s_UNKNOWN_ERROR", qosEVMErrorTypeStr)
	}

	return legacyRecord
}
