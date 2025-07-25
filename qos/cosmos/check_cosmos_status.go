package cosmos

import (
	"context"
	"fmt"
	"net/http"
)

// GET Cosmos SDK status including node info and block height.
// Reference: https://docs.cosmos.network/main/core/grpc_rest.html#status
const apiPathCosmosStatus = "/cosmos/base/node/v1beta1/status"

var (
	errNoCosmosStatusObs      = fmt.Errorf("endpoint has not had an observation of its response to a %q request", apiPathCosmosStatus)
	errInvalidCosmosStatusObs = fmt.Errorf("endpoint returned an invalid response to a %q request", apiPathCosmosStatus)
)

// endpointCheckCosmosStatus is a check that ensures the endpoint's Cosmos SDK status information is valid.
// It is used to verify the endpoint's current block height.
type endpointCheckCosmosStatus struct {
	// height stores the latest block height from the endpoint's response to a `/cosmos/base/node/v1beta1/status` request.
	// It is nil if there has NOT been an observation of the endpoint's response to a `/cosmos/base/node/v1beta1/status` request.
	latestBlockHeight *uint64
}

// GetRequest returns an HTTP request to check the Cosmos SDK status.
// e.g. GET /cosmos/base/node/v1beta1/status
// Uses a placeholder host that will be replaced with the actual endpoint URL later.
func (e *endpointCheckCosmosStatus) getRequest() *http.Request {
	// Use a placeholder URL to allow parsing of the URL.
	// Only the path component of the URL is used.
	// The actual URL is set when the endpoint is selected,
	// then the path is appended to the URL.
	fullURL := "http://placeholder.endpoint" + apiPathCosmosStatus
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, fullURL, nil)
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
