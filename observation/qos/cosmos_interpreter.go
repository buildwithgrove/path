package qos

import (
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
func (i *CosmosSDKObservationInterpreter) GetCosmosSdkChainID() string {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get Cosmos SDK chain ID: nil observations")
		return ""
	}
	return i.Observations.CosmosSdkChainId
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

// GetRequestMethod returns the CosmosSDK RPC method name from the request profile.
func (i *CosmosSDKObservationInterpreter) GetRequestMethod() string {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get request method: nil observations")
		return ""
	}

	// Check if request profile is available
	if i.Observations.RequestProfile == nil {
		return ""
	}

	// Handle different request types
	switch req := i.Observations.RequestProfile.ParsedRequest.(type) {
	case *CosmosRequestProfile_RestRequest:
		if req.RestRequest != nil {
			// For REST requests, use the API path as the method
			return req.RestRequest.ApiPath
		}
	case *CosmosRequestProfile_JsonrpcRequest:
		if req.JsonrpcRequest != nil {
			// For JSON-RPC requests, use the method name
			return req.JsonrpcRequest.Method
		}
	}

	return ""
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

	// Check if request profile and backend service details are available
	if i.Observations.RequestProfile == nil || i.Observations.RequestProfile.BackendServiceDetails == nil {
		return ""
	}

	return i.Observations.RequestProfile.BackendServiceDetails.BackendServiceType.String()
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
