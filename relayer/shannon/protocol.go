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

	"github.com/buildwithgrove/path/health"
	"github.com/buildwithgrove/path/relayer"
)

// relayer package's Protocol interface is fulfilled by the Protocol struct
// below using methods that are specific to Shannon.
var _ relayer.Protocol = &Protocol{}

// All components that report their ready status to /healthz must implement the health.Check interface.
var _ health.Check = &Protocol{}

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
		ticker := time.NewTicker(time.Second)
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

func (p *Protocol) Endpoints(serviceID relayer.ServiceID) (map[relayer.AppAddr][]relayer.Endpoint, error) {
	apps, found := p.serviceApps(serviceID)
	if !found {
		return nil, fmt.Errorf("endpoints: no apps found for service %s", serviceID)
	}

	allEndpoints := make(map[relayer.AppAddr][]relayer.Endpoint)
	for _, app := range apps {
		logger := p.logger.With(
			"service", string(serviceID),
			"address", app.Address,
		)

		session, found := p.getSession(serviceID, relayer.AppAddr(app.Address))
		if !found {
			logger.Warn().Msg("Endpoints: no sessions found for service,app combination. Skipping.")
			continue
		}

		sessionEndpoints, err := endpointsFromSession(session)
		if err != nil {
			p.logger.Warn().Err(err).Msg("Endpoints: error getting endpoints from the session")
			continue
		}

		endpoints := make([]relayer.Endpoint, len(sessionEndpoints))
		for i, sessionEndpoint := range sessionEndpoints {
			endpoints[i] = sessionEndpoint
		}
		allEndpoints[relayer.AppAddr(app.Address)] = endpoints
	}

	if len(allEndpoints) == 0 {
		return nil, fmt.Errorf("endpoints: no cached sessions found for service %s", serviceID)
	}

	return allEndpoints, nil
}

func (p *Protocol) SendRelay(req relayer.Request) (relayer.Response, error) {
	app, err := p.getApp(req.ServiceID, req.AppAddr)
	if err != nil {
		return relayer.Response{}, fmt.Errorf("sendRelay: app not found: %w", err)
	}

	session, found := p.getSession(req.ServiceID, req.AppAddr)
	if !found {
		return relayer.Response{}, fmt.Errorf("relay: session not found for service %s app %s", req.ServiceID, req.AppAddr)
	}

	endpoint, err := endpointFromSession(session, req.EndpointAddr)
	if err != nil {
		return relayer.Response{}, fmt.Errorf("relay: endpoint %s not found for service %s app %s: %w", req.EndpointAddr, req.ServiceID, req.AppAddr, err)
	}

	response, err := p.fullNode.SendRelay(app, session, endpoint, req.Payload)
	if err != nil {
		return relayer.Response{EndpointAddr: req.EndpointAddr},
			fmt.Errorf("relay: error sending relay for service %s app %s endpoint %s: %w",
				req.ServiceID, req.AppAddr, req.EndpointAddr, err,
			)
	}

	// The Payload field of the response received from the endpoint, i.e. the relay miner,
	// is a serialized http.Response struct. It needs to be deserialized into an HTTP Response struct
	// to access the Service's response body, status code, etc.
	relayResponse, err := deserializeRelayResponse(response.Payload)
	if err != nil {
		return relayer.Response{EndpointAddr: req.EndpointAddr},
			fmt.Errorf("relay: error unmarshalling endpoint response into a POKTHTTP response for service %s app %s endpoint %s: %w",
				req.ServiceID, req.AppAddr, req.EndpointAddr, err,
			)
	}

	relayResponse.EndpointAddr = req.EndpointAddr
	return relayResponse, nil
}

func (p *Protocol) serviceApps(serviceID relayer.ServiceID) ([]apptypes.Application, bool) {
	p.appCacheMu.RLock()
	defer p.appCacheMu.RUnlock()

	apps, found := p.appCache[serviceID]
	return apps, found
}

func (p *Protocol) getApp(serviceID relayer.ServiceID, appAddr relayer.AppAddr) (apptypes.Application, error) {
	apps, found := p.serviceApps(serviceID)
	if !found {
		return apptypes.Application{}, fmt.Errorf("getApp: service %s not found", serviceID)
	}

	app, found := appFromList(apps, appAddr)
	if found {
		return app, nil
	}

	return apptypes.Application{}, fmt.Errorf("getApp: service %s has no apps with address %s", serviceID, appAddr)
}

func (p *Protocol) getSession(serviceID relayer.ServiceID, appAddr relayer.AppAddr) (sessiontypes.Session, bool) {
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

	apps := p.AllApps()

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

			sessions[sessionCacheKey(serviceID, relayer.AppAddr(app.Address))] = session
			logger.Info().Msg("successfully fetched the session for service and app combination.")
		}
	}

	return sessions
}

func (p *Protocol) AllApps() map[relayer.ServiceID][]apptypes.Application {
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

func sessionCacheKey(serviceID relayer.ServiceID, appAddr relayer.AppAddr) string {
	return fmt.Sprintf("%s-%s", serviceID, appAddr)
}

func appFromList(apps []apptypes.Application, appAddr relayer.AppAddr) (apptypes.Application, bool) {
	for _, app := range apps {
		if app.Address == string(appAddr) {
			return app, true
		}
	}

	return apptypes.Application{}, false
}
