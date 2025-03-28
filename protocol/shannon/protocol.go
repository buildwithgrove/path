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
type FullNode interface {
	// GetApp returns the onchain application matching the application address
	GetApp(ctx context.Context, appAddr string) (*apptypes.Application, error)

	// GetSession returns the latest session matching the supplied service+app combination.
	// Sessions are solely used for sending relays, and therefore only the latest session for any service+app combination is needed.
	// Note: Shannon returns the latest session for a service+app combination if no blockHeight is provided.
	GetSession(serviceID protocol.ServiceID, appAddr string) (sessiontypes.Session, error)

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
		Logger: logger,

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
	Logger polylog.Logger
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

// BuildRequestContext builds and returns a Shannon-specific request context, which can be used to send relays.
func (p *Protocol) BuildRequestContext(
	serviceID protocol.ServiceID,
	httpReq *http.Request,
) (gateway.ProtocolRequestContext, error) {
	// TODO_TECHDEBT(@adshmh): validate "serviceID" is a valid onchain Shannon service.

	permittedApps, err := p.getGatewayModePermittedApps(context.TODO(), serviceID, httpReq)
	if err != nil {
		return nil, fmt.Errorf("BuildRequestContext: error building the permitted apps list for service %s gateway mode %s: %w", serviceID, p.gatewayMode, err)
	}

	endpoints, err := p.getAppsUniqueEndpoints(serviceID, permittedApps)
	if err != nil {
		return nil, fmt.Errorf("BuildRequestContext: error getting endpoints for service %s: %w", serviceID, err)
	}

	permittedSigner, err := p.getGatewayModePermittedRelaySigner(p.gatewayMode)
	if err != nil {
		return nil, fmt.Errorf("BuildRequestContext: error getting the permitted signer for gateway mode %s: %w", p.gatewayMode, err)
	}

	return &requestContext{
		fullNode:           p.FullNode,
		endpoints:          endpoints,
		serviceID:          serviceID,
		relayRequestSigner: permittedSigner,
	}, nil
}

// GetUniqueEndpoints returns a map of all unique endpoints for a given service ID.
// Implements the gateway.Protocol interface.
func (p *Protocol) GetUniqueEndpoints(serviceID protocol.ServiceID) ([]protocol.Endpoint, error) {
	permittedApps, err := p.getGatewayModePermittedApps(context.TODO(), serviceID, nil)
	if err != nil {
		return nil, fmt.Errorf("BuildRequestContext: error building the permitted apps list for service %s gateway mode %s: %w", serviceID, p.gatewayMode, err)
	}

	endpoints, err := p.getAppsUniqueEndpoints(serviceID, permittedApps)
	if err != nil {
		return nil, fmt.Errorf("getUniqueEndpoints: error getting endpoints for service %s: %w", serviceID, err)
	}

	uniqueEndpoints := make([]protocol.Endpoint, 0, len(endpoints))
	for _, endpoint := range endpoints {
		uniqueEndpoints = append(uniqueEndpoints, endpoint)
	}

	return uniqueEndpoints, nil
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
	serviceID protocol.ServiceID,
	permittedApps []*apptypes.Application,
) (map[protocol.EndpointAddr]endpoint, error) {
	logger := p.Logger.With("service", serviceID)

	var endpoints = make(map[protocol.EndpointAddr]endpoint)
	for _, app := range permittedApps {
		logger = logger.With("permitted_app_address", app.Address)
		logger.Debug().Msg("getAppsUniqueEndpoints: processing app.")

		session, err := p.FullNode.GetSession(serviceID, app.Address)
		if err != nil {
			return nil, fmt.Errorf("getAppsUniqueEndpoints: could not get the session for service %s app %s", serviceID, app.Address)
		}

		appEndpoints, err := endpointsFromSession(session)
		if err != nil {
			return nil, fmt.Errorf("getAppsUniqueEndpoints: error getting all endpoints for app %s session %s: %w", app.Address, session.SessionId, err)
		}

		for endpointAddr, endpoint := range appEndpoints {
			endpoints[endpointAddr] = endpoint
		}

		logger.With("num_endpoints", len(appEndpoints)).Info().Msg("Successfully fetched session for application.")
	}

	if len(endpoints) == 0 {
		return nil, fmt.Errorf("getAppsUniqueEndpoints: no endpoints found for service %s", serviceID)
	}

	return endpoints, nil
}
