package framework

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_FUTURE(@adshmh): Rank qualified endpoints, e.g. based on latency, for selection.

// TODO_MVP(@adshmh): Remove expired Sanctions from endpoints' results.
//

// EndpointSelectionContext provides context for selecting endpoints.
type EndpointSelectionContext struct {
	logger polylog.Logger

	// request's journal.
	// Provide read-only access to the request, e.g. the JSONRPC method.
	*requestJournal

	// Allows direct Get calls on the current service state.
	// It is read only: this is not the context for updating service state.
	*ServiceState

	// Supplied from the Custom QoS service definition.
	// Used to select an endpoint among those without an active sanction.
	// e.g. EVM QoS implements a selector that disqualifies endpoints that are out of sync.
	customSelector EndpointSelector

	// Endpoints disqualified from current selection context.
	disqualifiedEndpoints map[protocol.EndpointAddr]struct{}

	candidateEndpoints map[protocol.EndpointAddr]*Endpoint
}

func (ctx *EndpointSelectionContext) buildCandidateEndpointSet(availableEndpoints []protocol.Endpoint) {
	// Retrieve all the available endpoints from the endpoint store.
	candidates := make(map[protocol.EndpointAddr]*Endpoint)
	for _, availableEndpoint := range availableEndpoints {
		endpointAddr := availableEndpoint.Addr()

		// Use an empty endpoint struct for initialization if no entry was found in the store.
		candidates[endpointAddr] = ctx.getEndpoint(endpointAddr)
	}

	// Store the candidate endpoints.
	ctx.candidateEndpoints = candidates
}

// Entry method into the endpoint selection context.
// Called by gateway.requestContext.
// Implements protocol.EndpointSelector interface.
func (ctx *EndpointSelectionContext) Select(availableEndpoints []protocol.Endpoint) (protocol.EndpointAddr, error) {
	// No endpoints available: error out.
	if len(availableEndpoints) == 0 {
		errMsg := "No endpoints available for selection. Service request will fail."
		ctx.logger.Warn().Msg(errMsg)
		return protocol.EndpointAddr(""), errors.New(errMsg)
	}

	// Build a map of candidate endpoints for efficient lookup based on endpoint address.
	ctx.buildCandidateEndpointSet(availableEndpoints)

	// Drop endpoints with active sanctions from the list.
	ctx.disqualifySanctionedEndpoints()

	// If all endpoints were sanctioned, return early
	if len(ctx.candidateEndpoints) == len(ctx.disqualifiedEndpoints) {
		// TODO_IN_THIS_PR: should we error out here instead?
		ctx.logger.Info().Msg("All endpoints are currently sanctioned: returning a random endpoint.")
		return ctx.selectRandomEndpoint()
	}

	if ctx.customSelector == nil {
		return protocol.EndpointAddr(""), fmt.Errorf("Endpoint selection failed: custom QoS endpoint selection must be provided.")
	}

	// Call the custom selector.
	// Pass the context to provide helper methods.
	return ctx.customSelector(ctx)
}

type EndpointFilter func(*Endpoint) error

// Returns a random endpoint from the context.
// Only considers endpoints not dropped from the context.
func (ctx *EndpointSelectionContext) SelectRandomQualifiedEndpoint(endpointFilters ...EndpointFilter) (protocol.EndpointAddr, error) {
	for endpointAddr, endpoint := range ctx.candidateEndpoints {
		if _, isDisqualified := ctx.disqualifiedEndpoints[endpointAddr]; isDisqualified {
			// endpoint already disqualified. Skip further processing.
			continue
		}

		// Call the endpoint filters on the endpoint.
		// Endpoint will be disqualified if any filter reutrns an error.
		for _, filter := range endpointFilters {
			err := filter(endpoint)
			if err != nil {
				ctx.logger.With("endpoint_addr", endpointAddr).Debug().Err(err).Msg("endpoint has been disqualified.")
				// Mark the disqualified endpoint
				ctx.disqualifiedEndpoints[endpointAddr] = struct{}{}
			}
		}
	}
	
	// Disqualified endpoints have been marked.
	// return a random qualified endpoint.
	return ctx.selectRandomEndpoint()
}

func (ctx *EndpointSelectionContext) selectRandomEndpoint() (protocol.EndpointAddr, error) {
	// build the slice with addresses of qualified endpoints.
	var allEndpointsAddrs []protocol.EndpointAddr
	for endpointAddr := range ctx.candidateEndpoints {
		// disqualified endpoint: skip.
		if _, isDisqualified := ctx.disqualifiedEndpoints[endpointAddr]; isDisqualified {
			continue
		}

		allEndpointsAddrs = append(allEndpointsAddrs, endpointAddr)
	}

	// all endpoints have been disqualified: log a message.
	if len(allEndpointsAddrs) == 0 {
		ctx.logger.With("num_endpoints", len(ctx.candidateEndpoints)).Warn().Msg("all endpoints have been disqualified: returning a random endpoint.")
		// build the slice for random selection
		for endpointAddr := range ctx.candidateEndpoints {
			allEndpointsAddrs = append(allEndpointsAddrs, endpointAddr)
		}
	}

	// return a random endpoint from the slice.
	return allEndpointsAddrs[rand.Intn(len(allEndpointsAddrs))], nil
}

// disqualifySanctionedEndpoints marks endpoints with active sanctions as disqualified.
func (ctx *EndpointSelectionContext) disqualifySanctionedEndpoints() {
	if ctx.disqualifiedEndpoints == nil {
		ctx.disqualifiedEndpoints = make(map[protocol.EndpointAddr]struct{})
	}

	for endpointAddr, endpoint := range ctx.candidateEndpoints {
		// Check if the endpoint is sanctioned.
		activeSanction, isSanctioned := endpoint.GetActiveSanction()

		// endpoint has no active sanctions: skip further processing.
		if !isSanctioned {
			continue
		}

		ctx.logger.With(
			"endpoint_addr", string(endpointAddr),
			"sanction", activeSanction,
		).Debug().Msg("Dropping sanctioned endpoint")

		ctx.disqualifiedEndpoints[endpointAddr] = struct{}{}
	}
}
