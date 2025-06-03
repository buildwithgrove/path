package shannon

import (
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pokt-network/poktroll/pkg/polylog"

	protocolobservations "github.com/buildwithgrove/path/observation/protocol"
	"github.com/buildwithgrove/path/protocol"
)

// Constants for sanction expiration and cache cleanup
const (
	// Default TTL for session-limited sanctions
	// TODO_TECHDEBT(@olshansk): Align with protocol parameters for session length (may change in Shannon)
	defaultSessionSanctionExpiration = 1 * time.Hour

	// Interval for purging expired items from the cache
	// DEV_NOTE: Arbitrarily selected; can be changed as needed
	// TODO_TECHDEBT(@olshansk): Re-evaluate appropriate value
	defaultSanctionCacheCleanupInterval = 10 * time.Minute
)

// sanctionedEndpointsStore:
//   - Tracks sanctioned endpoints
//   - Supports both permanent and session-limited sanctions
//   - Session sanctions expire automatically via go-cache
type sanctionedEndpointsStore struct {
	logger polylog.Logger

	// permanentSanctions:
	//   - In-memory map of endpoints with permanent sanctions
	//   - Persists for process lifetime (not on disk)
	//   - Lost on PATH process restart; not shared across instances
	permanentSanctions      map[protocol.EndpointAddr]sanction
	permanentSanctionsMutex sync.RWMutex

	// sessionSanctionsCache:
	//   - Stores session-limited sanctions (auto-expire)
	//   - Key: app address + session key + endpoint address
	//   - Expire after defaultSessionSanctionExpiration
	//   - Lost on PATH process restart; not shared across instances
	// TODO_IMPROVE(@commoddity): Update this cache to use SturdyC.
	//   - https://github.com/viccon/sturdyc
	sessionSanctionsCache *cache.Cache
}

// newSanctionedEndpointsStore:
//   - Instantiates a new sanctionedEndpointsStore with logging and caches
func newSanctionedEndpointsStore(logger polylog.Logger) *sanctionedEndpointsStore {
	return &sanctionedEndpointsStore{
		logger:                logger,
		permanentSanctions:    make(map[protocol.EndpointAddr]sanction),
		sessionSanctionsCache: cache.New(defaultSessionSanctionExpiration, defaultSanctionCacheCleanupInterval),
	}
}

// ApplyObservations:
//   - Processes all provided observations and applies sanctions as needed
//   - Main public entry point for handling and sanctioning observations
func (ses *sanctionedEndpointsStore) ApplyObservations(shannonObservations []*protocolobservations.ShannonRequestObservations) {
	logger := ses.logger.With("method", "ApplyObservations")

	if len(shannonObservations) == 0 {
		logger.Warn().Msg("Skipping processing: received empty observation list")
		return
	}

	// For each observation set:
	for _, observationSet := range shannonObservations {
		// For each endpoint observation in the set:
		for _, endpointObservation := range observationSet.GetEndpointObservations() {
			// Build endpoint from observation
			endpoint := buildEndpointFromObservation(endpointObservation)

			// Hydrate logger with endpoint context
			logger := hydrateLoggerWithEndpoint(logger, endpoint)
			logger.Debug().Msg("processing endpoint observation.")

			// Skip if no sanction is recommended
			recommendedSanction := endpointObservation.GetRecommendedSanction()
			if recommendedSanction == protocolobservations.ShannonSanctionType_SHANNON_SANCTION_UNSPECIFIED {
				continue
			}

			// Build sanction from observation
			sanctionData := buildSanctionFromObservation(endpointObservation)

			// Apply appropriate type of sanction:
			switch recommendedSanction {
			case protocolobservations.ShannonSanctionType_SHANNON_SANCTION_PERMANENT:
				// Permanent sanction:
				//   - Persists for process lifetime (not on disk)
				//   - Lost on PATH restart; not shared
				logger.Info().Msg("Adding permanent sanction for endpoint")
				ses.addPermanentSanction(endpoint.Addr(), sanctionData)

			case protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION:
				// Session-based sanction:
				//   - Expires after set duration
				//   - More ephemeral than permanent
				//   - Lost on PATH restart; not shared
				logger.Info().Msg("Adding session sanction for endpoint")
				ses.addSessionSanction(endpoint, sanctionData)

			default:
				logger.Warn().Msg("sanction type not supported by the store. skipping.")
			}
		}
	}
}

