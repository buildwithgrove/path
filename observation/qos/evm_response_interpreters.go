package qos

import (
	"errors"
)

// errInvalidResponseType is returned when an observation doesn't match any registered type
var errInvalidResponseType = errors.New("endpoint response observation does not match any registered response type")

// evmResponseInterpreter defines an interpreter interface for EVM response observations.
// This interface decouples the rest of the codebase from proto-generated types by
// providing a consistent way to extract status and error information from various
// response type observations.
type evmResponseInterpreter interface {
	// extractValidityStatus interprets an observation and extracts standardized status information.
	// This method serves as a translation layer between proto-generated types and the rest of the system.
	// It's only used by other methods/functions within this package.
	//
	// Parameters:
	//   - obs: The typed observation from an EVM endpoint response
	//
	// Returns:
	//   - statusCode: The HTTP status code to return to the client
	//   - errorType: The specific error type from the proto definition (nil if no error)
	extractValidityStatus(obs *EVMEndpointObservation) (statusCode int, errorType *EVMResponseValidationError)
}

// responseInterpreters maps response type identifiers to their respective interpreter implementations.
// Each implementation translates a specific proto-generated response type into standardized
// status codes and error types, decoupling the rest of the codebase from these details.
//
// DEV_NOTE: To add a new response type, you MUST:
// 1. Add a new interpreter struct that implements the evmResponseInterpreter interface
// 2. Add a new entry to this responseInterpreters map
// 3. Add a new case in the getEVMResponseInterpreter function to recognize the type
// Example: To support a new eth_getBalance response, implement all three steps above
var responseInterpreters = map[string]evmResponseInterpreter{
	"chain_id":     &chainIDEVMResponseInterpreter{},
	"block_number": &blockNumberEVMResponseInterpreter{},
	"get_balance":  &getBalanceEVMResponseInterpreter{},
	"unrecognized": &unrecognizedEVMResponseInterpreter{},
	"empty":        &emptyEVMResponseInterpreter{},
	"no_response":  &noEVMResponseInterpreter{},
}

// getEVMResponseInterpreter returns the appropriate interpreter for a given observation type.
// This function selects the correct interpreter implementation based on the observation's
// proto-generated type, serving as part of the translation layer that shields the rest
// of the codebase from proto type details.
//
// Parameters:
//   - obs: The EVM endpoint observation to be interpreted
//
// Returns:
//   - An evmResponseInterpreter implementation specific to the observation type
//   - An error if the observation does not match any registered endpoint response type
func getEVMResponseInterpreter(obs *EVMEndpointObservation) (evmResponseInterpreter, error) {
	switch {

	// eth_chainId
	case obs.GetChainIdResponse() != nil:
		return responseInterpreters["chain_id"], nil

	// eth_blockNumber
	case obs.GetBlockNumberResponse() != nil:
		return responseInterpreters["block_number"], nil

	// eth_getBalance (used for archival checks)
	case obs.GetGetBalanceResponse() != nil:
		return responseInterpreters["get_balance"], nil

	// unrecognized response
	case obs.GetUnrecognizedResponse() != nil:
		return responseInterpreters["unrecognized"], nil

	// empty response
	case obs.GetEmptyResponse() != nil:
		return responseInterpreters["empty"], nil

	// no response
	case obs.GetNoResponse() != nil:
		return responseInterpreters["no_response"], nil

	default:
		return nil, errInvalidResponseType
	}
}

// chainIDEVMResponseInterpreter interprets eth_chainId response observations.
// It implements the evmResponseInterpreter interface to translate proto-generated
// chain ID response types into standardized status codes and error types.
type chainIDEVMResponseInterpreter struct{}

// extractValidityStatus extracts status information from chain ID response observations.
// It interprets the chain ID-specific proto type and translates it into standardized
// HTTP status codes and error types for the rest of the system.
func (i *chainIDEVMResponseInterpreter) extractValidityStatus(obs *EVMEndpointObservation) (int, *EVMResponseValidationError) {
	response := obs.GetChainIdResponse()
	validationErr := response.GetResponseValidationError()

	if validationErr != 0 {
		errType := EVMResponseValidationError(validationErr)
		return int(response.GetHttpStatusCode()), &errType
	}

	return int(response.GetHttpStatusCode()), nil
}

