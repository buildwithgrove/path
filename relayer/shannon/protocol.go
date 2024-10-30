package shannon

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"

	"github.com/buildwithgrove/path/relayer"
)

// relayer package's Protocol interface is fulfilled by the Protocol struct
// below using methods that are specific to Shannon.
var _ relayer.Protocol = &Protocol{}

type FullNode interface {
	GetApps(context.Context) ([]apptypes.Application, error)
	LatestBlockHeight() (int64, error)
	GetSession(serviceID, appAddr string, blockHeight int64) (sessiontypes.Session, error)
	SendRelay(apptypes.Application, sessiontypes.Session, endpoint, relayer.Payload) (*servicetypes.RelayResponse, error)
}

// TODO_UPNEXT(@adshmh): Add unit/E2E tests for the implementation of the Shannon relayer.
func NewProtocol(ctx context.Context, fullNode FullNode) (*Protocol, error) {
	protocol := &Protocol{
		fullNode: fullNode,
		logger:   polylog.Ctx(ctx),
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
	fullNode FullNode
	logger   polylog.Logger

	appCache   map[relayer.ServiceID][]apptypes.Application
	appCacheMu sync.RWMutex

	// TODO_IMPROVE: Add a sessionCacheKey type with the necessary helpers to concat a key
	// sessionCache caches sessions for use by the Relay function.
	// map keys are of the format "serviceID-appID"
	sessionCache   map[string]sessiontypes.Session
	sessionCacheMu sync.RWMutex
}

// Name satisfies the HealthCheck#Name interface function
func (p *Protocol) Name() string {
	return "pokt-shannon"
}

// IsAlive satisfies the HealthCheck#IsAlive interface function
func (p *Protocol) IsAlive() bool {
	p.appCacheMu.RLock()
	defer p.appCacheMu.RUnlock()
	p.sessionCacheMu.RLock()
	defer p.sessionCacheMu.RUnlock()

	return len(p.appCache) > 0 && len(p.sessionCache) > 0
}

// func (p *Protocol) Endpoints(serviceID relayer.ServiceID) (map[relayer.AppAddr][]relayer.Endpoint, error) {
func (p *Protocol) Endpoints(serviceID relayer.ServiceID) ([]relayer.Endpoint, error) {
	apps, found := p.serviceApps(serviceID)
	if !found {
		return nil, fmt.Errorf("endpoints: no apps found for service %s", serviceID)
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

// BuildRequestContext builds and returns a Shannon-specific request context, which can be used to send relays.
func (p *Protocol) BuildRequestContext(serviceID relayer.ServiceID) (relayer.ProtocolRequestContext, error) {
	apps, found := p.serviceApps(serviceID)
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

// TODO_FUTURE: Find a more optimized way of handling an overlap among endpoints
// matching multiple sessions of apps delegating to the gateway.
//
// getAppsUniqueEndpoints returns a map of all endpoints matching the provided service ID.
// If an endpoint matches a service ID through multiple apps/sessions, only a single entry
// matching one of the apps/sessions is returned.
func (p *Protocol) getAppsUniqueEndpoints(serviceID relayer.ServiceID, apps []apptypes.Application) (map[relayer.EndpointAddr]endpoint, error) {
	endpoints := make(map[relayer.EndpointAddr]endpoint)
	for _, app := range apps {
		session, found := p.getSession(serviceID, app.Address)
		if !found {
			return nil, fmt.Errorf("getAppsUniqueEndpoints: no session found for service %s app %s", serviceID, app.Address)
		}

		appEndpoints, err := endpointsFromSession(session)
		if err != nil {
			return nil, fmt.Errorf("getAppsUniqueEndpoints: error getting all endpoints for app %s session %s: %w", app.Address, session.SessionId, err)
		}

		for endpointAddr, endpoint := range appEndpoints {
			endpoints[endpointAddr] = endpoint
		}
	}

	return endpoints, nil
}

func (p *Protocol) serviceApps(serviceID relayer.ServiceID) ([]apptypes.Application, bool) {
	p.appCacheMu.RLock()
	defer p.appCacheMu.RUnlock()

	apps, found := p.appCache[serviceID]
	return apps, found
}

func (p *Protocol) getSession(serviceID relayer.ServiceID, appAddr string) (sessiontypes.Session, bool) {
	p.sessionCacheMu.RLock()
	defer p.sessionCacheMu.RUnlock()

	session, found := p.sessionCache[sessionCacheKey(serviceID, appAddr)]
	return session, found
}

func (p *Protocol) updateAppCache() {
	appData := p.fetchAppData()
	if len(appData) == 0 {
		p.logger.Warn().Msg("updateAppCache: received an empty app list; skipping update.")
		return
	}

	p.appCacheMu.Lock()
	defer p.appCacheMu.Unlock()
	p.appCache = appData
}

func (p *Protocol) fetchAppData() map[relayer.ServiceID][]apptypes.Application {
	onchainApps, err := p.fullNode.GetApps(context.Background())
	if err != nil {
		p.logger.Warn().Err(err).Msg("updateAppCache: error getting list of applications from the SDK")
		return nil
	}

	appData := make(map[relayer.ServiceID][]apptypes.Application)
	for _, onchainApp := range onchainApps {
		logger := p.logger.With("address", onchainApp.Address)

		if len(onchainApp.ServiceConfigs) == 0 {
			logger.Warn().Msg("updateAppCache: app has no services specified onchain. Skipping the app.")
			continue
		}

		for _, svcCfg := range onchainApp.ServiceConfigs {
			if svcCfg.ServiceId == "" {
				logger.Warn().Msg("updateAppCache: app has empty serviceId item in service config.")
				continue
			}

			serviceID := relayer.ServiceID(svcCfg.ServiceId)
			appData[serviceID] = append(appData[serviceID], onchainApp)
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

func (p *Protocol) fetchSessions() map[string]sessiontypes.Session {
	logger := p.logger.With("method", "fetchSessions")

	blockHeight, err := p.fullNode.LatestBlockHeight()
	if err != nil {
		logger.Warn().
			Err(err).
			Msg("error getting the latest block height. Skipping session update.")

		return nil
	}

	apps := p.allApps()

	sessions := make(map[string]sessiontypes.Session)
	// TODO_TECHDEBT: use multiple go routines.
	for serviceID, serviceApps := range apps {
		for _, app := range serviceApps {
			logger := logger.With(
				"service", string(serviceID),
				"address", app.Address,
			)

			session, err := p.fullNode.GetSession(string(serviceID), string(app.Address), blockHeight)
			if err != nil {
				logger.Warn().Err(err).Msg("could not get a session")
				continue
			}

			sessions[sessionCacheKey(serviceID, app.Address)] = session
			logger.Info().Msg("successfully fetched the session for service and app combination.")
		}
	}

	return sessions
}

func (p *Protocol) allApps() map[relayer.ServiceID][]apptypes.Application {
	p.appCacheMu.RLock()
	defer p.appCacheMu.RUnlock()

	allApps := make(map[relayer.ServiceID][]apptypes.Application)
	for serviceID, cachedApps := range p.appCache {
		apps := make([]apptypes.Application, len(cachedApps))
		copy(apps, cachedApps)
		allApps[serviceID] = apps
	}

	return allApps
}

func sessionCacheKey(serviceID relayer.ServiceID, appAddr string) string {
	return fmt.Sprintf("%s-%s", serviceID, appAddr)
}
