package shannon

import (
	"fmt"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"

	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/relayer"
)

// relayer package's Protocol interface is fulfilled by the Protocol struct
// below using methods that are specific to Shannon.
var _ relayer.Protocol = &Protocol{}

// FullNode defines the set of capabilities the Shannon protocol integration needs
// from a fullnode for sending relays.
type FullNode interface {
	GetServiceApps(relayer.ServiceID) ([]apptypes.Application, error)
	// Note: Shannon returns the latest session for a service+app combination if no blockHeight is provided.
	// This is used here because the gateway only needs the current session for any service+app combination.
	GetSession(serviceID relayer.ServiceID, appAddr string) (sessiontypes.Session, error)

	// ValidateRelayResponse validates the raw bytes returned from an endpoint (in response to a relay request) and returns the parsed response.
	ValidateRelayResponse(supplierAddr sdk.SupplierAddress, responseBz []byte) (*servicetypes.RelayResponse, error)

	// IsHealthy returns true if the FullNode instance is healthy.
	// A LazyFullNode will always return true.
	// A CachingFullNode will return true if it has data in app and session caches.
	IsHealthy() bool
}

// Protocol provides the functionality needed by the relayer and gateway packages
// for sending a relay to a specific endpoint.
type Protocol struct {
	FullNode
	OperationMode
	Logger polylog.Logger
}

// BuildRequestContext builds and returns a Shannon-specific request context, which can be used to send relays.
func (p *Protocol) BuildRequestContext(serviceID relayer.ServiceID, httpReq *http.Request) (relayer.ProtocolRequestContext, error) {
	operationModeInstance, err := p.OperationMode.BuildInstanceFromHTTPRequest(httpReq)
	if err != nil {
		return nil, fmt.Errorf("BuildRequestContext: error building an operation mode instance: %w", err)
	}

	endpoints, err := p.getAppsUniqueEndpoints(serviceID, operationModeInstance.GetAppFilterFn())
	if err != nil {
		return nil, fmt.Errorf("BuildRequestContext: error getting endpoints for service %s: %w", serviceID, err)
	}

	return &requestContext{
		fullNode:           p.FullNode,
		endpoints:          endpoints,
		serviceID:          serviceID,
		relayRequestSigner: operationModeInstance.GetRelayRequestSigner(),
	}, nil
}

// TODO_MVP(@adshmh): implement protocol state update using the observations by completing the ApplyObservations method below.
//
// ApplyObservations applies the Shannon protocol-level observations to the endpoint store.
// e.g. Observation showing an endpoint has maxed out on a specific app for the current session.
func (p *Protocol) ApplyObservations(protocolobservations.ProtocolDetails) error {
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

// TODO_FUTURE: Find a more optimized way of handling an overlap among endpoints
// matching multiple sessions of apps delegating to the gateway.
//
// getAppsUniqueEndpoints returns a map of all endpoints matching the provided service ID.
// If an endpoint matches a service ID through multiple apps/sessions, only a single entry
// matching one of the apps/sessions is returned.
func (p *Protocol) getAppsUniqueEndpoints(serviceID relayer.ServiceID, appFilter IsAppPermittedFn) (map[relayer.EndpointAddr]endpoint, error) {
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

	endpoints := make(map[relayer.EndpointAddr]endpoint)
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
