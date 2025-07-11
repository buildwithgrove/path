package shannon

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"

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
	config GatewayConfig,
	fullNode FullNode,
) (*Protocol, error) {
	shannonLogger := logger.With("protocol", "shannon")

	// Initialize the owned apps for the gateway mode. This value will be nil if gateway mode is not Centralized.
	ownedApps, err := getOwnedApps(logger, config.OwnedAppsPrivateKeysHex, fullNode)
	if err != nil {
		return nil, fmt.Errorf("failed to get app addresses from config: %w", err)
	}

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

		// ownedApps is the list of apps owned by the gateway operator
		ownedApps: ownedApps,
	}

	return protocolInstance, nil
}

// Protocol provides the functionality needed by the gateway package for sending a relay to a specific endpoint.
type Protocol struct {
	logger polylog.Logger
	FullNode

	// gatewayMode is the gateway mode in which the current instance of the Shannon protocol integration operates.
	// See protocol/shannon/gateway_mode.go for more details.
	gatewayMode protocol.GatewayMode

	// gatewayAddr is used by the SDK for selecting onchain applications which have delegated to the gateway.
	// The gateway can only sign relays on behalf of an application if the application has an active delegation to it.
	gatewayAddr string

	// gatewayPrivateKeyHex stores the private key of the gateway running this Shannon Gateway instance.
	// It is used for signing relay request in both Centralized and Delegated Gateway Modes.
	gatewayPrivateKeyHex string

	// ownedApps is the list of apps owned by the gateway operator
	ownedApps map[protocol.ServiceID][]string

	// sanctionedEndpointsStore tracks sanctioned endpoints
	sanctionedEndpointsStore *sanctionedEndpointsStore
}

// AvailableEndpoints returns the available endpoints for a given service ID.
//
// - Provides the list of endpoints that can serve the specified service ID.
// - Returns a list of valid endpoint addresses, protocol observations, and any error encountered.
//
// Usage:
//   - In Delegated mode, httpReq must contain the appropriate headers for app selection.
//   - In Centralized mode, httpReq may be nil.
//
// Returns:
//   - protocol.EndpointAddrList: the discovered endpoints for the service.
//   - protocolobservations.Observations: contextual observations (e.g., error context).
//   - error: if any error occurs during endpoint discovery or validation.
func (p *Protocol) AvailableEndpoints(
	ctx context.Context,
	serviceID protocol.ServiceID,
	httpReq *http.Request,
) (protocol.EndpointAddrList, protocolobservations.Observations, error) {
	// hydrate the logger.
	logger := p.logger.With(
		"service", serviceID,
		"method", "AvailableEndpoints",
		"gateway_mode", p.gatewayMode,
	)

	// TODO_TECHDEBT(@adshmh): validate "serviceID" is a valid onchain Shannon service.
	activeSessions, err := p.getActiveGatewaySessions(ctx, serviceID, httpReq)
	if err != nil {
		logger.Error().Err(err).Msg("Relay request will fail: error building the active sessions for service.")
		return nil, buildProtocolContextSetupErrorObservation(serviceID, err), err
	}

	logger = logger.With("number_of_valid_sessions", len(activeSessions))
	logger.Debug().Msg("fetched the set of active sessions.")

	// Retrieve a list of all unique endpoints for the given service ID filtered by the list of apps this gateway/application
	// owns and can send relays on behalf of.
	// The final boolean parameter sets whether to filter out sanctioned endpoints.
	endpoints, err := p.getSessionsUniqueEndpoints(ctx, serviceID, activeSessions, true)
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

// BuildRequestContextForEndpoint creates a new protocol request context for a specified service and endpoint.
//
// Parameters:
//   - ctx: Context for cancellation, deadlines, and logging.
//   - serviceID: The unique identifier of the target service.
//   - selectedEndpointAddr: The address of the endpoint to use for the request.
//   - httpReq: ONLY used in Delegated mode to extract the selected app from headers.
//   - TODO_TECHDEBT: Decouple context building for different gateway modes.
//
// Behavior:
//   - Retrieves active sessions for the given service ID from the full node.
//   - Retrieves unique endpoints available across all active sessions
//   - Filtering out sanctioned endpoints from list of unique endpoints.
//   - Obtains the relay request signer appropriate for the current gateway mode.
//   - Returns a fully initialized request context for use in downstream protocol operations.
//   - On failure, logs the error, returns a context setup observation, and a non-nil error.
//
// Implements the gateway.Protocol interface.
func (p *Protocol) BuildRequestContextForEndpoint(
	ctx context.Context,
	serviceID protocol.ServiceID,
	selectedEndpointAddr protocol.EndpointAddr,
	httpReq *http.Request,
) (gateway.ProtocolRequestContext, protocolobservations.Observations, error) {
	logger := p.logger.With(
		"method", "BuildRequestContextForEndpoint",
		"service_id", serviceID,
		"endpoint_addr", selectedEndpointAddr,
	)

	activeSessions, err := p.getActiveGatewaySessions(ctx, serviceID, httpReq)
	if err != nil {
		logger.Error().Err(err).Msgf("Relay request will fail due to error retrieving active sessions for service %s", serviceID)
		return nil, buildProtocolContextSetupErrorObservation(serviceID, err), err
	}

	// Retrieve the list of endpoints (i.e. backend service URLs by external operators)
	// that can service RPC requests for the given service ID for the given apps.
	// The final boolean parameter sets whether to filter out sanctioned endpoints.
	endpoints, err := p.getSessionsUniqueEndpoints(ctx, serviceID, activeSessions, true)
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
		p.logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: ApplyObservations called with nil input or nil Shannon observation list.")
		return nil
	}

	shannonObservations := observations.GetShannon().GetObservations()
	if len(shannonObservations) == 0 {
		p.logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msg("SHOULD RARELY HAPPEN: ApplyObservations called with nil set of Shannon request observations.")
		return nil
	}

	// hand over the observations to the sanctioned endpoints store for adding any applicable sanctions.
	p.sanctionedEndpointsStore.ApplyObservations(shannonObservations)

	return nil

}

