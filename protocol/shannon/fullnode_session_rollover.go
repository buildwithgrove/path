package shannon

import (
	"context"
	"sync"
	"time"

	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
)

const (
	// TODO_TECHDEBT(@commoddity): Make this configurable via config file
	// Grace period after session end where rollover issues may occur
	sessionRolloverGracePeriodBlocks = 10

	// How often to check for block height updates
	blockCheckInterval = 15 * time.Second
)

// sessionRolloverState tracks session rollover status for the LazyFullNode.
//
// The rollover window spans from 1 block before session start through
// `sessionRolloverGracePeriodBlocks` after session start. This provides
// early warning and extended monitoring of potentially problematic periods
// around session boundaries.
//
// To handle delays in fetching new sessions, we track both start and end heights:
// - currentSessionStartHeight: Start of the current session
// - currentSessionEndHeight: End of the current session (next session starts at end + 1)
// This allows rollover detection even when new session data is delayed.
//
// All access to this state is protected by rolloverStateMu for thread safety,
// allowing safe concurrent access from multiple goroutines.
type sessionRolloverState struct {
	currentBlockHeight        int64 // Latest block height from the blockchain
	currentSessionStartHeight int64 // Start height of the current session
	currentSessionEndHeight   int64 // End height of the current session
	isInSessionRollover       bool  // Cached rollover status (true = in rollover period)

	rolloverStateMu sync.RWMutex // Protects all fields above
}

// getSessionRolloverState returns whether we're currently in a session rollover period.
// Thread-safe.
func (lfn *LazyFullNode) getSessionRolloverState() bool {
	lfn.rolloverState.rolloverStateMu.RLock()
	defer lfn.rolloverState.rolloverStateMu.RUnlock()
	return lfn.rolloverState.isInSessionRollover
}

// startSessionRolloverMonitoring starts background monitoring for session rollovers.
// Called automatically when LazyFullNode is created.
func (lfn *LazyFullNode) startSessionRolloverMonitoring() {
	lfn.logger.Info().
		Dur("check_interval", blockCheckInterval).
		Int("grace_period_blocks", sessionRolloverGracePeriodBlocks).
		Msg("Starting session rollover monitoring")

	go lfn.monitorLoop()
}

// monitorLoop continuously checks block height to detect session rollovers
func (lfn *LazyFullNode) monitorLoop() {
	ticker := time.NewTicker(blockCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		lfn.updateBlockHeight()
	}
}

// updateSessionStartHeight updates session start and end heights when we fetch a new session.
// Called from GetSession() to keep rollover monitoring current.
func (lfn *LazyFullNode) updateSessionStartHeight(session sessiontypes.Session) {
	if session.Header == nil {
		lfn.logger.Warn().Msg("Session header is nil, cannot update session heights")
		return
	}

	lfn.rolloverState.rolloverStateMu.Lock()
	defer lfn.rolloverState.rolloverStateMu.Unlock()

	newSessionStartHeight := session.Header.SessionStartBlockHeight
	newSessionEndHeight := session.Header.SessionEndBlockHeight
	oldSessionStartHeight := lfn.rolloverState.currentSessionStartHeight

	// Update session heights and recalculate rollover status
	lfn.rolloverState.currentSessionStartHeight = newSessionStartHeight
	lfn.rolloverState.currentSessionEndHeight = newSessionEndHeight
	lfn.rolloverState.isInSessionRollover = lfn.calculateRolloverStatus()

	// Log session changes
	if oldSessionStartHeight != newSessionStartHeight {
		lfn.logger.Debug().
			Int64("session_start_height", newSessionStartHeight).
			Int64("session_end_height", newSessionEndHeight).
			Int64("current_block_height", lfn.rolloverState.currentBlockHeight).
			Bool("in_rollover", lfn.rolloverState.isInSessionRollover).
			Msg("Session heights updated")
	}
}

// updateBlockHeight fetches current block height and recalculates rollover status
func (lfn *LazyFullNode) updateBlockHeight() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	newHeight, err := lfn.GetCurrentBlockHeight(ctx)
	if err != nil {
		lfn.logger.Error().Err(err).Msg("Failed to get current block height")
		return
	}

	lfn.rolloverState.rolloverStateMu.Lock()
	defer lfn.rolloverState.rolloverStateMu.Unlock()

	previousHeight := lfn.rolloverState.currentBlockHeight
	lfn.rolloverState.currentBlockHeight = newHeight
	lfn.rolloverState.isInSessionRollover = lfn.calculateRolloverStatus()

	// Log block height changes
	if previousHeight != newHeight {
		lfn.logger.Debug().
			Int64("current_height", newHeight).
			Int64("session_start_height", lfn.rolloverState.currentSessionStartHeight).
			Int64("session_end_height", lfn.rolloverState.currentSessionEndHeight).
			Bool("in_rollover", lfn.rolloverState.isInSessionRollover).
			Msg("Block height updated")
	}
}

// calculateRolloverStatus determines if we're in a session rollover period using both
// session start and end heights for robust detection even when session data is delayed.
func (lfn *LazyFullNode) calculateRolloverStatus() bool {
	blockHeight := lfn.rolloverState.currentBlockHeight
	sessionEndHeight := lfn.rolloverState.currentSessionEndHeight

	if blockHeight == 0 {
		return false
	}

	// Fallback: If session start is stale or unavailable, use session end height
	// to detect next session rollover. Next session starts at sessionEndHeight + 1
	if sessionEndHeight != 0 && sessionEndHeight > blockHeight {
		return lfn.isInRolloverWindow(blockHeight)
	}

	// No session data available
	return false
}

// isInRolloverWindow checks if blockHeight falls within the rollover window around sessionEndHeight.
// Rollover window = [sessionEndHeight - 1, sessionEndHeight + gracePeriod]
func (lfn *LazyFullNode) isInRolloverWindow(blockHeight int64) bool {
	sessionEndHeight := lfn.rolloverState.currentSessionEndHeight

	if sessionEndHeight == 0 || blockHeight == 0 {
		return false
	}

	rolloverStart := sessionEndHeight - 1
	rolloverEnd := sessionEndHeight + sessionRolloverGracePeriodBlocks

	return blockHeight >= rolloverStart && blockHeight <= rolloverEnd
}
