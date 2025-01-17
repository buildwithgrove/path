package shannon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"

	"github.com/buildwithgrove/path/protocol"
)

// TODO_IMPROVE: make the refresh interval configurable.
const cacheRefreshIntervalSeconds = 60

// TODO_IMPROVE: make this configurable.
const maxSessionFetchWorkers = 100

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

// CachingFullNode single responsibility is to add a caching layer around a LazyFullNode.
type CachingFullNode struct {
	Logger polylog.Logger

	*LazyFullNode

	// sessionCache caches sessions to be used to build the `endpointCache`.
	// Map keys are of the format "{serviceID}-{appID}"
	sessionCache   map[sessionCacheKey]sessiontypes.Session
	sessionCacheMu sync.RWMutex

	// endpointCache caches endpoints to be used by the `protocol.BuildRequestContext` method.
	endpointCache   map[protocol.ServiceID]map[protocol.EndpointAddr]endpoint
	endpointCacheMu sync.RWMutex

	// once is used to ensure the cache update go routine of the `start` method is only run once.
	once sync.Once
}

// sessionCacheKey returns a string to be used as the key for storing the session matching the supplied service ID and application address.
// e.g. for service with ID `svc1` and app with address `appAddress1`, the key is `svc1-appAddress1`.
type sessionCacheKey struct {
	serviceID protocol.ServiceID
	appAddr   string
}

func newSessionCacheKey(serviceID protocol.ServiceID, appAddr string) sessionCacheKey {
	return sessionCacheKey{
		serviceID: serviceID,
		appAddr:   appAddr,
	}
}

// start launches a goroutine, only once per instance of CachingFullNode in order to update the cached items at a fixed interval.
func (cfn *CachingFullNode) start() error {
	if cfn.LazyFullNode == nil {
		return errors.New("start: CachingFullNode needs a LazyFullNode to operate")
	}

	if cfn.Logger == nil {
		return errors.New("start: CachingFullNode needs a Logger to operate")
	}

	// Start the cache update process, using a sync.Once to ensure it is only run once.
	cfn.once.Do(func() {
		go func() {
			// TODO_IMPROVE: make the refresh interval configurable.
			ticker := time.NewTicker(cacheRefreshIntervalSeconds * time.Second)

			for {
				cfn.Logger.Info().Msg("Starting the cache update process.")

				// 1. Fetch and filter the onchain apps.
				appsDataByService, err := cfn.fetchAndFilterApps()
				if err != nil {
					cfn.Logger.Warn().Err(err).Msg("cache update: error fetching and filtering apps; skipping update.")
					continue
				}

				// 2. Fetch and cache the sessions for the filtered apps.
				cfn.updateSessionCache(appsDataByService)

				// 3. Update the endpoint cache using the cached sessions.
				cfn.updateEndpointCache(appsDataByService)

				<-ticker.C
			}
		}()
	})

	return nil
}

/* ------------------------------- FullNode Interface Methods ------------------------------- */

// SetPermittedAppFilter sets the permitted app filter for the protocol instance.
func (cfn *CachingFullNode) SetPermittedAppFilter(permittedAppFilter permittedAppFilter) {
	cfn.LazyFullNode.SetPermittedAppFilter(permittedAppFilter)
}

// GetServiceEndpoints returns (from the cache) the set of endpoints which delegate to the gateway, matching the supplied service ID.
// It is required to fulfill the FullNode interface.
func (cfn *CachingFullNode) GetServiceEndpoints(serviceID protocol.ServiceID, req *http.Request) (map[protocol.EndpointAddr]endpoint, error) {
	cfn.endpointCacheMu.RLock()
	defer cfn.endpointCacheMu.RUnlock()

	cachedEndpoints, found := cfn.endpointCache[serviceID]
	if !found {
		return nil, fmt.Errorf("getServiceEndpoints: no endpoints found for service %s", serviceID)
	}

	return cachedEndpoints, nil
}

// GetSession returns the cached session matching (serviceID, appAddr) combination.
// It is required to fulfill the FullNode interface.
func (cfn *CachingFullNode) GetSession(serviceID protocol.ServiceID, appAddr string) (sessiontypes.Session, error) {
	cfn.sessionCacheMu.RLock()
	defer cfn.sessionCacheMu.RUnlock()

	session, found := cfn.sessionCache[newSessionCacheKey(serviceID, appAddr)]
	if !found {
		return session, fmt.Errorf("getSession: no cached sessions found for service %s, app %s", serviceID, appAddr)
	}

	return session, nil
}

// ValidateRelayResponse validates the raw response bytes received from an endpoint using the SDK and the account client.
// This method delegates to the underlying LazyFullNode.
// It is required to fulfill the FullNode interface.
func (cfn *CachingFullNode) ValidateRelayResponse(supplierAddr sdk.SupplierAddress, responseBz []byte) (*servicetypes.RelayResponse, error) {
	return cfn.LazyFullNode.ValidateRelayResponse(supplierAddr, responseBz)
}

// IsHealthy indicates the health status of the caching full node.
// It is required to fulfill the health.Check interface.
func (cfn *CachingFullNode) IsHealthy() bool {
	cfn.endpointCacheMu.RLock()
	defer cfn.endpointCacheMu.RUnlock()
	cfn.sessionCacheMu.RLock()
	defer cfn.sessionCacheMu.RUnlock()

	return len(cfn.endpointCache) > 0 && len(cfn.sessionCache) > 0
}

