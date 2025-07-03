package shannon

import (
	"context"
	"maps"
	"sync"
	"time"

	"github.com/buildwithgrove/path/protocol"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	"github.com/viccon/sturdyc"
)

const (
	// Network timing constants
	blockTime            = 30 * time.Second // Estimated time per block
	blockPollingInterval = 1 * time.Second  // Polling frequency during intensive monitoring
	retryBaseDelay       = 100 * time.Millisecond
)

// sessionKeyInfo holds metadata for active sessions during background refresh
type sessionKeyInfo struct {
	serviceID protocol.ServiceID
	appAddr   string
}

// sessionRefreshState holds the state for block-based session monitoring
//
// Session refresh lifecycle:
//  1. Normal monitoring: Check every 15 seconds
//  2. Intensive polling: 1-second checks starting at SessionEndBlockHeight
//  3. Cache refresh: Triggered at SessionEndBlockHeight+1
//  4. Background refresh: New cache populated while old cache serves requests
//  5. Atomic swap: New cache replaces old cache with zero downtime
//
// Documentation: https://github.com/viccon/sturdyc
type sessionRefreshState struct {
	currentSessionEndHeight int64
	sessionEndHeightMu      sync.RWMutex

	blockMonitorMu sync.Mutex
	isMonitoring   bool

	// Active session tracking for background refresh
	activeSessionKeys map[string]sessionKeyInfo
	activeSessionMu   sync.RWMutex
}

// ============================================================================
// Session Monitoring Lifecycle
// ============================================================================

// startSessionMonitoring begins background block monitoring
func (cfn *cachingFullNode) startSessionMonitoring() {
	cfn.logger.Debug().
		Dur("check_interval", blockTime/2).
		Msg("Starting session monitoring background process")

	go cfn.monitorBlockHeights()
}

// updateSessionEndHeight updates the global session end height from a fetched session
func (cfn *cachingFullNode) updateSessionEndHeight(session sessiontypes.Session) {
	if session.Header == nil {
		cfn.logger.Warn().Msg("Session header is nil, cannot update session end height")
		return
	}

	sessionEndHeight := session.Header.SessionEndBlockHeight

	cfn.sessionRefreshState.sessionEndHeightMu.Lock()
	defer cfn.sessionRefreshState.sessionEndHeightMu.Unlock()

	// Get the previous session end height
	previousHeight := cfn.sessionRefreshState.currentSessionEndHeight

	// Update the current session end height
	cfn.sessionRefreshState.currentSessionEndHeight = sessionEndHeight

	if previousHeight != sessionEndHeight {
		cfn.logger.Debug().
			Int64("previous_session_end_height", previousHeight).
			Int64("new_session_end_height", sessionEndHeight).
			Msg("Updated session end height for monitoring")
	}
}

// ============================================================================
// Background Monitoring Logic
// ============================================================================

// monitorBlockHeights runs the main monitoring loop that checks every 15 seconds
func (cfn *cachingFullNode) monitorBlockHeights() {
	checkCount := 0
	for {
		time.Sleep(blockTime / 2) // Check every 15 seconds
		checkCount++

		cfn.logger.Debug().
			Int("check_count", checkCount).
			Msg("Background monitoring check")

		cfn.checkAndHandleSessionRefresh()
	}
}