// blockNumberEVMResponseInterpreter interprets eth_blockNumber response observations.
// It implements the evmResponseInterpreter interface to translate proto-generated
// block number response types into standardized status codes and error types.
type blockNumberEVMResponseInterpreter struct{}

// extractValidityStatus extracts status information from block number response observations.
// It interprets the block number-specific proto type and translates it into standardized
// HTTP status codes and error types for the rest of the system.
func (i *blockNumberEVMResponseInterpreter) extractValidityStatus(obs *EVMEndpointObservation) (int, *EVMResponseValidationError) {
	response := obs.GetBlockNumberResponse()
	validationErr := response.GetResponseValidationError()

	if validationErr != 0 {
		errType := EVMResponseValidationError(validationErr)
		return int(response.GetHttpStatusCode()), &errType
	}

	return int(response.GetHttpStatusCode()), nil
}

// getBalanceEVMResponseInterpreter interprets eth_getBalance response observations.
// It implements the evmResponseInterpreter interface to translate proto-generated
// getBalance response types into standardized status codes and error types.
type getBalanceEVMResponseInterpreter struct{}

// extractValidityStatus extracts status information from getBalance response observations.
// It interprets the getBalance response-specific proto type and translates it into
// standardized HTTP status codes and error types for the rest of the system.
func (i *getBalanceEVMResponseInterpreter) extractValidityStatus(obs *EVMEndpointObservation) (int, *EVMResponseValidationError) {
	response := obs.GetGetBalanceResponse()
	validationErr := response.GetResponseValidationError()

	if validationErr != 0 {
		errType := EVMResponseValidationError(validationErr)
		return int(response.GetHttpStatusCode()), &errType
	}

	return int(response.GetHttpStatusCode()), nil
}

// unrecognizedEVMResponseInterpreter interprets unrecognized response observations.
// It implements the evmResponseInterpreter interface to translate proto-generated
// unrecognized response types into standardized status codes and error types.
type unrecognizedEVMResponseInterpreter struct{}

// extractValidityStatus extracts status information from unrecognized response observations.
// It interprets the unrecognized response-specific proto type and translates it into
// standardized HTTP status codes and error types for the rest of the system.
func (i *unrecognizedEVMResponseInterpreter) extractValidityStatus(obs *EVMEndpointObservation) (int, *EVMResponseValidationError) {
	response := obs.GetUnrecognizedResponse()
	validationErr := response.GetResponseValidationError()

	if validationErr != 0 {
		errType := EVMResponseValidationError(validationErr)
		return int(response.GetHttpStatusCode()), &errType
	}

	return int(response.GetHttpStatusCode()), nil
}

// emptyEVMResponseInterpreter interprets empty response observations.
// It implements the evmResponseInterpreter interface to translate proto-generated
// empty response types into standardized status codes and error types.
type emptyEVMResponseInterpreter struct{}

// extractValidityStatus extracts status information from empty response observations.
// It interprets the empty response-specific proto type and provides a standardized
// error type that indicates an empty response was received.
func (i *emptyEVMResponseInterpreter) extractValidityStatus(obs *EVMEndpointObservation) (int, *EVMResponseValidationError) {
	response := obs.GetEmptyResponse()
	// Empty responses are always errors
	errType := EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_EMPTY
	return int(response.GetHttpStatusCode()), &errType
}

// noEVMResponseInterpreter interprets no-response observations.
// It implements the evmResponseInterpreter interface to translate proto-generated
// no-response types into standardized status codes and error types.
type noEVMResponseInterpreter struct{}

// extractValidityStatus extracts status information from no-response observations.
// It interprets the no-response-specific proto type and provides a standardized
// error type that indicates no response was received.
func (i *noEVMResponseInterpreter) extractValidityStatus(obs *EVMEndpointObservation) (int, *EVMResponseValidationError) {
	response := obs.GetNoResponse()
	// No response is always an error
	errType := EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_NO_RESPONSE
	return int(response.GetHttpStatusCode()), &errType
}
