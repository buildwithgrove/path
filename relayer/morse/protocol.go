package morse

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pokt-foundation/pocket-go/provider"
	sdkrelayer "github.com/pokt-foundation/pocket-go/relayer"
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/health"
	"github.com/buildwithgrove/path/relayer"
)

// relayer package's Protocol interface is fulfilled by the Protocol struct
// below using Morse-specific methods.
var _ relayer.Protocol = &Protocol{}

// All components that report their ready status to /healthz must implement the health.Check interface.
var _ health.Check = &Protocol{}

// TODO_TECHDEBT: Make this configurable via an env variable.
const defaultRelayTimeoutMillisec = 5000

// OffChainBackend allows enhancing an onchain application with extra fields that are required to sign/send relays.
// This is used to supply AAT data to a Morse application, which is needed for sending relays on behalf of the application.
type OffChainBackend interface {
	// GetSignedAAT returns the AAT created by AppID offchain
	GetSignedAAT(appAddr string) (provider.PocketAAT, bool)
}

// FullNode defines the functionality expected by the Protocol struct
// from a Morse full node.
type FullNode interface {
	GetAllApps(context.Context) ([]provider.App, error)
	GetSession(ctx context.Context, chainID, appPublicKey string) (provider.Session, error)
	SendRelay(context.Context, *sdkrelayer.Input) (*sdkrelayer.Output, error)
}

func NewProtocol(ctx context.Context, fullNode FullNode, offChainBackend OffChainBackend) (*Protocol, error) {
	protocol := &Protocol{
		fullNode:        fullNode,
		offChainBackend: offChainBackend,
		logger:          polylog.Ctx(ctx),
	}

	go func() {
		// TODO_IMPROVE: make the refresh interval configurable.
		ticker := time.NewTicker(time.Minute)
		for {
			protocol.updateAppCache()
			protocol.updateSessionCache()

			<-ticker.C
		}
	}()

	return protocol, nil
}

type Protocol struct {
	logger polylog.Logger

	fullNode        FullNode
	offChainBackend OffChainBackend

	appCache   map[relayer.ServiceID][]app
	appCacheMu sync.RWMutex
	// TODO_IMPROVE: Add a sessionCacheKey type with the necessary helpers to concat a key
	// sessionCache caches sessions for use by the Relay function.
	// map keys are of the format "serviceID-appID"
	sessionCache   map[string]provider.Session
	sessionCacheMu sync.RWMutex
}

// BuildRequestContext builds and returns a Morse-specific request context, which can be used to send relays.
func (p *Protocol) BuildRequestContext(serviceID relayer.ServiceID) (relayer.ProtocolRequestContext, error) {
	apps, found := p.getServiceApps(serviceID)
	if !found {
		return nil, fmt.Errorf("buildRequestContext: no apps found for service %s", serviceID)
	}

	endpoints, err := p.getAppsUniqueEndpoints(serviceID, apps)
	if err != nil {
		return nil, fmt.Errorf("buildRequestContext: error getting endpoints for service %s: %w", serviceID, err)
	}

	return &requestContext{
		fullNode:  p.fullNode,
		endpoints: endpoints,
		serviceID: serviceID,
	}, nil
}

