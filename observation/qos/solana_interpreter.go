package qos

import (
	"github.com/pokt-network/poktroll/pkg/polylog"
)

// SolanaObservationInterpreter extracts and interprets data from Solana request observations.
// It provides methods to access metrics-relevant information for Prometheus reporting.
type SolanaObservationInterpreter struct {
	// Logger for reporting issues during interpretation
	Logger polylog.Logger

	// Observations contains the raw Solana request data
	Observations *SolanaRequestObservations
}

// GetChainID returns the blockchain identifier from observations.
func (i *SolanaObservationInterpreter) GetChainID() string {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get chain ID: nil observations")
		return ""
	}
	return i.Observations.ChainId
}

// GetServiceID returns the service identifier from observations.
func (i *SolanaObservationInterpreter) GetServiceID() string {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get service ID: nil observations")
		return ""
	}
	return i.Observations.ServiceId
}

// GetRequestMethod returns the JSON-RPC method name from the request.
func (i *SolanaObservationInterpreter) GetRequestMethod() string {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get request method: nil observations")
		return ""
	}

	// JsonrpcRequest can be nil in case of internal errors or parsing errors - this is expected
	if i.Observations.JsonrpcRequest == nil {
		return ""
	}

	return i.Observations.JsonrpcRequest.Method
}

// IsRequestSuccessful determines if the request completed without errors.
func (i *SolanaObservationInterpreter) IsRequestSuccessful() bool {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot determine request success: nil observations")
		return false
	}

	// RequestError being nil is normal for successful requests
	return i.Observations.RequestError == nil
}

// GetRequestErrorType returns the error type if request failed or empty string if successful.
func (i *SolanaObservationInterpreter) GetRequestErrorType() string {
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

// GetRequestHTTPStatus returns the HTTP status code from the last endpoint observation.
// Returns 0 if observations are nil or no endpoint observations exist.
func (i *SolanaObservationInterpreter) GetRequestHTTPStatus() int32 {
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get HTTP status: nil observations")
		return 0 // Return 0 to indicate observation issues to metrics
	}

	// If there's a request error, return its HTTP status code
	if i.Observations.RequestError != nil {
		return i.Observations.RequestError.HttpStatusCode
	}

	// Loop through endpoint observations and return the HTTP status code from the last one
	if len(i.Observations.EndpointObservations) > 0 {
		lastObservation := i.Observations.EndpointObservations[len(i.Observations.EndpointObservations)-1]
		return lastObservation.HttpStatusCode
	}

	// No endpoint observations available
	i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cannot get HTTP status: no endpoint observations available")
	return 0
}
