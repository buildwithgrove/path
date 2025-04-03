package evm

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseToGetBalance provides the functionality required from a response by a requestContext instance.
var _ response = responseToGetBalance{}

// responseUnmarshallerGetBalance deserializes the provided JSONRPC payload into
// a responseToGetBalance struct, adding any encountered errors to the returned struct.
//
// The results from this method are used to update the endpoint's archival state
// only if they are for the currently selected archival block number.
//
// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_getbalance
func responseUnmarshallerGetBalance(
	logger polylog.Logger,
	jsonrpcReq jsonrpc.Request,
	jsonrpcResp jsonrpc.Response,
) (response, error) {
	// The endpoint returned an error: no need to do further processing of the response.
	if jsonrpcResp.IsError() {
		return responseToGetBalance{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
			validationError: nil, // Intentionally set to nil to indicate a valid JSONRPC error response.
		}, nil
	}

	responseBz, err := jsonrpcResp.GetResultAsBytes()
	if err != nil {
		validationError := qosobservations.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_UNMARSHAL
		return responseToGetBalance{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
			validationError: &validationError,
		}, err
	}

	var validationError *qosobservations.EVMResponseValidationError
	var balanceResponse string
	if err := json.Unmarshal(responseBz, &balanceResponse); err != nil {
		errValue := qosobservations.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_UNMARSHAL
		validationError = &errValue
	}

	// Valid eth_getBalance response: no error and a valid balance.
	return responseToGetBalance{
		logger:          logger,
		jsonRPCResponse: jsonrpcResp,
		balance:         balanceResponse,
		// Extract the block number from the JSON-RPC request and attach it to the observation.
		blockNumber:     getBlockNumberFromRequest(jsonrpcReq),
		validationError: validationError,
	}, nil
}

// responseToGetBalance captures the fields expected in a response to an `eth_getBalance` request.
type responseToGetBalance struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSONRPC response parsed from an endpoint's response bytes.
	jsonRPCResponse jsonrpc.Response

	// balance string
	balance string

	// blockNumber string
	blockNumber string

	// validationError indicates why the response failed validation, if it did.
	validationError *qosobservations.EVMResponseValidationError
}

// GetObservation returns an observation based on the archival response.
// Implements the response interface.
func (r responseToGetBalance) GetObservation() qosobservations.EVMEndpointObservation {
	return qosobservations.EVMEndpointObservation{
		ResponseObservation: &qosobservations.EVMEndpointObservation_GetBalanceResponse{
			GetBalanceResponse: &qosobservations.EVMGetBalanceResponse{
				HttpStatusCode:          int32(r.getHTTPStatusCode()),
				Balance:                 r.balance,
				BlockNumber:             r.blockNumber,
				ResponseValidationError: r.validationError,
			},
		},
	}
}

// GetHTTPResponse returns the HTTP response corresponding to the JSON-RPC response.
// Implements the response interface.
func (r responseToGetBalance) GetHTTPResponse() httpResponse {
	return httpResponse{
		responsePayload: r.getResponsePayload(),
		httpStatusCode:  r.getHTTPStatusCode(),
	}
}

// getResponsePayload returns the JSON-RPC response payload as a byte slice.
func (r responseToGetBalance) getResponsePayload() []byte {
	responseBz, err := json.Marshal(r.jsonRPCResponse)
	if err != nil {
		r.logger.Warn().Err(err).Msg("responseToGetBalance: Marshaling JSONRPC response failed.")
	}
	return responseBz
}

// getHTTPStatusCode returns an HTTP status code corresponding to the underlying JSON-RPC response.
func (r responseToGetBalance) getHTTPStatusCode() int {
	return r.jsonRPCResponse.GetRecommendedHTTPStatusCode()
}

// getBlockNumberFromRequest extracts the block number (hex string) from the JSONRPC request.
// For 'eth_getBalance', the block number is the second parameter in the params array.
//
// For example, for the JSON-RPC request:
// `{"jsonrpc": "2.0", "method": "eth_getBalance", "params": ["0x407d73d8a49eeb85d32cf465507dd71d507100c1", "0x59E8A"]}`
// the block number is "0x59E8A"
//
// Returns an empty string if the block number cannot be extracted.
func getBlockNumberFromRequest(req jsonrpc.Request) string {
	if req.Params.IsEmpty() {
		return ""
	}

	paramsBz, err := json.Marshal(req.Params)
	if err != nil {
		return ""
	}

	// eth_getBalance params are always an array of two strings
	var paramsArray [2]string
	if err := json.Unmarshal(paramsBz, &paramsArray); err != nil {
		return ""
	}

	// Extract the block parameter (second item in array)
	blockParam := paramsArray[1]
	if blockParam == "" {
		return ""
	}

	return blockParam
}
