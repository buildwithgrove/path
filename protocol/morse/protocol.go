package morse

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pokt-foundation/pocket-go/provider"
	sdkrelayer "github.com/pokt-foundation/pocket-go/relayer"
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/health"
	"github.com/buildwithgrove/path/metrics/devtools"
	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
)

// gateway package's Protocol interface is fulfilled by the Protocol struct
// below using Morse-specific methods.
var _ gateway.Protocol = &Protocol{}

// Morse protocol implements the health.Check and health.ServiceIDReporter interfaces.
// This allows the protocol to report its health status and the list of service IDs it is configured for.
var (
	_ health.Check             = &Protocol{}
	_ health.ServiceIDReporter = &Protocol{}
)

// TODO_TECHDEBT(@adshmh): make the apps and sessions cache refresh interval configurable.
var appsAndSessionsCacheRefreshInterval = time.Minute

// OffChainBackend allows enhancing an onchain application with extra fields that are required to sign/send relays.
// This is used to supply AAT data to a Morse application, which is needed for sending relays on behalf of the application.
type OffChainBackend interface {
	// GetSignedAAT returns the AAT created by AppID offchain
	GetSignedAAT(appAddr string) (provider.PocketAAT, bool)
}

// FullNode defines the interface for what a "full node" needs to expose
type FullNode interface {
	GetAllApps(context.Context) ([]provider.App, error)
	GetSession(ctx context.Context, chainID, appPublicKey string) (provider.Session, error)
	SendRelay(context.Context, *sdkrelayer.Input) (*sdkrelayer.Output, error)
}

// NewProtocol creates a new Protocol.
func NewProtocol(logger polylog.Logger, fullNode FullNode, offChainBackend OffChainBackend) *Protocol {
	morseLogger := logger.With("protocol", "morse")

	protocol := &Protocol{
		appCache:                 make(map[protocol.ServiceID][]app),
		sessionCache:             make(map[string]provider.Session),
		logger:                   morseLogger,
		fullNode:                 fullNode,
		offChainBackend:          offChainBackend,
		sanctionedEndpointsStore: newSanctionedEndpointsStore(logger),
	}

	go func() {
		// Start the initial refresh
		protocol.refreshAll()
		// TODO_TECHDEBT(@adshmh): make the refresh interval configurable.
		ticker := time.NewTicker(appsAndSessionsCacheRefreshInterval)
		defer ticker.Stop()

		for {
			<-ticker.C
			protocol.refreshAll()
		}
	}()

	return protocol
}

// Protocol is Gateway protocol adapter for Morse protocol. It adapts Gateway interface to Morse interface.
type Protocol struct {
	logger          polylog.Logger
	fullNode        FullNode
	offChainBackend OffChainBackend

	// sanctionedEndpointsStore tracks sanctioned endpoints
	sanctionedEndpointsStore *sanctionedEndpointsStore

	// appCache caches applications associated with the services supported.
	appCache   map[protocol.ServiceID][]app
	appCacheMu sync.RWMutex

	// sessionCache caches sessions for use by the Relay function.
	// map keys are of the format "serviceID:appAddr"
	sessionCache   map[string]provider.Session
	sessionCacheMu sync.RWMutex
}

// TODO_TECHDEBT(@adshmh): Enforce the deadline from the supplied context.
// This is needed to avoid triggering a timeout on the HTTP server handling the service request.
//
// AvailableEndpoints returns the list of available endpoints for a given service ID.
//
// Implements the gateway.Protocol interface.
func (p *Protocol) AvailableEndpoints(_ context.Context, serviceID protocol.ServiceID, _ *http.Request) (protocol.EndpointAddrList, error) {
	endpoints, err := p.getEndpoints(serviceID)
	if err != nil {
		return nil, fmt.Errorf("AvailableEndpoints: error getting endpoints for service %s: %w", serviceID, err)
	}

	// Convert the list of endpoints to a list of endpoint addresses
	endpointAddrs := make(protocol.EndpointAddrList, 0, len(endpoints))
	for endpointAddr := range endpoints {
		endpointAddrs = append(endpointAddrs, endpointAddr)
	}

	return endpointAddrs, nil
}

