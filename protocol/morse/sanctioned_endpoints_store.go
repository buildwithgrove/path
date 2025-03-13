package morse

import (
	"fmt"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pokt-network/poktroll/pkg/polylog"

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
	// TODO_MVP(@adshmh): validate the session key format. Should not create a sanction entry if the session key is not valid.
	key := sessionSanctionKey(appAddr, sessionKey, string(endpointAddr))
	ses.sessionSanctions.Set(key, sanctionData, defaultSessionSanctionExpiration)
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
	key := sessionSanctionKey(appAddr, sessionKey, string(endpointAddr))
	sessionSanctionObj, hasSessionSanction := ses.sessionSanctions.Get(key)
	if hasSessionSanction {
		sanctionRecord := sessionSanctionObj.(sanction)
		return true, fmt.Sprintf("session sanction: %s", sanctionRecord.reason)
	}

	return false, ""
}

// sessionSanctionKey creates a unique key for session-based sanctions
// Format: app_addr:session_key:endpoint_addr
func sessionSanctionKey(appAddr, sessionKey, endpointAddr string) string {
	return fmt.Sprintf("%s:%s:%s", appAddr, sessionKey, endpointAddr)
}
