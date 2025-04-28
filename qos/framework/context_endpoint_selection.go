package framework

import (
	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_FUTURE: consider ranking filtered endpoints, e.g. based on latency, rather than randomization.

// TODO_MVP(@adshmh): Remove expired Sanctions from endpoints' results.
//

// EndpointSelectionContext provides context for selecting endpoints.
type EndpointSelectionContext struct {
	// Allows direct Get calls on the current service state.
	// It is read only: this is not the context for updating service state.
	*ServiceState

	// Supplied from the Custom QoS service definition.
	// Used to select an endpoint among those without an active sanction.
	// e.g. EVM QoS implements a selector that disqualifies endpoints that are out of sync.
	customSelector EndpointSelector

	// Used to retrieve stored endpoints to examine their attributes during selection.
	endpointStore *endpointStore

	// TODO_TECHDEBT(@adshmh): make this readonly and allow access through a getter method to prevent accidental modification by user.
	// The JSONRPC request for which an endpoint is to be selected.
	Request *jsonrpc.Request

	// Endpoints disqualified from current selection context.
	disqualifiedEndpoints map[protocol.EndpointAddr]struct{}
}

// Entry method into the endpoint selection context.
// Called by gateway.requestContext.
// Implements protocol.EndpointSelector interface.
func (ctx *EndpointSelectionContext) Select(availableEndpoints []protocol.Endpoint) (protocol.EndpointAddr, error) {
	// Retrieve all the available endpoints from the endpoint store.
	candidates := make(map[protocol.EndpointAddr]Endpoint)
	for _, availableEndpoint := range availableEndpoints {
		endpointAddr := availableEndpoint.Addr()

		// Use an empty endpoint struct for initialization if no entry was found in the store.
		candidates[endpointAddr] = ctx.endpointStore.GetEndpoint(endpointAddr)
	}

	// Drop endpoints with active sanctions from the list.
	filteredEndpoints := ctx.filterSanctionedEndpoints(candidates)

	// If all endpoints were sanctioned, return early
	if len(unsanctionedEndpoints) == 0 {
		// TODO_IN_THIS_PR: should we error out here instead?
		s.logger.Info().Msg("All endpoints are currently sanctioned: returning a random endpoint.")
		return ctx.selectRandomEndpoint(availableEndpoints)
	}

	// TODO_IN_THIS_PR: implement this in the EndpointSelectionContext struct.
	// If no custom selector is provided, use a random selector.
	if s.customSelector == nil {
		return ctx.SelectRandom()
	}

	// Call the custom selector with the context
	return s.customSelector(ctx)
}

// Select marks an endpoint as selected.
func (ctx *EndpointSelectionContext) SelectEndpoint(endpoint Endpoint) *EndpointSelectionContext {
	ctx.selected = append(ctx.selected, endpoint)
	return ctx
}

// SelectAll marks all available endpoints as selected.
func (ctx *EndpointSelectionContext) SelectAll() *EndpointSelectionContext {
	ctx.selected = append(ctx.selected[:0], ctx.Endpoints...)
	return ctx
}

// SelectIf selects endpoints that match a predicate function.
func (ctx *EndpointSelectionContext) SelectIf(predicate func(Endpoint) bool) *EndpointSelectionContext {
	for _, endpoint := range ctx.Endpoints {
		if predicate(endpoint) {
			ctx.selected = append(ctx.selected, endpoint)
		}
	}
	return ctx
}

// Selected returns the selected endpoints.
func (ctx *EndpointSelectionContext) Selected() []Endpoint {
	return ctx.selected
}

// filterSanctionedEndpoints removes endpoints with active sanctions
// This is an internal function used by the framework before passing
// endpoints to the custom service
func (ctx *EndpointSelectionContext) filterSanctionedEndpoints(endpoints map[protocol.EndpointAddr]Endpoint) []Endpoint {
	filteredEndpoints := make([]Endpoint, 0, len(endpoints))

	for endpointAddr, endpoint := range endpoints {
		// Check if the endpoint is sanctioned.
		activeSanction, isSanctioned := endpoint.GetActiveSanction()

		// Skip sanctioned endpoints.
		if isSanctioned {
			ctx.logger.With(
				"endpoint_addr", string(endpointAddr),
				"sanction", activeSanction.String(),
			).Debug().Msg("Dropping sanctioned endpoint")

			continue
		}

		filteredEndpoints = append(filteredEndpoints, endpoint)
	}

	return filteredEndpoints
}
