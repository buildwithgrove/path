package shannon

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/health"
	"github.com/buildwithgrove/path/metrics/devtools"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
)

// gateway package's Protocol interface is fulfilled by the Protocol struct
// below using methods that are specific to Shannon.
var _ gateway.Protocol = &Protocol{}

// Shannon protocol implements the health.Check and health.ServiceIDReporter interfaces.
// This allows the protocol to report its health status and the list of service IDs it is configured for.
var (
	_ health.Check             = &Protocol{}
	_ health.ServiceIDReporter = &Protocol{}
)

// devtools.ProtocolDisqualifiedEndpointsReporter is fulfilled by the Protocol struct below.
// This allows the protocol to report its sanctioned endpoints data to the devtools.DisqualifiedEndpointReporter.
var _ devtools.ProtocolDisqualifiedEndpointsReporter = &Protocol{}

// NewProtocol instantiates an instance of the Shannon protocol integration.
func NewProtocol(
	logger polylog.Logger,
	fullNode sdk.FullNode,
	config GatewayConfig,
	// ownedApps is the list of apps owned by the gateway operator running PATH in Centralized gateway mode.
	// If PATH is not running in Centralized mode, this field is nil.
	ownedApps map[sdk.ServiceID][]string,
) (*Protocol, error) {
	shannonLogger := logger.With("protocol", "shannon")

	protocolInstance := &Protocol{
		logger: shannonLogger,

		FullNode: fullNode,

		// TODO_MVP(@adshmh): verify the gateway address and private key are valid, by completing the following:
		// 1. Query onchain data for a gateway with the supplied address.
		// 2. Query onchain data for app(s) matching the derived addresses.
		gatewayAddr:          config.GatewayAddress,
		gatewayPrivateKeyHex: config.GatewayPrivateKeyHex,
		gatewayMode:          config.GatewayMode,
		// tracks sanctioned endpoints
		sanctionedEndpointsStore: newSanctionedEndpointsStore(logger),

		// ownedApps is the list of apps owned by the gateway operator running PATH in Centralized gateway mode.
		// If PATH is not running in Centralized mode, this field is nil.
		ownedApps: ownedApps,
	}

	return protocolInstance, nil
}

// Protocol provides the functionality needed by the gateway package for sending a relay to a specific endpoint.
type Protocol struct {
	logger polylog.Logger

	// TODO_IN_THIS_PR(@commoddity): Create GatewayClient interface which embeds
	// sdk.FullNode into it and is implemented in the Shannon SDK.
	sdk.FullNode

	// gatewayMode is the gateway mode in which the current instance of the Shannon protocol integration operates.
	// See protocol/shannon/gateway_mode.go for more details.
	gatewayMode protocol.GatewayMode

	// gatewayAddr is used by the SDK for selecting onchain applications which have delegated to the gateway.
	// The gateway can only sign relays on behalf of an application if the application has an active delegation to it.
	gatewayAddr string

	// gatewayPrivateKeyHex stores the private key of the gateway running this Shannon Gateway instance.
	// It is used for signing relay request in both Centralized and Delegated Gateway Modes.
	gatewayPrivateKeyHex string

	// ownedApps holds the addresses and staked service IDs of all apps owned by the gateway operator running
	// PATH in Centralized mode. If PATH is not running in Centralized mode, this field is nil.
	ownedApps map[sdk.ServiceID][]string

	// sanctionedEndpointsStore tracks sanctioned endpoints
	sanctionedEndpointsStore *sanctionedEndpointsStore
}

