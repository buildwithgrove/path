package evm

import (
	"github.com/buildwithgrove/path/observation/qos"
)

const (
	// TODO_TECHDEBT(@adshmh): Revisit the decision to use 0 as the unknown status code if one cannot be found.
	unknownHTTPStatusCode = 0
)

// getHTTPStatusCodeFromObservations tries to extract an HTTP status code from the observations.
// Returns the found status code by checking (in order of precedence):
// 1. Request validation failures for an HTTP status code
// 2. Then, check endpoint response observations for an HTTP status code
// 3. Return the unknownHTTPStatusCode (0) if no status code can be determined
func getHTTPStatusCodeFromObservations(observations *qos.EVMRequestObservations) int {
	if observations == nil {
		return unknownHTTPStatusCode
	}

	// Use request interface to check for validation failures
	req := newRequestAdapter(observations)
	if httpStatusCode := req.GetHTTPStatusCode(); httpStatusCode > 0 {
		return httpStatusCode
	}

	endpointObservations := observations.GetEndpointObservations()

	// No status code could be determined since there are no endpoint observations
	if len(endpointObservations) == 0 {
		return unknownHTTPStatusCode
	}

	// Check endpoint observations for status code if they are present
	// The status code from the final endpoint observation is returned to the user.
	// Subsequent endpoints are only selected if previous endpoints fail for any reason.
	lastObsIndex := len(endpointObservations) - 1
	lastObs := endpointObservations[lastObsIndex]

	// Use the response interface to get the HTTP status code
	if resp := extractEndpointResponseFromObservation(lastObs); resp != nil {
		return resp.GetHTTPStatusCode()
	}

	// No status code could be determined
	return unknownHTTPStatusCode
}
