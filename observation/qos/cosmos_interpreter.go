package qos

import (
	"errors"

	"github.com/pokt-network/poktroll/pkg/polylog"
)

// CosmosSDKObservationInterpreter extracts and interprets data from CosmosSDK request observations.
// It provides methods to access metrics-relevant information for Prometheus reporting.
type CosmosSDKObservationInterpreter struct {
	// Logger for reporting issues during interpretation
	Logger polylog.Logger

	// Observations contains the raw CosmosSDK request data
	Observations *CosmosRequestObservations
}

// GetChainID returns the blockchain identifier from observations.
func (i *CosmosSDKObservationInterpreter) GetCosmosChainID() string {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get Cosmos SDK chain ID: nil observations")
		return ""
	}
	return i.Observations.CosmosChainId
}

// GetEVMChainID returns the EVM chain identifier from observations.
func (i *CosmosSDKObservationInterpreter) GetEVMChainID() string {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get EVM chain ID: nil observations")
		return ""
	}
	return i.Observations.EvmChainId
}

// GetServiceID returns the service identifier from observations.
func (i *CosmosSDKObservationInterpreter) GetServiceID() string {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get service ID: nil observations")
		return ""
	}
	return i.Observations.ServiceId
}

// TODO_TECHDEBT: For batch requests, this will only return one of the methods in the batch.
// GetRequestMethod returns the CosmosSDK RPC method name from the request profile.
func (i *CosmosSDKObservationInterpreter) GetRequestMethods() ([]string, bool) {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get request method: nil observations")
		return nil, false
	}

	// Check if request profiles are available
	if len(i.Observations.RequestProfiles) == 0 {
		return nil, false
	}

	requestMethods := []string{}

	// Iterate over each request profile
	for _, requestProfile := range i.Observations.RequestProfiles {
		// Handle different request types
		switch req := requestProfile.ParsedRequest.(type) {
		case *CosmosRequestProfile_RestRequest:
			if req.RestRequest != nil {
				// For REST requests, use the API path as the method
				requestMethods = append(requestMethods, req.RestRequest.ApiPath)
			}
		case *CosmosRequestProfile_JsonrpcRequest:
			if req.JsonrpcRequest != nil {
				// For JSON-RPC requests, use the method name
				requestMethods = append(requestMethods, req.JsonrpcRequest.Method)
			}
		}
	}

	return requestMethods, true
}

// IsRequestSuccessful determines if the request completed without errors.
func (i *CosmosSDKObservationInterpreter) IsRequestSuccessful() bool {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot determine request success: nil observations")
		return false
	}

	// RequestLevelError being nil is normal for successful requests
	return i.Observations.RequestLevelError == nil
}

// GetRequestErrorType returns the error type if request failed or empty string if successful.
func (i *CosmosSDKObservationInterpreter) GetRequestErrorType() string {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get error type: nil observations")
		return ""
	}

	// RequestLevelError being nil is normal for successful requests
	if i.Observations.RequestLevelError == nil {
		return ""
	}

	return i.Observations.RequestLevelError.ErrorKind.String()
}

// GetRPCType returns the RPC type from the backend service details.
func (i *CosmosSDKObservationInterpreter) GetRPCType() string {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get RPC type: nil observations")
		return ""
	}

	// Check if request profiles are available
	if len(i.Observations.RequestProfiles) == 0 {
		return ""
	}

	// Use the first request profile for RPC type extraction
	requestProfile := i.Observations.RequestProfiles[0]
	if requestProfile == nil || requestProfile.BackendServiceDetails == nil {
		return ""
	}

	return requestProfile.BackendServiceDetails.BackendServiceType.String()
}

// GetRequestHTTPStatus returns the HTTP status code from the request error or endpoint responses.
// Returns 200 if request was successful, 0 if observations are nil.
func (i *CosmosSDKObservationInterpreter) GetRequestHTTPStatus() int32 {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get HTTP status: nil observations")
		return 0 // Return 0 to indicate observation issues to metrics
	}

	// If there's a request-level error, return its HTTP status
	if i.Observations.RequestLevelError != nil {
		return i.Observations.RequestLevelError.HttpStatusCode
	}

	// If there are endpoint observations, return the HTTP status from the first endpoint response
	if len(i.Observations.EndpointObservations) > 0 {
		if i.Observations.EndpointObservations[0].EndpointResponseValidationResult != nil {
			return i.Observations.EndpointObservations[0].EndpointResponseValidationResult.HttpStatusCode
		}
	}

	// Default to 200 for successful requests with no specific status
	return 200
}

// GetTotalRequestPayloadLength calculates the total payload length across all request profiles.
// For Cosmos SDK requests, this aggregates payload lengths from both REST and JSON-RPC requests.
func (i *CosmosSDKObservationInterpreter) GetTotalRequestPayloadLength() uint32 {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get payload length: nil observations")
		return 0
	}

	var totalLength uint32
	for _, requestProfile := range i.Observations.RequestProfiles {
		if requestProfile == nil {
			continue
		}

		// Handle different request types
		switch req := requestProfile.ParsedRequest.(type) {
		case *CosmosRequestProfile_RestRequest:
			if req.RestRequest != nil {
				totalLength += req.RestRequest.PayloadLength
			}
		case *CosmosRequestProfile_JsonrpcRequest:
			// JSON-RPC requests don't have explicit payload length in the current structure
			// The payload length would need to be calculated from the serialized request
			// For now, we'll use a nominal value or skip
			// TODO_TECHDEBT: Add payload length tracking to JSON-RPC requests if needed
		}
	}

	return totalLength
}

// GetRequestStatus interprets the observations to determine request status information.
// Returns: (httpStatusCode, requestError, err)
// - httpStatusCode: the suggested HTTP status code to return to the client
// - requestError: error details (nil if successful)  
// - err: error if interpreter cannot determine status (e.g., nil observations)
func (i *CosmosSDKObservationInterpreter) GetRequestStatus() (int32, *RequestError, error) {
	if i.Observations == nil {
		return 0, nil, errors.New("no observations available")
	}

	// Check for request-level error first
	if i.Observations.RequestLevelError != nil {
		return i.Observations.RequestLevelError.HttpStatusCode, i.Observations.RequestLevelError, nil
	}

	// If no request-level error, check endpoint observations for response errors
	httpStatusCode := i.GetRequestHTTPStatus()
	
	// Request is successful if no request-level error and HTTP status indicates success
	if httpStatusCode >= 200 && httpStatusCode < 300 {
		return httpStatusCode, nil, nil
	}

	// For non-success status codes, create a generic error
	// Note: Unlike EVM, Cosmos doesn't have specific error categorization yet
	// TODO_TECHDEBT: Add specific Cosmos error categorization similar to EVM
	return httpStatusCode, nil, nil
}