// AvailableEndpoints returns the list available endpoints for a given service ID.
// Takes the HTTP request as an argument for Delegated mode to get permitted apps from the HTTP request's headers.
//
// Implements the gateway.Protocol interface.
func (p *Protocol) AvailableEndpoints(
	ctx context.Context,
	serviceID sdk.ServiceID,
	httpReq *http.Request,
) (protocol.EndpointAddrList, protocolobservations.Observations, error) {
	// hydrate the logger.
	logger := p.logger.With(
		"service", serviceID,
		"method", "AvailableEndpoints",
		"gateway_mode", p.gatewayMode,
	)

	// TODO_TECHDEBT(@adshmh): validate "serviceID" is a valid onchain Shannon service.
	permittedSessions, err := p.getGatewayModePermittedSessions(ctx, serviceID, httpReq)
	if err != nil {
		logger.Error().Err(err).Msg("Relay request will fail: error building the permitted apps list for service.")
		return nil, buildProtocolContextSetupErrorObservation(serviceID, err), err
	}

	logger = logger.With("number_of_permitted_sessions", len(permittedSessions))
	logger.Debug().Msg("fetched the set of permitted apps.")

	// Retrieve a list of all unique endpoints for the given service ID filtered by the list of apps this gateway/application
	// owns and can send relays on behalf of.
	// The final boolean parameter sets whether to filter out sanctioned endpoints.
	endpoints, err := p.getSessionsUniqueEndpoints(ctx, serviceID, permittedSessions, true)
	if err != nil {
		logger.Error().Err(err).Msg(err.Error())
		return nil, buildProtocolContextSetupErrorObservation(serviceID, err), err
	}

	logger = logger.With("number_of_unique_endpoints", len(endpoints))
	logger.Debug().Msg("Successfully fetched the set of available endpoints for the selected apps.")

	// Convert the list of endpoints to a list of endpoint addresses
	endpointAddrs := make(protocol.EndpointAddrList, 0, len(endpoints))
	for endpointAddr := range endpoints {
		endpointAddrs = append(endpointAddrs, endpointAddr)
	}

	return endpointAddrs, buildSuccessfulEndpointLookupObservation(serviceID), nil
}

// BuildRequestContextForEndpoint builds a new request context for a given service ID and endpoint address.
// Takes the HTTP request as an argument for Delegate mode to get permitted apps from the HTTP request's headers.
// Returns:
// - An initialized request context.
// - An observation to use if the context initialization failed.
// - An error if the context initialization failed.
//
// Implements the gateway.Protocol interface.
func (p *Protocol) BuildRequestContextForEndpoint(
	ctx context.Context,
	serviceID sdk.ServiceID,
	selectedEndpointAddr protocol.EndpointAddr,
	httpReq *http.Request,
) (gateway.ProtocolRequestContext, protocolobservations.Observations, error) {
	logger := p.logger.With(
		"method", "BuildRequestContextForEndpoint",
		"service_id", serviceID,
		"endpoint_addr", selectedEndpointAddr,
	)

	permittedSessions, err := p.getGatewayModePermittedSessions(ctx, serviceID, httpReq)
	if err != nil {
		logger.Error().Err(err).Msg("BuildRequestContextForEndpoint: error building the permitted apps list for service. Relay request will fail")
		return nil, buildProtocolContextSetupErrorObservation(serviceID, err), err
	}

	// Retrieve the list of endpoints (i.e. backend service URLs by external operators)
	// that can service RPC requests for the given service ID for the given apps.
	// The final boolean parameter sets whether to filter out sanctioned endpoints.
	endpoints, err := p.getSessionsUniqueEndpoints(ctx, serviceID, permittedSessions, true)
	if err != nil {
		logger.Error().Err(err).Msg(err.Error())
		return nil, buildProtocolContextSetupErrorObservation(serviceID, err), err
	}

	// Select the endpoint that matches the pre-selected address.
	// This ensures QoS checks are performed on the selected endpoint.
	selectedEndpoint, ok := endpoints[selectedEndpointAddr]
	if !ok {
		// Wrap the context setup error.
		// Used to generate the observation.
		err := fmt.Errorf("%w: service %s endpoint %s", errRequestContextSetupInvalidEndpointSelected, serviceID, selectedEndpointAddr)
		logger.Error().Err(err).Msg("Selected endpoint is not available.")
		return nil, buildProtocolContextSetupErrorObservation(serviceID, err), err
	}

	// Retrieve the relay request signer for the current gateway mode.
	permittedSigner, err := p.getGatewayModePermittedRelaySigner(p.gatewayMode)
	if err != nil {
		// Wrap the context setup error.
		// Used to generate the observation.
		err = fmt.Errorf("%w: gateway mode %s: %w", errRequestContextSetupErrSignerSetup, p.gatewayMode, err)
		return nil, buildProtocolContextSetupErrorObservation(serviceID, err), err
	}

	// Return new request context for the pre-selected endpoint
	return &requestContext{
		logger:             p.logger,
		fullNode:           p.FullNode,
		selectedEndpoint:   &selectedEndpoint,
		serviceID:          serviceID,
		relayRequestSigner: permittedSigner,
	}, protocolobservations.Observations{}, nil
}

