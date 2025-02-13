package main

import (
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/config"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/cometbft"
	"github.com/buildwithgrove/path/qos/evm"
	"github.com/buildwithgrove/path/qos/solana"
)

// getServiceQoSInstances returns all QoS instances to be used by the Gateway and the EndpointHydrator.
func getServiceQoSInstances(logger polylog.Logger) (map[protocol.ServiceID]gateway.QoSService, error) {
	// TODO_TECHDEBT(@adshmh): refactor this function to remove the
	// need to manually add entries for every new QoS implementation.
	qosServices := make(map[protocol.ServiceID]gateway.QoSService)

	logger = logger.With("module", "qos")

	// Initialize QoS services for all service IDs with a corresponding QoS
	// implementation, as defined in the `config/service_qos.go` file.
	for serviceID, serviceQoSType := range config.ServiceQoSTypes {
		switch serviceQoSType {

		case config.ServiceIDEVM:
			evmQoS := evm.NewQoSInstance(logger, config.GetEVMChainID(serviceID))
			qosServices[serviceID] = evmQoS

		case config.ServiceIDSolana:
			solanaQoS := solana.NewQoSInstance(logger)
			qosServices[serviceID] = solanaQoS

		case config.ServiceIDPOKT:
			// TODO_TECHDEBT: add pokt qos service here

		case config.ServiceIDCometBFT:
			cometBFTQoS := cometbft.NewQoSInstance(logger)
			qosServices[serviceID] = cometBFTQoS

		default: // this should never happen
			return nil, fmt.Errorf("error building QoS instances: service ID %q not supported by PATH", serviceID)
		}
	}

	return qosServices, nil
}
