package morse

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

const (
	// defaultSessionSanctionExpiration is the default TTL for session-limited sanctions
	// TODO_TECHDEBT(@adshmh): Align this with protocol specific parameters determining a session
	// length which may change and vary in Shannon.
	defaultSessionSanctionExpiration = 1 * time.Hour

	// defaultCacheCleanupInterval is how often the cache purges expired items
	// DEV_NOTE: This was arbitrarily selected and can be changed in the future.
	defaultCacheCleanupInterval = 10 * time.Minute
)

// sanctionedEndpointsStore maintains records of endpoints that have been sanctioned
// It provides both permanent sanctions and time-limited session-based sanctions
// with automatic expiration through go-cache.
type sanctionedEndpointsStore struct {
	logger polylog.Logger

	// Mutex for synchronized access to the permanent sanctions map
	mu sync.RWMutex

	// permanentSanctions is an in-memory map storing endpoints with permanent sanctions.
	// These sanctions persist for the lifetime of the process but are not persisted to disk.
	// They will be lost on PATH process restart and are not shared across multiple instances.
	permanentSanctions map[protocol.EndpointAddr]sanction

	// sessionSanctions stores endpoints with session-limited sanctions
	// The key is a combination of app address, session key, and endpoint address
	// These sanctions automatically expire after defaultSessionSanctionExpiration
	// They will be lost on PATH process restart and are not shared across multiple instances.
	sessionSanctions *cache.Cache
}

// newSanctionedEndpointsStore creates a new sanctionedEndpointsStore
func newSanctionedEndpointsStore(logger polylog.Logger) *sanctionedEndpointsStore {
	return &sanctionedEndpointsStore{
		logger:             logger,
		permanentSanctions: make(map[protocol.EndpointAddr]sanction),
		sessionSanctions:   cache.New(defaultSessionSanctionExpiration, defaultCacheCleanupInterval),
	}
}

// ApplyObservations processes all observations in the list and applies appropriate sanctions.
// This is the main public method of the sanctioned endpoints store that should be called by external components
// to handle observations and apply sanctions as needed.
func (ses *sanctionedEndpointsStore) ApplyObservations(morseObservations []*protocolobservations.MorseRequestObservations) {
	if len(morseObservations) == 0 {
		ses.logger.With("method", "ApplyObservations").Warn().Msg("Skipping processing: received empty observation list")
		return
	}

	// Process each observation set
	for _, observationSet := range morseObservations {
		// Process each endpoint observation in the set
		for _, endpointObservation := range observationSet.GetEndpointObservations() {
			logger := loggerWithEndpointObservation(ses.logger, endpointObservation)
			logger.Debug().Msg("processing endpoint observation.")

			// Skip if no sanction is recommended
			recommendedSanction := endpointObservation.GetRecommendedSanction()
			if recommendedSanction == protocolobservations.MorseSanctionType_MORSE_SANCTION_UNSPECIFIED {
				continue
			}

			// Create sanction based on the observation
			sanctionData := buildSanctionFromObservation(endpointObservation)

			// Apply the appropriate type of sanction
			switch recommendedSanction {

			// Permanent sanction: these persist for the lifetime of the process but are not persisted to disk.
			// They will be lost on PATH process restart and are not shared across multiple instances.
			case protocolobservations.MorseSanctionType_MORSE_SANCTION_PERMANENT:
				logger.Info().Msg("Adding permanent sanction for endpoint")

				ses.addPermanentSanction(
					protocol.EndpointAddr(endpointObservation.GetEndpointAddr()),
					sanctionData,
				)

			// Session-based sanctions: These automatically expire after a set duration.
			// They are more ephemeral than permanent session, and are also
			// lost on PATH process restart and are not shared across multiple instances.
			case protocolobservations.MorseSanctionType_MORSE_SANCTION_SESSION:
				logger.Info().Msg("Adding session sanction for endpoint")

				ses.addSessionSanction(
					protocol.EndpointAddr(endpointObservation.GetEndpointAddr()),
					endpointObservation.GetAppAddress(),
					endpointObservation.GetSessionKey(),
					sanctionData,
				)
			default:
				logger.Warn().Msg("sanction type not supported by the store. skipping.")
			}
		}
	}
}

// FilterSanctionedEndpoints removes sanctioned endpoints from the provided list.
// It returns a filtered list containing only endpoints that don't have active sanctions.
// This method is used during endpoint selection to avoid selecting endpoints that have been sanctioned.
func (ses *sanctionedEndpointsStore) FilterSanctionedEndpoints(
	allEndpoints []endpoint,
	appAddr string,
	sessionKey string,
) []endpoint {
	var filteredEndpoints []endpoint

	for _, endpoint := range allEndpoints {
		sanctioned, reason := ses.isSanctioned(endpoint.Addr(), appAddr, sessionKey)
		if sanctioned {
			logger := loggerWithEndpoint(ses.logger, appAddr, sessionKey, endpoint.Addr(), reason)
			logger.Debug().Msg("Filtering out sanctioned endpoint")
			continue
		}
		filteredEndpoints = append(filteredEndpoints, endpoint)
	}

	return filteredEndpoints
}

// addPermanentSanction adds a permanent sanction for an endpoint.
// Permanent sanctions never expire and require manual intervention to remove.
// These are used for serious errors like validation failures that may indicate malicious behavior.
func (ses *sanctionedEndpointsStore) addPermanentSanction(
	endpointAddr protocol.EndpointAddr,
	sanctionData sanction,
) {
	ses.mu.Lock()
	ses.permanentSanctions[endpointAddr] = sanctionData
	ses.mu.Unlock()
}