// ApplyObservations updates protocol instance state based on endpoint observations.
// Examples:
// - Mark endpoints as invalid based on response quality
// - Disqualify endpoints for a time period
//
// Implements gateway.Protocol interface.
func (p *Protocol) ApplyObservations(observations *protocolobservations.Observations) error {
	// Sanity check the input
	if observations == nil || observations.GetShannon() == nil {
		p.logger.Warn().Msg("SHOULD NEVER HAPPEN: ApplyObservations called with nil input or nil Shannon observation list.")
		return nil
	}

	shannonObservations := observations.GetShannon().GetObservations()
	if len(shannonObservations) == 0 {
		p.logger.Warn().Msg("SHOULD NEVER HAPPEN: ApplyObservations called with nil set of Shannon request observations.")
		return nil
	}

	// hand over the observations to the sanctioned endpoints store for adding any applicable sanctions.
	p.sanctionedEndpointsStore.ApplyObservations(shannonObservations)

	return nil

}

// ConfiguredServiceIDs returns the list of all all service IDs for all configured AATs.
// This is used by the hydrator to determine which service IDs to run QoS checks on.
func (p *Protocol) ConfiguredServiceIDs() map[sdk.ServiceID]struct{} {
	// Currently hydrator is only enabled for Centralized gateway mode.
	// TODO_FUTURE(@adshmh): support specifying the app(s) used for sending/signing synthetic relay requests by the hydrator.
	if p.gatewayMode != protocol.GatewayModeCentralized {
		return nil
	}

	configuredServiceIDs := make(map[sdk.ServiceID]struct{})
	for serviceID := range p.ownedApps {
		configuredServiceIDs[serviceID] = struct{}{}
	}

	return configuredServiceIDs
}

// Name satisfies the HealthCheck#Name interface function
func (p *Protocol) Name() string {
	return "pokt-shannon"
}

// IsAlive satisfies the HealthCheck#IsAlive interface function
func (p *Protocol) IsAlive() bool {
	return p.FullNode.IsHealthy()
}

