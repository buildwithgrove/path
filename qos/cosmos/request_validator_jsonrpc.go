package cosmos

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

const (
	// maximum length of the error message stored in request validation failure observations and logs.
	maxErrMessageLen = 1000

	// defaultJSONRPCRequestTimeoutMillisec is the default timeout when sending a request to a Cosmos blockchain endpoint.
	// TODO_IMPROVE(@adshmh): Support method level specific timeouts and allow the user to configure them.
	defaultJSONRPCRequestTimeoutMillisec = 10_000
)

// validateJSONRPCRequest validates a JSONRPC request by:
// 1. Reading and parsing the JSONRPC request
// 2. Determining the specific RPC type from the method
// 3. Checking if the RPC type is supported
// 4. Creating the request context with all necessary information
func (rv *requestValidator) validateJSONRPCRequest(
	body []byte,
) (gateway.RequestQoSContext, bool) {

	logger := rv.logger.With("validator", "JSONRPC")

	// Parse JSONRPC request
	var jsonrpcReq jsonrpc.Request
	if err := json.Unmarshal(body, &jsonrpcReq); err != nil {
		logger.Warn().Err(err).Msg("Failed to parse JSONRPC request")
		return rv.createJSONRPCParseFailureContext(err), false
	}

	// Determine service type based on JSONRPC request's method
	method := string(jsonrpcReq.Method)
	rpcType := detectJSONRPCServiceType(method)

	// Hydrate the logger with data extracted from the request.
	logger = logger.With(
		"rpc_type", rpcType.String(),
		"jsonrpc_method", method,
	)

	// Check if this RPC type is supported by the service
	if _, supported := rv.supportedAPIs[rpcType]; !supported {
		logger.Warn().Msg("Request uses unsupported RPC type")
		return rv.createJSONRPCUnsupportedRPCTypeContext(jsonrpcReq, rpcType), false
	}

	// Build and return the request context
	return rv.buildJSONRPCRequestContext(
		rpcType,
		jsonrpcReq,
		qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
	)
}

func (rv *requestValidator) buildJSONRPCRequestContext(
	rpcType sharedtypes.RPCType,
	jsonrpcReq jsonrpc.Request,
	requestOrigin qosobservations.RequestOrigin,
) (gateway.RequestQoSContext, bool) {
	logger := rv.logger.With(
		"method", "buildJSONRPCRequestContext",
	)

	// Build service payload
	servicePayload, err := buildJSONRPCServicePayload(rpcType, jsonrpcReq)
	if err != nil {
		logger.Warn().Err(err).Msg("SHOULD NEVER HAPPEN: failed to marshal JSONRPC service payload")

		return rv.createJSONRPCServicePayloadBuildFailureContext(jsonrpcReq, err), false
	}

	// Generate the QoS observation for the request.
	// requestContext will amend this with endpoint observation(s).
	requestObservation := rv.buildJSONRPCRequestObservations(
		rpcType,
		jsonrpcReq,
		servicePayload,
		requestOrigin,
	)

	logger.Debug().
		Str("id", jsonrpcReq.ID.String()).
		Int("payload_length", len(servicePayload.Data)).
		Msg("JSONRPC request validation successful.")

	// Hydrate the logger with JSONRPC method.
	logger = logger.With(
		"jsonrpc_request", jsonrpcReq,
	)

	// Create specialized JSONRPC context
	return &requestContext{
		logger:                       logger,
		serviceState:                 rv.serviceState,
		servicePayload:               servicePayload,
		observations:                 requestObservation,
		endpointResponseValidator:    getJSONRPCRequestEndpointResponseValidator(jsonrpcReq),
		protocolErrorResponseBuilder: buildJSONRPCProtocolErrorResponse(jsonrpcReq.ID),
		// Protocol-level request error observation is the same for JSONRPC and REST.
		protocolErrorObservationBuilder: buildProtocolErrorObservation,
	}, true
}

// buildJSONRPCServicePayload builds a protocol payload for a JSONRPC request.
func buildJSONRPCServicePayload(rpcType sharedtypes.RPCType, jsonrpcReq jsonrpc.Request) (protocol.Payload, error) {
	// DEV_NOTE: marshaling the request, rather than using the original payload, is necessary.
	// Otherwise, a request missing `id` field could fail.
	// See the Request struct in `jsonrpc` package for the details.
	reqBz, err := json.Marshal(jsonrpcReq)
	if err != nil {
		return protocol.EmptyErrorPayload(), err
	}

	return protocol.Payload{
		Data:            string(reqBz),
		Method:          http.MethodPost, // JSONRPC always uses POST
		Path:            "",              // JSONRPC does not use paths
		Headers:         map[string]string{},
		TimeoutMillisec: defaultJSONRPCRequestTimeoutMillisec,
		RPCType:         rpcType, // Add the RPCType hint the so protocol sets correct HTTP headers for the endpoint.
	}, nil
}

