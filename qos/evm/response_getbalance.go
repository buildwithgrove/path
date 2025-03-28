package evm

import (
	"encoding/json"
	"strings"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseToGetBalance provides the functionality required from a response by a requestContext instance.
var _ response = responseToGetBalance{}

const trieNodeErrorString = "missing trie node"

// responseUnmarshallerGetBalance deserializes the provided payload
// into a responseToGetBalance struct, adding any encountered errors
// to the returned struct.
func responseUnmarshallerGetBalance(
	logger polylog.Logger,
	jsonrpcReq jsonrpc.Request,
	jsonrpcResp jsonrpc.Response,
) (response, error) {
	// The endpoint returned an error: no need to do further processing of the response.
	if jsonrpcResp.IsError() {
		if strings.Contains(jsonrpcResp.Error.Message, trieNodeErrorString) {
			return responseToGetBalance{
				logger:          logger,
				jsonRPCResponse: jsonrpcResp,
				trieNodeError:   true,
			}, nil
		}

		return responseToGetBalance{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
			validationError: nil, // Valid JSON-RPC error response.
		}, nil
	}

	resultBz, err := jsonrpcResp.GetResultAsBytes()
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
	if err := json.Unmarshal(resultBz, &balanceResponse); err != nil {
		errValue := qosobservations.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_UNMARSHAL
		validationError = &errValue
	}

	return responseToGetBalance{
		logger:          logger,
		jsonRPCResponse: jsonrpcResp,
		balance:         balanceResponse,
		validationError: validationError,
	}, nil
}

// responseToGetBalance captures the fields expected in a
// response to an `eth_getBalance` request.
type responseToGetBalance struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSONRPC response parsed from an endpoint's response bytes.
	jsonRPCResponse jsonrpc.Response

	// trieNodeError tracks whether the endpoint has returned a trie node error.
	trieNodeError bool

	// balance string
	balance string

	// validationError indicates why the response failed validation, if it did.
	validationError *qosobservations.EVMResponseValidationError
}

// GetObservation returns an observation based on the archival response.
// Implements the response interface.
func (r responseToGetBalance) GetObservation() qosobservations.EVMEndpointObservation {
	return qosobservations.EVMEndpointObservation{
		ResponseObservation: &qosobservations.EVMEndpointObservation_ArchivalResponse{
			ArchivalResponse: &qosobservations.EVMArchivalResponse{
				HttpStatusCode:           int32(r.getHTTPStatusCode()),
				HasReturnedTrieNodeError: r.trieNodeError,
				Balance:                  r.balance,
				ResponseValidationError:  r.validationError,
			},
		},
	}
}

func (r responseToGetBalance) GetHTTPResponse() httpResponse {
	return httpResponse{
		responsePayload: r.getResponsePayload(),
		httpStatusCode:  r.getHTTPStatusCode(),
	}
}

func (r responseToGetBalance) getResponsePayload() []byte {
	bz, err := json.Marshal(r.jsonRPCResponse)
	if err != nil {
		r.logger.Warn().Err(err).Msg("responseToGetBalance: Marshaling JSONRPC response failed.")
	}
	return bz
}

// getHTTPStatusCode returns an HTTP status code corresponding to the underlying JSON-RPC response.
func (r responseToGetBalance) getHTTPStatusCode() int {
	return r.jsonRPCResponse.GetRecommendedHTTPStatusCode()
}
