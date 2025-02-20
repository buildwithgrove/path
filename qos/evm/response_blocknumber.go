package evm

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseToBlockNumber provides the functionality required from a response by a requestContext instance.
var _ response = responseToBlockNumber{}

// responseUnmarshallerBlockNumber deserializes the provided payload
// into a responseToBlockNumber struct, adding any encountered errors
// to the returned struct.
func responseUnmarshallerBlockNumber(
	logger polylog.Logger,
	jsonrpcReq jsonrpc.Request,
	jsonrpcResp jsonrpc.Response,
) (response, error) {
	// The endpoint returned an error: no need to do further processing of the response.
	if jsonrpcResp.IsError() {
		// TODO_TECHDEBT(@adshmh): validate the `eth_blockNumber`
		// request that was sent to the endpoint.
		return responseToBlockNumber{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,

			// DEV_NOTE: A valid JSONRPC error response is considered a valid response.
			valid: true,
		}, nil
	}

	resultBz, err := jsonrpcResp.GetResultAsBytes()
	if err != nil {
		validationError := qosobservations.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_UNMARSHAL
		return responseToBlockNumber{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
			validationError: &validationError,
		}, err
	}

	var result string
	validationError := qosobservations.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_UNSPECIFIED

	err = json.Unmarshal(resultBz, &result)
	if err != nil {
		validationError = qosobservations.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_UNMARSHAL
	}

	return responseToBlockNumber{
		logger:          logger,
		jsonRPCResponse: jsonrpcResp,
		result:          result,

		// TODO_MVP(@adshmh): use the contents of the result field to determine the validity of the response.
		// e.g. a response that fails parsing as a number is not valid.
		valid: err == nil,

		validationError: &validationError,
	}, err
}

// responseToBlockNumber captures the fields expected in a
// response to an `eth_blockNumber` request.
type responseToBlockNumber struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSONRPC response parsed from an endpoint's response bytes.
	jsonRPCResponse jsonrpc.Response

	// result stores the result field of a response to a `eth_blockNumber` request.
	result string

	// valid is set to true if the endpoint response is deemed valid.
	// As of PR #152, a response is valid if either of the following holds:
	//	- It is a valid JSONRPC error response
	//	- It is a valid JSONRPC response with any string value in `result` field.
	valid bool

	// Why the response has failed validation.
	// Used when generating observations.
	validationError *qosobservations.EVMResponseValidationError
}

// GetObservation returns an observation using an `eth_blockNumber` request's response.
// Implements the response interface.
// Reference: https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_blocknumber
func (r responseToBlockNumber) GetObservation() qosobservations.EVMEndpointObservation {
	return qosobservations.EVMEndpointObservation{
		ResponseObservation: &qosobservations.EVMEndpointObservation_BlockNumberResponse{
			BlockNumberResponse: &qosobservations.EVMBlockNumberResponse{
				BlockNumberResponse:     r.result,
				Valid:                   r.valid,
				ResponseValidationError: r.validationError,
			},
		},
	}
}

func (r responseToBlockNumber) GetResponsePayload() []byte {
	// TODO_MVP(@adshmh): return a JSONRPC response indicating the error if unmarshaling failed.
	bz, err := json.Marshal(r.jsonRPCResponse)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.logger.Warn().Err(err).Msg("responseToBlockNumber: Marshaling JSONRPC response failed.")
	}
	return bz
}
