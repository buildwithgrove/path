package evm

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseToGetBlockByNumber provides the functionality required from a response by a requestContext instance.
var _ response = responseToGetBlockByNumber{}

// responseUnmarshallerGetBlockByNumber deserializes the provided payload
// into a responseToGetBlockByNumber struct, adding any encountered errors
// to the returned struct.
func responseUnmarshallerGetBlockByNumber(
	logger polylog.Logger,
	jsonrpcReq jsonrpc.Request,
	jsonrpcResp jsonrpc.Response,
) (response, error) {
	// The endpoint returned an error: no need to do further processing of the response.
	if jsonrpcResp.IsError() {
		// TODO_TECHDEBT: validate the `eth_getBlockByNumber`
		// request that was sent to the endpoint.
		return responseToGetBlockByNumber{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,

			// DEV_NOTE: A valid JSONRPC error response is considered a valid response.
			valid: true,
		}, nil
	}

	resultBz, err := jsonrpcResp.GetResultAsBytes()
	if err != nil {
		validationError := qosobservations.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_UNMARSHAL
		return responseToGetBlockByNumber{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
			validationError: &validationError,
		}, err
	}

	var result map[string]interface{}
	validationError := qosobservations.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_UNSPECIFIED

	// Handle null result which means block not found
	if string(resultBz) == "null" {
		validationError = qosobservations.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_INVALID_RESULT
		return responseToGetBlockByNumber{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
			valid:           false,
			validationError: &validationError,
		}, nil
	}

	err = json.Unmarshal(resultBz, &result)
	if err != nil {
		validationError = qosobservations.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_UNMARSHAL
		return responseToGetBlockByNumber{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
			valid:           false,
			validationError: &validationError,
		}, err
	}

	// Validate the required fields in the block
	valid := true
	if _, ok := result["number"]; !ok {
		valid = false
		validationError = qosobservations.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_INVALID_RESULT
	} else if _, ok := result["hash"]; !ok {
		valid = false
		validationError = qosobservations.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_INVALID_RESULT
	} else if _, ok := result["parentHash"]; !ok {
		valid = false
		validationError = qosobservations.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_INVALID_RESULT
	}

	return responseToGetBlockByNumber{
		logger:          logger,
		jsonRPCResponse: jsonrpcResp,
		result:          result,
		valid:           valid,
		validationError: &validationError,
	}, nil
}

// responseToGetBlockByNumber captures the fields expected in a
// response to an `eth_getBlockByNumber` request.
type responseToGetBlockByNumber struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSONRPC response parsed from an endpoint's response bytes.
	jsonRPCResponse jsonrpc.Response

	// result stores the result field of a response to a `eth_getBlockByNumber` request.
	result map[string]interface{}

	// valid is set to true if the endpoint response is deemed valid.
	// A response is valid if either of the following holds:
	//	- It is a valid JSONRPC error response
	//	- It is a valid JSONRPC response with a non-null block containing required fields (number, hash, parentHash)
	valid bool

	// Why the response has failed validation.
	// Used when generating observations.
	validationError *qosobservations.EVMResponseValidationError
}

// GetObservation returns an observation using an `eth_getBlockByNumber` request's response.
// Implements the response interface.
// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_getblockbynumber
func (r responseToGetBlockByNumber) GetObservation() qosobservations.EVMEndpointObservation {
	blockNumber := ""
	blockHash := ""
	txCount := 0

	if r.result != nil {
		if num, ok := r.result["number"].(string); ok {
			blockNumber = num
		}
		if hash, ok := r.result["hash"].(string); ok {
			blockHash = hash
		}
		if txs, ok := r.result["transactions"].([]interface{}); ok {
			txCount = len(txs)
		}
	}

	return qosobservations.EVMEndpointObservation{
		ResponseObservation: &qosobservations.EVMEndpointObservation_GetBlockByNumberResponse{
			GetBlockByNumberResponse: &qosobservations.EVMGetBlockByNumberResponse{
				BlockNumber:             blockNumber,
				BlockHash:               blockHash,
				TransactionCount:        int32(txCount),
				Valid:                   r.valid,
				ResponseValidationError: r.validationError,
			},
		},
	}
}

func (r responseToGetBlockByNumber) GetResponsePayload() []byte {
	// TODO_MVP: return a JSONRPC response indicating the error if unmarshaling failed.
	bz, err := json.Marshal(r.jsonRPCResponse)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.logger.Warn().Err(err).Msg("responseToGetBlockByNumber: Marshaling JSONRPC response failed.")
	}
	return bz
}
