package shannon

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/metrics/devtools"
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
			logger := hydrateLoggerWithEndpoint(logger, endpoint).With("method", "ApplyObservations")
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
	key := newSanctionKey(endpoint).string()
	ses.sessionSanctionsCache.Set(key, sanction, defaultSessionSanctionExpiration)
}

// isSanctioned checks if an endpoint has any active sanction (permanent or session-based)
func (ses *sanctionedEndpointsStore) isSanctioned(endpoint *endpoint) (bool, string) {
	// Check permanent sanctions first - these apply regardless of session
	ses.permanentSanctionsMutex.RLock()
	defer ses.permanentSanctionsMutex.RUnlock()
	sanctionRecord, hasPermanentSanction := ses.permanentSanctions[endpoint.Addr()]

	if hasPermanentSanction {
		return true, fmt.Sprintf("permanent sanction: %s", sanctionRecord.reason)
	}

	// Check session sanctions - these are specific to app+session+endpoint
	key := newSanctionKey(endpoint).string()
	sessionSanctionObj, hasSessionSanction := ses.sessionSanctionsCache.Get(key)
	if hasSessionSanction {
		sanctionRecord := sessionSanctionObj.(sanction)
		return true, fmt.Sprintf("session sanction: %s", sanctionRecord.reason)
	}

	return false, ""
}

// --------- Session Sanction Key ---------

// TODO_TECHDEBT(@commoddity,adshmh): update this to be composed of only
//     - endpoint URL
//     AND
//     - either supplier address or session ID
// Discord conversation: https://discord.com/channels/824324475256438814/1273320783547990160/1378346761151844465

// sanctionKey:
//   - Creates a unique key for session-based sanctions
//   - Format: "<app_address>:<session_id>:<supplier_address>:<endpoint_url>"
type sanctionKey struct {
	supplier    string
	endpointURL string
	appAddr     string
	sessionID   string
}

func newSanctionKey(endpoint *endpoint) sanctionKey {
	// Session header is never nil:
	// Ref: https://github.com/pokt-network/shannon-sdk/blob/64d83f85e7e3f8e7d6ddee98ced276203cf5475f/session.go#L134
	header := endpoint.session.Header

	appAddr := header.ApplicationAddress
	sessionID := header.SessionId
	return sanctionKey{
		supplier:    endpoint.supplier,
		endpointURL: endpoint.url,
		appAddr:     appAddr,
		sessionID:   sessionID,
	}
}

// sanctionKeyFromCacheKeyString decomposes the cache key into its components
// in order to populate the details map with the data required.
func sanctionKeyFromCacheKeyString(cacheKey string) sanctionKey {
	// Only split for 4 parts, as final part is URL which contains a ":" character.
	parts := strings.SplitN(cacheKey, ":", 4)
	fmt.Println("🚀parts", parts)
	fmt.Println("cacheKey", cacheKey)
	return sanctionKey{
		appAddr:     parts[0],
		sessionID:   parts[1],
		supplier:    parts[2],
		endpointURL: parts[3],
	}
}

// string returns the string representation of the sanction key.
//   - Format: "<app_address>:<session_id>:<supplier_address>:<endpoint_url>"
func (s sanctionKey) string() string {
	return fmt.Sprintf(
		"%s:%s:%s:%s",
		s.appAddr,
		s.sessionID,
		s.supplier,
		s.endpointURL,
	)
}

// endpointAddr returns the endpoint address for the sanction key.
//   - Format: "<supplier_address>-<endpoint_url>"
//   - Example: "pokt13771d0a403a599ee4a3812321e2fabc509e7f3-https://us-west-test-endpoint-1.demo"
func (s sanctionKey) endpointAddr() protocol.EndpointAddr {
	return protocol.EndpointAddr(fmt.Sprintf("%s-%s", s.supplier, s.endpointURL))
}

// --------- Sanction Details ---------

// getSanctionDetails returns the sanctioned endpoints for a given service ID.
// It provides information about:
//   - the currently sanctioned endpoints, including the reason
//   - counts for valid and sanctioned endpoints
//
// It is called by the router to allow quick information about currently sanctioned endpoints.
func (ses *sanctionedEndpointsStore) getSanctionDetails(serviceID protocol.ServiceID) devtools.ProtocolLevelDataResponse {
	permanentSanctionDetails := make(map[protocol.EndpointAddr]devtools.SanctionedEndpoint)
	sessionSanctionDetails := make(map[protocol.EndpointAddr]devtools.SanctionedEndpoint)

	// First get permanent sanctions
	for key, sanction := range ses.permanentSanctions {
		sanctionServiceID := protocol.ServiceID(sanction.sessionServiceID)

		// Only return sanctions for the provided service ID
		// Filter out all sanctions for other service IDs.
		if sanctionServiceID != serviceID {
			continue
		}

		ses.processSanctionIntoDetailsMap(string(key), sanction, permanentSanctionDetails)
	}

	// Then get session sanctions
	for key, cachedSanction := range ses.sessionSanctionsCache.Items() {
		sanction, ok := cachedSanction.Object.(sanction)
		if !ok {
			ses.logger.Error().Msg("SHOULD NEVER HAPPEN: cached sanction is not a sanction")
			continue
		}

		sanctionServiceID := protocol.ServiceID(sanction.sessionServiceID)

		// Only return sanctions for the provided service ID
		// Filter out all sanctions for other service IDs.
		if sanctionServiceID != serviceID {
			continue
		}

		ses.processSanctionIntoDetailsMap(string(key), sanction, sessionSanctionDetails)
	}

	permanentSanctionedEndpointsCount := len(permanentSanctionDetails)
	sessionSanctionedEndpointsCount := len(sessionSanctionDetails)
	totalSanctionedEndpointsCount := permanentSanctionedEndpointsCount + sessionSanctionedEndpointsCount

	return devtools.ProtocolLevelDataResponse{
		PermanentlySanctionedEndpoints:    permanentSanctionDetails,
		SessionSanctionedEndpoints:        sessionSanctionDetails,
		PermamentSanctionedEndpointsCount: permanentSanctionedEndpointsCount,
		SessionSanctionedEndpointsCount:   sessionSanctionedEndpointsCount,
		TotalSanctionedEndpointsCount:     totalSanctionedEndpointsCount,
	}
}

// processSanctionIntoDetailsMap decomposes the sanction key into its components
// in order to populate the details map with the data required for the devtools.ProtocolLevelDataResponse.
func (ses *sanctionedEndpointsStore) processSanctionIntoDetailsMap(
	cacheKey string,
	sanction sanction,
	detailsMap map[protocol.EndpointAddr]devtools.SanctionedEndpoint,
) {
	sanctionKey := sanctionKeyFromCacheKeyString(cacheKey)

	detailsMap[sanctionKey.endpointAddr()] = sanction.toSanctionDetails(
		sanctionKey.supplier,
		sanctionKey.endpointURL,
		sanctionKey.appAddr,
		sanctionKey.sessionID,
		protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION,
	)
}