func (p *Protocol) Endpoints(serviceID relayer.ServiceID) ([]relayer.Endpoint, error) {
	apps, found := p.getServiceApps(serviceID)
	if !found {
		return nil, fmt.Errorf("buildRequestContext: no apps found for service %s", serviceID)
	}

	endpointsIdx, err := p.getAppsUniqueEndpoints(serviceID, apps)
	if err != nil {
		return nil, fmt.Errorf("endpoints: error getting endpoints for service %s: %w", serviceID, err)
	}

	var endpoints []relayer.Endpoint
	for _, endpoint := range endpointsIdx {
		endpoints = append(endpoints, endpoint)
	}

	if len(endpoints) == 0 {
		return nil, fmt.Errorf("endpoints: no endpoints found for service %s", serviceID)
	}

	return endpoints, nil
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

func (p *Protocol) getServiceApps(serviceID relayer.ServiceID) ([]app, bool) {
	p.appCacheMu.RLock()
	defer p.appCacheMu.RUnlock()

	cachedApps, found := p.appCache[serviceID]
	if !found {
		return nil, false
	}

	apps := make([]app, len(cachedApps))
	for i, app := range cachedApps {
		apps[i] = app
	}

	return apps, true
}

func (p *Protocol) getSession(serviceID relayer.ServiceID, appAddr string) (provider.Session, bool) {
	p.sessionCacheMu.RLock()
	defer p.sessionCacheMu.RUnlock()

	session, found := p.sessionCache[sessionCacheKey(serviceID, appAddr)]
	return session, found
}

func (p *Protocol) updateAppCache() {
	appData := p.fetchAppData()

	if len(appData) == 0 {
		p.logger.Warn().Msg("updateAppCache: received an empty app list; skipping update")
		return
	}

	p.appCacheMu.Lock()
	defer p.appCacheMu.Unlock()
	p.appCache = appData
}

func (p *Protocol) fetchAppData() map[relayer.ServiceID][]app {
	logger := p.logger.With(
		"protocol", "Morse",
		"method", "fetchAppData",
	)

	onchainApps, err := p.fullNode.GetAllApps(context.Background())
	if err != nil {
		logger.Warn().Err(err).Msg("error getting list of onchain applications")
		return nil
	}

	appData := make(map[relayer.ServiceID][]app)
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
			logger.Info().Msg("no AAT configured for app. Skipping the app.")
			continue
		}

		app := app{
			address:   onchainApp.Address,
			publicKey: onchainApp.PublicKey,
			aat:       signedAAT,
		}

		for _, chainID := range onchainApp.Chains {
			serviceID := relayer.ServiceID(chainID)
			appData[serviceID] = append(appData[serviceID], app)
			logger.With("serviceID", serviceID).Info().Msg("Found matching AAT, adding the app/service combination to the cache.")
		}

	}

	return appData
}

func (p *Protocol) updateSessionCache() {
	sessions := p.fetchSessions()
	if len(sessions) == 0 {
		p.logger.Warn().Msg("updateSessionCache: received empty session list; skipping update.")
		return
	}

	p.sessionCacheMu.Lock()
	defer p.sessionCacheMu.Unlock()
	p.sessionCache = sessions
}

func (p *Protocol) fetchSessions() map[string]provider.Session {
	p.appCacheMu.RLock()
	defer p.appCacheMu.RUnlock()

	sessions := make(map[string]provider.Session)
	// TODO_TECHDEBT: use multiple go routines.
	for serviceID, apps := range p.appCache {
		for _, app := range apps {
			// NOTE: We use the application's public key here because that is what Morse full nodes require to return a session,
			// but we use an application's address to cache it and its corresponding session(s).
			session, err := p.fullNode.GetSession(context.Background(), string(serviceID), app.publicKey)
			if err != nil {
				p.logger.Warn().
					Err(err).
					Str("service", string(serviceID)).
					Str("appPublicKey", string(app.publicKey)).
					Msg("fetchSessions: error getting a session")

				continue
			}
			sessions[sessionCacheKey(serviceID, app.address)] = session
		}
	}

	return sessions
}

// TODO_UPNEXT(@adshmh): Refactor all caching out of the Protocol struct, and use an interface to access Apps and Sessions, and send relays.
// Then add 2 implementations of the FullNode interface:
// CachingFullNode
// LazyFullNode
//
// getAppsUniqueEndpoints returns a map of all endpoints matching the provided service ID.
// If an endpoint matches a service ID through multiple apps/sessions, only a single entry
// matching one of the apps/sessions is returned.
func (p *Protocol) getAppsUniqueEndpoints(serviceID relayer.ServiceID, apps []app) (map[relayer.EndpointAddr]endpoint, error) {
	endpoints := make(map[relayer.EndpointAddr]endpoint)
	for _, app := range apps {
		session, found := p.getSession(serviceID, app.Addr())
		if !found {
			return nil, fmt.Errorf("getAppsUniqueEndpoints: no session found for service %s app %s", serviceID, app.Addr())
		}

		for _, endpoint := range getEndpointsFromAppSession(app, session) {
			endpoints[endpoint.Addr()] = endpoint
		}
	}

	return endpoints, nil
}

func sessionCacheKey(serviceID relayer.ServiceID, appAddr string) string {
	return fmt.Sprintf("%s-%s", serviceID, appAddr)
}
