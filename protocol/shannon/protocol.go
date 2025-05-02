package shannon

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/gateway"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
)

// gateway package's Protocol interface is fulfilled by the Protocol struct
// below using methods that are specific to Shannon.
var _ gateway.Protocol = &Protocol{}

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
	fullNode FullNode,
	config GatewayConfig,
) (*Protocol, error) {
	shannonLogger := logger.With("protocol", "shannon")

	// Derive the address of apps owned by the gateway operator using the supplied apps' private keys.
	// This only applies to Centralized gateway mode and needs to be done during initialization to ensure it is possible to send relays in Centralized mode.
	ownedAppsAddr, err := getCentralizedModeOwnedAppsAddr(config.OwnedAppsPrivateKeysHex)
	if err != nil {
		return nil, fmt.Errorf("NewProtocol: error parsing the supplied private keys: %w", err)
	}

	ownedAppsAddrIdx := make(map[string]struct{})
	for _, appAddr := range ownedAppsAddr {
		ownedAppsAddrIdx[appAddr] = struct{}{}
	}

	return &Protocol{
		logger: shannonLogger,

		FullNode: fullNode,

		// TODO_MVP(@adshmh): verify the gateway address and private key are valid, by completing the following:
		// 1. Query onchain data for a gateway with the supplied address.
		// 2. Query onchain data for app(s) matching the derived addresses.
		gatewayAddr:          config.GatewayAddress,
		gatewayPrivateKeyHex: config.GatewayPrivateKeyHex,
		gatewayMode:          config.GatewayMode,
		ownedAppsAddr:        ownedAppsAddrIdx,
	}, nil
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

	// ownedAppsAddr holds the addresss of all apps owned by the gateway operator running PATH in Centralized mode.
	// This data is stored as a map for efficiency, since this field is only used to lookup app addresses.
	ownedAppsAddr map[string]struct{}
}

// AvailableEndpoints returns the list available endpoints for a given service ID.
// Takes the HTTP request as an argument for Delegated mode to get permitted apps from the HTTP request's headers.
//
// Implements the gateway.Protocol interface.
func (p *Protocol) AvailableEndpoints(
	ctx context.Context,
	serviceID protocol.ServiceID,
	httpReq *http.Request,
) (protocol.EndpointAddrList, error) {
	// hydrate the logger.
	logger := p.logger.With("service", serviceID)

	// TODO_TECHDEBT(@adshmh): validate "serviceID" is a valid onchain Shannon service.
	permittedApps, err := p.getGatewayModePermittedApps(ctx, serviceID, httpReq)
	if err != nil {
		logger.Error().Err(err).Msg("error getting the permitted apps list: relay request will fail.")
		return nil, fmt.Errorf("AvailableEndpoints: error building the permitted apps list for service '%s' and gateway mode '%s': %w", serviceID, p.gatewayMode, err)
	}

	logger = logger.With("number_of_permitted_apps", len(permittedApps))
	logger.Debug().Msg("fetched the set of permitted apps.")

	// Retrieve a list of all unique endpoints for the given service ID filtered by the list of apps this gateway/application
	// owns and can send relays on behalf of.
	endpoints, err := p.getAppsUniqueEndpoints(ctx, serviceID, permittedApps)
	if err != nil {
		logger.Error().Err(err).Msg("error getting the set of available endpoints: relay request will fail.")
		return nil, fmt.Errorf("AvailableEndpoints: error getting endpoints for service %s: %w", serviceID, err)
	}

	logger = logger.With("number_of_unique_endpoints", len(endpoints))
	logger.Debug().Msg("Successfully fetched the set of available endpoints for the selected apps.")

	// Convert the list of endpoints to a list of endpoint addresses
	endpointAddrs := make(protocol.EndpointAddrList, 0, len(endpoints))
	for endpointAddr := range endpoints {
		endpointAddrs = append(endpointAddrs, endpointAddr)
	}

	return endpointAddrs, nil
}

