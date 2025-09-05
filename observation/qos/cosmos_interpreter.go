package qos

import (
	"strings"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
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

// GetEndpointDomain returns the domain of the endpoint that served the request.
//
// If multiple endpoint observations are present, it returns the domain of the first endpoint observation.
// If no endpoint observations are present, it returns an empty string.
//
// TODO_TECHDEBT: Consolidate this with the business logic of other "GetEndpointDomain" implementations.
func (i *CosmosSDKObservationInterpreter) GetEndpointDomain() string {
	// Ensure observations are not nil
	if i.Observations == nil {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cosmos observations are nil")
		return ""
	}

	// Ensure endpoint observations are not empty
	endpointObservations := i.Observations.GetEndpointObservations()
	if len(endpointObservations) == 0 {
		i.Logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: Cosmos endpoint observations are empty")
		return ""
	}

	// Build a set of unique endpoint addresses
	uniqueEndpointAddrs := make(map[string]struct{})
	endpointAddrs := make([]string, 0, len(endpointObservations))
	for _, eo := range endpointObservations {
		endpointAddr := eo.GetEndpointAddr()
		if _, seen := uniqueEndpointAddrs[endpointAddr]; !seen {
			uniqueEndpointAddrs[endpointAddr] = struct{}{}
			endpointAddrs = append(endpointAddrs, endpointAddr)
		}
	}

	// If multiple endpoint addresses are observed, log a warning and use the first one for domain extraction
	// TODO_DISCUSS: Decide how we want to handle this case in the future.
	numUniqueEndpointAddrs := len(uniqueEndpointAddrs)
	if numUniqueEndpointAddrs > 1 {
		i.Logger.With(
			"num_unique_endpoint_addrs", numUniqueEndpointAddrs,
			"unique_endpoint_addrs", strings.Join(endpointAddrs, ", "),
		).Warn().Msg("Multiple endpoint addresses observed for a single request. Using the first one for metrics domain.")
	}

	// Use the first observed endpoint address for domain extraction
	domain, err := protocol.EndpointAddr(endpointAddrs[0]).GetDomain()
	if err != nil {
		i.Logger.Error().Err(err).Msg("SHOULD NEVER HAPPEN: Cannot get endpoint domain: empty endpoint observations")
	}
	return domain
}
