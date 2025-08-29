package qos

// TODO_REFACTOR(@adshmh): Extract patterns from this package into a shared location to enable reuse across other observation interpreters (e.g., solana, cometbft).
// This would establish a consistent interpretation pattern across all QoS services while maintaining service-specific interpreters.

import (
	"errors"

	"github.com/pokt-network/poktroll/pkg/polylog"
)

var (
	// TODO_REFACTOR(@adshmh): Consider consolidating all errors in the qos package into a single file.
	ErrEVMNoObservations              = errors.New("no observations available")
	ErrEVMNoEndpointObservationsFound = errors.New("no endpoint observations listed")
)

// EVMObservationInterpreter provides interpretation helpers for EVM QoS observations.
// It serves as a utility layer for the EVMRequestObservations protobuf type, making
// the relationships and meaning of different observation fields clear while shielding
// the rest of the codebase from proto type details.
//
// The EVMRequestObservations type contains:
// - Various metadata (e.g., ChainID)
// - A single JSON-RPC request (exactly one)
// - A list of endpoint observations (zero or more)
//
// This interpreter allows the rest of the code to draw conclusions about the observations
// without needing to understand the structure of the proto-generated types.
type EVMObservationInterpreter struct {
	Logger polylog.Logger

	// TODO_TECHDEBT(@adshmh): Missing a logger
	Observations *EVMRequestObservations
}

// GetRequestMethods extracts all JSON-RPC methods from the request observations.
// Returns (methods, true) if extraction succeeded
// Returns (nil, false) if requests are invalid or missing
func (i *EVMObservationInterpreter) GetRequestMethods() ([]string, bool) {
	if i.Observations == nil {
		return nil, false
	}

	// Check for validation failures using the shared method
	if _, reqError := i.checkRequestValidationFailures(); reqError != nil {
		return nil, false
	}

	// Get the JSON-RPC requests from the observations.
	// One request observation per JSON-RPC request.
	//   - In the case of EVM batch requests, this will return multiple request observations.
	//   - Non-EVM batch requests will return a single request observation.
	requestObservations := i.Observations.GetRequestObservations()
	if len(requestObservations) == 0 {
		return nil, false
	}

	// Extract all JSONRPC request methods for the request observations.
	var methods []string
	for _, reqObs := range requestObservations {
		if jsonrpcReq := reqObs.GetJsonrpcRequest(); jsonrpcReq != nil {
			if method := jsonrpcReq.GetMethod(); method != "" {
				methods = append(methods, method)
			}
		}
	}

	if len(methods) == 0 {
		return nil, false
	}

	// Return the methods and success flag
	return methods, true
}

// GetChainID extracts the chain ID associated with the EVM observations.
// Returns (chainID, true) if available
// Returns ("", false) if chain ID is missing or observations are nil
func (i *EVMObservationInterpreter) GetChainID() (string, bool) {
	if i.Observations == nil {
		return "", false
	}

	chainID := i.Observations.GetChainId()
	if chainID == "" {
		return "", false
	}

	return chainID, true
}

// GetServiceID extracts the service ID associated with the EVM observations.
// Returns (serviceID, true) if available
// Returns ("", false) if service ID is missing or observations are nil
func (i *EVMObservationInterpreter) GetServiceID() (string, bool) {
	if i.Observations == nil {
		return "", false
	}

	serviceID := i.Observations.GetServiceId()
	if serviceID == "" {
		return "", false
	}

	return serviceID, true
}

// GetRequestOrigin returns the Origin of the request:
// Organic, i.e. User requests.
// Synthetic, i.e. requests build by the QoS sytem to get observations on endpoints.
func (i *EVMObservationInterpreter) GetRequestOrigin() string {
	// Nil observations: log a warning and skip further processing.
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: got nil EVM QoS observations.")
		return RequestOrigin_REQUEST_ORIGIN_UNSPECIFIED.String()
	}

	return i.Observations.GetRequestOrigin().String()
}

