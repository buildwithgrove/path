package evm

import (
	"errors"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/metrics/devtools"
	qosobservations "github.com/buildwithgrove/path/observation/qos"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

var (
	errNilApplyObservations    = errors.New("ApplyObservations: received nil")
	errNilApplyEVMObservations = errors.New("ApplyObservations: received nil EVM observation")
)

// The serviceState struct maintains the expected current state of the EVM blockchain
// based on the endpoints' responses to different requests.
//
// It is responsible for the following:
//  1. Generate QoS endpoint checks for the hydrator
//  2. Select a valid endpoint for a service request
//  3. Update the stored endpoints from observations.
//  4. Update the stored service state from observations.
type serviceState struct {
	logger polylog.Logger

	// serviceStateLock is a read-write mutex used to synchronize access to this struct
	serviceStateLock sync.RWMutex

	// serviceQoSConfig maintains the QoS configs for this service
	serviceQoSConfig EVMServiceQoSConfig

	// endpointStore maintains the set of available endpoints and their quality data
	endpointStore *endpointStore

	// perceivedBlockNumber is the perceived current block number
	// based on endpoints' responses to `eth_blockNumber` requests.
	// It is calculated as the maximum of block height reported by
	// any of the endpoints for the service.
	//
	// See the following link for more details:
	// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_blocknumber
	perceivedBlockNumber uint64

	// archivalState contains the current state of the EVM archival check for the service.
	archivalState archivalState
}

/* -------------------- QoS Endpoint Check Generator -------------------- */

// serviceState provides the endpoint check generator required by
// the gateway package to augment endpoints' quality data,
// using synthetic service requests.
var _ gateway.QoSEndpointCheckGenerator = &serviceState{}

// GetRequiredQualityChecks returns the list of quality checks required for an endpoint.
// It is called in the `gateway/hydrator.go` file on each run of the hydrator.
func (ss *serviceState) GetRequiredQualityChecks(endpointAddr protocol.EndpointAddr) []gateway.RequestQoSContext {
	ss.endpointStore.endpointsMu.RLock()
	defer ss.endpointStore.endpointsMu.RUnlock()

	endpoint := ss.endpointStore.endpoints[endpointAddr]

	var checks = []gateway.RequestQoSContext{
		// Block number check should always run
		ss.getEndpointCheck(endpoint.checkBlockNumber.getRequest()),
	}

	// Chain ID check runs infrequently as an endpoint's EVM chain ID is very unlikely to change regularly.
	if ss.shouldChainIDCheckRun(endpoint.checkChainID) {
		checks = append(checks, ss.getEndpointCheck(endpoint.checkChainID.getRequest()))
	}

	// Archival check runs infrequently as the result of a request for an archival block is not expected to change regularly.
	// Additionally, this check will only run if the service is configured to perform archival checks.
	if ss.archivalState.shouldArchivalCheckRun(endpoint.checkArchival) {
		checks = append(
			checks,
			ss.getEndpointCheck(endpoint.checkArchival.getRequest(ss.archivalState)),
		)
	}

	return checks
}

// getEndpointCheck prepares a request context for a specific endpoint check.
// The pre-selected endpoint address is assigned to the request context in the `endpoint.getChecks` method.
// It is called in the individual `check_*.go` files to build the request context.
// getEndpointCheck prepares a request context for a specific endpoint check.
func (ss *serviceState) getEndpointCheck(jsonrpcReq jsonrpc.Request) *requestContext {
	return &requestContext{
		logger:       ss.logger,
		serviceState: ss,
		jsonrpcReq:   jsonrpcReq,
		// Set the chain and Service ID: this is required to generate observations with the correct chain ID.
		chainID:   ss.serviceQoSConfig.getEVMChainID(),
		serviceID: ss.serviceQoSConfig.GetServiceID(),
		// Set the origin of the request as Synthetic.
		// The request is generated by the QoS service to collect extra observations on endpoints.
		requestOrigin: qosobservations.RequestOrigin_REQUEST_ORIGIN_SYNTHETIC,
	}
}

// shouldChainIDCheckRun returns true if the chain ID check is not yet initialized or has expired.
func (ss *serviceState) shouldChainIDCheckRun(check endpointCheckChainID) bool {
	return check.expiresAt.IsZero() || check.expiresAt.Before(time.Now())
}

/* -------------------- QoS Endpoint State Updater -------------------- */

// ApplyObservations updates endpoint storage and blockchain state from observations.
func (ss *serviceState) ApplyObservations(observations *qosobservations.Observations) error {
	if observations == nil {
		return errNilApplyObservations
	}

	evmObservations := observations.GetEvm()
	if evmObservations == nil {
		return errNilApplyEVMObservations
	}

	updatedEndpoints := ss.endpointStore.updateEndpointsFromObservations(
		evmObservations,
		ss.archivalState.blockNumberHex,
	)

	return ss.updateFromEndpoints(updatedEndpoints)
}

// updateFromEndpoints updates the service state based on new observations from endpoints.
// - Only endpoints with received observations are considered.
// - Estimations are derived from these updated endpoints.
func (ss *serviceState) updateFromEndpoints(updatedEndpoints map[protocol.EndpointAddr]endpoint) error {
	ss.serviceStateLock.Lock()
	defer ss.serviceStateLock.Unlock()

	for endpointAddr, endpoint := range updatedEndpoints {
		logger := ss.logger.With(
			"endpoint_addr", endpointAddr,
			"perceived_block_number", ss.perceivedBlockNumber,
		)

		// Do not update the perceived block number if the chain ID is invalid.
		if err := ss.isChainIDValid(endpoint.checkChainID); err != nil {
			logger.Error().Err(err).Msgf("❌ Skipping endpoint because it has an invalid chain id: %s", endpointAddr)
			continue
		}

		// Retrieve the block number from the endpoint.
		blockNumber, err := endpoint.checkBlockNumber.getBlockNumber()
		if err != nil {
			logger.Error().Err(err).Msgf("❌ Skipping endpoint because it has an invalid block number: %s", endpointAddr)
			continue
		}

		// Update perceived block number to maximum instead of overwriting with last endpoint.
		// Per perceivedBlockNumber field documentation, it should be "the maximum of block height reported by any endpoint"
		// but code was incorrectly overwriting with each endpoint, causing validation failures.
		if blockNumber > ss.perceivedBlockNumber {
			logger.Debug().Msgf("Updating perceived block number from %d to %d", ss.perceivedBlockNumber, blockNumber)
			ss.perceivedBlockNumber = blockNumber
		}
	}

	// If archival checks are enabled for the service, update the archival state.
	if ss.archivalState.isEnabled() {
		// Update the archival state based on the perceived block number.
		// When the expected balance at the archival block number is known, this becomes a no-op.
		ss.archivalState.updateArchivalState(ss.perceivedBlockNumber, updatedEndpoints)
	}

	return nil
}

// getDisqualifiedEndpointsResponse gets the QoSLevelDisqualifiedEndpoints map for a devtools.DisqualifiedEndpointResponse.
// It checks the current service state and populates a map with QoS-level disqualified endpoints.
// This data is useful for creating a snapshot of the current QoS state for a given service.
func (ss *serviceState) getDisqualifiedEndpointsResponse(serviceID protocol.ServiceID) devtools.QoSLevelDataResponse {
	qosLevelDataResponse := devtools.QoSLevelDataResponse{
		DisqualifiedEndpoints: make(map[protocol.EndpointAddr]devtools.QoSDisqualifiedEndpoint),
	}

	// Populate the data response object using the endpoints in the endpoint store.
	for endpointAddr, endpoint := range ss.endpointStore.endpoints {
		if err := ss.basicEndpointValidation(endpoint); err != nil {
			qosLevelDataResponse.DisqualifiedEndpoints[endpointAddr] = devtools.QoSDisqualifiedEndpoint{
				EndpointAddr: endpointAddr,
				Reason:       err.Error(),
				ServiceID:    serviceID,
			}

			// DEV_NOTE: if new checks are added to a service, we need to add them here.
			switch {
			// Endpoint is disqualified due to an empty qosLevelDataResponse.
			case errors.Is(err, errEmptyResponseObs):
				qosLevelDataResponse.EmptyResponseCount++

			// Endpoint is disqualified due to a missing or invalid block number.
			case errors.Is(err, errNoBlockNumberObs),
				errors.Is(err, errInvalidBlockNumberObs):
				qosLevelDataResponse.BlockNumberCheckErrorsCount++

			// Endpoint is disqualified due to a missing or invalid chain ID.
			case errors.Is(err, errNoChainIDObs),
				errors.Is(err, errInvalidChainIDObs):
				qosLevelDataResponse.ChainIDCheckErrorsCount++

			// Endpoint is disqualified due to a missing or invalid archival balance.
			case errors.Is(err, errNoArchivalBalanceObs),
				errors.Is(err, errInvalidArchivalBalanceObs):
				qosLevelDataResponse.ArchivalCheckErrorsCount++

			default:
				ss.logger.Error().Err(err).Msgf("SHOULD NEVER HAPPEN: unknown error for endpoint: %s", endpointAddr)
			}

			continue
		}
	}

	return qosLevelDataResponse
}
