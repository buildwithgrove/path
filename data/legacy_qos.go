package data

// TODO_MVP(@adshmh): handle QoS observations for:
// - Solana
// - CometBFT

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservation "github.com/buildwithgrove/path/observation/qos"
)

// setLegacyFieldsFromQoSObservations populates legacy records with QoS observation data.
// Currently supports:
// - EVM observations (returns multiple records based on RequestObservations)
// - Solana observations (returns single record)
// Future support planned for CometBFT
//
// Parameters:
// - logger: logging interface
// - baseLegacyRecord: the base record to populate
// - observations: QoS observations data
// Returns: slice of populated legacy records
func setLegacyFieldsFromQoSObservations(
	logger polylog.Logger,
	baseLegacyRecord *legacyRecord,
	observations *qosobservation.Observations,
) []*legacyRecord {
	if evmObservations := observations.GetEvm(); evmObservations != nil {
		return setLegacyFieldsFromQoSEVMObservations(logger, baseLegacyRecord, evmObservations)
	}

	// Use Solana observations to update the legacy record's fields.
	if solanaObservations := observations.GetSolana(); solanaObservations != nil {
		populatedRecord := setLegacyFieldsFromQoSSolanaObservations(logger, baseLegacyRecord, solanaObservations)
		return []*legacyRecord{populatedRecord}
	}

	return []*legacyRecord{baseLegacyRecord}
}

// qosEVMErrorTypeStr defines the prefix for EVM QoS error types in legacy records
const qosEVMErrorTypeStr = "QOS_EVM"

// setLegacyFieldsFromQoSEVMObservations populates legacy records with EVM-specific QoS data.
// It captures:
// - Request payload size
// - JSONRPC method information
// - Error details (when applicable)
// Creates one legacy record per RequestObservation
//
// Parameters:
// - logger: logging interface
// - baseLegacyRecord: the base record to copy for each method
// - observations: EVM-specific QoS observations
// Returns: slice of populated legacy records
func setLegacyFieldsFromQoSEVMObservations(
	logger polylog.Logger,
	baseLegacyRecord *legacyRecord,
	observations *qosobservation.EVMRequestObservations,
) []*legacyRecord {
	// Set common fields from observations
	baseLegacyRecord.RequestDataSize = float64(observations.RequestPayloadLength)

	evmInterpreter := &qosobservation.EVMObservationInterpreter{
		Observations: observations,
	}

	// Extract all JSONRPC request methods
	jsonrpcRequestMethods, ok := evmInterpreter.GetRequestMethods()
	if !ok || len(jsonrpcRequestMethods) == 0 {
		// If no methods found, return single record with base data
		populateErrorFields(baseLegacyRecord, evmInterpreter)
		return []*legacyRecord{baseLegacyRecord}
	}

	// Create a separate legacy record for each method
	var legacyRecords []*legacyRecord
	for _, method := range jsonrpcRequestMethods {
		// Create a copy of the base record
		recordCopy := *baseLegacyRecord
		legacyRecord := &recordCopy

		// Set the method for this record
		legacyRecord.ChainMethod = method

		// Populate error fields if needed
		populateErrorFields(legacyRecord, evmInterpreter)

		legacyRecords = append(legacyRecords, legacyRecord)
	}

	return legacyRecords
}

// populateErrorFields sets error-related fields in the legacy record based on QoS observations
func populateErrorFields(legacyRecord *legacyRecord, evmInterpreter *qosobservation.EVMObservationInterpreter) {
	// ErrorType is already set at gateway or protocol level.
	// Skip updating the error fields to preserve the original error.
	if legacyRecord.ErrorType != "" {
		return
	}

	_, requestErr, err := evmInterpreter.GetRequestStatus()
	// Could not extract request error details, skip the rest of the updates.
	if err != nil || requestErr == nil {
		return
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
}

// setLegacyFieldsFromQoSSolanaObservations populates legacy record with Solana-specific QoS data.
// It captures:
// - Request payload size
// - JSONRPC method information
// - Error details (when applicable)
//
// Parameters:
// - logger: logging interface
// - legacyRecord: the record to populate
// - observations: Solana-specific QoS observations
// Returns: the populated legacy record
func setLegacyFieldsFromQoSSolanaObservations(
	logger polylog.Logger,
	legacyRecord *legacyRecord,
	observations *qosobservation.SolanaRequestObservations,
) *legacyRecord {
	logger = logger.With("method", "setLegacyFieldsFromQoSSolanaObservations")

	// In bytes: the length of the request: float64 type is for compatibility with the legacy data pipeline.
	legacyRecord.RequestDataSize = float64(observations.RequestPayloadLength)

	// Initialize the Solana observations interpreter.
	// Used to extract required fields from the observations.
	solanaInterpreter := &qosobservation.SolanaObservationInterpreter{
		Logger:       logger,
		Observations: observations,
	}

	// Extract the JSONRPC request's method.
	legacyRecord.ChainMethod = solanaInterpreter.GetRequestMethod()

	// ErrorType is already set at gateway or protocol level.
	// Skip updating the error fields to preserve the original error.
	if legacyRecord.ErrorType != "" {
		return legacyRecord
	}

	errType := solanaInterpreter.GetRequestErrorType()
	legacyRecord.ErrorType = errType
	legacyRecord.ErrorMessage = errType

	return legacyRecord
}
