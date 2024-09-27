package evm

import (
	"github.com/buildwithgrove/path/gateway"
)

const (
	// TODO_IN_THIS_COMMIT: provide a basic JSONRPC package to meet
	// the current needs of the evm package: e.g. unmarshalling a
	// JSONRPC ID.
	idChainIDCheck     = jsonrpc.IDFromNumber(1001)
	idBlockNumberCheck = jsonrpc.IDFromNumber(1002)
)

// EndpointStore provides the endpoint check generator required by
// the gateway package to augment endpoints' quality data,
// using synthetic service requests.
var _ gateway.QoSEndpointCheckGenerator = &EndpointStore{}

func (es *EndpointStore) GetRequiredQualityChecks(endpointAddr relayer.EndpointAddr) []gateway.ServiceRequestContext {
	// TODO_IMPROVE: skip any checks for which the endpoint already has
	// a valid (e.g. not expired) quality data point.
	return []gateway.ServiceRequestContext{
		getChainIDCheck(es.chainID),
		getBlockHeightCheck(),
		// TODO_FUTURE: add an archival endpoint check.
	}
}

func getChainIDCheck(chainID string) serviceRequestContext {
	return serviceRequestContext{
		method:  methodChainID,
		id:      idChainIDCheck,
		isValid: true,
	}
}

func getBlockHeightCheck() serviceRequestContext {
	return serviceRequestContext{
		method:  methodBlockNumber,
		id:      idBlockNumberCheck,
		isValid: true,
	}
}
