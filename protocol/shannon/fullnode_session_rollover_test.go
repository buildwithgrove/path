package shannon

import (
	"sync"
	"testing"

	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
)

// newMockLazyFullNode creates a minimal LazyFullNode for testing
func newMockLazyFullNode() *LazyFullNode {
	return &LazyFullNode{
		logger: polyzero.NewLogger(),
		rolloverState: &sessionRolloverState{
			rolloverStateMu: sync.RWMutex{},
		},
	}
}

func Test_getSessionRolloverState(t *testing.T) {
	tests := []struct {
		name           string
		initialState   bool
		expectedResult bool
	}{
		{
			name:           "returns true when in rollover",
			initialState:   true,
			expectedResult: true,
		},
		{
			name:           "returns false when not in rollover",
			initialState:   false,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lfn := newMockLazyFullNode()
			lfn.rolloverState.isInSessionRollover = tt.initialState

			result := lfn.getSessionRolloverState()

			if result != tt.expectedResult {
				t.Errorf("getSessionRolloverState() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

func Test_updateSessionValues(t *testing.T) {
	tests := []struct {
		name                   string
		session                sessiontypes.Session
		initialRolloverStart   int64
		initialRolloverEnd     int64
		initialBlockHeight     int64
		expectedRolloverStart  int64
		expectedRolloverEnd    int64
		expectedInRollover     bool
		shouldReturn           bool // whether function should return early
		shouldUpdateBoundaries bool // whether rollover boundaries should be updated
	}{
		{
			name: "nil session header returns early",
			session: sessiontypes.Session{
				Header: nil,
			},
			initialRolloverStart:   100,
			initialRolloverEnd:     170,
			initialBlockHeight:     105,
			expectedRolloverStart:  100, // unchanged
			expectedRolloverEnd:    170, // unchanged
			expectedInRollover:     false,
			shouldReturn:           true,
			shouldUpdateBoundaries: false,
		},
		{
			name: "first session - boundaries not set yet",
			session: sessiontypes.Session{
				Header: &sessiontypes.SessionHeader{
					SessionStartBlockHeight: 100,
					SessionEndBlockHeight:   160,
				},
			},
			initialRolloverStart:   0, // not set yet
			initialRolloverEnd:     0, // not set yet
			initialBlockHeight:     105,
			expectedRolloverStart:  159,   // 160 - 1
			expectedRolloverEnd:    170,   // 160 + 10
			expectedInRollover:     false, // 105 is not in [159, 170]
			shouldReturn:           false,
			shouldUpdateBoundaries: true,
		},
		{
			name: "rollover period ended - update boundaries",
			session: sessiontypes.Session{
				Header: &sessiontypes.SessionHeader{
					SessionStartBlockHeight: 200,
					SessionEndBlockHeight:   260,
				},
			},
			initialRolloverStart:   159, // previous session
			initialRolloverEnd:     170, // previous session
			initialBlockHeight:     175, // past previous rollover end
			expectedRolloverStart:  259, // 260 - 1
			expectedRolloverEnd:    270, // 260 + 10
			expectedInRollover:     false,
			shouldReturn:           false,
			shouldUpdateBoundaries: true,
		},
		{
			name: "rollover period active - don't update boundaries",
			session: sessiontypes.Session{
				Header: &sessiontypes.SessionHeader{
					SessionStartBlockHeight: 200,
					SessionEndBlockHeight:   260,
				},
			},
			initialRolloverStart:   159, // current session
			initialRolloverEnd:     170, // current session
			initialBlockHeight:     165, // within current rollover window
			expectedRolloverStart:  159, // unchanged
			expectedRolloverEnd:    170, // unchanged
			expectedInRollover:     true,
			shouldReturn:           false,
			shouldUpdateBoundaries: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lfn := newMockLazyFullNode()
			lfn.rolloverState.sessionRolloverStart = tt.initialRolloverStart
			lfn.rolloverState.sessionRolloverEnd = tt.initialRolloverEnd
			lfn.rolloverState.currentBlockHeight = tt.initialBlockHeight

			lfn.updateSessionValues(tt.session)

			if tt.shouldReturn && (lfn.rolloverState.sessionRolloverStart != tt.expectedRolloverStart || lfn.rolloverState.sessionRolloverEnd != tt.expectedRolloverEnd) {
				t.Errorf("updateSessionValues() should have returned early, but state changed")
				return
			}

			if lfn.rolloverState.sessionRolloverStart != tt.expectedRolloverStart {
				t.Errorf("sessionRolloverStart = %v, want %v", lfn.rolloverState.sessionRolloverStart, tt.expectedRolloverStart)
			}

			if lfn.rolloverState.sessionRolloverEnd != tt.expectedRolloverEnd {
				t.Errorf("sessionRolloverEnd = %v, want %v", lfn.rolloverState.sessionRolloverEnd, tt.expectedRolloverEnd)
			}

			if !tt.shouldReturn && lfn.rolloverState.isInSessionRollover != tt.expectedInRollover {
				t.Errorf("isInSessionRollover = %v, want %v", lfn.rolloverState.isInSessionRollover, tt.expectedInRollover)
			}
		})
	}
}

func Test_calculateRolloverStatus(t *testing.T) {
	tests := []struct {
		name               string
		currentBlockHeight int64
		rolloverStart      int64
		rolloverEnd        int64
		expected           bool
	}{
		{
			name:               "no block height returns false",
			currentBlockHeight: 0,
			rolloverStart:      159,
			rolloverEnd:        170,
			expected:           false,
		},
		{
			name:               "no rollover start returns false",
			currentBlockHeight: 165,
			rolloverStart:      0,
			rolloverEnd:        170,
			expected:           false,
		},
		{
			name:               "no rollover end returns false",
			currentBlockHeight: 165,
			rolloverStart:      159,
			rolloverEnd:        0,
			expected:           false,
		},
		{
			name:               "block height within rollover window",
			currentBlockHeight: 165,
			rolloverStart:      159, // session end height - 1
			rolloverEnd:        170, // session end height + grace period
			expected:           true,
		},
		{
			name:               "block height at rollover start",
			currentBlockHeight: 159,
			rolloverStart:      159,
			rolloverEnd:        170,
			expected:           true,
		},
		{
			name:               "block height at rollover end",
			currentBlockHeight: 170,
			rolloverStart:      159,
			rolloverEnd:        170,
			expected:           true,
		},
		{
			name:               "block height before rollover window",
			currentBlockHeight: 158,
			rolloverStart:      159,
			rolloverEnd:        170,
			expected:           false,
		},
		{
			name:               "block height after rollover window",
			currentBlockHeight: 171,
			rolloverStart:      159,
			rolloverEnd:        170,
			expected:           false,
		},
		{
			name:               "large block heights",
			currentBlockHeight: 1000005,
			rolloverStart:      999999,
			rolloverEnd:        1000010,
			expected:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lfn := newMockLazyFullNode()
			lfn.rolloverState.currentBlockHeight = tt.currentBlockHeight
			lfn.rolloverState.sessionRolloverStart = tt.rolloverStart
			lfn.rolloverState.sessionRolloverEnd = tt.rolloverEnd

			result := lfn.calculateRolloverStatus()

			if result != tt.expected {
				t.Errorf("calculateRolloverStatus() = %v, want %v", result, tt.expected)
			}
		})
	}
}
