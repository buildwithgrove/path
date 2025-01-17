package shannon

import (
	"fmt"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"
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
	// GetServiceEndpoints returns all the onchain applications staked for the supplied Service ID.
	GetServiceEndpoints(protocol.ServiceID, *http.Request) (map[protocol.EndpointAddr]endpoint, error)

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

	// SetPermittedAppFilter sets the permitted app filter for the protocol instance.
	SetPermittedAppFilter(permittedAppFilter permittedAppFilter)
}

// NewProtocol instantiates an instance of the Shannon protocol integration.
func NewProtocol(
	fullNode FullNode,
	logger polylog.Logger,
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

	protocol := &Protocol{
		FullNode: fullNode,
		Logger:   logger,

		// TODO_MVP(@adshmh): verify the gateway address and private key are valid, by completing the following:
		// 1. Query onchain data for a gateway with the supplied address.
		// 2. Query onchain data for app(s) matching the derived addresses.
		gatewayAddr:          config.GatewayAddress,
		gatewayPrivateKeyHex: config.GatewayPrivateKeyHex,
		gatewayMode:          config.GatewayMode,
		ownedAppsAddr:        ownedAppsAddrIdx,
	}

	// TODO_IMPROVE(@commoddity): move permittedAppFilter initialization to the FullNode initialization.
	permittedAppFilter, err := protocol.getGatewayModePermittedAppFilter(protocol.gatewayMode)
	if err != nil {
		return nil, fmt.Errorf("NewProtocol: error building the permitted apps filter for gateway mode %s: %w", protocol.gatewayMode, err)
	}

	fullNode.SetPermittedAppFilter(permittedAppFilter)

	return protocol, nil
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

	// gatewayPrivateKeyHex stores the private key of the gateway running this Shannon Gateway instance.
	// It is used for signing relay request in both Centralized and Delegated Gateway Modes.
	gatewayPrivateKeyHex string

	// ownedAppsAddr holds the addresss of all apps owned by the gateway operator running PATH in Centralized mode.
	// This data is stored as a map for efficiency, since this field is only used to lookup app addresses.
	ownedAppsAddr map[string]struct{}
}

// BuildRequestContext builds and returns a Shannon-specific request context, which can be used to send relays.
// TODO_TECHDEBT(@dashmh): validate the provided request's service ID is supported by the Shannon protocol.
func (p *Protocol) BuildRequestContext(
	serviceID protocol.ServiceID,
	httpReq *http.Request,
) (gateway.ProtocolRequestContext, error) {
	endpoints, err := p.FullNode.GetServiceEndpoints(serviceID, httpReq)
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

// TODO_MVP(@adshmh): complete the ApplyObservations method by implementing:
//  1. An endpoint store to maintain a status for each endpoint.
//  2. Validation logic that updates the endpoint store based on the supplied observations.
//  3. Use the endpoint store to filter out invalid endpoints before setting them on any requestContexts.
//     e.g. an endpoint that is maxed out for an app should be dropped for the remaining of the current session.
//
// ApplyObservations updates the Morse protocol instance's internal state using the supplied observations.
// e.g. an invalid response from an endpoint could be used to disqualify it for a set period of time.
// This method implements the gateway.Protocol interface.
func (p *Protocol) ApplyObservations(_ *protocolobservations.Observations) error {
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
