package evm

import (
	"encoding/json"
	"strconv"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseToArchival provides the functionality required from a response by a requestContext instance.
var _ response = responseToArchival{}

type ArchivalResponse struct {
	Number     string `json:"number"`
	Hash       string `json:"hash"`
	ParentHash string `json:"parentHash"`
	Timestamp  string `json:"timestamp"`
	Difficulty string `json:"difficulty"`
}

// responseUnmarshallerArchival deserializes the provided payload
// into a responseToArchival struct, adding any encountered errors
// to the returned struct.
func responseUnmarshallerArchival(
	logger polylog.Logger,
	jsonrpcReq jsonrpc.Request,
	jsonrpcResp jsonrpc.Response,
) (response, error) {
	// The endpoint returned an error: no need to do further processing of the response.
	if jsonrpcResp.IsError() {
		return responseToArchival{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
			validationError: nil, // Valid JSON-RPC error response.
		}, nil
	}

	resultBz, err := jsonrpcResp.GetResultAsBytes()
	if err != nil {
		validationError := qosobservations.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_UNMARSHAL
		return responseToArchival{
			logger:          logger,
			jsonRPCResponse: jsonrpcResp,
			validationError: &validationError,
		}, err
	}

	var validationError *qosobservations.EVMResponseValidationError

	var block ArchivalResponse
	if err := json.Unmarshal(resultBz, &block); err != nil {
		errValue := qosobservations.EVMResponseValidationError_EVM_RESPONSE_VALIDATION_ERROR_UNMARSHAL
		validationError = &errValue
	}

	// Validate the archival block data using our custom helper.
	validArchivalNodeResponse := validateArchivalBlockResult(block)

	return responseToArchival{
		logger:                    logger,
		jsonRPCResponse:           jsonrpcResp,
		validArchivalNodeResponse: validArchivalNodeResponse,
		validationError:           validationError,
	}, nil
}

// responseToArchival captures the fields expected in a
// response to an `eth_getBlockByNumber` request.
type responseToArchival struct {
	logger polylog.Logger

	// jsonRPCResponse stores the JSONRPC response parsed from an endpoint's response bytes.
	jsonRPCResponse jsonrpc.Response

	// validArchivalNodeResponse indicates whether the response is from a valid archival node.
	validArchivalNodeResponse bool

	// validationError indicates why the response failed validation, if it did.
	validationError *qosobservations.EVMResponseValidationError
}

// GetObservation returns an observation based on the archival response.
// Implements the response interface.
func (r responseToArchival) GetObservation() qosobservations.EVMEndpointObservation {
	return qosobservations.EVMEndpointObservation{
		ResponseObservation: &qosobservations.EVMEndpointObservation_ArchivalResponse{
			ArchivalResponse: &qosobservations.EVMArchivalResponse{
				HttpStatusCode:            int32(r.getHTTPStatusCode()),
				ValidArchivalNodeResponse: r.validArchivalNodeResponse,
				ResponseValidationError:   r.validationError,
			},
		},
	}
}

func (r responseToArchival) GetHTTPResponse() httpResponse {
	return httpResponse{
		responsePayload: r.getResponsePayload(),
		httpStatusCode:  r.getHTTPStatusCode(),
	}
}

func (r responseToArchival) getResponsePayload() []byte {
	bz, err := json.Marshal(r.jsonRPCResponse)
	if err != nil {
		r.logger.Warn().Err(err).Msg("responseToArchival: Marshaling JSONRPC response failed.")
	}
	return bz
}

// getHTTPStatusCode returns an HTTP status code corresponding to the underlying JSON-RPC response.
func (r responseToArchival) getHTTPStatusCode() int {
	return r.jsonRPCResponse.GetRecommendedHTTPStatusCode()
}

// validateArchivalBlockResult validates the integrity of the archival block response.
// It checks that critical fields are non-empty, validates the block number format,
// and verifies that the block hash appears to be the correct length.
// TODO_IN_THIS_PR(@commoddity): Finalize checks to perform on archival check to validate that node is archival.
func validateArchivalBlockResult(block ArchivalResponse) bool {
	// Ensure critical fields are present.
	if block.Number == "" || block.Hash == "" || block.ParentHash == "" || block.Timestamp == "" {
		return false
	}
	// Validate that block number is a valid hex value.
	if _, err := strconv.ParseUint(block.Number, 0, 64); err != nil {
		return false
	}
	// Check that the block hash has the expected length (e.g., "0x" plus 64 hex digits).
	if len(block.Hash) != 66 {
		return false
	}
	return true
}