func getJSONRPCRequestEndpointResponseValidator(
	jsonrpcReq jsonrpc.Request,
) func(polylog.Logger, []byte) response {

	// Delegate the unmarshaling/validation of endpoint response to the specialized JSONRPC unmarshaler.
	return func(logger polylog.Logger, endpointResponseBz []byte) response {
		return unmarshalJSONRPCRequestEndpointResponse(logger, jsonrpcReq, endpointResponseBz)
	}
}

func buildJSONRPCProtocolErrorResponse(
	jsonrpcRequestID jsonrpc.ID,
) func(logger polylog.Logger) gateway.HTTPResponse {
	return func(logger polylog.Logger) gateway.HTTPResponse {
		errorResp := jsonrpc.NewErrResponseInternalErr(
			jsonrpcRequestID,
			errors.New("protocol-level error: no endpoint responses received"),
		)
		return qos.BuildHTTPResponseFromJSONRPCResponse(logger, errorResp)
	}
}

func (rv *requestValidator) buildJSONRPCRequestObservations(
	rpcType sharedtypes.RPCType,
	jsonrpcReq jsonrpc.Request,
	servicePayload protocol.Payload,
	requestOrigin qosobservations.RequestOrigin,
) *qosobservations.CosmosRequestObservations {

	return &qosobservations.CosmosRequestObservations{
		CosmosChainId: rv.cosmosChainID,
		EvmChainId:    rv.evmChainID,
		ServiceId:     string(rv.serviceID),
		RequestOrigin: requestOrigin,
		RequestProfile: &qosobservations.CosmosRequestProfile{
			BackendServiceDetails: &qosobservations.BackendServiceDetails{
				BackendServiceType: convertToProtoBackendServiceType(rpcType),
				SelectionReason:    "JSONRPC method detection",
			},
			ParsedRequest: &qosobservations.CosmosRequestProfile_JsonrpcRequest{
				JsonrpcRequest: jsonrpcReq.GetObservation(),
			},
		},
	}
}

// Used for both JSONRPC and REST requests.
func buildProtocolErrorObservation() *qosobservations.RequestError {
	return &qosobservations.RequestError{
		ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_INTERNAL_PROTOCOL_ERROR,
		ErrorDetails:   "No endpoint responses received",
		HttpStatusCode: int32(http.StatusInternalServerError),
	}
}

// convertToProtoBackendServiceType converts sharedtypes.RPCType to proto BackendServiceType
func convertToProtoBackendServiceType(rpcType sharedtypes.RPCType) qosobservations.BackendServiceType {
	switch rpcType {
	case sharedtypes.RPCType_JSON_RPC:
		return qosobservations.BackendServiceType_BACKEND_SERVICE_TYPE_JSONRPC
	case sharedtypes.RPCType_REST:
		return qosobservations.BackendServiceType_BACKEND_SERVICE_TYPE_REST
	default:
		return qosobservations.BackendServiceType_BACKEND_SERVICE_TYPE_UNSPECIFIED
	}
}

// createJSONRPCParseFailureContext creates an error context for JSONRPC parsing failures
func (rv *requestValidator) createJSONRPCParseFailureContext(err error) gateway.RequestQoSContext {
	// Create the JSON-RPC error response with empty ID since we couldn't parse the request
	response := jsonrpc.NewErrResponseInvalidRequest(jsonrpc.ID{}, err)

	// Create the observations object with the parse failure observation
	observations := rv.createJSONRPCParseFailureObservation(err, response)

	// Build and return the error context
	return &qos.RequestErrorContext{
		Logger:   rv.logger,
		Response: response,
		Observations: &qosobservations.Observations{
			ServiceObservations: &qosobservations.Observations_Cosmos{
				Cosmos: observations,
			},
		},
	}
}

func (rv *requestValidator) createJSONRPCParseFailureObservation(
	err error,
	jsonrpcResponse jsonrpc.Response,
) *qosobservations.CosmosRequestObservations {
	return &qosobservations.CosmosRequestObservations{
		CosmosChainId: rv.cosmosChainID,
		EvmChainId:    rv.evmChainID,
		ServiceId:     string(rv.serviceID),
		RequestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
		RequestLevelError: &qosobservations.RequestError{
			ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_JSONRPC_PARSE_ERROR,
			ErrorDetails:   truncateErrorMessage(err.Error()),
			HttpStatusCode: int32(jsonrpcResponse.GetRecommendedHTTPStatusCode()),
		},
	}
}

