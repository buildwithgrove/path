package evm

import (
	"github.com/buildwithgrove/path/observation/qos"
)

// getHTTPStatusCodeFromObservations extracts the HTTP status code that would be returned to the user.
// Returns the appropriate status code by checking:
// 1. First check request validation failures (takes precedence)
// 2. Then check endpoint response observations
// 3. Return 0 if no status code can be determined
func getHTTPStatusCodeFromObservations(observations *qos.EVMRequestObservations) int {
	if observations == nil {
		return 0
	}

	// Use request interface to check for validation failures
	req := newRequestAdapter(observations)
	if httpStatusCode := req.GetHTTPStatusCode(); httpStatusCode > 0 {
		return httpStatusCode
	}

	endpointObservations := observations.GetEndpointObservations()

	// No status code could be determined
	if len(endpointObservations) == 0 {
		return 0
	}

	// Check endpoint observations
	// The last endpoint observation's status code is what's returned to the user
	lastObs := endpointObservations[len(endpointObservations)-1]

	// Use the response interface to get the HTTP status code
	if resp := extractEndpointResponseFromObservation(lastObs); resp != nil {
		return resp.GetHTTPStatusCode()
	}

	// No status code could be determined
	return 0
}
