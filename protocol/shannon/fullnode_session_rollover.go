package shannon

import (
	"context"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"
)

const (
	// How often to check for block height updates
	blockCheckInterval = 15 * time.Second
)

// sessionRolloverState tracks session rollover status for the LazyFullNode.
//
// The rollover window spans from 1 block before session end through
// `sessionRolloverBlocks` after session end. This provides
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
	logger polylog.Logger // Logger for rollover operations

	blockClient *sdk.BlockClient // Block client for getting current block height

	sessionRolloverBlocks int64 // Grace period after session end where rollover issues may occur

	currentBlockHeight   int64 // Latest block height from the blockchain
	sessionRolloverStart int64 // Start height of the rollover window
	sessionRolloverEnd   int64 // End height of the rollover window

	isInSessionRollover bool // Cached rollover status (true = in rollover period)

	rolloverStateMu sync.RWMutex // Protects all fields above
}

// newSessionRolloverState creates a new sessionRolloverState with the provided logger, block client, and rollover blocks
func newSessionRolloverState(logger polylog.Logger, blockClient *sdk.BlockClient, sessionRolloverBlocks int64) *sessionRolloverState {
	srs := &sessionRolloverState{
		logger:                logger.With("component", "session_rollover_state"),
		blockClient:           blockClient,
		sessionRolloverBlocks: sessionRolloverBlocks,
	}

	go srs.blockHeightMonitorLoop()

	srs.logger.Info().
		Dur("check_interval", blockCheckInterval).
		Int64("session_rollover_blocks", sessionRolloverBlocks).
		Msg("Starting session rollover monitoring")

	return srs
}

// getSessionRolloverState returns whether we're currently in a session rollover period.
// Thread-safe.
func (srs *sessionRolloverState) getSessionRolloverState() bool {
	srs.rolloverStateMu.RLock()
	defer srs.rolloverStateMu.RUnlock()
	return srs.isInSessionRollover
}

// blockHeightMonitorLoop continuously checks block height to detect session rollovers
func (srs *sessionRolloverState) blockHeightMonitorLoop() {
	srs.logger.Info().
		Bool("block_client_available", srs.blockClient != nil).
		Dur("check_interval", blockCheckInterval).
		Msg("Block height monitor loop starting")

	ticker := time.NewTicker(blockCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		srs.updateBlockHeight()
	}
}

// updateBlockHeight fetches current block height and recalculates rollover status
// Runs on a regular interval to keep the rollover status up to date.
func (srs *sessionRolloverState) updateBlockHeight() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use the block client to get the current block height
	newHeight, err := srs.blockClient.LatestBlockHeight(ctx)
	if err != nil {
		srs.logger.Error().Err(err).Msg("Failed to get current block height")
		return
	}

	srs.rolloverStateMu.Lock()
	defer srs.rolloverStateMu.Unlock()

	// Record the previous block height
	previousHeight := srs.currentBlockHeight

	// Skip if block height hasn't increased
	if previousHeight >= newHeight {
		return
	}

	// Update the current block height
	srs.currentBlockHeight = newHeight

	// Update the cached rollover status based on the new block height
	srs.isInSessionRollover = srs.calculateRolloverStatus()

	logEvent := srs.logger.Debug().
		Int64("current_height", newHeight).
		Int64("rollover_start", srs.sessionRolloverStart).
		Int64("rollover_end", srs.sessionRolloverEnd).
		Bool("in_rollover", srs.isInSessionRollover)

	logEvent.Msg("Block height updated")
}

// updateSessionRolloverBoundaries updates rollover boundaries when we fetch a new session.
// Called from GetSession() to keep rollover monitoring current.
// Only updates rollover boundaries if the current rollover period has ended.
func (srs *sessionRolloverState) updateSessionRolloverBoundaries(session sessiontypes.Session) {
	if session.Header == nil {
		srs.logger.Warn().Msg("Session header is nil, cannot update session values")
		return
	}

	srs.rolloverStateMu.Lock()
	defer srs.rolloverStateMu.Unlock()

	sessionEndHeight := session.Header.SessionEndBlockHeight
	currentBlockHeight := srs.currentBlockHeight

	// Only update rollover boundaries if current rollover period has ended
	// or if we don't have rollover boundaries set yet
	if srs.sessionRolloverEnd == 0 || currentBlockHeight > srs.sessionRolloverEnd {
		newRolloverStart := sessionEndHeight - 1
		newRolloverEnd := sessionEndHeight + srs.sessionRolloverBlocks

		oldRolloverStart := srs.sessionRolloverStart

		// Update rollover boundaries
		srs.sessionRolloverStart = newRolloverStart
		srs.sessionRolloverEnd = newRolloverEnd

		// Log rollover boundary changes
		if oldRolloverStart != newRolloverStart {
			srs.logger.Debug().
				Int64("session_end_height", sessionEndHeight).
				Int64("rollover_start", newRolloverStart).
				Int64("rollover_end", newRolloverEnd).
				Int64("current_block_height", currentBlockHeight).
				Msg("Rollover boundaries updated")
		}
	}

	// Recalculate rollover status based on current state
	srs.isInSessionRollover = srs.calculateRolloverStatus()
}

// calculateRolloverStatus determines if we're in a session rollover period using
// pre-calculated rollover boundaries for efficient and stable detection.
func (srs *sessionRolloverState) calculateRolloverStatus() bool {
	blockHeight := srs.currentBlockHeight
	rolloverStart := srs.sessionRolloverStart
	rolloverEnd := srs.sessionRolloverEnd

	if blockHeight == 0 || rolloverStart == 0 || rolloverEnd == 0 {
		return false
	}

	// Check if current block height is within the rollover window
	return blockHeight >= rolloverStart && blockHeight <= rolloverEnd
}