// addSessionSanction adds a session-based sanction for an endpoint.
// Session-based sanctions automatically expire after defaultSessionSanctionExpiration.
// These are used for temporary issues like timeouts or connection problems.
func (ses *sanctionedEndpointsStore) addSessionSanction(
	endpointAddr protocol.EndpointAddr,
	appAddr string,
	sessionKey string,
	sanctionData sanction,
) {
	key := newSessionSanctionKey(appAddr, sessionKey, endpointAddr)
	ses.sessionSanctions.Set(key.string(), sanctionData, defaultSessionSanctionExpiration)
}

// isSanctioned checks if an endpoint has an active sanction.
// It checks both permanent and session-based sanctions.
func (ses *sanctionedEndpointsStore) isSanctioned(
	endpointAddr protocol.EndpointAddr,
	appAddr string,
	sessionKey string,
) (bool, string) {
	// Check permanent sanctions first - these apply regardless of session
	ses.mu.RLock()
	defer ses.mu.RUnlock()
	sanctionRecord, hasPermanentSanction := ses.permanentSanctions[endpointAddr]

	if hasPermanentSanction {
		return true, fmt.Sprintf("permanent sanction: %s", sanctionRecord.reason)
	}

	// Check session sanctions - these are specific to app+session+endpoint
	key := newSessionSanctionKey(appAddr, sessionKey, endpointAddr)
	sessionSanctionObj, hasSessionSanction := ses.sessionSanctions.Get(key.string())
	if hasSessionSanction {
		sanctionRecord := sessionSanctionObj.(sanction)
		return true, fmt.Sprintf("session sanction: %s", sanctionRecord.reason)
	}

	return false, ""
}

// --------- Session Sanction Key ---------

// sessionSanctionKey creates a unique key for session-based sanctions
// Format: app_addr:session_key:endpoint_addr
type sessionSanctionKey struct {
	appAddr      string
	sessionKey   string
	endpointAddr protocol.EndpointAddr
}

// TODO_MVP(@adshmh): return an error if any of the composite key components are empty.
func newSessionSanctionKey(appAddr, sessionKey string, endpointAddr protocol.EndpointAddr) sessionSanctionKey {
	return sessionSanctionKey{
		appAddr:      appAddr,
		sessionKey:   sessionKey,
		endpointAddr: endpointAddr,
	}
}

func newSessionSanctionKeyFromCacheKey(cacheKey string) sessionSanctionKey {
	parts := strings.Split(cacheKey, ":")
	return sessionSanctionKey{
		appAddr:      parts[0],
		sessionKey:   parts[1],
		endpointAddr: protocol.EndpointAddr(parts[2]),
	}
}

func (s sessionSanctionKey) string() string {
	return fmt.Sprintf("%s:%s:%s", s.appAddr, s.sessionKey, s.endpointAddr)
}

// --------- Sanction Details ---------

// getSanctionDetails returns the sanctioned endpoints for a given service ID.
// It provides information about:
//   - the currently sanctioned endpoints, including the reason
//   - counts for valid and sanctioned endpoints
//
// It is called by the router to allow quick information about currently sanctioned endpoints.
func (ses *sanctionedEndpointsStore) getSanctionDetails(serviceID protocol.ServiceID) devtools.ProtocolLevelDataResponse {
	permanentSanctionDetails := make(map[protocol.EndpointAddr]devtools.DisqualifiedEndpoint)
	permanentSanctionedEndpointsCount := 0

	sessionSanctionDetails := make(map[protocol.EndpointAddr]devtools.DisqualifiedEndpoint)
	sessionSanctionedEndpointsCount := 0

	// First get permanent sanctions
	for endpointAddr, sanction := range ses.permanentSanctions {
		sanctionServiceID := protocol.ServiceID(sanction.sessionServiceID)

		// If serviceID is provided, skip sanctions for other service IDs
		if serviceID != "" && sanctionServiceID != serviceID {
			continue
		}

		sanctionDetails := sanction.toSanctionDetails(
			endpointAddr, protocolobservations.MorseSanctionType_MORSE_SANCTION_PERMANENT,
		)

		permanentSanctionDetails[endpointAddr] = sanctionDetails
		permanentSanctionedEndpointsCount++
	}

	// Then get session sanctions
	for keyStr, cachedSanction := range ses.sessionSanctions.Items() {
		key := newSessionSanctionKeyFromCacheKey(keyStr)

		sanction, ok := cachedSanction.Object.(sanction)
		if !ok {
			ses.logger.Warn().Msg("cached sanction is not a sanction")
			continue
		}

		sanctionServiceID := protocol.ServiceID(sanction.sessionServiceID)

		// If serviceID is provided, skip sanctions for other service IDs
		if serviceID != "" && sanctionServiceID != serviceID {
			continue
		}

		sanctionDetails := sanction.toSanctionDetails(
			key.endpointAddr, protocolobservations.MorseSanctionType_MORSE_SANCTION_SESSION,
		)

		sessionSanctionDetails[key.endpointAddr] = sanctionDetails
		sessionSanctionedEndpointsCount++
	}

	return devtools.ProtocolLevelDataResponse{
		PermanentlySanctionedEndpoints:    permanentSanctionDetails,
		SessionSanctionedEndpoints:        sessionSanctionDetails,
		SanctionedEndpointsCount:          permanentSanctionedEndpointsCount + sessionSanctionedEndpointsCount,
		PermamentSanctionedEndpointsCount: permanentSanctionedEndpointsCount,
		SessionSanctionedEndpointsCount:   sessionSanctionedEndpointsCount,
	}
}
