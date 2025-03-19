package qos

import (
	"errors"
)

var errInvalidResponseType = errors.New("endpont response observation does not match any registered response type")

// EVMResponseHandler defines the interface for handling different response types
type EVMResponseHandler interface {
	// ExtractValidityStatus extracts the validation status from an observation
	// Returns:
	// - statusCode: the HTTP status code to return
	// - errorType: the specific error type from the proto definition (nil if no error)
	ExtractValidityStatus(obs *EVMEndpointObservation) (statusCode int, errorType *EVMResponseValidationError)
}

// DEV_NOTE: To add a new response type, you MUST:
// 1. Add a new handler struct that implements the EVMResponseHandler interface
// 2. Add a new entry to the responseHandlers map below
// 3. Add a new case in the getEVMResponseHandler function to recognize the type
// Example: To support a new eth_getBalance response, implement all three steps above
var responseHandlers = map[string]EVMResponseHandler{
	"chain_id":     &chainIDEVMResponseHandler{},
	"block_number": &blockNumberEVMResponseHandler{},
	"unrecognized": &unrecognizedEVMResponseHandler{},
	"empty":        &emptyEVMResponseHandler{},
	"no_response":  &noEVMResponseHandler{},
}

// getEVMResponseHandler returns the appropriate handler for the given observation
// Returns an error if the observation does not match any registered endpoint response type.
func getEVMResponseHandler(obs *EVMEndpointObservation) (EVMResponseHandler, error) {
	switch {
	case obs.GetChainIdResponse() != nil:
		return responseHandlers["chain_id"], nil
	case obs.GetBlockNumberResponse() != nil:
		return responseHandlers["block_number"], nil
	case obs.GetUnrecognizedResponse() != nil:
		return responseHandlers["unrecognized"], nil
	case obs.GetEmptyResponse() != nil:
		return responseHandlers["empty"], nil
	case obs.GetNoResponse() != nil:
		return responseHandlers["no_response"], nil
	default:
		return nil, errInvalidResponseType
	}
}

// chainIDEVMResponseHandler handles eth_chainId responses
type chainIDEVMResponseHandler struct{}

func (h *chainIDEVMResponseHandler) ExtractValidityStatus(obs *EVMEndpointObservation) (int, *EVMResponseValidationError) {
	response := obs.GetChainIdResponse()
	validationErr := response.GetResponseValidationError()

	if validationErr != 0 {
		errType := EVMResponseValidationError(validationErr)
		return int(response.GetHttpStatusCode()), &errType
	}

	return int(response.GetHttpStatusCode()), nil
}

// blockNumberEVMResponseHandler handles eth_blockNumber responses
type blockNumberEVMResponseHandler struct{}

func (h *blockNumberEVMResponseHandler) ExtractValidityStatus(obs *EVMEndpointObservation) (int, *EVMResponseValidationError) {
	response := obs.GetBlockNumberResponse()
	validationErr := response.GetResponseValidationError()

	if validationErr != 0 {
		errType := EVMResponseValidationError(validationErr)
		return int(response.GetHttpStatusCode()), &errType
	}

	return int(response.GetHttpStatusCode()), nil
}

// unrecognizedEVMResponseHandler handles unrecognized responses
type unrecognizedEVMResponseHandler struct{}

func (h *unrecognizedEVMResponseHandler) ExtractValidityStatus(obs *EVMEndpointObservation) (int, *EVMResponseValidationError) {
	response := obs.GetUnrecognizedResponse()
	validationErr := response.GetResponseValidationError()

	if validationErr != 0 {
		errType := EVMResponseValidationError(validationErr)
		return int(response.GetHttpStatusCode()), &errType
	}

	return int(response.GetHttpStatusCode()), nil
}

// emptyEVMResponseHandler handles empty responses
type emptyEVMResponseHandler struct{}

func (h *emptyEVMResponseHandler) ExtractValidityStatus(obs *EVMEndpointObservation) (int, *EVMResponseValidationError) {
	response := obs.GetEmptyResponse()
	// Empty responses are always errors
	errType := EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_EMPTY
	return int(response.GetHttpStatusCode()), &errType
}

// noEVMResponseHandler handles no response scenarios
type noEVMResponseHandler struct{}

func (h *noEVMResponseHandler) ExtractValidityStatus(obs *EVMEndpointObservation) (int, *EVMResponseValidationError) {
	response := obs.GetNoResponse()
	// No response is always an error
	errType := EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_NO_RESPONSE
	return int(response.GetHttpStatusCode()), &errType
}