// TODO_TECHDEBT(@adshmh): Enforce the deadline from the supplied context.
//
// BuildRequestContextForEndpoint builds a new request context for a given service ID and endpoint address.
// Implements the gateway.Protocol interface.
func (p *Protocol) BuildRequestContextForEndpoint(
	_ context.Context,
	serviceID protocol.ServiceID,
	selectedEndpointAddr protocol.EndpointAddr,
	_ *http.Request,
) (gateway.ProtocolRequestContext, error) {
	// Create a logger specifically for this request context
	ctxLogger := p.logger.With(
		"service_id", string(serviceID),
		"component", "request_context_for_endpoint",
	)

	// Retrieve the list of endpoints (i.e. backend service URLs by external operators)
	// that can service RPC requests for the given service ID.
	endpoints, err := p.getEndpoints(serviceID)
	if err != nil {
		return nil, fmt.Errorf("BuildRequestContextForEndpoint: error getting endpoints for service %s: %w", serviceID, err)
	}

	// Select the endpoint that matches the pre-selected address.
	// This ensures QoS checks are performed on the selected endpoint.
	selectedEndpoint, ok := endpoints[selectedEndpointAddr]
	if !ok {
		return nil, fmt.Errorf("BuildRequestContextForEndpoint: could not find endpoint for service %s and endpoint address %s", serviceID, selectedEndpointAddr)
	}

	// Return new request context for the pre-selected endpoint
	return &requestContext{
		logger:           ctxLogger,
		fullNode:         p.fullNode,
		selectedEndpoint: &selectedEndpoint,
		serviceID:        serviceID,
	}, nil
}

// ApplyObservations updates the Morse protocol instance's internal state using the supplied observations.
// It processes endpoint error observations to apply appropriate sanctions.
// Implements the gateway.Protocol interface.
func (p *Protocol) ApplyObservations(observations *protocolobservations.Observations) error {
	// Sanity check the input
	if observations == nil || observations.GetMorse() == nil {
		p.logger.Warn().Msg("ApplyObservations called with nil input or nil Morse observation list.")
		return nil
	}
	morseObservations := observations.GetMorse().GetObservations()
	if len(morseObservations) == 0 {
		p.logger.Warn().Msg("ApplyObservations called with nil set of Morse request observations.")
		return nil
	}

	// hand over the observations to the sanctioned endpoints store for adding any applicable sanctions.
	p.sanctionedEndpointsStore.ApplyObservations(morseObservations)

	return nil
}

// ConfiguredServiceIDs returns the list of all service IDs with configured AATs.
// This is used by the hydrator to determine which service IDs to run QoS checks on.
func (p *Protocol) ConfiguredServiceIDs() map[protocol.ServiceID]struct{} {
	p.appCacheMu.RLock()
	defer p.appCacheMu.RUnlock()

	configuredServiceIDs := make(map[protocol.ServiceID]struct{}, len(p.appCache))
	for serviceID := range p.appCache {
		configuredServiceIDs[serviceID] = struct{}{}
	}

	return configuredServiceIDs
}

// Name satisfies the HealthCheck#Name interface function
func (p *Protocol) Name() string {
	return "pokt-morse"
}

// IsAlive satisfies the HealthCheck#IsAlive interface function
func (p *Protocol) IsAlive() bool {
	p.appCacheMu.RLock()
	defer p.appCacheMu.RUnlock()
	p.sessionCacheMu.RLock()
	defer p.sessionCacheMu.RUnlock()

	return len(p.appCache) > 0 && len(p.sessionCache) > 0
}

