package cosmos

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/buildwithgrove/path/gateway"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
	"github.com/pokt-network/poktroll/pkg/polylog"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
)

// Default timeout for REST requests
const defaultRESTRequestTimeoutMillisec = 10_000

// validateRESTRequest validates a REST request by:
// 1. Validating HTTP method and path
// 2. Determining the specific RPC type from the path
// 3. Checking if the RPC type is supported
// 4. Creating the request context with all necessary information
func (rv *requestValidator) validateRESTRequest(
	httpRequestURL *url.URL,
	httpRequestMethod string,
	httpRequestBody []byte,
) (gateway.RequestQoSContext, bool) {
	httpRequestPath := httpRequestURL.Path

	logger := rv.logger.With(
		"validator", "REST",
		"http_path", httpRequestPath,
	)

	// Determine the specific RPC type based on path patterns - delegate to specialized detection
	rpcType, err := determineRESTRPCType(httpRequestPath)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to identify the target backend service using the request path")
		// TODO_TECHDEBT(@adshmh): Review error context creation for REST requests.
		return rv.createRESTServiceDetectionFailureContext(httpRequestURL, httpRequestMethod, httpRequestBody, err), false
	}

	logger = logger.With(
		"detected_rpc_type", rpcType.String(),
		"request_path", httpRequestPath,
	)

	// Check if this RPC type is supported by the service
	if _, supported := rv.supportedAPIs[rpcType]; !supported {
		logger.Warn().Msg("Request uses unsupported RPC type")
		// TODO_TECHDEBT(@adshmh): Review error context creation for REST requests.
		return rv.createRESTUnsupportedRPCTypeContext(httpRequestURL, httpRequestMethod, httpRequestBody, rpcType), false
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
	requestObservation := rv.buildRESTRequestObservations(
		rpcType,
		httpRequestURL,
		httpRequestMethod,
		httpRequestBody,
		servicePayload,
	)

	logger.With(
		"payload_length", len(servicePayload.Data),
		"request_path", servicePayload.Path,
	).Debug().Msg("REST request validation successful.")

	// Create specialized REST context
	return &requestContext{
		logger:                       logger,
		servicePayload:               servicePayload,
		observations:                 requestObservation,
		endpointResponseValidator:    getRESTRequestEndpointResponseValidator(),
		protocolErrorResponseBuilder: buildRESTProtocolErrorResponse(),
		// Protocol-level request error observation is the same for JSONRPC and REST.
		protocolErrorObservationBuilder: buildProtocolErrorObservation,
	}, true
}

func buildRESTServicePayload(
	rpcType sharedtypes.RPCType,
	httpRequestURL *url.URL,
	httpRequestMethod string,
	httpRequestBody []byte,
) protocol.Payload {
	path := httpRequestURL.Path
	if httpRequestURL.RawQuery != "" {
		path += "?" + httpRequestURL.RawQuery
	}

	return protocol.Payload{
		Data:            string(httpRequestBody),
		Method:          httpRequestMethod,
		TimeoutMillisec: defaultRESTRequestTimeoutMillisec,
		// Add the RPCType hint, so protocol sets correct HTTP headers for the endpoint.
		RPCType: rpcType,
		// Set the request path, including raw query, if used.
		Path: path,
	}
}

func getRESTRequestEndpointResponseValidator() func(polylog.Logger, []byte) response {
	// Delegate the unmarshaling/validation of endpoint response to the specialized REST unmarshaler.
	return func(logger polylog.Logger, endpointResponseBz []byte) response {
		return unmarshalRESTRequestEndpointResponse(logger, endpointResponseBz)
	}
}

func (rv *requestValidator) buildRESTRequestObservations(
	rpcType sharedtypes.RPCType,
	httpRequestURL *url.URL,
	httpRequestMethod string,
	httpRequestBody []byte,
	servicePayload protocol.Payload,
) *qosobservations.CosmosRequestObservations {

	// Determine content type from headers if available, otherwise empty
	contentType := ""
	// Note: We don't have access to headers here, but this would be where we'd extract it

	return &qosobservations.CosmosRequestObservations{
		ChainId:       rv.chainID,
		ServiceId:     string(rv.serviceID),
		RequestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
		RequestProfile: &qosobservations.CosmosRequestProfile{
			BackendServiceDetails: &qosobservations.BackendServiceDetails{
				BackendServiceType: convertToProtoBackendServiceType(rpcType),
				SelectionReason:    "REST path detection",
			},
			ParsedRequest: &qosobservations.CosmosRequestProfile_RestRequest{
				RestRequest: &qosobservations.RESTRequest{
					ApiPath:       httpRequestURL.Path,
					HttpMethod:    httpRequestMethod,
					ContentType:   contentType,
					PayloadLength: uint32(len(httpRequestBody)),
				},
			},
		},
	}
}

// TODO_TECHDEBT(@adshmh): Review the expected user experience on protocol errors in REST requests.
func buildRESTProtocolErrorResponse() func(logger polylog.Logger) gateway.HTTPResponse {
	return func(logger polylog.Logger) gateway.HTTPResponse {
		// For REST requests, we return a JSON-RPC error response with null ID
		// TODO_TECHDEBT(@adshmh): Consider returning proper REST error response format
		errorResp := jsonrpc.NewErrResponseInternalErr(
			jsonrpc.ID{}, // use null as ID.
			errors.New("protocol-level error: no endpoint responses received"),
		)
		return qos.BuildHTTPResponseFromJSONRPCResponse(logger, errorResp)
	}
}

