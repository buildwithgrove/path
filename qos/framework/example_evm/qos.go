package evm

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	framework "github.com/buildwithgrove/path/qos/framework/jsonrpc"
)

// NewQoSInstance builds and returns an instance of the EVM QoS service.
func NewQoSInstance(logger polylog.Logger, evmChainID string) *QoS {
	logger = logger.With(
		"qos_instance", "evm",
		"evm_chain_id", evmChainID,
	)

	// Setup the QoS definitions for EVM blockchains QoS service
	qosDefinition := framework.QoSDefinition {
		Logger: logger,

		// ServiceInfo for the EVM blockchain QoS service
		ServiceInfo: framework.ServiceInfo{
			Name: "EVM-QoS",
			Description: "QoS service for EVM blockchains, built using PATH's QoS framework",
		},

		// ResultBuilders for JSONRPC request methods used for endpoint attributes.
		ResultBuilders: getJSONRPCMethodEndpointResultBuilders(),

		// StateUpdater uses the endpoint attributes to update the service state.
		StateUpdater: 

		// custom endpoint selection logic
		EndpointSelector: 

		// TODO_FUTURE(@adshmh): implement and supply a custom request validator to control the set of allowed JSONRPC methods.
		//
		// Use the framework's default request validator: i.e. accept any valid JSONRPC request.
		RequestValidator: nil, 
	}

	return framework.NewQoSService(qosDefinition)
}

// Return the set of endpoint result builders to be called by the framework.
// A result builder will be called if it matches the method of the JSONRPC request from the client.
func getJSONRPCMethodEndpointResultBuilders() map[jsonrpc.Method]EndpointResultBuilder {
	// EVM QoS service collects endpoint attributes based on responses to the following methods.
	return map[jsonrpc.Method]EndpointResultBuilder {
		jsonrpc.Method(methodETHChainID):  endpointResultBuilderChainID,
		jsonrpc.Method(methodETHBlockNumber): endpointResultBuilderBlockNumber,
		// TODO_IN_THIS_PR: add eth_getBalance 
	}
}