// TODO_FUTURE(@adshmh): If multiple apps (across different sessions) are delegating
// to this gateway, optimize how the endpoints are managed/organized/cached.
//
// getAppsUniqueEndpoints returns a map of all endpoints matching serviceID and passing appFilter.
// If an endpoint matches a serviceID across multiple apps/sessions, only a single entry
// matching one of the apps/sessions is returned.
func (p *Protocol) getSessionsUniqueEndpoints(
	_ context.Context,
	serviceID sdk.ServiceID,
	permittedSessions []sessiontypes.Session,
	filterSanctioned bool, // will be true for calls to getAppsUniqueEndpoints made by service request handling.
) (map[protocol.EndpointAddr]endpoint, error) {
	logger := p.logger.With(
		"method", "getAppsUniqueEndpoints",
		"service", serviceID,
		"num_permitted_sessions", len(permittedSessions),
	)

	endpoints := make(map[protocol.EndpointAddr]endpoint)
	for _, session := range permittedSessions {
		app := session.Application

		// Using a single iteration scope for this logger.
		// Avoids adding all apps in the loop to the logger's fields.
		logger := logger.With("permitted_app_address", app.Address)
		logger.Debug().Msg("processing app.")

		appEndpoints, err := endpointsFromSession(session)
		if err != nil {
			logger.Error().Err(err).Msgf("Internal error: error getting all endpoints for service %s app %s and session: skipping the app.", serviceID, app.Address)
			continue
		}

		qualifiedEndpoints := appEndpoints
		// In calls to getAppsUniqueEndpoints made by service request handling, we filter out sanctioned endpoints.
		if filterSanctioned {
			logger.Debug().Msgf("app %s has %d endpoints before filtering sanctioned endpoints.", app.Address, len(appEndpoints))

			// Filter out any sanctioned endpoints
			filteredEndpoints := p.sanctionedEndpointsStore.FilterSanctionedEndpoints(qualifiedEndpoints)
			// All endpoints are sanctioned: log a warning and skip this app.
			if len(filteredEndpoints) == 0 {
				logger.Error().Msgf("All %d app endpoints are sanctioned on service %s, app %s. Skipping the app.",
					len(appEndpoints), serviceID, app.Address,
				)
				continue
			}
			qualifiedEndpoints = filteredEndpoints

			logger.Debug().Msgf("app %s has %d endpoints after filtering sanctioned endpoints.", app.Address, len(qualifiedEndpoints))
		}

		// Log the number of endpoints before and after filtering
		logger.Info().Msgf("Filtered number of endpoints for app %s from %d to %d.", app.Address, len(appEndpoints), len(qualifiedEndpoints))

		for endpointAddr, endpoint := range qualifiedEndpoints {
			endpoints[endpointAddr] = endpoint
		}

		logger.Info().Msg("Successfully fetched session for application.")
	}

	// Ensure at least one endpoint is available for the requested service.
	if len(endpoints) == 0 {
		// Wrap the context setup error.
		// Used for generating observations.
		err := fmt.Errorf("%w: service %s", errProtocolContextSetupNoEndpoints, serviceID)
		logger.Warn().Err(err).Msg("No endpoints available after filtering sanctioned endpoints: relay request will fail.")
		return nil, err
	}

	logger.With("num_endpoints", len(endpoints)).Debug().Msg("Successfully fetched endpoints for permitted apps.")

	return endpoints, nil
}

// GetTotalProtocolEndpointsCount returns the count of all unique endpoints for a service ID
// without filtering sanctioned endpoints.
func (p *Protocol) GetTotalServiceEndpointsCount(serviceID sdk.ServiceID, httpReq *http.Request) (int, error) {
	ctx := context.Background()

	// Get the list of permitted sessions for the service ID.
	permittedSessions, err := p.getGatewayModePermittedSessions(ctx, serviceID, httpReq)
	if err != nil {
		return 0, err
	}

	// Get all endpoints for the service ID without filtering sanctioned endpoints.
	endpoints, err := p.getSessionsUniqueEndpoints(ctx, serviceID, permittedSessions, false)
	if err != nil {
		return 0, err
	}

	return len(endpoints), nil
}

// HydrateDisqualifiedEndpointsResponse hydrates the disqualified endpoint response with the protocol-specific data.
//   - takes a pointer to the DisqualifiedEndpointResponse
//   - called by the devtools.DisqualifiedEndpointReporter to fill it with the protocol-specific data.
func (p *Protocol) HydrateDisqualifiedEndpointsResponse(serviceID sdk.ServiceID, details *devtools.DisqualifiedEndpointResponse) {
	p.logger.Info().Msgf("hydrating disqualified endpoints response for service ID: %s", serviceID)
	details.ProtocolLevelDisqualifiedEndpoints = p.sanctionedEndpointsStore.getSanctionDetails(serviceID)
}
