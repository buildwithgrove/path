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
		return errContext..., false
	}

	// Determine service type based on JSONRPC request's method
	method := string(jsonrpcReq.Method)
	rpcType, err := detectJSONRPCServiceType(method)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to identify the target backend service using the request")
		return errContext..., false
	}

	// Hydrate the logger with data extracted from the request.
	rv.logger = logger.With(
		"detected_rpc_type", rpcType.String(), 
		"jsonrpc_method", method,
	)

	// Check if this RPC type is supported by the service
	if _, supported := rv.supportedAPIs[rpcType]; !supported {
		rv.logger.Warn().Msg("Request uses unsupported RPC type")
		return errContext..., false
	}

	// Build and return the request context
	return rv.buildJSONRPCRequestContext(
		rpcType,
		jsonrpcReq,
	)
}

func (rv *requestValidator) buildJSONRPCRequestContext(
	rpcType sharedtypes.RPCType,
	jsonrpcReq jsonrpc.Request,
) (gateway.RequestQoSContext, bool) {
	logger := rv.logger.With(
		"method", "buildJSONRPCRequestContext",
	)

	// Build service payload
	servicePayload, err := buildJSONRPCServicePayload(rpcType, jsonrpcReq)
	if err != nil {
		logger.Warn().Err(err).Msg("SHOULD NEVER HAPPEN: failed to marshal JSONRPC service payload")

		return errContext..., false
	}

	// Generate the QoS observation for the request.
	// requestContext will amend this with endpoint observation(s).
	requestObservation := rv.buildJSONRPCRequestObservations(rpcType, jsonrpcReq, servicePayload)

	logger.Debug().
		Str("id", jsonrpcReq.ID.String()).
		Int("payload_length", len(servicePayload.Data)).
		Msg("JSONRPC request validation successful.")


	// Create specialized JSONRPC context
	return &requestContext{
		logger:               logger,
		servicePayload: servicePayload,
		observations:   requestObservations,
		endpointResponseValidator: getJSONRPCRequestEndpointResponseValidator(jsonrpcReq),
		protocolErrorResponseBuilder: buildJSONRPCProtocolErrorResponse(jsonrpcReq.ID),
		// Protocol-level request error observation is the same for JSONRPC and REST.
		protocolErrorObservationBuilder: buildProtocolErrorObservation,
	}, true
}

func buildJSONRPCServicePayload(jsonrpcReq jsonrpc.Request) (protocol.Payload, error) {
	// DEV_NOTE: marshaling the request, rather than using the original payload, is necessary.
	// Otherwise, a request missing `id` field could fail.
	// See the Request struct in `jsonrpc` package for the details.
	reqBz, err := json.Marshal(jsonrpcReq)
	if err != nil {
		return protocol.Payload{}, err
	}

	return protocol.Payload{
		Data:            string(reqBz),
		// JSONRPC always uses POST
		Method:          http.MethodPost,
		TimeoutMillisec: defaultJSONRPCRequestTimeoutMillisec,
		// Add the RPCType hint, so protocol sets correct HTTP headers for the endpoint.
		RPCType:         rpcType,
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
) func (logger polylog.Logger) gateway.HTTPResponse {
	return func() response {
		errorResp := jsonrpc.NewErrResponseInternalErr(
			jc.jsonrpcReq.ID,
			errors.New("protocol-level error: no endpoint responses received"),
		)
		return qos.BuildHTTPResponseFromJSONRPCResponse(logger, errorResp)
	}
}

func (rv *requestValidator) buildJSONRPCRequestObservations(
	rpcType sharedtypes.RPCType,
	jsonrpcReq jsonrpc.Request,
	servicePayload protocol.Payload,
) *qosobservations.CosmosRequestObservations {

	return &qosobservations.CosmosRequestObservations{
		ChainId:              rv.chainID,
		ServiceId:            string(rv.serviceID),
		RequestPayloadLength: uint32(len(servicePayload.Data)),
		requestOrigin:        qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
		RpcType:              convertToProtoRPCType(rpcType),
		JsonrpcRequest:       jsonrpcReq.GetObservation(),
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
