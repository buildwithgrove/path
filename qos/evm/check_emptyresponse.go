package evm

import (
	"errors"
	"time"
)

const (
	endpointCheckNameEmptyResponse endpointCheckName = "empty_response"
	// TODO_IMPROVE: determine an appropriate interval for checking the empty response.
	emptyResponseCheckInterval = 60 * time.Minute
)

var (
	errHasReturnedEmptyResponse = errors.New("endpoint is invalid: history of empty responses")
)

// endpointCheckEmptyResponse is a check that ensures the endpoint has not returned an empty response.
// It is used to ensure that the endpoint is not returning empty responses.
type endpointCheckEmptyResponse struct {
	hasReturnedEmptyResponse bool
	expiresAt                time.Time
}

func (e *endpointCheckEmptyResponse) CheckName() string {
	return string(endpointCheckNameEmptyResponse)
}

func (e *endpointCheckEmptyResponse) IsValid(serviceState *ServiceState) error {
	if e.hasReturnedEmptyResponse {
		return errHasReturnedEmptyResponse
	}
	return nil
}

func (e *endpointCheckEmptyResponse) ExpiresAt() time.Time {
	return e.expiresAt
}
