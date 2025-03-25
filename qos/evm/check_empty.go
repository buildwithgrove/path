package evm

import (
	"errors"
)

var _ check = &endpointCheckEmptyResponse{}

// TODO_MVP(@commoddity): should we provide for a mechanism to un-sanction an endpoint that has returned an empty response?
// Currently it will be removed for the entire session but perhaps we want a mechanism to un-sanction it.
// For example if it returns a valid response to either of the other checks after a given period has elapsed?

const checkNameEmptyResponse endpointCheckName = "empty_response"

var errHasReturnedEmptyResponse = errors.New("endpoint is invalid: history of empty responses")

// endpointCheckEmptyResponse is a check that ensures the endpoint has not returned an empty response.
// It is used to ensure that the endpoint is not returning empty responses.
type endpointCheckEmptyResponse struct {
	// hasReturnedEmptyResponse stores whether the endpoint has returned an empty response.
	hasReturnedEmptyResponse bool
}

func (e *endpointCheckEmptyResponse) name() endpointCheckName {
	return checkNameEmptyResponse
}

// isValid returns an error if the endpoint has returned an empty response.
func (e *endpointCheckEmptyResponse) isValid(serviceState *ServiceState) error {
	if e.hasReturnedEmptyResponse {
		return errHasReturnedEmptyResponse
	}
	return nil
}

// shouldRun will always return false as returning an empty response disqualifies an endpoint
// for the entire session and there is no requestContext to run for this check.
// TODO_MVP(@commoddity): should we provide for a mechanism to un-sanction an endpoint that has returned an empty response?
func (e *endpointCheckEmptyResponse) shouldRun() bool {
	return false
}
