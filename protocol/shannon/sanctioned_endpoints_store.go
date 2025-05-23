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

const (
	// defaultSessionSanctionExpiration is the default TTL for session-limited sanctions
	// TODO_TECHDEBT(@adshmh): Align this with protocol specific parameters determining a session
	// length which may change and vary in Shannon.
	defaultSessionSanctionExpiration = 1 * time.Hour

	// defaultSanctionCacheCleanupInterval is how often the cache purges expired items
	// DEV_NOTE: This was arbitrarily selected and can be changed in the future.
	defaultSanctionCacheCleanupInterval = 10 * time.Minute
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
		sessionSanctions:   cache.New(defaultSessionSanctionExpiration, defaultSanctionCacheCleanupInterval),
	}
}

// ApplyObservations processes all observations in the list and applies appropriate sanctions.
// This is the main public method of the sanctioned endpoints store that should be called by external components
// to handle observations and apply sanctions as needed.
func (ses *sanctionedEndpointsStore) ApplyObservations(shannonObservations []*protocolobservations.ShannonRequestObservations) {
	if len(shannonObservations) == 0 {
		ses.logger.With("method", "ApplyObservations").Warn().Msg("Skipping processing: received empty observation list")
		return
	}

	// Process each observation set
	for _, observationSet := range shannonObservations {
		// Process each endpoint observation in the set
		for _, endpointObservation := range observationSet.GetEndpointObservations() {
			// Construct the endpoint specified by the observation.
			endpoint := buildEndpointFromObservation(endpointObservation)

			// Hydrate the logger with the endpoint specified by the observation.
			logger := hydrateLoggerWithEndpoint(ses.logger, endpoint)
			logger.Debug().Msg("processing endpoint observation.")

			// Skip if no sanction is recommended
			recommendedSanction := endpointObservation.GetRecommendedSanction()
			if recommendedSanction == protocolobservations.ShannonSanctionType_SHANNON_SANCTION_UNSPECIFIED {
				continue
			}

			// Create sanction based on the observation
			sanctionData := buildSanctionFromObservation(endpointObservation)

			// Apply the appropriate type of sanction
			switch recommendedSanction {

			// Permanent sanction: these persist for the lifetime of the process but are not persisted to disk.
			// They will be lost on PATH process restart and are not shared across multiple instances.
			case protocolobservations.ShannonSanctionType_SHANNON_SANCTION_PERMANENT:
				logger.Info().Msg("Adding permanent sanction for endpoint")

				ses.addPermanentSanction(
					endpoint.Addr(),
					sanctionData,
				)

			// Session-based sanctions: These automatically expire after a set duration.
			// They are more ephemeral than permanent session, and are also
			// lost on PATH process restart and are not shared across multiple instances.
			case protocolobservations.ShannonSanctionType_SHANNON_SANCTION_SESSION:
				logger.Info().Msg("Adding session sanction for endpoint")

				ses.addSessionSanction(
					endpoint,
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
	allEndpoints map[protocol.EndpointAddr]endpoint,
) map[protocol.EndpointAddr]endpoint {
	filteredEndpoints := make(map[protocol.EndpointAddr]endpoint)

	for endpointAddr, endpoint := range allEndpoints {
		sanctioned, reason := ses.isSanctioned(&endpoint)
		if sanctioned {
			hydratedLogger := hydrateLoggerWithEndpoint(ses.logger, &endpoint)
			hydratedLogger.With("sanction_reason", reason).Debug().Msg("Filtering out sanctioned endpoint")
			continue
		}
		filteredEndpoints[endpointAddr] = endpoint
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
	endpoint *endpoint,
	sanction sanction,
) {
	key := sessionSanctionKey(endpoint)
	ses.sessionSanctions.Set(key, sanction, defaultSessionSanctionExpiration)
}

// isSanctioned checks if an endpoint has an active sanction.
// It checks both permanent and session-based sanctions.
func (ses *sanctionedEndpointsStore) isSanctioned(
	endpoint *endpoint,
) (bool, string) {
	// Check permanent sanctions first - these apply regardless of session
	ses.mu.RLock()
	defer ses.mu.RUnlock()
	sanctionRecord, hasPermanentSanction := ses.permanentSanctions[endpoint.Addr()]

	if hasPermanentSanction {
		return true, fmt.Sprintf("permanent sanction: %s", sanctionRecord.reason)
	}

	// Check session sanctions - these are specific to app+session+endpoint
	key := sessionSanctionKey(endpoint)
	sessionSanctionObj, hasSessionSanction := ses.sessionSanctions.Get(key)
	if hasSessionSanction {
		sanctionRecord := sessionSanctionObj.(sanction)
		return true, fmt.Sprintf("session sanction: %s", sanctionRecord.reason)
	}

	return false, ""
}

// sessionSanctionKey creates a unique key for session-based sanctions
// Format: appAddr:sessionID:supplier:endpoint_url
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