// TODO_TECHDEBT(@adshmh): Review error context creation for REST requests.
// createRESTServiceDetectionFailureContext creates an error context for REST service detection failures
func (rv *requestValidator) createRESTServiceDetectionFailureContext(
	httpRequestURL *url.URL,
	httpRequestMethod string,
	httpRequestBody []byte,
	err error,
) gateway.RequestQoSContext {
	// Create the JSON-RPC error response (reusing JSONRPC error format for now)
	response := jsonrpc.NewErrResponseMethodNotFound(jsonrpc.ID{}, err)

	// Create the observations object with the service detection failure observation
	observations := rv.createRESTServiceDetectionFailureObservation(
		httpRequestURL,
		httpRequestMethod,
		httpRequestBody,
		err,
		response,
	)

	// Build and return the error context
	return &qos.RequestErrorContext{
		Logger:   rv.logger,
		Response: qos.BuildHTTPResponseFromJSONRPCResponse(rv.logger, response),
		Observations: &qosobservations.Observations{
			ServiceObservations: &qosobservations.Observations_Cosmos{
				Cosmos: observations,
			},
		},
	}
}

func (rv *requestValidator) createRESTServiceDetectionFailureObservation(
	httpRequestURL *url.URL,
	httpRequestMethod string,
	httpRequestBody []byte,
	err error,
	jsonrpcResponse jsonrpc.Response,
) *qosobservations.CosmosRequestObservations {
	return &qosobservations.CosmosRequestObservations{
		ServiceId:     string(rv.serviceID),
		ChainId:       rv.chainID,
		RequestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
		RequestProfile: &qosobservations.CosmosRequestProfile{
			ParsedRequest: &qosobservations.CosmosRequestProfile_RestRequest{
				RestRequest: &qosobservations.RESTRequest{
					ApiPath:       httpRequestURL.Path,
					HttpMethod:    httpRequestMethod,
					ContentType:   "", // Not available at this point
					PayloadLength: uint32(len(httpRequestBody)),
				},
			},
		},
		RequestLevelError: &qosobservations.RequestError{
			ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_REST_SERVICE_DETECTION_ERROR,
			ErrorDetails:   truncateErrorMessage(err.Error()),
			HttpStatusCode: int32(jsonrpcResponse.GetRecommendedHTTPStatusCode()),
		},
	}
}

// TODO_TECHDEBT(@adshmh): Review error context creation for REST requests.
// createRESTUnsupportedRPCTypeContext creates an error context for unsupported RPC type in REST requests
func (rv *requestValidator) createRESTUnsupportedRPCTypeContext(
	httpRequestURL *url.URL,
	httpRequestMethod string,
	httpRequestBody []byte,
	rpcType sharedtypes.RPCType,
) gateway.RequestQoSContext {
	// Create the JSON-RPC error response (reusing JSONRPC error format for now)
	err := errors.New("unsupported RPC type: " + rpcType.String())
	response := jsonrpc.NewErrResponseMethodNotFound(jsonrpc.ID{}, err)

	// Create the observations object with the unsupported RPC type observation
	observations := rv.createRESTUnsupportedRPCTypeObservation(
		httpRequestURL,
		httpRequestMethod,
		httpRequestBody,
		rpcType,
		response,
	)

	// Build and return the error context
	return &qos.RequestErrorContext{
		Logger:   rv.logger,
		Response: qos.BuildHTTPResponseFromJSONRPCResponse(rv.logger, response),
		Observations: &qosobservations.Observations{
			ServiceObservations: &qosobservations.Observations_Cosmos{
				Cosmos: observations,
			},
		},
	}
}

func (rv *requestValidator) createRESTUnsupportedRPCTypeObservation(
	httpRequestURL *url.URL,
	httpRequestMethod string,
	httpRequestBody []byte,
	rpcType sharedtypes.RPCType,
	jsonrpcResponse jsonrpc.Response,
) *qosobservations.CosmosRequestObservations {
	return &qosobservations.CosmosRequestObservations{
		ServiceId:     string(rv.serviceID),
		ChainId:       rv.chainID,
		RequestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_ORGANIC,
		RequestProfile: &qosobservations.CosmosRequestProfile{
			BackendServiceDetails: &qosobservations.BackendServiceDetails{
				BackendServiceType: convertToProtoBackendServiceType(rpcType),
				SelectionReason:    "REST path detection (unsupported)",
			},
			ParsedRequest: &qosobservations.CosmosRequestProfile_RestRequest{
				RestRequest: &qosobservations.RESTRequest{
					ApiPath:       httpRequestURL.Path,
					HttpMethod:    httpRequestMethod,
					ContentType:   "", // Not available at this point
					PayloadLength: uint32(len(httpRequestBody)),
				},
			},
		},
		RequestLevelError: &qosobservations.RequestError{
			ErrorKind:      qosobservations.RequestErrorKind_REQUEST_ERROR_USER_ERROR_REST_UNSUPPORTED_RPC_TYPE,
			ErrorDetails:   "Unsupported RPC type: " + rpcType.String(),
			HttpStatusCode: int32(jsonrpcResponse.GetRecommendedHTTPStatusCode()),
		},
	}
}
