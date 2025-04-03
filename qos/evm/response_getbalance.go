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

	// validationError indicates why the response failed validation, if it did.
	validationError *qosobservations.EVMResponseValidationError
}

// GetObservation returns an observation based on the archival response.
// Implements the response interface.
func (r responseToGetBalance) GetObservation() qosobservations.EVMEndpointObservation {
	return qosobservations.EVMEndpointObservation{
		ResponseObservation: &qosobservations.EVMEndpointObservation_ArchivalResponse{
			ArchivalResponse: &qosobservations.EVMArchivalResponse{
				HttpStatusCode:          int32(r.getHTTPStatusCode()),
				Balance:                 r.balance,
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
