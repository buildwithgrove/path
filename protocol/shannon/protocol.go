package shannon

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
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

	// GetGatewayAddr returns the gateway address configured for the fullnode, to be used in filtering apps.
	GetGatewayAddr() string

	// GetAccountClient returns the account client from the fullnode, to be used in building relay request signers.
	GetAccountClient() *sdk.AccountClient
}

// NewProtocol instantiates an instance of the Shannon protocol integration.
func NewProtocol(
	fullNode FullNode,
	logger polylog.Logger,
	gatewayPrivateKeyHex string,
	ownedAppsPrivateKeys []*secp256k1.PrivKey,
) (*Protocol, error) {
	}
	ownedAppsAddr, err := getCentralizedModeOwnedAppsAddr(ownedAppsPrivateKeys)
	if err != nil {
		return nil, fmt.Errorf("NewProtocol: error parsing the supplied private keys: %w", err)
	}

	ownedAppsAddrIdx := make(map[string]struct{})
	for _, appAddr := range ownedAppsAddr {
		ownedAppsAddrIdx[appAddr] = struct{}{}
	}

	return &Protocol{
		FullNode: fullNode,
		Logger: logger,
		ownedAppsAddr: ownedAppsAddrIdx,
	}, nil
}

// Protocol provides the functionality needed by the gateway packag for sending a relay to a specific endpoint.
type Protocol struct {
	FullNode
	Logger polylog.Logger

	// ownedAppsAddr holds the addresss of all apps owned by the gateway operator running PATH in centralized mode.
	// This data is stored as a map for efficiency, since this field is only used to lookup app addresses.
	ownedAppsAddr map[string]struct{}
}

// BuildRequestContext builds and returns a Shannon-specific request context, which can be used to send relays.
func (p *Protocol) BuildRequestContext(
	serviceID protocol.ServiceID,
	gatewayMode protocol.GatewayMode, 
	httpReq *http.Request,
) (gateway.ProtocolRequestContext, error) {

	permittedAppsFilter, err := p.getGatewayModePermittedAppsFilter(gatewayMode, httpReq)
	if err != nil {
		return nil, fmt.Errorf("BuildRequestContext: error building the permitted apps filter for gateway mode %s: %w", gatewayMode, err)
	}

	endpoints, err := p.getAppsUniqueEndpoints(serviceID, permittedAppsFilter)
	if err != nil {
		return nil, fmt.Errorf("BuildRequestContext: error getting endpoints for service %s: %w", serviceID, err)
	}

	permittedSigner, err := p.getGatewayModePermittedSigner(gatewayMode)
	if err != nil {
		return nil, fmt.Errorf("BuildRequestContext: error getting the permitted signer for gateway mode %s: %w", gatewayMode, err)
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

// TODO_FUTURE: Find a more optimized way of handling an overlap among endpoints
// matching multiple sessions of apps delegating to the gateway.
//
// getAppsUniqueEndpoints returns a map of all endpoints matching the provided service ID.
// If an endpoint matches a service ID through multiple apps/sessions, only a single entry
// matching one of the apps/sessions is returned.
func (p *Protocol) getAppsUniqueEndpoints(serviceID protocol.ServiceID, appFilter permittedAppFilter) (map[protocol.EndpointAddr]endpoint, error) {
	apps, err := p.FullNode.GetServiceApps(serviceID)
	if err != nil {
		return nil, fmt.Errorf("getAppsUniqueEndpoints: no apps found for service %s: %w", serviceID, err)
	}

	var filteredApps []apptypes.Application
	for _, app := range apps {
		if isPermitted := appFilter(&app); isPermitted {
			filteredApps = append(filteredApps, app)
		}
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
