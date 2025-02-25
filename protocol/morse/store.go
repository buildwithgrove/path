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
	defaultSessionSanctionExpiration = 1 * time.Hour

	// defaultCacheCleanupInterval is how often the cache purges expired items
	defaultCacheCleanupInterval = 10 * time.Minute
)

// sessionSanctionKey creates a unique key for session-based sanctions
// Format: app_addr:session_key:endpoint_addr
func sessionSanctionKey(appAddress, sessionKey, endpointAddr string) string {
	return fmt.Sprintf("%s:%s:%s", appAddress, sessionKey, endpointAddr)
}

// EndpointStore maintains sanctions for Morse endpoints
// It provides both permanent sanctions and time-limited session-based sanctions
// with automatic expiration through go-cache.
type EndpointStore struct {
	logger polylog.Logger

	// Mutex for synchronized access to the permanent sanctions map
	mu sync.RWMutex

	// permanentSanctions stores endpoints with permanent sanctions
	// These sanctions never expire and will require manual intervention to remove
	permanentSanctions map[protocol.EndpointAddr]sanction

	// sessionSanctions stores endpoints with session-limited sanctions
	// The key is a combination of app address, session key, and endpoint address
	// These sanctions automatically expire after defaultSessionSanctionExpiration
	sessionSanctions *cache.Cache
}

// NewEndpointStore creates a new EndpointStore
func NewEndpointStore(logger polylog.Logger) *EndpointStore {
	return &EndpointStore{
		logger:             logger,
		permanentSanctions: make(map[protocol.EndpointAddr]sanction),
		sessionSanctions:   cache.New(defaultSessionSanctionExpiration, defaultCacheCleanupInterval),
	}
}

// AddSanction adds a sanction for an endpoint based on observed behavior
// Two types of sanctions can be added:
//  1. Permanent sanctions: These never expire and require manual intervention to remove.
//     Used for serious errors like validation failures that may indicate malicious behavior.
//  2. Session-based sanctions: These automatically expire after a set duration (1 hour by default).
//     Used for temporary issues like timeouts or connection problems.
func (es *EndpointStore) AddSanction(
	endpointAddr protocol.EndpointAddr,
	appAddress string,
	sessionKey string,
	errorType protocolobservations.MorseEndpointErrorType,
	sanctionType protocolobservations.MorseSanctionType,
	reason string,
	sessionChain string,
	sessionHeight int,
) {
	// Create a sanction record with detailed information
	newSanction := sanction{
		Type:          sanctionType,
		Reason:        reason,
		ErrorType:     errorType,
		CreatedAt:     time.Now(),
		SessionChain:  sessionChain,
		SessionHeight: sessionHeight,
	}

	logger := es.logger.With(
		"endpoint", string(endpointAddr),
		"app_addr", appAddress,
		"reason", reason,
		"session_chain", sessionChain,
		"session_height", sessionHeight,
		"session_key", sessionKey,
	)

	switch sanctionType {
	case protocolobservations.MorseSanctionType_MORSE_SANCTION_PERMANENT:
		// Permanent sanctions are stored in a map and never expire
		// They require manual intervention to remove
		es.mu.Lock()
		es.permanentSanctions[endpointAddr] = newSanction
		es.mu.Unlock()
		logger.Info().Msg("Added permanent sanction for endpoint")

	// TODO_MVP(@adshmh): validate the session key format. Should not create a sanction entry if the session key is not valid.
	case protocolobservations.MorseSanctionType_MORSE_SANCTION_SESSION:
		// Session-based sanctions automatically expire after defaultSessionSanctionExpiration
		// They are specific to a combination of app, session, and endpoint
		key := sessionSanctionKey(appAddress, sessionKey, string(endpointAddr))
		es.sessionSanctions.Set(key, newSanction, defaultSessionSanctionExpiration)
		logger.Info().Msg("Added session sanction for endpoint")
	default:
		logger.With("sanction_type", sanctionType).Info().Msg("sanction type not supported by the store. skipping.")
	}
}

// IsSanctioned checks if an endpoint has an active sanction
// It checks both permanent and session-based sanctions
func (es *EndpointStore) IsSanctioned(
	endpointAddr protocol.EndpointAddr,
	appAddress string,
	sessionKey string,
) (bool, string) {
	// Check permanent sanctions first - these apply regardless of session
	es.mu.RLock()
	sanctionRecord, hasPermanentSanction := es.permanentSanctions[endpointAddr]
	es.mu.RUnlock()

	if hasPermanentSanction {
		return true, fmt.Sprintf("permanent sanction: %s", sanctionRecord.Reason)
	}

	// Check session sanctions - these are specific to app+session+endpoint
	key := sessionSanctionKey(appAddress, sessionKey, string(endpointAddr))
	sessionSanctionObj, found := es.sessionSanctions.Get(key)
	if found {
		sanctionRecord := sessionSanctionObj.(sanction)
		return true, fmt.Sprintf("session sanction: %s", sanctionRecord.Reason)
	}

	return false, ""
}

// FilterSanctionedEndpoints removes sanctioned endpoints from the provided list
// It returns a filtered list containing only endpoints that don't have active sanctions
// This is used during endpoint selection to avoid selecting endpoints that have been sanctioned
func (es *EndpointStore) FilterSanctionedEndpoints(
	allEndpoints []endpoint,
	appAddress string,
	sessionKey string,
) []endpoint {
	var filteredEndpoints []endpoint

	for _, endpoint := range allEndpoints {
		sanctioned, reason := es.IsSanctioned(endpoint.Addr(), appAddress, sessionKey)
		if sanctioned {
			es.logger.Debug().
				Str("endpoint", string(endpoint.Addr())).
				Str("reason", reason).
				Str("app", appAddress).
				Str("session_key", sessionKey).
				Msg("Filtering out sanctioned endpoint")
			continue
		}
		filteredEndpoints = append(filteredEndpoints, endpoint)
	}

	return filteredEndpoints
}