// BuildRequestContextForEndpoint builds a new request context for a given service ID and endpoint address.
// Takes the HTTP request as an argument for Delegate mode to get permitted apps from the HTTP request's headers.
//
// Implements the gateway.Protocol interface.
func (p *Protocol) BuildRequestContextForEndpoint(
	ctx context.Context,
	serviceID protocol.ServiceID,
	selectedEndpointAddr protocol.EndpointAddr,
	httpReq *http.Request,
) (gateway.ProtocolRequestContext, error) {
	permittedApps, err := p.getGatewayModePermittedApps(ctx, serviceID, httpReq)
	if err != nil {
		return nil, fmt.Errorf("BuildRequestContextForEndpoint: error building the permitted apps list for service %s gateway mode %s: %w", serviceID, p.gatewayMode, err)
	}

	// Retrieve the list of endpoints (i.e. backend service URLs by external operators)
	// that can service RPC requests for the given service ID for the given apps.
	endpoints, err := p.getAppsUniqueEndpoints(ctx, serviceID, permittedApps)
	if err != nil {
		return nil, fmt.Errorf("BuildRequestContextForEndpoint: error getting endpoints for service %s: %w", serviceID, err)
	}

	// Select the endpoint that matches the pre-selected address.
	// This ensures QoS checks are performed on the selected endpoint.
	selectedEndpoint, ok := endpoints[selectedEndpointAddr]
	if !ok {
		return nil, fmt.Errorf("BuildRequestContextForEndpoint: could not find endpoint for service %s and endpoint address %s", serviceID, selectedEndpointAddr)
	}

	// Retrieve the relay request signer for the current gateway mode.
	permittedSigner, err := p.getGatewayModePermittedRelaySigner(p.gatewayMode)
	if err != nil {
		return nil, fmt.Errorf("BuildRequestContextForEndpoint: error getting the permitted signer for gateway mode %s: %w", p.gatewayMode, err)
	}

	// Return new request context for the pre-selected endpoint
	return &requestContext{
		logger:             p.logger,
		fullNode:           p.FullNode,
		selectedEndpoint:   &selectedEndpoint,
		serviceID:          serviceID,
		relayRequestSigner: permittedSigner,
	}, nil
}

// ApplyObservations updates protocol instance state based on endpoint observations.
// Examples:
// - Mark endpoints as invalid based on response quality
// - Disqualify endpoints for a time period
//
// Implements gateway.Protocol interface.
func (p *Protocol) ApplyObservations(_ *protocolobservations.Observations) error {
	// TODO_MVP(@adshmh):
	//  1. Implement endpoint store for status tracking
	//  2. Add validation logic to update store based on observations
	//  3. Filter invalid endpoints before setting on requestContexts
	//     (e.g., drop maxed-out endpoints for current session)
	return nil
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
func (p *Protocol) getAppsUniqueEndpoints(
	ctx context.Context,
	serviceID protocol.ServiceID,
	permittedApps []*apptypes.Application,
) (map[protocol.EndpointAddr]endpoint, error) {
	logger := p.logger.With(
		"service", serviceID,
		"num_permitted_apps", len(permittedApps),
	)

	endpoints := make(map[protocol.EndpointAddr]endpoint)
	for _, app := range permittedApps {
		// Using a single iteration scope for this logger.
		// Avoids adding all apps in the loop to the logger's fields.
		logger := logger.With("permitted_app_address", app.Address)
		logger.Debug().Msg("getAppsUniqueEndpoints: processing app.")

		session, err := p.FullNode.GetSession(ctx, serviceID, app.Address)
		if err != nil {
			logger.Warn().Err(err).Msg("Internal error: error getting a session for the app. Service request will fail.")
			return nil, fmt.Errorf("getAppsUniqueEndpoints: could not get the session for service %s app %s", serviceID, app.Address)
		}

		appEndpoints, err := endpointsFromSession(session)
		if err != nil {
			logger.Warn().Err(err).Msg("Internal error: error getting all endpoints for app and session. Service request will fail.")
			return nil, fmt.Errorf("getAppsUniqueEndpoints: error getting all endpoints for app %s session %s: %w", app.Address, session.SessionId, err)
		}

		for endpointAddr, endpoint := range appEndpoints {
			endpoints[endpointAddr] = endpoint
		}

		logger.With("num_endpoints", len(appEndpoints)).Info().Msg("Successfully fetched session for application.")
	}

	if len(endpoints) == 0 {
		logger.Warn().Msg("Internal error: no endpoints available for permitted apps. Service request will fail.")
		return nil, fmt.Errorf("getAppsUniqueEndpoints: no endpoints found for service %s", serviceID)
	}

	logger.With("num_endpoints", len(endpoints)).Debug().Msg("Successfully fetched endpoints for permitted apps.")

	return endpoints, nil
}