// checkAndHandleSessionRefresh determines if session refresh is needed and takes action
func (cfn *cachingFullNode) checkAndHandleSessionRefresh() {
	targetHeight := cfn.getCurrentSessionEndHeight()
	if targetHeight == 0 {
		cfn.logger.Debug().Msg("No sessions to monitor yet")
		return
	}

	currentHeight, err := cfn.getCurrentBlockHeight()
	if err != nil {
		cfn.logger.Error().
			Err(err).
			Int64("session_end_height", targetHeight).
			Msg("Failed to get current block height")
		return
	}

	cfn.logger.Debug().
		Int64("current_height", currentHeight).
		Int64("session_end_height", targetHeight).
		Int64("blocks_until_session_end", targetHeight-currentHeight).
		Msg("Checking session end proximity")

	// Refresh immediately if we're past SessionEndBlockHeight + 1
	if currentHeight >= targetHeight+1 {
		cfn.logger.Debug().
			Int64("current_height", currentHeight).
			Int64("session_end_height", targetHeight).
			Msg("SessionEndBlockHeight+1 reached, refreshing sessions")

		cfn.refreshAllSessions()
		cfn.resetMonitoring()
		return
	}

	// Start intensive polling if we've reached SessionEndBlockHeight
	if currentHeight >= targetHeight && cfn.tryStartIntensiveMonitoring() {
		cfn.logger.Debug().
			Int64("session_end_height", targetHeight).
			Msg("Starting intensive polling for SessionEndBlockHeight+1")

		cfn.runIntensivePolling(targetHeight)
	}
}

// tryStartIntensiveMonitoring attempts to start intensive monitoring (prevents duplicates)
func (cfn *cachingFullNode) tryStartIntensiveMonitoring() bool {
	cfn.sessionRefreshState.blockMonitorMu.Lock()
	defer cfn.sessionRefreshState.blockMonitorMu.Unlock()

	if cfn.sessionRefreshState.isMonitoring {
		return false // Already monitoring intensively
	}

	cfn.sessionRefreshState.isMonitoring = true
	return true
}

// runIntensivePolling performs 1-second polling until SessionEndBlockHeight+1 is reached
func (cfn *cachingFullNode) runIntensivePolling(targetHeight int64) {
	ticker := time.NewTicker(blockPollingInterval)
	defer ticker.Stop()

	pollCount := 0
	for {
		<-ticker.C
		pollCount++

		if cfn.shouldRefreshNow(targetHeight, pollCount) {
			cfn.refreshAllSessions()
			cfn.resetMonitoring()
			return
		}
	}
}

// shouldRefreshNow checks if we've reached SessionEndBlockHeight+1 during intensive polling
func (cfn *cachingFullNode) shouldRefreshNow(targetHeight int64, pollCount int) bool {
	currentHeight, err := cfn.getCurrentBlockHeight()
	if err != nil {
		cfn.logger.Error().
			Err(err).
			Int64("session_end_height", targetHeight).
			Int("poll_count", pollCount).
			Msg("Failed to get current block height during intensive polling")
		return false
	}

	cfn.logger.Debug().
		Int64("current_height", currentHeight).
		Int64("target_refresh_height", targetHeight+1).
		Int64("blocks_until_refresh", (targetHeight+1)-currentHeight).
		Int("poll_count", pollCount).
		Msg("Intensive polling check")

	if currentHeight >= targetHeight+1 {
		cfn.logger.Debug().
			Int64("current_height", currentHeight).
			Int64("session_end_height", targetHeight).
			Int("total_polls", pollCount).
			Msg("SessionEndBlockHeight+1 reached, refreshing sessions")
		return true
	}

	return false
}

// refreshAllSessions triggers background refresh of all active sessions
func (cfn *cachingFullNode) refreshAllSessions() {
	activeKeys := cfn.getActiveSessionKeys()
	if len(activeKeys) == 0 {
		cfn.logger.Debug().Msg("No active sessions to refresh")
		return
	}

	cfn.logger.Info().
		Int("session_count", len(activeKeys)).
		Msg("Starting background session refresh")

	cfn.refreshSessionsInBackground(activeKeys)
}

// resetMonitoring resets monitoring state for the next session cycle
func (cfn *cachingFullNode) resetMonitoring() {
	cfn.sessionRefreshState.blockMonitorMu.Lock()
	defer cfn.sessionRefreshState.blockMonitorMu.Unlock()

	cfn.sessionRefreshState.sessionEndHeightMu.Lock()
	defer cfn.sessionRefreshState.sessionEndHeightMu.Unlock()

	cfn.sessionRefreshState.isMonitoring = false
	cfn.sessionRefreshState.currentSessionEndHeight = 0

	cfn.logger.Debug().Msg("Reset session monitoring for next cycle")
}

