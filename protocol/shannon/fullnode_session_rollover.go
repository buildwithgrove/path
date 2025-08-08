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
// The rollover window spans from 1 block before session end through
// sessionRolloverGracePeriodBlocks after session end. This provides
// early warning and extended monitoring of potentially problematic periods.
//
// All access to this state is protected by rolloverStateMu for thread safety,
// allowing safe concurrent access from multiple goroutines.
type sessionRolloverState struct {
	currentBlockHeight       int64 // Latest block height from the blockchain
	currentSessionEndHeight  int64 // End height of the current/latest session
	previousSessionEndHeight int64 // End height of the previous session (for rollover checking)
	isInSessionRollover      bool  // Cached rollover status (true = in rollover period)

	rolloverStateMu sync.RWMutex // Protects all fields above
}

// getSessionRolloverState returns whether we're currently in a session rollover period.
// Thread-safe.
func (lfn *LazyFullNode) getSessionRolloverState() bool {
	lfn.rolloverState.rolloverStateMu.RLock()
	defer lfn.rolloverState.rolloverStateMu.RUnlock()
	return lfn.rolloverState.isInSessionRollover
}

// updateSessionEndHeight updates session end height when we fetch a new session.
// Called from GetSession() to keep rollover monitoring current.
func (lfn *LazyFullNode) updateSessionEndHeight(session sessiontypes.Session) {
	if session.Header == nil {
		lfn.logger.Warn().Msg("Session header is nil, cannot update session end height")
		return
	}

	newSessionEndHeight := session.Header.SessionEndBlockHeight

	lfn.rolloverState.rolloverStateMu.Lock()
	defer lfn.rolloverState.rolloverStateMu.Unlock()

	oldSessionEndHeight := lfn.rolloverState.currentSessionEndHeight

	// Update session end heights
	if oldSessionEndHeight != 0 && oldSessionEndHeight != newSessionEndHeight {
		// We have a new session - the old current becomes the previous
		lfn.rolloverState.previousSessionEndHeight = oldSessionEndHeight
	}
	lfn.rolloverState.currentSessionEndHeight = newSessionEndHeight

	// For rollover checking, use the previous session if we have one, otherwise use current
	rolloverCheckHeight := lfn.rolloverState.previousSessionEndHeight
	if rolloverCheckHeight == 0 {
		rolloverCheckHeight = newSessionEndHeight
	}

	lfn.rolloverState.isInSessionRollover = lfn.isInRolloverWindow(
		lfn.rolloverState.currentBlockHeight,
		rolloverCheckHeight,
	)

	// Log session changes
	if oldSessionEndHeight != newSessionEndHeight {
		lfn.logger.Debug().
			Int64("previous_session_end_height", lfn.rolloverState.previousSessionEndHeight).
			Int64("new_session_end_height", newSessionEndHeight).
			Int64("current_block_height", lfn.rolloverState.currentBlockHeight).
			Int64("rollover_check_height", rolloverCheckHeight).
			Bool("in_rollover", lfn.rolloverState.isInSessionRollover).
			Msg("Session end height updated")
	}
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

	// For rollover checking, use the previous session if we have one, otherwise use current
	rolloverCheckHeight := lfn.rolloverState.previousSessionEndHeight
	if rolloverCheckHeight == 0 {
		rolloverCheckHeight = lfn.rolloverState.currentSessionEndHeight
	}

	// Recalculate rollover status with new block height
	lfn.rolloverState.isInSessionRollover = lfn.isInRolloverWindow(
		newHeight,
		rolloverCheckHeight,
	)

	// Log block height changes
	if previousHeight != newHeight {
		lfn.logger.Debug().
			Int64("current_height", newHeight).
			Int64("session_end_height", lfn.rolloverState.currentSessionEndHeight).
			Int64("previous_session_end_height", lfn.rolloverState.previousSessionEndHeight).
			Int64("rollover_check_height", rolloverCheckHeight).
			Bool("in_rollover", lfn.rolloverState.isInSessionRollover).
			Msg("Block height updated")
	}
}

// isInRolloverWindow checks if blockHeight falls within the rollover window around sessionEndHeight.
// Rollover window = [sessionEndHeight - 1, sessionEndHeight + gracePeriod]
func (lfn *LazyFullNode) isInRolloverWindow(blockHeight, sessionEndHeight int64) bool {
	if sessionEndHeight == 0 || blockHeight == 0 {
		return false
	}

	rolloverStart := sessionEndHeight - 1
	rolloverEnd := sessionEndHeight + sessionRolloverGracePeriodBlocks

	return blockHeight >= rolloverStart && blockHeight <= rolloverEnd
}