// GetRequestStatus interprets the observations to determine request status information:
// - httpStatusCode: the suggested HTTP status code to return to the client
// - requestError: error details (nil if successful)
// - err: error if interpreter cannot determine status (e.g., nil observations)
func (i *EVMObservationInterpreter) GetRequestStatus() (httpStatusCode int, requestError *EVMRequestError, err error) {
	// Unknown status if no observations are available
	if i.Observations == nil {
		return 0, nil, ErrEVMNoObservations
	}

	// First, check for request validation failures
	if httpStatusCode, requestError := i.checkRequestValidationFailures(); requestError != nil {
		return httpStatusCode, requestError, nil
	}

	// Then, interpret endpoint response status
	return i.getEndpointResponseStatus()
}

// GetEndpointObservations extracts endpoint observations and indicates success
// Returns (nil, false) if observations are missing or validation failed
// Returns (observations, true) if observations are available
func (i *EVMObservationInterpreter) GetEndpointObservations() ([]*EVMEndpointObservation, bool) {
	if i.Observations == nil {
		return nil, false
	}

	// Check for validation failures using the shared method
	if _, reqError := i.checkRequestValidationFailures(); reqError != nil {
		return nil, false
	}

	// Get endpoint observations from request observations
	requestObservations := i.Observations.GetRequestObservations()
	if len(requestObservations) == 0 {
		return nil, false
	}

	var allEndpointObservations []*EVMEndpointObservation
	for _, reqObs := range requestObservations {
		endpointObs := reqObs.GetEndpointObservations()
		allEndpointObservations = append(allEndpointObservations, endpointObs...)
	}

	if len(allEndpointObservations) == 0 {
		return nil, false
	}

	return allEndpointObservations, true
}

// checkRequestValidationFailures examines observations for request validation failures
// Returns (httpStatusCode, requestError) where requestError is non-nil if a validation failure was found
func (i *EVMObservationInterpreter) checkRequestValidationFailures() (int, *EVMRequestError) {
	// Check for HTTP body read failure
	if failure := i.Observations.GetEvmHttpBodyReadFailure(); failure != nil {
		errType := EVMRequestValidationError_EVM_REQUEST_VALIDATION_ERROR_HTTP_BODY_READ_FAILURE
		return int(failure.GetHttpStatusCode()), &EVMRequestError{
			requestValidationError: &errType,
		}
	}

	// Check for unmarshaling failure
	if failure := i.Observations.GetEvmRequestUnmarshalingFailure(); failure != nil {
		errType := EVMRequestValidationError_EVM_REQUEST_VALIDATION_ERROR_REQUEST_UNMARSHALING_FAILURE
		return int(failure.GetHttpStatusCode()), &EVMRequestError{
			requestValidationError: &errType,
		}
	}

	// No validation failures found
	return 0, nil
}

// getEndpointResponseStatus interprets endpoint response observations to extract status information
// Returns (httpStatusCode, requestError, error) tuple
func (i *EVMObservationInterpreter) getEndpointResponseStatus() (int, *EVMRequestError, error) {
	observations, ok := i.GetEndpointObservations()
	if !ok {
		return 0, nil, ErrEVMNoEndpointObservationsFound
	}

	// No endpoint observations indicates no responses were received
	if len(observations) == 0 {
		return 0, nil, ErrEVMNoEndpointObservationsFound
	}

	// Use only the last observation (latest response)
	lastObs := observations[len(observations)-1]
	responseInterpreter, err := getEVMResponseInterpreter(lastObs)
	if err != nil {
		return 0, nil, err
	}

	// Extract the status code and error type
	statusCode, errType := responseInterpreter.extractValidityStatus(lastObs)
	if errType == nil {
		return statusCode, nil, nil
	}

	// Create appropriate EVMRequestError based on the observed error type
	reqError := &EVMRequestError{
		responseValidationError: errType,
	}

	return statusCode, reqError, nil
}