// ConfiguredServiceIDs returns the list of all all service IDs for all configured AATs.
// This is used by the hydrator to determine which service IDs to run QoS checks on.
func (p *Protocol) ConfiguredServiceIDs() map[protocol.ServiceID]struct{} {
	// Currently hydrator is only enabled for Centralized gateway mode.
	// TODO_FUTURE(@adshmh): support specifying the app(s) used for sending/signing synthetic relay requests by the hydrator.
	if p.gatewayMode != protocol.GatewayModeCentralized {
		return nil
	}

	configuredServiceIDs := make(map[protocol.ServiceID]struct{})
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
	return p.IsHealthy()
}

// TODO_FUTURE(@adshmh): If multiple apps (across different sessions) are delegating
// to this gateway, optimize how the endpoints are managed/organized/cached.
//
// getAppsUniqueEndpoints returns a map of all endpoints matching serviceID and passing appFilter.
// If an endpoint matches a serviceID across multiple apps/sessions, only a single entry
// matching one of the apps/sessions is returned.
func (p *Protocol) getSessionsUniqueEndpoints(
	_ context.Context,
	serviceID protocol.ServiceID,
	activeSessions []sessiontypes.Session,
	filterSanctioned bool, // will be true for calls to getAppsUniqueEndpoints made by service request handling.
) (map[protocol.EndpointAddr]endpoint, error) {
	logger := p.logger.With(
		"method", "getAppsUniqueEndpoints",
		"service", serviceID,
		"num_valid_sessions", len(activeSessions),
	)
	logger.Info().Msgf(
		"About to fetch all unique endpoints for service %s given %d active sessions.",
		serviceID, len(activeSessions),
	)

	endpoints := make(map[protocol.EndpointAddr]endpoint)
	// Iterate over all active sessions for the service ID.
	for _, session := range activeSessions {
		app := session.Application

		// Using a single iteration scope for this logger.
		// Avoids adding all apps in the loop to the logger's fields.
		// Hydrate the logger with session details.
		logger := logger.With("valid_app_address", app.Address)
		logger = hydrateLoggerWithSession(logger, &session)
		logger.ProbabilisticDebugInfo(polylog.ProbabilisticDebugInfoProb).Msgf("Finding unique endpoints for session %s for app %s for service %s.", session.SessionId, app.Address, serviceID)

		// Retrieve all endpoints for the session.
		sessionEndpoints, err := endpointsFromSession(session)
		if err != nil {
			logger.Error().Err(err).Msgf("Internal error: error getting all endpoints for service %s app %s and session: skipping the app.", serviceID, app.Address)
			continue
		}

		qualifiedEndpoints := sessionEndpoints
		// In calls to getAppsUniqueEndpoints made by service request handling, we filter out sanctioned endpoints.
		if filterSanctioned {
			logger.Debug().Msgf(
				"app %s has %d endpoints before filtering sanctioned endpoints.",
				app.Address, len(sessionEndpoints),
			)

			// Filter out any sanctioned endpoints
			filteredEndpoints := p.sanctionedEndpointsStore.FilterSanctionedEndpoints(qualifiedEndpoints)
			// All endpoints are sanctioned: log a warning and skip this app.
			if len(filteredEndpoints) == 0 {
				logger.Error().Msgf(
					"All %d session endpoints are sanctioned for service %s, app %s. Skipping the app.",
					len(sessionEndpoints), serviceID, app.Address,
				)
				continue
			}
			qualifiedEndpoints = filteredEndpoints

			logger.Debug().Msgf("app %s has %d endpoints after filtering sanctioned endpoints.", app.Address, len(qualifiedEndpoints))
		}

		// Log the number of endpoints before and after filtering
		logger.Info().Msgf("Filtered session endpoints for app %s from %d to %d.", app.Address, len(sessionEndpoints), len(qualifiedEndpoints))

		for endpointAddr, endpoint := range qualifiedEndpoints {
			endpoints[endpointAddr] = endpoint
		}

		logger.Info().Msgf(
			"Successfully fetched %d endpoints for session %s for application %s for service %s.",
			len(qualifiedEndpoints), session.SessionId, app.Address, serviceID,
		)
	}

	// Ensure at least one endpoint is available for the requested service.
	if len(endpoints) == 0 {
		// Wrap the context setup error. Used for generating observations.
		err := fmt.Errorf("%w: service %s", errProtocolContextSetupNoEndpoints, serviceID)
		logger.Warn().Err(err).Msg("No endpoints available after filtering sanctioned endpoints: relay request will fail.")
		return nil, err
	}

	logger.Info().Msgf("Successfully fetched %d endpoints for active sessions.", len(endpoints))

	return endpoints, nil
}

