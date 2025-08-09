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
// `sessionRolloverGracePeriodBlocks` after session end. This provides
// early warning and extended monitoring of potentially problematic periods
// around session boundaries.
//
// We store pre-calculated rollover boundaries for efficiency:
// - sessionRolloverStart: Start of the rollover window (session end height - 1)
// - sessionRolloverEnd: End of the rollover window (session end height + grace period)
// These boundaries are only updated when the current rollover period has ended,
// preventing frequent recalculation and ensuring stable rollover detection.
//
// All access to this state is protected by rolloverStateMu for thread safety,
// allowing safe concurrent access from multiple goroutines.
type sessionRolloverState struct {
	currentBlockHeight   int64 // Latest block height from the blockchain
	sessionRolloverStart int64 // Start height of the rollover window
	sessionRolloverEnd   int64 // End height of the rollover window
	isInSessionRollover  bool  // Cached rollover status (true = in rollover period)

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

// updateSessionValues updates rollover boundaries when we fetch a new session.
// Called from GetSession() to keep rollover monitoring current.
// Only updates rollover boundaries if the current rollover period has ended.
func (lfn *LazyFullNode) updateSessionValues(session sessiontypes.Session) {
	if session.Header == nil {
		lfn.logger.Warn().Msg("Session header is nil, cannot update session values")
		return
	}

	lfn.rolloverState.rolloverStateMu.Lock()
	defer lfn.rolloverState.rolloverStateMu.Unlock()

	sessionEndHeight := session.Header.SessionEndBlockHeight
	currentBlockHeight := lfn.rolloverState.currentBlockHeight

	// Only update rollover boundaries if current rollover period has ended
	// or if we don't have rollover boundaries set yet
	if lfn.rolloverState.sessionRolloverEnd == 0 || currentBlockHeight > lfn.rolloverState.sessionRolloverEnd {
		newRolloverStart := sessionEndHeight - 1
		newRolloverEnd := sessionEndHeight + sessionRolloverGracePeriodBlocks

		oldRolloverStart := lfn.rolloverState.sessionRolloverStart

		// Update rollover boundaries and recalculate rollover status
		lfn.rolloverState.sessionRolloverStart = newRolloverStart
		lfn.rolloverState.sessionRolloverEnd = newRolloverEnd
		lfn.rolloverState.isInSessionRollover = lfn.calculateRolloverStatus()

		// Log rollover boundary changes
		if oldRolloverStart != newRolloverStart {
			lfn.logger.Debug().
				Int64("session_end_height", sessionEndHeight).
				Int64("rollover_start", newRolloverStart).
				Int64("rollover_end", newRolloverEnd).
				Int64("current_block_height", currentBlockHeight).
				Bool("in_rollover", lfn.rolloverState.isInSessionRollover).
				Msg("Rollover boundaries updated")
		}
	} else {
		// Even if we don't update boundaries, recalculate rollover status
		// in case block height has changed
		lfn.rolloverState.isInSessionRollover = lfn.calculateRolloverStatus()
	}
}

// updateBlockHeight fetches current block height and recalculates rollover status
// Runs on a regular interval to keep the rollover status up to date.
func (lfn *LazyFullNode) updateBlockHeight() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use the lazy full node's block client to get the current block height
	newHeight, err := lfn.GetCurrentBlockHeight(ctx)
	if err != nil {
		lfn.logger.Error().Err(err).Msg("Failed to get current block height")
		return
	}

	lfn.rolloverState.rolloverStateMu.Lock()
	defer lfn.rolloverState.rolloverStateMu.Unlock()

	// Record the previous block height for comparison
	previousHeight := lfn.rolloverState.currentBlockHeight

	// Update the current block height
	lfn.rolloverState.currentBlockHeight = newHeight

	// Update the cached rollover status based on the new block height
	lfn.rolloverState.isInSessionRollover = lfn.calculateRolloverStatus()

	// Log block height changes only if the block height has changed
	if previousHeight != newHeight {
		lfn.logger.Debug().
			Int64("current_height", newHeight).
			Int64("rollover_start", lfn.rolloverState.sessionRolloverStart).
			Int64("rollover_end", lfn.rolloverState.sessionRolloverEnd).
			Bool("in_rollover", lfn.rolloverState.isInSessionRollover).
			Msg("Block height updated")
	}
}

// calculateRolloverStatus determines if we're in a session rollover period using
// pre-calculated rollover boundaries for efficient and stable detection.
func (lfn *LazyFullNode) calculateRolloverStatus() bool {
	blockHeight := lfn.rolloverState.currentBlockHeight
	rolloverStart := lfn.rolloverState.sessionRolloverStart
	rolloverEnd := lfn.rolloverState.sessionRolloverEnd

	if blockHeight == 0 || rolloverStart == 0 || rolloverEnd == 0 {
		return false
	}

	// Check if current block height is within the rollover window
	return blockHeight >= rolloverStart && blockHeight <= rolloverEnd
}
