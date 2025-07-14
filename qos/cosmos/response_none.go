package cosmos

import (
	"encoding/json"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseNone satisfies the response interface and handles the case
// where no response has been received from any endpoint.
// This is not the same as empty responses (where an endpoint responded with empty data).
var _ response = responseNone{}

// responseNone represents the absence of any endpoint response.
// This can occur due to protocol-level failures or when no endpoint was selected.
type responseNone struct {
	logger     polylog.Logger
	httpReq    http.Request     // HTTP request context for determining request type
	jsonrpcReq *jsonrpc.Request // JSON-RPC request (can be nil for REST requests)
}

// isJsonRpcRequest checks if this is a JSON-RPC request
// Uses the same logic as requestContext.isJsonRpcRequest()
func (r responseNone) isJsonRpcRequest() bool {
	return r.httpReq.Method == http.MethodPost && r.jsonrpcReq != nil
}

// GetObservation returns an observation indicating no endpoint provided a response.
// This allows tracking metrics for scenarios where endpoint selection or communication failed.
// Implements the response interface.
func (r responseNone) GetObservation() qosobservations.CosmosSDKEndpointObservation {
	if r.isJsonRpcRequest() {
		return qosobservations.CosmosSDKEndpointObservation{
			ResponseObservation: &qosobservations.CosmosSDKEndpointObservation_UnrecognizedResponse{
				UnrecognizedResponse: &qosobservations.CosmosSDKUnrecognizedResponse{
					JsonrpcResponse: &qosobservations.JsonRpcResponse{
						Id: r.jsonrpcReq.ID.String(),
					},
				},
			},
		}
	}

	// For REST requests or when jsonrpcReq is nil
	return qosobservations.CosmosSDKEndpointObservation{
		ResponseObservation: &qosobservations.CosmosSDKEndpointObservation_UnrecognizedResponse{
			UnrecognizedResponse: &qosobservations.CosmosSDKUnrecognizedResponse{},
		},
	}
}

// GetHTTPResponse creates and returns a predefined httpResponse for cases when QoS has received no responses from the protocol.
// Implements the response interface.
func (r responseNone) GetHTTPResponse() httpResponse {
	return httpResponse{
		responsePayload: r.GetResponsePayload(),
		httpStatusCode:  r.GetResponseStatusCode(),
	}
}

// getResponsePayload constructs an appropriate error response based on request type.
// For JSON-RPC requests: returns a JSONRPC error response with request ID
// For REST requests: returns a simple JSON error message
func (r responseNone) GetResponsePayload() []byte {
	var responsePayload []byte
	var err error

	if r.isJsonRpcRequest() {
		// JSON-RPC error response with proper ID
		userResponse := newErrResponseNoEndpointResponse(r.jsonrpcReq.ID)
		responsePayload, err = json.Marshal(userResponse)
	} else {
		// REST error response - simple JSON error message
		restErrorResponse := map[string]any{
			"error": map[string]any{
				"code":    -1,
				"message": "No endpoint response received",
				"data":    "The request could not be processed because no endpoint provided a response",
			},
		}
		responsePayload, err = json.Marshal(restErrorResponse)
	}

	if err != nil {
		// This should never happen: log an entry but return a fallback response
		r.logger.Warn().Err(err).Msg("responseNone: Marshaling error response failed.")
		return []byte(`{"error":{"code":-1,"message":"Internal error: failed to marshal error response"}}`)
	}

	return responsePayload
}

// getHTTPStatusCode returns the HTTP status code to be returned to the client.
// Always a 500 Internal Server Error for the responseNone struct.
func (r responseNone) GetResponseStatusCode() int {
	return httpStatusResponseValidationFailureNoResponse
}