// refreshAll refreshes all caches
func (p *Protocol) refreshAll() {
	p.logger.Debug().Msg("refreshAll: starting cache refresh")
	err := p.refreshAppsCache()
	if err != nil {
		p.logger.Error().Err(err).Msg("refreshAll: error refreshing apps cache")
	}
	err = p.refreshSessionCache()
	if err != nil {
		p.logger.Error().Err(err).Msg("refreshAll: error refreshing session cache")
	}
	p.logger.Debug().Msg("refreshAll: finished cache refresh")
}

// refreshAppsCache refreshes the app cache
func (p *Protocol) refreshAppsCache() error {
	appData := p.fetchAppData()
	if len(appData) == 0 {
		return errors.New("refreshAppCache: received an empty app list; skipping update")
	}

	p.appCacheMu.Lock()
	defer p.appCacheMu.Unlock()
	p.appCache = appData

	p.logger.Debug().Int("apps_count", len(appData)).Msg("refreshAppsCache: refreshed app cache")
	return nil
}

func (p *Protocol) fetchAppData() map[protocol.ServiceID][]app {
	logger := p.logger.With(
		"protocol", "Morse",
		"method", "fetchAppData",
	)

	onchainApps, err := p.fullNode.GetAllApps(context.Background())
	if err != nil {
		logger.Warn().Err(err).Msg("error getting list of onchain applications")
		return nil
	}

	appData := make(map[protocol.ServiceID][]app)
	for _, onchainApp := range onchainApps {
		logger := logger.With(
			"publicKey", onchainApp.PublicKey,
			"address", onchainApp.Address,
		)

		if len(onchainApp.Chains) == 0 {
			logger.Warn().Msg("app has no chains specified onchain. Skipping the app.")
			continue
		}

		// TODO_IMPROVE: validate the AAT received from the offChainBackend
		signedAAT, ok := p.offChainBackend.GetSignedAAT(onchainApp.Address)
		if !ok {
			logger.Debug().Msg("no AAT configured for app. Skipping the app.")
			continue
		}

		app := app{
			address:   onchainApp.Address,
			publicKey: onchainApp.PublicKey,
			aat:       signedAAT,
		}

		for _, chainID := range onchainApp.Chains {
			serviceID := protocol.ServiceID(chainID)
			appData[serviceID] = append(appData[serviceID], app)
			logger.With("service_iD", serviceID).Info().Msg("Found matching AAT, adding the app/service combination to the cache.")
		}

	}

	return appData
}

// refreshSessionCache refreshes the session cache
func (p *Protocol) refreshSessionCache() error {
	p.appCacheMu.RLock()
	defer p.appCacheMu.RUnlock()

	sessions := make(map[string]provider.Session)
	for serviceID, apps := range p.appCache {
		for _, app := range apps {
			session, err := p.fullNode.GetSession(context.Background(), string(serviceID), app.publicKey)
			if err != nil {
				// Log the error but continue processing other sessions
				p.logger.Warn().
					Err(err).
					Str("service", string(serviceID)).
					Str("appPublicKey", string(app.publicKey)).
					Msg("refreshSessionCache: error getting a session")

				continue
			}

			key := sessionCacheKey(serviceID, app.address)
			sessions[key] = session
		}
	}

	p.sessionCacheMu.Lock()
	p.sessionCache = sessions
	p.sessionCacheMu.Unlock()

	p.logger.Debug().Int("count", len(sessions)).Msg("refreshSessionCache: refreshed session cache")
	return nil
}

