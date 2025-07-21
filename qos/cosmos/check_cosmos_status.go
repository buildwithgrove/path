package cosmos

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pokt-network/poktroll/pkg/relayer/proxy"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
)

// GET Cosmos SDK status including node info and block height.
// Reference: https://docs.cosmos.network/main/core/grpc_rest.html#status
const apiPathCosmosStatus = "/cosmos/base/node/v1beta1/status"

// TODO_IMPROVE(@commoddity): determine an appropriate interval for checking the status and/or make it configurable.
const checkCosmosStatusInterval = 10 * time.Second

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

	// expiresAt stores the time at which the last check expires.
	expiresAt time.Time
}

// GetRequest returns an HTTP request to check the Cosmos SDK status.
// e.g. GET /cosmos/base/node/v1beta1/status
func (e *endpointCheckCosmosStatus) GetRequest() *http.Request {
	req, _ := http.NewRequest(http.MethodGet, apiPathCosmosStatus, nil)
	req.Header.Set(proxy.RPCTypeHeader, strconv.Itoa(int(sharedtypes.RPCType_REST)))
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

// IsExpired returns true if the check has expired and needs to be refreshed.
func (e *endpointCheckCosmosStatus) IsExpired() bool {
	return time.Now().After(e.expiresAt)
}
