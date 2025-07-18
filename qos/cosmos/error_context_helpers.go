package cosmos

import (
	"net/http"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/pokt-network/poktroll/pkg/polylog"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
)

// createValidationErrorContext creates a standardized error context for validation failures
func createValidationErrorContext(
	logger polylog.Logger,
	chainID string,
	serviceID protocol.ServiceID,
	rpcType qosobservations.RPCType,
	errorKind qosobservations.RequestErrorKind,
	errorDetails string,
	httpStatusCode int,
) gateway.RequestQoSContext {
	observations := &qosobservations.Observations_Cosmos{
		Cosmos: &qosobservations.CosmosSDKRequestObservations{
			ChainId:       chainID,
			ServiceId:     string(serviceID),
			RequestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
			RpcType:       rpcType,
			RequestError: &qosobservations.RequestError{
				ErrorKind:      errorKind,
				ErrorDetails:   errorDetails,
				HttpStatusCode: int32(httpStatusCode),
			},
		},
	}

	return &errorContext{
		logger:                 logger,
		responseHTTPStatusCode: httpStatusCode,
		cosmosSDKObservations:  observations,
	}
}

// createInvalidMethodError creates error context for invalid HTTP methods
func createInvalidMethodError(
	method string,
	logger polylog.Logger,
	chainID string,
	serviceID protocol.ServiceID,
	validatorType string, // "JSONRPC" or "REST"
) gateway.RequestQoSContext {
	var rpcType qosobservations.RPCType
	var errorDetails string

	if validatorType == "JSONRPC" {
		rpcType = qosobservations.RPCType_RPC_TYPE_UNSPECIFIED
		errorDetails = "JSONRPC requests must use POST method, got: " + method
	} else {
		rpcType = qosobservations.RPCType_RPC_TYPE_REST
		errorDetails = "Invalid HTTP method for REST API: " + method
	}

	return createValidationErrorContext(
		logger,
		chainID,
		serviceID,
		rpcType,
		qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_JSONRPC_PARSE_ERROR,
		errorDetails,
		http.StatusMethodNotAllowed,
	)
}

// createUnsupportedRPCTypeError creates error context for unsupported RPC types
func createUnsupportedRPCTypeError(
	rpcType sharedtypes.RPCType,
	logger polylog.Logger,
	chainID string,
	serviceID protocol.ServiceID,
) gateway.RequestQoSContext {
	return createValidationErrorContext(
		logger,
		chainID,
		serviceID,
		convertToProtoRPCType(rpcType),
		qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_JSONRPC_PARSE_ERROR,
		"RPC type not supported by this service: "+rpcType.String(),
		httpStatusRequestValidationFailureUnsupportedRPCType,
	)
}

// createJSONRPCParseError creates error context for JSONRPC parsing failures
func createJSONRPCParseError(
	err error,
	logger polylog.Logger,
	chainID string,
	serviceID protocol.ServiceID,
) gateway.RequestQoSContext {
	return createValidationErrorContext(
		logger,
		chainID,
		serviceID,
		qosobservations.RPCType_RPC_TYPE_UNSPECIFIED,
		qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_JSONRPC_PARSE_ERROR,
		"Failed to parse JSONRPC request: "+err.Error(),
		http.StatusBadRequest,
	)
}

// createInvalidPathError creates error context for invalid REST paths
func createInvalidPathError(
	path string,
	logger polylog.Logger,
	chainID string,
	serviceID protocol.ServiceID,
) gateway.RequestQoSContext {
	return createValidationErrorContext(
		logger,
		chainID,
		serviceID,
		qosobservations.RPCType_RPC_TYPE_REST,
		qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_JSONRPC_PARSE_ERROR,
		"Invalid path for REST API: "+path,
		http.StatusNotFound,
	)
}

// createBodyReadError creates error context for request body read failures
func createBodyReadError(
	err error,
	logger polylog.Logger,
	chainID string,
	serviceID protocol.ServiceID,
) gateway.RequestQoSContext {
	return createValidationErrorContext(
		logger,
		chainID,
		serviceID,
		qosobservations.RPCType_RPC_TYPE_REST,
		qosobservations.RequestErrorKind_REQUEST_ERROR_INTERNAL_READ_HTTP_ERROR,
		"Failed to read REST request body: "+err.Error(),
		http.StatusInternalServerError,
	)
}
