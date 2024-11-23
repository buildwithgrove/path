package shannon

import (
	"fmt"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
)

// gateway package's Protocol interface is fulfilled by the Protocol struct
// below using methods that are specific to Shannon.
var _ gateway.Protocol = &Protocol{}

// FullNode defines the set of capabilities the Shannon protocol integration needs
// from a fullnode for sending relays.
type FullNode interface {
	GetServiceApps(protocol.ServiceID) ([]apptypes.Application, error)

		// Note: Shannon returns the latest session for a service+app combination if no blockHeight is provided.
	// This is used here because the gateway only needs the current session for any service+app combination.
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
	fullNode FullNode,
	logger polylog.Logger,
	config GatewayConfig,
) (*Protocol, error) {
	ownedAppsAddr, err := getCentralizedModeOwnedAppsAddr(config.OwnedAppsPrivateKeysHex)
	if err != nil {
		return nil, fmt.Errorf("NewProtocol: error parsing the supplied private keys: %w", err)
	}

	ownedAppsAddrIdx := make(map[string]struct{})
	for _, appAddr := range ownedAppsAddr {
		ownedAppsAddrIdx[appAddr] = struct{}{}
	}

	return &Protocol{
		FullNode: fullNode,
		Logger:   logger,

		// TODO_MVP(@adshmh): verify the gateway address and private key are valid.
		gatewayAddr:          config.GatewayAddress,
		gatewayPrivateKeyHex: config.GatewayPrivateKeyHex,
		gatewayMode:          config.GatewayMode,
		ownedAppsAddr:        ownedAppsAddrIdx,
	}, nil
}

// Protocol provides the functionality needed by the gateway package for sending a relay to a specific endpoint.
type Protocol struct {
	FullNode
	Logger polylog.Logger

	// gatewayMode is the gateway mode in which the current instance of the Shannon protocol integration operates.
	// See protocol/shannon/gateway_mode.go for more details.
	gatewayMode protocol.GatewayMode

	// gatewayAddr is used by the SDK for selecting onchain applications which have delegated to the gateway.
	// The gateway can only sign relays on behalf of an application if the application has an active delegation to it.
	gatewayAddr string

	// gatewayPrivateKeyHex stores the private key of the gateway running this Shannon integration instance.
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
	permittedAppFilter, err := p.getGatewayModePermittedAppFilter(p.gatewayMode, httpReq)
	if err != nil {
		return nil, fmt.Errorf("BuildRequestContext: error building the permitted apps filter for gateway mode %s: %w", p.gatewayMode, err)
	}

	endpoints, err := p.getAppsUniqueEndpoints(serviceID, permittedAppFilter)
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

// Name satisfies the HealthCheck#Name interface function
func (p *Protocol) Name() string {
	return "pokt-shannon"
}

// IsAlive satisfies the HealthCheck#IsAlive interface function
func (p *Protocol) IsAlive() bool {
	return p.FullNode.IsHealthy()
}

// TODO_FUTURE(@adshmh): Find a more optimized way of handling an overlap among endpoints
// matching multiple sessions of apps delegating to the gateway.
//
// getAppsUniqueEndpoints returns a map of all endpoints which match the provided service ID and pass the supplied app filter.
// If an endpoint matches a service ID through multiple apps/sessions, only a single entry
// matching one of the apps/sessions is returned.
func (p *Protocol) getAppsUniqueEndpoints(serviceID protocol.ServiceID, appFilter permittedAppFilter) (map[protocol.EndpointAddr]endpoint, error) {
	apps, err := p.FullNode.GetServiceApps(serviceID)
	if err != nil {
		return nil, fmt.Errorf("getAppsUniqueEndpoints: no apps found for service %s: %w", serviceID, err)
	}

	logger := p.Logger.With("service", serviceID)
	var filteredApps []apptypes.Application
	for _, app := range apps {
		logger = logger.With("app_address", app.Address)

		if errSelectingApp := appFilter(&app); errSelectingApp != nil {
			logger.Warn().Err(errSelectingApp).Msg("App filter rejected the app: skipping the app.")
			continue
		}

		filteredApps = append(filteredApps, app)
	}

	endpoints := make(map[protocol.EndpointAddr]endpoint)
	for _, app := range filteredApps {
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
	}

	if len(endpoints) == 0 {
		return nil, fmt.Errorf("getAppsUniqueEndpoints: no endpoints found for service %s", serviceID)
	}

	return endpoints, nil
}
