package shannon

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sharedtypes "github.com/pokt-network/poktroll/x/shared/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/health"
	"github.com/buildwithgrove/path/metrics/devtools"
	shannonmetrics "github.com/buildwithgrove/path/metrics/protocol/shannon"
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

// FullNode defines the set of capabilities the Shannon protocol integration needs
// from a fullnode for sending relays.
//
// A properly initialized fullNode struct can:
// 1. Return the onchain apps matching a service ID.
// 2. Fetch a session for a (service,app) combination.
// 3. Validate a relay response.
// 4. Etc...
type FullNode interface {
	// GetApp returns the onchain application matching the application address
	GetApp(ctx context.Context, appAddr string) (*apptypes.Application, error)

	// GetSession returns the latest session matching the supplied service+app combination.
	// Sessions are solely used for sending relays, and therefore only the latest session for any service+app combination is needed.
	// Note: Shannon returns the latest session for a service+app combination if no blockHeight is provided.
	GetSession(ctx context.Context, serviceID protocol.ServiceID, appAddr string) (sessiontypes.Session, error)

	// GetSessionWithExtendedValidity implements session retrieval with support for
	// Pocket Network's native "session grace period" business logic.
	//
	// At the protocol level, it is used to account for the case when:
	// - RelayMiner.FullNode.Height > Gateway.FullNode.Height
	// AND
	// - RelayMiner.FullNode.Session > Gateway.FullNodeSession
	//
	// PATH leverages it by accounting for the case when:
	// - RelayMiner.FullNode.Height < Gateway.FullNode.Height
	// AND
	// - Gateway.FullNode.Session > RelayMiner.FullNodeSession
	//
	// This enables signing and sending relays to Suppliers who are behind the Gateway.
	//
	// The recommendation usage is to use both GetSession and GetSessionWithExtendedValidity
	// in order to account for both cases when selecting the pool of available Suppliers.
	//
	// Protocol References:
	// - https://github.com/pokt-network/poktroll/blob/main/proto/pocket/shared/params.proto
	// - https://dev.poktroll.com/protocol/governance/gov_params
	// - https://dev.poktroll.com/protocol/primitives/claim_and_proof_lifecycle

	// GetSessionWithExtendedValidity returns the appropriate session considering grace period logic.
	// If within grace period of a session rollover, it may return the previous session.
	GetSessionWithExtendedValidity(ctx context.Context, serviceID protocol.ServiceID, appAddr string) (sessiontypes.Session, error)

	// GetSharedParams returns the shared module parameters from the blockchain.
	GetSharedParams(ctx context.Context) (*sharedtypes.Params, error)

	// GetCurrentBlockHeight returns the current block height from the blockchain.
	GetCurrentBlockHeight(ctx context.Context) (int64, error)

	// ValidateRelayResponse validates the raw bytes returned from an endpoint (in response to a relay request) and returns the parsed response.
	ValidateRelayResponse(supplierAddr sdk.SupplierAddress, responseBz []byte) (*servicetypes.RelayResponse, error)

	// IsHealthy returns true if the FullNode instance is healthy.
	// A LazyFullNode will always return true.
	// A CachingFullNode will return true if it has data in app and session caches.
	IsHealthy() bool

	// GetAccountClient returns the account client from the fullnode, to be used in building relay request signers.
	GetAccountClient() *sdk.AccountClient
}

// NewProtocol instantiates an instance of the Shannon protocol integration.
func NewProtocol(
	logger polylog.Logger,
	config GatewayConfig,
	fullNode FullNode,
) (*Protocol, error) {
	shannonLogger := logger.With("protocol", "shannon")

	// Initialize the owned apps for the gateway mode. This value will be nil if gateway mode is not Centralized.
	var ownedApps map[protocol.ServiceID][]string
	if config.GatewayMode == protocol.GatewayModeCentralized {
		var err error
		if ownedApps, err = getCentralizedModeOwnedApps(logger, config.OwnedAppsPrivateKeysHex, fullNode); err != nil {
			return nil, fmt.Errorf("failed to get app addresses from config: %w", err)
		}
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

		// ownedApps is the list of apps owned by the gateway operator running PATH in Centralized gateway mode.
		// If PATH is not running in Centralized mode, this field is nil.
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

	// ownedApps holds the addresses and staked service IDs of all apps owned by the gateway operator running
	// PATH in Centralized mode. If PATH is not running in Centralized mode, this field is nil.
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
	startTime := time.Now()

	// hydrate the logger.
	logger := p.logger.With(
		"service", serviceID,
		"method", "AvailableEndpoints",
		"gateway_mode", p.gatewayMode,
	)

	// TODO_TECHDEBT(@adshmh): validate "serviceID" is a valid onchain Shannon service.
	sessionsStartTime := time.Now()
	activeSessions, err := p.getActiveGatewaySessions(ctx, serviceID, httpReq)
	sessionsDuration := time.Since(sessionsStartTime).Seconds()

	if err != nil {
		logger.Error().Err(err).Msg("Relay request will fail: error building the active sessions for service.")
		shannonmetrics.RecordSessionOperationDuration(serviceID, "get_active_sessions", "error", false, sessionsDuration)
		return nil, buildProtocolContextSetupErrorObservation(serviceID, err), err
	}
	shannonmetrics.RecordSessionOperationDuration(serviceID, "get_active_sessions", "success", false, sessionsDuration)

	logger = logger.With("number_of_valid_sessions", len(activeSessions))
	logger.Debug().Msg("fetched the set of active sessions.")

	// Retrieve a list of all unique endpoints for the given service ID filtered by the list of apps this gateway/application
	// owns and can send relays on behalf of.
	// The final boolean parameter sets whether to filter out sanctioned endpoints.
	endpointsStartTime := time.Now()
	// TODO_IN_THIS_PR: Maintain the session metadata alongside the endpoint so we can
	// make sure we make requests across two different sessions.
	endpoints, err := p.getSessionsUniqueEndpoints(ctx, serviceID, activeSessions, true)
	endpointsDuration := time.Since(endpointsStartTime).Seconds()

	if err != nil {
		logger.Error().Err(err).Msg(err.Error())
		shannonmetrics.RecordSessionOperationDuration(serviceID, "get_unique_endpoints", "error", false, endpointsDuration)
		return nil, buildProtocolContextSetupErrorObservation(serviceID, err), err
	}
	shannonmetrics.RecordSessionOperationDuration(serviceID, "get_unique_endpoints", "success", false, endpointsDuration)

	logger = logger.With("number_of_unique_endpoints", len(endpoints))
	logger.Debug().Msg("Successfully fetched the set of available endpoints for the selected apps.")

	// Convert the list of endpoints to a list of endpoint addresses
	endpointAddrs := make(protocol.EndpointAddrList, 0, len(endpoints))
	for endpointAddr := range endpoints {
		endpointAddrs = append(endpointAddrs, endpointAddr)
	}

	// Record overall AvailableEndpoints duration
	duration := time.Since(startTime).Seconds()
	shannonmetrics.RecordSessionOperationDuration(serviceID, "available_endpoints", "success", false, duration)

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
		context:            ctx,
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
		logger := logger.With("valid_app_address", app.Address).With("method", "getSessionsUniqueEndpoints")
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