// getAppsUniqueEndpoints returns a map of all endpoints matching the provided service ID.
// It also filters out sanctioned endpoints from the endpoint store.
func (p *Protocol) getAppsUniqueEndpoints(serviceID protocol.ServiceID, apps []app) (map[protocol.EndpointAddr]endpoint, error) {
	endpoints := make(map[protocol.EndpointAddr]endpoint)

	// Get a logger specifically for this operation
	logger := p.logger.With("method", "getAppsUniqueEndpoints")

	for _, app := range apps {
		session, found := p.getSession(serviceID, app.Addr())
		if !found {
			return nil, fmt.Errorf("getAppsUniqueEndpoints: no session found for service %s app %s", serviceID, app.Addr())
		}

		logger := loggerWithSession(logger, app.Addr(), session)

		// Log session information for debugging
		logger.Debug().Msg("Processing app-session combination")

		// Get all endpoints for this app-session combination
		allAppEndpoints := getEndpointsFromAppSession(app, session)
		logger.Debug().
			Int("endpoint_count", len(allAppEndpoints)).
			Msg("Found endpoints for app")

		// Filter out any sanctioned endpoints
		filteredEndpoints := p.sanctionedEndpointsStore.FilterSanctionedEndpoints(allAppEndpoints, app.Addr(), session.Key)
		logger.Debug().
			Int("original_count", len(allAppEndpoints)).
			Int("filtered_count", len(filteredEndpoints)).
			Msg("Filtered sanctioned endpoints")

		// Add remaining endpoints to the map
		for _, endpoint := range filteredEndpoints {
			endpoints[endpoint.Addr()] = endpoint
		}
	}

	return endpoints, nil
}

// getSession gets a session from the session cache for the given service ID and application address
func (p *Protocol) getSession(serviceID protocol.ServiceID, appAddr string) (provider.Session, bool) {
	p.sessionCacheMu.RLock()
	defer p.sessionCacheMu.RUnlock()

	key := sessionCacheKey(serviceID, appAddr)
	session, found := p.sessionCache[key]
	return session, found
}

// getEndpoints returns all endpoints for a given service ID
func (p *Protocol) getEndpoints(serviceID protocol.ServiceID) (map[protocol.EndpointAddr]endpoint, error) {
	apps, found := p.getApps(serviceID)
	if !found || len(apps) == 0 {
		return nil, fmt.Errorf("getEndpoints: no apps found for service %s", serviceID)
	}

	return p.getAppsUniqueEndpoints(serviceID, apps)
}

// getApps gets apps from the app cache for a given service Id
func (p *Protocol) getApps(serviceID protocol.ServiceID) ([]app, bool) {
	p.appCacheMu.RLock()
	defer p.appCacheMu.RUnlock()

	apps, found := p.appCache[serviceID]
	return apps, found
}

// sessionCacheKey generates a cache key for a (serviceID, appAddr) pair
func sessionCacheKey(serviceID protocol.ServiceID, appAddr string) string {
	return fmt.Sprintf("%s:%s", serviceID, appAddr)
}

// GetSanctionedEndpoints returns all Morse sanctioned endpoints
// If serviceID is provided, it returns the sanctioned endpoints for that service ID.
// If serviceID is not provided, it returns the sanctioned endpoints for all configured service IDs.
// It also returns the total number of endpoints, the number of valid endpoints, and the number of sanctioned endpoints.
func (p *Protocol) GetSanctionedEndpoints(serviceID protocol.ServiceID) devtools.SanctionDetailsResponse {
	validEndpoints := 0

	var serviceIDsToCount []protocol.ServiceID
	if serviceID == "" {
		for serviceID := range p.ConfiguredServiceIDs() {
			serviceIDsToCount = append(serviceIDsToCount, serviceID)
		}
	} else {
		serviceIDsToCount = []protocol.ServiceID{serviceID}
	}

	for _, serviceID := range serviceIDsToCount {
		endpoints, err := p.getEndpoints(serviceID)
		if err != nil {
			p.logger.Error().Err(err).Msg("GetSanctionedEndpoints: error getting endpoints")
			continue
		}
		validEndpoints += len(endpoints)
	}

	sanctionDetails := p.sanctionedEndpointsStore.getSanctionDetails(serviceID)

	sanctionDetails.ValidEndpointsCount = validEndpoints
	sanctionDetails.TotalEndpointsCount = validEndpoints + sanctionDetails.SanctionedEndpointsCount

	return sanctionDetails
}
