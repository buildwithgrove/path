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
	Observations *CosmosSDKRequestObservations
}

// GetChainID returns the blockchain identifier from observations.
func (i *CosmosSDKObservationInterpreter) GetChainID() string {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get chain ID: nil observations")
		return ""
	}
	return i.Observations.ChainId
}

// GetServiceID returns the service identifier from observations.
func (i *CosmosSDKObservationInterpreter) GetServiceID() string {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get service ID: nil observations")
		return ""
	}
	return i.Observations.ServiceId
}

// GetRequestMethod returns the CosmosSDK RPC method name from the route request.
func (i *CosmosSDKObservationInterpreter) GetRequestMethod() string {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get request method: nil observations")
		return ""
	}

	// RouteRequest can be empty in case of internal errors or parsing errors - this is expected
	if i.Observations.RouteRequest == "" {
		return ""
	}

	// TODO_IMPROVE: Parse the route_request to extract the actual method name
	// For now, return the full route request as the method
	// Example: "/health" -> "health", "/status" -> "status"
	return i.Observations.RouteRequest
}

// IsRequestSuccessful determines if the request completed without errors.
func (i *CosmosSDKObservationInterpreter) IsRequestSuccessful() bool {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot determine request success: nil observations")
		return false
	}

	// RequestError being nil is normal for successful requests
	return i.Observations.RequestError == nil
}

// GetRequestErrorType returns the error type if request failed or empty string if successful.
func (i *CosmosSDKObservationInterpreter) GetRequestErrorType() string {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get error type: nil observations")
		return ""
	}

	// RequestError being nil is normal for successful requests
	if i.Observations.RequestError == nil {
		return ""
	}

	return i.Observations.RequestError.ErrorKind.String()
}

// GetRequestHTTPStatus returns the HTTP status code from the request error.
// Returns 200 if request was successful, 0 if observations are nil.
func (i *CosmosSDKObservationInterpreter) GetRequestHTTPStatus() int32 {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get HTTP status: nil observations")
		return 0 // Return 0 to indicate observation issues to metrics
	}

	// RequestError being nil is normal for successful requests
	if i.Observations.RequestError == nil {
		return 200 // OK status for successful requests
	}

	return i.Observations.RequestError.HttpStatusCode
}
