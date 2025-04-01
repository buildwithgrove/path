package evm

import (
	"errors"
)

var _ evmQualityCheck = &endpointCheckEmptyResponse{}

var errHasReturnedEmptyResponse = errors.New("endpoint is invalid: history of empty responses")

// endpointCheckEmptyResponse is a check that ensures the endpoint has not returned an empty response.
// It is used to ensure that the endpoint is not returning empty responses.
type endpointCheckEmptyResponse struct {
	// hasReturnedEmptyResponse stores whether the endpoint has returned an empty response.
	hasReturnedEmptyResponse bool
}

// isValid returns an error if the endpoint has returned an empty response.
func (e *endpointCheckEmptyResponse) isValid(serviceState *ServiceState) error {
	if e.hasReturnedEmptyResponse {
		return errHasReturnedEmptyResponse
	}
	return nil
}

// shouldRun will always return false as returning an empty response disqualifies an endpoint
// for the entire session and there is no requestContext to set for this check.
func (e *endpointCheckEmptyResponse) shouldRun() bool {
	return false
}

// setRequestContext is a no-op for this check as it does not require a request context.
func (e *endpointCheckEmptyResponse) setRequestContext(_ *requestContext) {}
