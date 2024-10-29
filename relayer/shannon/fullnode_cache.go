package shannon

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"

	"github.com/buildwithgrove/path/relayer"
)

// TODO_IMPROVE: make the refresh interval configurable.
const cacheRefreshIntervalSeconds = 60

// The Shannon Relayer's FullNode interface is implemented by the CachingFullNode struct below,
// which provides the full node capabilities required by the Shannon relayer.
var _ FullNode = &CachingFullNode{}

func NewCachingFullNode(lazyFullNode *LazyFullNode, logger polylog.Logger) (*CachingFullNode, error) {
	cachingFullNode := CachingFullNode{
		LazyFullNode: lazyFullNode,
		Logger:       logger,
	}

	if err := cachingFullNode.start(); err != nil {
		return nil, err
	}

	return &cachingFullNode, nil
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

	// once is used to ensure the cache update go routine of the `start` method is only run once.
	once sync.Once
}

// start launches a goroutine, only once per instance of FullNodeCache, to
func (cfn *CachingFullNode) start() error {
	if cfn.LazyFullNode == nil {
		return errors.New("CachingFullNode needs a LazyFullNode to operate.")
	}

	if cfn.Logger == nil {
		return errors.New("CachingFullNode needs a Logger to operate.")
	}

	cfn.once.Do(func() {
		go func() {
			// TODO_IMPROVE: make the refresh interval configurable.
			ticker := time.NewTicker(cacheRefreshIntervalSeconds * time.Second)
			for {
				cfn.Logger.Info().Msg("Starting the cache update process.")

				cfn.updateAppCache()
				cfn.updateSessionCache()

				<-ticker.C
			}
		}()
	})

	return nil
}

// GetServiceApps returns (from the cache) the set of onchain applications which delegate to the gateway, matching the supplied service ID.
// It is required to fulfill the FullNode interface.
func (cfn *CachingFullNode) GetServiceApps(serviceID relayer.ServiceID) ([]apptypes.Application, error) {
	cfn.appCacheMu.RLock()
	defer cfn.appCacheMu.RUnlock()

	apps, found := cfn.appCache[serviceID]
	if !found {
		return nil, fmt.Errorf("getServiceApps: no apps found for service %s", serviceID)
	}

	return apps, nil
}

// GetSession returns the cached session matching (serviceID, appAddr) combination.
// It is required to fulfill the FullNode interface.
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
// It is required to fulfill the FullNode interface.
func (cfn *CachingFullNode) SendRelay(app apptypes.Application, session sessiontypes.Session, endpoint endpoint, payload relayer.Payload) (*servicetypes.RelayResponse, error) {
	return cfn.LazyFullNode.SendRelay(app, session, endpoint, payload)
}

// IsHealthy indicates the health status of the caching full node.
// It is required to fulfill the FullNode interface.
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

// sessionCacheKey returns a string to be used as the key for storing the session matching the supplied service ID and application address.
// e.g. for service with ID `svc1` and app with address `appAddress1`, the key is `svc1-appAddress1`.
func sessionCacheKey(serviceID relayer.ServiceID, appAddr string) string {
	return fmt.Sprintf("%s-%s", string(serviceID), appAddr)
}
