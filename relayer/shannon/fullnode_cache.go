package shannon

import (
	"fmt"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"

	"github.com/buildwithgrove/path/relayer"
)

// The Shannon Relayer's FullNode interface is implemented by the CachingFullNode struct below,
// which provides the full node capabilities required by the Shannon relayer.
var _ FullNode = &CachingFullNode{}

func NewCachingFullNode(lazyFullNode *LazyFullNode, logger polylog.Logger) *CachingFullNode {
	cachingFullNode := CachingFullNode{
		LazyFullNode: lazyFullNode,
		Logger:       logger,
	}

	cachingFullNode.start()
	return &cachingFullNode
}

// FullNodeCache's single responsibility is to add a caching layer around a LazyFullNode.
type CachingFullNode struct {
	LazyFullNode *LazyFullNode
	Logger       polylog.Logger

	appCache   map[relayer.ServiceID][]apptypes.Application
	appCacheMu sync.RWMutex

	// TODO_IMPROVE: Add a sessionCacheKey type with the necessary helpers to concat a key
	// sessionCache caches sessions for use by the Relay function.
	// map keys are of the format "serviceID-appID"
	sessionCache   map[string]sessiontypes.Session
	sessionCacheMu sync.RWMutex
}

// start launches a goroutine, only once per instance of FullNodeCache, to
func (cfn *CachingFullNode) start() {
	go func() {
		// TODO_IMPROVE: make the refresh interval configurable.
		ticker := time.NewTicker(time.Minute)
		for {
			cfn.updateAppCache()
			cfn.updateSessionCache()

			<-ticker.C
		}
	}()
}

func (cfn *CachingFullNode) GetServiceApps(serviceID relayer.ServiceID) ([]apptypes.Application, error) {
	cfn.appCacheMu.RLock()
	defer cfn.appCacheMu.RUnlock()

	apps, found := cfn.appCache[serviceID]
	if !found {
		return nil, fmt.Errorf("getServiceApps: no apps found for service %s", serviceID)
	}

	return apps, nil
}

func (cfn *CachingFullNode) GetSession(serviceID relayer.ServiceID, appAddr string) (sessiontypes.Session, error) {
	cfn.sessionCacheMu.RLock()
	defer cfn.sessionCacheMu.RUnlock()

	session, found := cfn.sessionCache[sessionCacheKey(serviceID, appAddr)]
	if !found {
		return session, fmt.Errorf("getSession: no cached sessions found for service %s, app %s", serviceID, appAddr)
	}

	return session, nil
}

// SendRelay delegates the sending of the relay to the LazyFullNode.
func (cfn *CachingFullNode) SendRelay(app apptypes.Application, session sessiontypes.Session, endpoint endpoint, payload relayer.Payload) (*servicetypes.RelayResponse, error) {
	return cfn.LazyFullNode.SendRelay(app, session, endpoint, payload)
}

func (cfn *CachingFullNode) IsHealthy() bool {
	cfn.appCacheMu.RLock()
	defer cfn.appCacheMu.RUnlock()
	cfn.sessionCacheMu.RLock()
	defer cfn.sessionCacheMu.RUnlock()

	return len(cfn.appCache) > 0 && len(cfn.sessionCache) > 0
}

func (cfn *CachingFullNode) updateAppCache() {
	appData, err := cfn.LazyFullNode.GetAllServicesApps()
	if err != nil {
		cfn.Logger.Warn().Err(err).Msg("updateAppCache: error getting the list of apps; skipping update.")
		return
	}

	cfn.appCacheMu.Lock()
	defer cfn.appCacheMu.Unlock()
	cfn.appCache = appData
}

func (cfn *CachingFullNode) updateSessionCache() {
	sessions := cfn.fetchSessions()
	if len(sessions) == 0 {
		cfn.Logger.Warn().Msg("updateSessionCache: received empty session list; skipping update.")
		return
	}

	cfn.sessionCacheMu.Lock()
	defer cfn.sessionCacheMu.Unlock()
	cfn.sessionCache = sessions
}

func (cfn *CachingFullNode) fetchSessions() map[string]sessiontypes.Session {
	logger := cfn.Logger.With("method", "fetchSessions")

	apps, err := cfn.LazyFullNode.GetAllServicesApps()
	if err != nil {
		logger.Warn().Err(err).Msg("fetchSession: error listing applications")
	}

	sessions := make(map[string]sessiontypes.Session)
	// TODO_TECHDEBT: use multiple go routines.
	for serviceID, serviceApps := range apps {
		for _, app := range serviceApps {
			logger := logger.With(
				"service", string(serviceID),
				"address", app.Address,
			)

			session, err := cfn.LazyFullNode.GetSession(serviceID, string(app.Address))
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

func sessionCacheKey(serviceID relayer.ServiceID, appAddr string) string {
	return fmt.Sprintf("%s-%s", string(serviceID), appAddr)
}
