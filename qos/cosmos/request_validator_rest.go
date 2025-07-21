package cosmos

import (
	"io"
	"net/http"
	"net/url"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/pokt-network/poktroll/pkg/polylog"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
)

// restRequestValidator handles validation of REST API requests
// Determines RPC type from request path and creates appropriate context
type restRequestValidator struct{}

// validateRESTRequest validates a REST request by:
// 1. Validating HTTP method and path
// 2. Determining the specific RPC type from the path
// 3. Checking if the RPC type is supported
// 4. Creating the request context with all necessary information
func (rv *restRequestValidator) validateRESTRequest(
	httpRequestURL *url.URL,
	httpRequestMethod string,
	httpRequestBody []byte,
) (gateway.RequestQoSContext, bool) {
	// Determine the specific RPC type based on path patterns - delegate to specialized detection
	rpcType := determineRESTRPCType(httpRequestPath)
	logger = logger.With(
		"validator", "REST",
		"detected_rpc_type", rpcType.String(),
	)

	// Check if this RPC type is supported by the service
	if _, supported := rv.supportedAPIs[rpcType]; !supported {
		logger.Warn().Msg("Request uses unsupported RPC type")
		return createUnsupportedRPCTypeError(rpcType, logger, chainID, serviceID), false
	}

	logger.Debug().
		Int("body_length", len(httpRequestBody)).
		Msg("REST request validation successful")

	// Build and return the request context
	return rv.buildRESTRequestContext(
		rpcType,
		httpRequestURL,
		httpRequestMethod,
		httpRequestBody,
	)
}

// TODO_TECHDEBT(@adshmh): Validate the REST payload based on HTTP request's path.
//
func (rv *requestValidator) buildRESTRequestContext(
	rpcType sharedtypes.RPCType,
	httpRequestURL *url.URL,
	httpRequestMethod string,
	httpRequestBody []byte,
) (gateway.RequestQoSContext, bool) {
	logger := rv.logger.With(
		"method", "buildRESTRequestContext",
	)

	// Build service payload
	servicePayload := buildRESTServicePayload(
		rpcType,
		httpRequestURL,
		httpRequestMethod,
		httpRequestBody,
	)

	// Generate the QoS observation for the request.
	// requestContext will amend this with endpoint observation(s).
	requestObservation := rv.buildRESTRequestObservations(rpcType, servicePayload)

	logger.With(
		"payload_length", len(servicePayload.Data),
		"request_path", servicePayload.Path,
	).Debug().Msg("REST request validation successful.")

	// Create specialized REST context
	return &requestContext{
		logger:               logger,
		servicePayload: servicePayload,
		observations:   requestObservations,
		endpointResponseValidator: getRESTRequestEndpointResponseValidator(jsonrpcReq),
		protocolErrorResponseBuilder: buildRESTProtocolErrorResponse,
		// Protocol-level request error observation is the same for JSONRPC and REST.
		protocolErrorObservationBuilder: buildProtocolErrorObservation,
	}, true
}

func buildRESTServicePayload(
	rpcType sharedtypes.RpcType,
	httpRequestURL *url.URL,
	httpRequestMethod string,
	httpRequestBody []byte
) (protocol.Payload, error) {
	path := httpRequestURL.Path
	if httpRequestURL.RawQuery != "" {
		path += "?" + httpRequestURL.RawQuery
	}

	return protocol.Payload{
		Data:            string(reqBz),
		Method:          httpRequestMethod,
		TimeoutMillisec: defaultJSONRPCRequestTimeoutMillisec,
		// Add the RPCType hint, so protocol sets correct HTTP headers for the endpoint.
		RPCType:         rpcType,
		// Set the request path, including raw query, if used.
		Path:            path,
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

func (rv *requestValidator) buildRESTRequestObservations(
	rpcType sharedtypes.RPCType,
	servicePayload protocol.Payload,
) *qosobservations.CosmosRequestObservations {

	return &qosobservations.CosmosRequestObservations{
		ChainId:              rv.chainID,
		ServiceId:            string(rv.serviceID),
		RequestPayloadLength: uint32(len(servicePayload.Data)),
		requestOrigin:        qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
		RpcType:              convertToProtoRPCType(rpcType),
		RestRequest:          ...
	}
}

// TODO_TECHDEBT(@adshmh): Review the expected user experience on protocol errors in REST requests.
//
func buildRESTProtocolErrorResponse() func(logger polylog.Logger) gateway.HTTPResponse {
	return func() response {
		errorResp := jsonrpc.NewErrResponseInternalErr(
			jsonrpc.ID{}, // use null as ID.
			errors.New("protocol-level error: no endpoint responses received"),
		)
		return qos.BuildHTTPResponseFromJSONRPCResponse(logger, errorResp)
	}
}

func buildRESTProtocolErrorResponse() func(logger polylog.Logger) gateway.HTTPResponse {
	return func() response {
		errorResp := jsonrpc.NewErrResponseInternalErr(
			jsonrpc.ID{}, // use null as ID.
			errors.New("protocol-level error: no endpoint responses received"),
		)
		return qos.BuildHTTPResponseFromJSONRPCResponse(logger, errorResp)
	}
}