// ============================================================================
// Session Key Tracking - For background refresh
// ============================================================================

// trackActiveSession records a session key for background refresh tracking
func (cfn *cachingFullNode) trackActiveSession(sessionKey string, serviceID protocol.ServiceID, appAddr string) {
	cfn.sessionRefreshState.activeSessionMu.Lock()
	defer cfn.sessionRefreshState.activeSessionMu.Unlock()

	cfn.sessionRefreshState.activeSessionKeys[sessionKey] = sessionKeyInfo{
		serviceID: serviceID,
		appAddr:   appAddr,
	}

	cfn.logger.Debug().
		Str("session_key", sessionKey).
		Msg("Tracking session for background refresh")
}

// getActiveSessionKeys returns a thread-safe copy of all active session keys
func (cfn *cachingFullNode) getActiveSessionKeys() map[string]sessionKeyInfo {
	cfn.sessionRefreshState.activeSessionMu.RLock()
	defer cfn.sessionRefreshState.activeSessionMu.RUnlock()

	activeKeys := make(map[string]sessionKeyInfo, len(cfn.sessionRefreshState.activeSessionKeys))
	maps.Copy(activeKeys, cfn.sessionRefreshState.activeSessionKeys)

	return activeKeys
}

// refreshSessionsInBackground creates a new cache with fresh sessions and atomically swaps it
func (cfn *cachingFullNode) refreshSessionsInBackground(activeKeys map[string]sessionKeyInfo) {
	go func() {
		cfn.logger.Debug().
			Int("session_count", len(activeKeys)).
			Msg("Creating new cache with fresh sessions")

		// Create a new empty cache to populate with fresh sessions
		newCache := getCache[sessiontypes.Session]()

		// Populate the new cache with fresh sessions based on the active keys
		successCount, errorCount := cfn.populateNewCache(newCache, activeKeys)

		// Atomically swap to the new cache
		cfn.sessionCacheMu.Lock()
		cfn.sessionCache = newCache
		cfn.sessionCacheMu.Unlock()

		cfn.logger.Info().
			Int("total_sessions", len(activeKeys)).
			Int("successful_refreshes", successCount).
			Int("failed_refreshes", errorCount).
			Msg("Background session refresh completed")
	}()
}

// populateNewCache fetches all sessions concurrently and populates the new cache
func (cfn *cachingFullNode) populateNewCache(newCache *sturdyc.Client[sessiontypes.Session], activeKeys map[string]sessionKeyInfo) (int, int) {
	var wg sync.WaitGroup
	var successCount, errorCount int
	var countMu sync.Mutex

	for sessionKey, keyInfo := range activeKeys {
		wg.Add(1)
		go func(key string, info sessionKeyInfo) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			_, err := newCache.GetOrFetch(ctx, key, func(fetchCtx context.Context) (sessiontypes.Session, error) {
				session, err := cfn.onchainDataFetcher.GetSession(fetchCtx, info.serviceID, info.appAddr)
				if err == nil {
					cfn.updateSessionEndHeight(session)
				}
				return session, err
			})

			countMu.Lock()
			if err != nil {
				errorCount++
				cfn.logger.Warn().Err(err).Str("session_key", key).Msg("Failed to refresh session")
			} else {
				successCount++
			}
			countMu.Unlock()
		}(sessionKey, keyInfo)
	}

	wg.Wait()
	return successCount, errorCount
}

// ============================================================================
// HELPER METHODS - Thread-safe getters and utilities
// ============================================================================

// getCurrentSessionEndHeight safely gets the current session end height
func (cfn *cachingFullNode) getCurrentSessionEndHeight() int64 {
	cfn.sessionRefreshState.sessionEndHeightMu.RLock()
	defer cfn.sessionRefreshState.sessionEndHeightMu.RUnlock()
	return cfn.sessionRefreshState.currentSessionEndHeight
}

// getCurrentBlockHeight gets the current blockchain height with a 10-second timeout
func (cfn *cachingFullNode) getCurrentBlockHeight() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return cfn.onchainDataFetcher.LatestBlockHeight(ctx)
}
