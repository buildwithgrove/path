// Package qos handles exporting of all qos-related metrics.
package qos

import (
	"github.com/pokt-network/poktroll/pkg/polylog"

	metricshttp "github.com/buildwithgrove/path/metrics/http"
	"github.com/buildwithgrove/path/metrics/qos/cosmos"
	"github.com/buildwithgrove/path/metrics/qos/evm"
	"github.com/buildwithgrove/path/metrics/qos/solana"
	"github.com/buildwithgrove/path/observation/qos"
)

// PublishMetrics builds and exports all qos-related metrics using qos-level observations.
func PublishQoSMetrics(
	logger polylog.Logger,
	qosObservations *qos.Observations,
) {
	hydratedLogger := logger.With("method", "PublishQoSMetrics")

	if qosObservations == nil {
		hydratedLogger.Warn().Msg("received nil set of QoS observations.")
		return
	}

	// Publish EVM metrics.
	if evmObservations := qosObservations.GetEvm(); evmObservations != nil {
		endpointDomain := extractEndpointDomainFromEVM(hydratedLogger, evmObservations)
		evm.PublishMetrics(hydratedLogger, evmObservations, endpointDomain)
		hydratedLogger.Debug().Msg("published EVM metrics.")
		return
	}

	// Publish CometBFT metrics.
	if cosmosObservations := qosObservations.GetCosmos(); cosmosObservations != nil {
		endpointDomain := extractEndpointDomainFromCosmos(hydratedLogger, cosmosObservations)
		cosmos.PublishMetrics(hydratedLogger, cosmosObservations, endpointDomain)
		hydratedLogger.Debug().Msg("published Cosmos SDK metrics.")
		return
	}

	// Publish Solana metrics.
	if solanaObservations := qosObservations.GetSolana(); solanaObservations != nil {
		endpointDomain := extractEndpointDomainFromSolana(hydratedLogger, solanaObservations)
		solana.PublishMetrics(hydratedLogger, solanaObservations, endpointDomain)
		hydratedLogger.Debug().Msg("published Solana metrics.")
		return
	}

	// Log warning if no matching observation types were found
	hydratedLogger.Warn().Msgf("SHOULD RARELY HAPPEN: supplied observations do not match any known QoS service: '%+v'", qosObservations)
}

// extractEndpointDomainFromEVM extracts the endpoint domain from EVM observations.
// Returns "unknown" if domain cannot be determined.
func extractEndpointDomainFromEVM(logger polylog.Logger, observations *qos.EVMRequestObservations) string {
	// Create an interpreter for the observations
	interpreter := &qos.EVMObservationInterpreter{
		Logger:       logger,
		Observations: observations,
	}

	// Get endpoint observations and extract domain from the last one used
	endpointObservations, found := interpreter.GetEndpointObservations()
	if !found || len(endpointObservations) == 0 {
		return "unknown"
	}

	// Use the last endpoint observation (most recent endpoint used, similar to Shannon metrics pattern)
	lastObs := endpointObservations[len(endpointObservations)-1]
	return metricshttp.ExtractDomainFromEndpointAddr(logger, lastObs.GetEndpointAddr())
}

// extractEndpointDomainFromCosmos extracts the endpoint domain from Cosmos observations.
// Returns "unknown" if domain cannot be determined.
func extractEndpointDomainFromCosmos(logger polylog.Logger, observations *qos.CosmosRequestObservations) string {
	// Get endpoint observations and extract domain from the last one used
	endpointObservations := observations.GetEndpointObservations()
	if len(endpointObservations) == 0 {
		return "unknown"
	}

	// Use the last endpoint observation (most recent/final endpoint used)
	lastObs := endpointObservations[len(endpointObservations)-1]
	return metricshttp.ExtractDomainFromEndpointAddr(logger, lastObs.GetEndpointAddr())
}

// extractEndpointDomainFromSolana extracts the endpoint domain from Solana observations.
// Returns "unknown" if domain cannot be determined.
func extractEndpointDomainFromSolana(logger polylog.Logger, observations *qos.SolanaRequestObservations) string {
	// Get endpoint observations and extract domain from the last one used
	endpointObservations := observations.GetEndpointObservations()
	if len(endpointObservations) == 0 {
		return "unknown"
	}

	// Use the last endpoint observation (most recent endpoint used, similar to Shannon metrics pattern)
	lastObs := endpointObservations[len(endpointObservations)-1]
	return metricshttp.ExtractDomainFromEndpointAddr(logger, lastObs.GetEndpointAddr())
}