// createJSONRPCUnsupportedRPCTypeContext creates an error context for unsupported RPC type
func (rv *requestValidator) createJSONRPCUnsupportedRPCTypeContext(jsonrpcReq jsonrpc.Request, rpcType sharedtypes.RPCType) gateway.RequestQoSContext {
	// Create the JSON-RPC error response
	err := errors.New("unsupported RPC type: " + rpcType.String())
	response := jsonrpc.NewErrResponseInvalidRequest(jsonrpcReq.ID, err)

	// Create the observations object with the unsupported RPC type observation
	observations := rv.createJSONRPCUnsupportedRPCTypeObservation(jsonrpcReq, rpcType, response)

	// Build and return the error context
	return &qos.RequestErrorContext{
		Logger:   rv.logger,
		Response: response,
		Observations: &qosobservations.Observations{
			ServiceObservations: &qosobservations.Observations_Cosmos{
				Cosmos: observations,
			},
		},
	}
}

func (rv *requestValidator) createJSONRPCUnsupportedRPCTypeObservation(
	jsonrpcReq jsonrpc.Request,
	rpcType sharedtypes.RPCType,
	jsonrpcResponse jsonrpc.Response,
) *qosobservations.CosmosRequestObservations {
	return &qosobservations.CosmosRequestObservations{
		CosmosChainId: rv.cosmosChainID,
		EvmChainId:    rv.evmChainID,
		ServiceId:     string(rv.serviceID),
		RequestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
		RequestProfile: &qosobservations.CosmosRequestProfile{
			BackendServiceDetails: &qosobservations.BackendServiceDetails{
				BackendServiceType: convertToProtoBackendServiceType(rpcType),
				SelectionReason:    "JSONRPC method detection (unsupported)",
			},
			ParsedRequest: &qosobservations.CosmosRequestProfile_JsonrpcRequest{
				JsonrpcRequest: jsonrpcReq.GetObservation(),
			},
		},
		RequestLevelError: &qosobservations.RequestError{
			ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_JSONRPC_UNSUPPORTED_RPC_TYPE,
			ErrorDetails:   fmt.Sprintf("Unsupported RPC type %s for service %s", rpcType.String(), string(rv.serviceID)),
			HttpStatusCode: int32(jsonrpcResponse.GetRecommendedHTTPStatusCode()),
		},
	}
}

// createJSONRPCServicePayloadBuildFailureContext creates an error context for service payload build failures
func (rv *requestValidator) createJSONRPCServicePayloadBuildFailureContext(jsonrpcReq jsonrpc.Request, err error) gateway.RequestQoSContext {
	// Create the JSON-RPC error response
	response := jsonrpc.NewErrResponseInternalErr(jsonrpcReq.ID, err)

	// Create the observations object with the payload build failure observation
	observations := rv.createJSONRPCServicePayloadBuildFailureObservation(jsonrpcReq, err, response)

	// Build and return the error context
	return &qos.RequestErrorContext{
		Logger:   rv.logger,
		Response: response,
		Observations: &qosobservations.Observations{
			ServiceObservations: &qosobservations.Observations_Cosmos{
				Cosmos: observations,
			},
		},
	}
}

func (rv *requestValidator) createJSONRPCServicePayloadBuildFailureObservation(
	jsonrpcReq jsonrpc.Request,
	err error,
	jsonrpcResponse jsonrpc.Response,
) *qosobservations.CosmosRequestObservations {
	return &qosobservations.CosmosRequestObservations{
		CosmosChainId: rv.cosmosChainID,
		EvmChainId:    rv.evmChainID,
		ServiceId:     string(rv.serviceID),
		RequestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
		RequestProfile: &qosobservations.CosmosRequestProfile{
			ParsedRequest: &qosobservations.CosmosRequestProfile_JsonrpcRequest{
				JsonrpcRequest: jsonrpcReq.GetObservation(),
			},
		},
		RequestLevelError: &qosobservations.RequestError{
			ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_INTERNAL_JSONRPC_PAYLOAD_BUILD_ERROR,
			ErrorDetails:   truncateErrorMessage(err.Error()),
			HttpStatusCode: int32(jsonrpcResponse.GetRecommendedHTTPStatusCode()),
		},
	}
}

// truncateErrorMessage truncates error message to maxErrMessageLen
func truncateErrorMessage(errMsg string) string {
	if len(errMsg) <= maxErrMessageLen {
		return errMsg
	}
	return errMsg[:maxErrMessageLen]
}