/* ------------------------------- 1. Fetch Permitted Onchain Apps Data ------------------------------- */

// fetchAndFilterApps fetches the all apps for all services and filters them using the gateway mode's permittedAppFilter.
func (cfn *CachingFullNode) fetchAndFilterApps() (map[protocol.ServiceID][]apptypes.Application, error) {
	// TODO_MVP(@adshmh): remove this once poktroll supports querying the onchain apps.
	// More specifically, once we can filter by apps delegating to a gateway address.
	appsData, err := cfn.appClient.GetAllApplications(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("fetchAndFilterApps: error getting the list of apps: %w", err)
	}

	// A nil request is passed to filterPermittedApps as the caching full node does not
	// need to filter based on the app address specified in the HTTP request's headers.
	filteredAppsData := cfn.filterPermittedApps(appsData, nil)

	appsDataByService, err := cfn.buildAppsServiceMap(filteredAppsData, nil)
	if err != nil {
		return nil, fmt.Errorf("fetchAndFilterApps: %w", err)
	}

	return appsDataByService, nil
}

/* ------------------------------- 2. Fetch and Cache Sessions ------------------------------- */

// updateSessionCache fetches the sessions for the supplied apps and caches them.
func (cfn *CachingFullNode) updateSessionCache(appsDataByService map[protocol.ServiceID][]apptypes.Application) {
	sessions := cfn.fetchSessions(appsDataByService)
	if len(sessions) == 0 {
		cfn.Logger.Warn().Msg("updateSessionCache: received empty session list; skipping update.")
		return
	}

	cfn.sessionCacheMu.Lock()
	defer cfn.sessionCacheMu.Unlock()
	cfn.sessionCache = sessions
}

// fetchSessions fetches the sessions for the supplied apps and returns them,
// using a worker pool pattern to fetch the sessions concurrently.
func (cfn *CachingFullNode) fetchSessions(appsDataByService map[protocol.ServiceID][]apptypes.Application) map[sessionCacheKey]sessiontypes.Session {
	if len(appsDataByService) == 0 {
		cfn.Logger.Warn().Msg("fetchSessions: received empty app list; skipping update.")
		return nil
	}

	sessions := make(map[sessionCacheKey]sessiontypes.Session)
	var sessionsMu sync.Mutex

	// Use a worker pool to fetch the sessions concurrently.
	jobs := make(chan sessionCacheKey, len(appsDataByService))
	var wg sync.WaitGroup
	for i := 0; i < maxSessionFetchWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for sessionCacheKey := range jobs {
				logger := cfn.Logger.With(
					"service", string(sessionCacheKey.serviceID),
					"address", sessionCacheKey.appAddr,
				)

				session, err := cfn.LazyFullNode.GetSession(sessionCacheKey.serviceID, sessionCacheKey.appAddr)
				if err != nil {
					logger.Warn().Err(err).Msg("could not get a session")
					continue
				}

				sessionsMu.Lock()
				sessions[sessionCacheKey] = session
				sessionsMu.Unlock()

				logger.Info().Msg("fetchSessions: successfully fetched the session for service and app combination.")
			}
		}()
	}

	// Send jobs to the workers
	for serviceID, serviceApps := range appsDataByService {
		for _, app := range serviceApps {
			jobs <- newSessionCacheKey(serviceID, app.Address)
		}
	}
	close(jobs)

	wg.Wait()

	return sessions
}

/* ------------------------------- 3. Update Endpoint Cache ------------------------------- */

// updateEndpointCache updates the endpoint cache by fetching the sessions for all apps and then
// using the getAppsUniqueEndpoints method to get the unique endpoints for each service ID.
func (cfn *CachingFullNode) updateEndpointCache(appsDataByService map[protocol.ServiceID][]apptypes.Application) {
	if len(appsDataByService) == 0 {
		cfn.Logger.Warn().Msg("updateEndpointCache: received empty app list; skipping update.")
		return
	}

	endpointData := make(map[protocol.ServiceID]map[protocol.EndpointAddr]endpoint)

	for serviceID, apps := range appsDataByService {
		endpointsForService, err := cfn.getAppsUniqueEndpoints(serviceID, apps)
		if err != nil {
			continue
		}

		endpointData[serviceID] = endpointsForService
	}

	cfn.endpointCacheMu.Lock()
	defer cfn.endpointCacheMu.Unlock()
	cfn.endpointCache = endpointData
}

// TODO_FUTURE(@adshmh): Find a more optimized way of handling an overlap among endpoints
// matching multiple sessions of apps delegating to the gateway.
//
// getAppsUniqueEndpoints returns a map of all endpoints which match the provided service ID and pass the supplied app filter.
// If an endpoint matches a service ID through multiple apps/sessions, only a single entry
// matching one of the apps/sessions is returned.
func (cfn *CachingFullNode) getAppsUniqueEndpoints(serviceID protocol.ServiceID, apps []apptypes.Application) (map[protocol.EndpointAddr]endpoint, error) {
	cfn.sessionCacheMu.RLock()
	sessions := cfn.sessionCache
	cfn.sessionCacheMu.RUnlock()

	endpoints := make(map[protocol.EndpointAddr]endpoint)

	for _, app := range apps {
		session, found := sessions[newSessionCacheKey(serviceID, app.Address)]
		if !found {
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
