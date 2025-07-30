package cosmos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// GET Cosmos SDK status including node info and block height.
// Reference: https://docs.cosmos.network/api#tag/Service/operation/Status
const apiPathCosmosStatus = "/cosmos/base/node/v1beta1/status"

// TODO_IMPROVE(@commoddity): Consider adding check interval and expiry like CometBFT checks
// if periodic validation becomes necessary for Cosmos SDK status checks.

var (
	errNoCosmosStatusObs      = fmt.Errorf("endpoint has not had an observation of its response to a Cosmos SDK '%q' request", apiPathCosmosStatus)
	errInvalidCosmosStatusObs = fmt.Errorf("endpoint returned an invalid response to a Cosmos SDK '%q' request", apiPathCosmosStatus)
)

// endpointCheckCosmosStatus is a check that ensures the endpoint's Cosmos SDK status information is valid.
// It is used to verify the endpoint's current block height.
//
// Note: Unlike CometBFT checks which have expiry, this check does not expire
// as it's only used for basic height validation.
type endpointCheckCosmosStatus struct {
	// latestBlockHeight stores the latest block height from the endpoint's response to a `/cosmos/base/node/v1beta1/status` request.
	// It is nil if there has NOT been an observation of the endpoint's response to a `/cosmos/base/node/v1beta1/status` request.
	latestBlockHeight *uint64
}

// getRequest returns an HTTP request to check the Cosmos SDK status.
// e.g. GET /cosmos/base/node/v1beta1/status
//
// It is called in `request_validator_checks.go` to generate the endpoint checks.
func (e *endpointCheckCosmosStatus) getRequest() *http.Request {
	// Create URL with just the path component.
	// No need to check the errors here as the `apiPathCosmosStatus`
	// is a constant and guaranteed to be valid.
	url, _ := url.Parse(apiPathCosmosStatus)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "", nil)
	req.URL = url // Set the URL to the parsed path.
	return req
}

// GetHeight returns the parsed height value for the endpoint.
func (e *endpointCheckCosmosStatus) GetHeight() (uint64, error) {
	if e.latestBlockHeight == nil {
		return 0, errNoCosmosStatusObs
	}
	if *e.latestBlockHeight == 0 {
		return 0, errInvalidCosmosStatusObs
	}
	return *e.latestBlockHeight, nil
}