// FilterSanctionedEndpoints:
//   - Removes sanctioned endpoints from the provided list
//   - Returns only endpoints without active sanctions
//   - Used during endpoint selection to avoid sanctioned endpoints
func (ses *sanctionedEndpointsStore) FilterSanctionedEndpoints(
	allEndpoints map[protocol.EndpointAddr]endpoint,
) map[protocol.EndpointAddr]endpoint {
	filteredEndpoints := make(map[protocol.EndpointAddr]endpoint)

	for endpointAddr, endpoint := range allEndpoints {
		sanctioned, reason := ses.isSanctioned(&endpoint)
		if sanctioned {
			// Log and skip sanctioned endpoints
			hydratedLogger := hydrateLoggerWithEndpoint(ses.logger, &endpoint)
			hydratedLogger.With("sanction_reason", reason).Debug().Msg("Filtering out sanctioned endpoint")
			continue
		}
		filteredEndpoints[endpointAddr] = endpoint
	}

	return filteredEndpoints
}

// addPermanentSanction:
//   - Adds a permanent sanction for an endpoint
//   - Never expires; requires manual removal
//   - Used for serious errors (e.g., validation failures, suspected malicious behavior)
func (ses *sanctionedEndpointsStore) addPermanentSanction(
	endpointAddr protocol.EndpointAddr,
	sanctionData sanction,
) {
	ses.permanentSanctionsMutex.Lock()
	defer ses.permanentSanctionsMutex.Unlock()
	ses.permanentSanctions[endpointAddr] = sanctionData
}

// addSessionSanction:
//   - Adds a session-based sanction for an endpoint
//   - Sanction expires after defaultSessionSanctionExpiration
//   - Used for temporary issues (e.g., timeouts, connection problems)
func (ses *sanctionedEndpointsStore) addSessionSanction(
	endpoint *endpoint,
	sanction sanction,
) {
	key := sessionSanctionKey(endpoint)
	ses.sessionSanctionsCache.Set(key, sanction, defaultSessionSanctionExpiration)
}

// isSanctioned checks if an endpoint has any active sanction (permanent or session-based)
func (ses *sanctionedEndpointsStore) isSanctioned(
	endpoint *endpoint,
) (bool, string) {
	// Check permanent sanctions first - these apply regardless of session
	ses.permanentSanctionsMutex.RLock()
	defer ses.permanentSanctionsMutex.RUnlock()
	sanctionRecord, hasPermanentSanction := ses.permanentSanctions[endpoint.Addr()]

	if hasPermanentSanction {
		return true, fmt.Sprintf("permanent sanction: %s", sanctionRecord.reason)
	}

	// Check session sanctions - these are specific to app+session+endpoint
	key := sessionSanctionKey(endpoint)
	sessionSanctionObj, hasSessionSanction := ses.sessionSanctionsCache.Get(key)
	if hasSessionSanction {
		sanctionRecord := sessionSanctionObj.(sanction)
		return true, fmt.Sprintf("session sanction: %s", sanctionRecord.reason)
	}

	return false, ""
}

// sessionSanctionKey:
//   - Creates a unique key for session-based sanctions
//   - Format: appAddr:sessionID:supplier:endpoint_url
func sessionSanctionKey(
	endpoint *endpoint,
) string {
	// Session header is never nil:
	// Ref: https://github.com/pokt-network/shannon-sdk/blob/64d83f85e7e3f8e7d6ddee98ced276203cf5475f/session.go#L134
	header := endpoint.session.Header

	appAddr := header.ApplicationAddress
	sessionID := header.SessionId
	return fmt.Sprintf(
		"%s:%s:%s:%s",
		appAddr,
		sessionID,
		endpoint.supplier,
		endpoint.url,
	)
}