// GetTotalProtocolEndpointsCount returns the count of all unique endpoints for a service ID
// without filtering sanctioned endpoints.
func (p *Protocol) GetTotalServiceEndpointsCount(serviceID protocol.ServiceID, httpReq *http.Request) (int, error) {
	ctx := context.Background()

	// Get the list of active sessions for the service ID.
	activeSessions, err := p.getActiveGatewaySessions(ctx, serviceID, httpReq)
	if err != nil {
		return 0, err
	}

	// Get all endpoints for the service ID without filtering sanctioned endpoints.
	endpoints, err := p.getSessionsUniqueEndpoints(ctx, serviceID, activeSessions, false)
	if err != nil {
		return 0, err
	}

	return len(endpoints), nil
}

// HydrateDisqualifiedEndpointsResponse hydrates the disqualified endpoint response with the protocol-specific data.
//   - takes a pointer to the DisqualifiedEndpointResponse
//   - called by the devtools.DisqualifiedEndpointReporter to fill it with the protocol-specific data.
func (p *Protocol) HydrateDisqualifiedEndpointsResponse(serviceID protocol.ServiceID, details *devtools.DisqualifiedEndpointResponse) {
	p.logger.Info().Msgf("hydrating disqualified endpoints response for service ID: %s", serviceID)
	details.ProtocolLevelDisqualifiedEndpoints = p.sanctionedEndpointsStore.getSanctionDetails(serviceID)
}
