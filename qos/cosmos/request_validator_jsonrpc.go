package cosmos

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/pokt-network/poktroll/pkg/polylog"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
)

// maximum length of the error message stored in request validation failure observations and logs.
const maxErrMessageLen = 1000

// jsonrpcRequestValidator handles validation of JSONRPC requests
// Determines RPC type from JSONRPC method and creates appropriate context
type jsonrpcRequestValidator struct{}

// validateJSONRPCRequest validates a JSONRPC request by:
// 1. Reading and parsing the JSONRPC request
// 2. Determining the specific RPC type from the method
// 3. Checking if the RPC type is supported
// 4. Creating the request context with all necessary information
func (jv *jsonrpcRequestValidator) validateJSONRPCRequest(
	req *http.Request,
	supportedAPIs map[sharedtypes.RPCType]struct{},
	logger polylog.Logger,
	chainID string,
	serviceID protocol.ServiceID,
) (gateway.RequestQoSContext, bool) {

	logger = logger.With("validator", "JSONRPC")

	// JSONRPC requests must be POST
	if req.Method != http.MethodPost {
		logger.Warn().Str("method", req.Method).Msg("JSONRPC requests must use POST method")
		return jv.createInvalidMethodContext(req.Method, logger, chainID, serviceID), false
	}

	// Read and parse JSONRPC request
	jsonrpcReq, rpcType, err := jv.parseJSONRPCAndDetectService(req, logger)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to parse JSONRPC request")
		return jv.createJSONRPCParseErrorContext(err, logger, chainID, serviceID), false
	}

	logger = logger.With("detected_rpc_type", rpcType.String(), "method", jsonrpcReq.Method)

	// Check if this RPC type is supported by the service
	if _, supported := supportedAPIs[rpcType]; !supported {
		logger.Warn().Msg("Request uses unsupported RPC type")
		return jv.createUnsupportedRPCTypeContext(rpcType, logger, chainID, serviceID), false
	}

	// Calculate payload length for metrics
	reqBz, _ := json.Marshal(jsonrpcReq)
	payloadLength := len(reqBz)

	logger.Debug().
		Str("id", jsonrpcReq.ID.String()).
		Int("payload_length", payloadLength).
		Msg("JSONRPC request validation successful")

	// Create request context with detected RPC type
	return &requestContext{
		logger:               logger,
		httpReq:              *req,
		chainID:              chainID,
		serviceID:            serviceID,
		rpcType:              rpcType,
		requestPayloadLength: uint(payloadLength),
		requestOrigin:        qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
		jsonrpcReq:           &jsonrpcReq,
	}, true
}

// parseJSONRPCAndDetectService parses the JSONRPC request and determines service type
func (jv *jsonrpcRequestValidator) parseJSONRPCAndDetectService(req *http.Request, logger polylog.Logger) (jsonrpc.Request, sharedtypes.RPCType, error) {
	// Read body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return jsonrpc.Request{}, sharedtypes.RPCType_UNSPECIFIED, err
	}

	// Restore body for potential later use
	req.Body = io.NopCloser(bytes.NewReader(body))

	if len(body) == 0 {
		return jsonrpc.Request{}, sharedtypes.RPCType_UNSPECIFIED, &jsonrpc.Error{
			Code:    -32600,
			Message: "Request body is empty",
		}
	}

	// Parse JSONRPC request
	var jsonrpcReq jsonrpc.Request
	if err := json.Unmarshal(body, &jsonrpcReq); err != nil {
		return jsonrpc.Request{}, sharedtypes.RPCType_UNSPECIFIED, err
	}

	// Basic validation
	if jsonrpcReq.Method == "" {
		return jsonrpc.Request{}, sharedtypes.RPCType_UNSPECIFIED, &jsonrpc.Error{
			Code:    -32600,
			Message: "Invalid Request: missing method field",
		}
	}

	// Determine service type based on method - delegate to specialized detection
	method := string(jsonrpcReq.Method)
	rpcType := detectJSONRPCServiceType(method)

	return jsonrpcReq, rpcType, nil
}

// Error context creation methods

func (jv *jsonrpcRequestValidator) createInvalidMethodContext(method string, logger polylog.Logger, chainID string, serviceID protocol.ServiceID) gateway.RequestQoSContext {
	observations := &qosobservations.Observations_Cosmos{
		Cosmos: &qosobservations.CosmosSDKRequestObservations{
			ChainId:       chainID,
			ServiceId:     string(serviceID),
			RequestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
			RpcType:       qosobservations.RPCType_RPC_TYPE_UNSPECIFIED,
			RequestError: &qosobservations.RequestError{
				ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_JSONRPC_PARSE_ERROR,
				ErrorDetails:   "JSONRPC requests must use POST method, got: " + method,
				HttpStatusCode: http.StatusMethodNotAllowed,
			},
		},
	}

	return &errorContext{
		logger:                 logger,
		responseHTTPStatusCode: http.StatusMethodNotAllowed,
		cosmosSDKObservations:  observations,
	}
}

func (jv *jsonrpcRequestValidator) createJSONRPCParseErrorContext(err error, logger polylog.Logger, chainID string, serviceID protocol.ServiceID) gateway.RequestQoSContext {
	observations := &qosobservations.Observations_Cosmos{
		Cosmos: &qosobservations.CosmosSDKRequestObservations{
			ChainId:       chainID,
			ServiceId:     string(serviceID),
			RequestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
			RpcType:       qosobservations.RPCType_RPC_TYPE_UNSPECIFIED,
			RequestError: &qosobservations.RequestError{
				ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_JSONRPC_PARSE_ERROR,
				ErrorDetails:   "Failed to parse JSONRPC request: " + err.Error(),
				HttpStatusCode: http.StatusBadRequest,
			},
		},
	}

	return &errorContext{
		logger:                 logger,
		responseHTTPStatusCode: http.StatusBadRequest,
		cosmosSDKObservations:  observations,
	}
}

func (jv *jsonrpcRequestValidator) createUnsupportedRPCTypeContext(rpcType sharedtypes.RPCType, logger polylog.Logger, chainID string, serviceID protocol.ServiceID) gateway.RequestQoSContext {
	observations := &qosobservations.Observations_Cosmos{
		Cosmos: &qosobservations.CosmosSDKRequestObservations{
			ChainId:       chainID,
			ServiceId:     string(serviceID),
			RequestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
			RpcType:       convertToProtoRPCType(rpcType),
			RequestError: &qosobservations.RequestError{
				ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_JSONRPC_PARSE_ERROR,
				ErrorDetails:   "RPC type not supported by this service: " + rpcType.String(),
				HttpStatusCode: httpStatusRequestValidationFailureUnsupportedRPCType,
			},
		},
	}
	return &errorContext{
		logger:                 logger,
		responseHTTPStatusCode: httpStatusRequestValidationFailureUnsupportedRPCType,
		cosmosSDKObservations:  observations,
	}
}
