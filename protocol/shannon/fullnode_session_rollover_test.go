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

func Test_updateSessionEndHeight(t *testing.T) {
	tests := []struct {
		name                    string
		session                 sessiontypes.Session
		initialCurrentSession   int64
		initialPreviousSession  int64
		initialBlockHeight      int64
		expectedCurrentSession  int64
		expectedPreviousSession int64
		expectedInRollover      bool
		shouldReturn            bool // whether function should return early
	}{
		{
			name: "nil session header returns early",
			session: sessiontypes.Session{
				Header: nil,
			},
			initialCurrentSession:   100,
			initialPreviousSession:  90,
			initialBlockHeight:      105,
			expectedCurrentSession:  100, // unchanged
			expectedPreviousSession: 90,  // unchanged
			expectedInRollover:      false,
			shouldReturn:            true,
		},
		{
			name: "first session - no previous session",
			session: sessiontypes.Session{
				Header: &sessiontypes.SessionHeader{
					SessionEndBlockHeight: 100,
				},
			},
			initialCurrentSession:   0,
			initialPreviousSession:  0,
			initialBlockHeight:      95,
			expectedCurrentSession:  100,
			expectedPreviousSession: 0,
			expectedInRollover:      false, // 95 is NOT in rollover window [99, 110] of session 100
		},
		{
			name: "new session - previous session tracked",
			session: sessiontypes.Session{
				Header: &sessiontypes.SessionHeader{
					SessionEndBlockHeight: 200,
				},
			},
			initialCurrentSession:   100,
			initialPreviousSession:  0,
			initialBlockHeight:      105,
			expectedCurrentSession:  200,
			expectedPreviousSession: 100,
			expectedInRollover:      true, // 105 is in rollover window [99, 110] of previous session 100
		},
		{
			name: "same session - no change",
			session: sessiontypes.Session{
				Header: &sessiontypes.SessionHeader{
					SessionEndBlockHeight: 100,
				},
			},
			initialCurrentSession:   100,
			initialPreviousSession:  90,
			initialBlockHeight:      105,
			expectedCurrentSession:  100,
			expectedPreviousSession: 90,    // unchanged
			expectedInRollover:      false, // 105 is NOT in rollover window [89, 100] of previous session 90
		},
		{
			name: "block height outside rollover window",
			session: sessiontypes.Session{
				Header: &sessiontypes.SessionHeader{
					SessionEndBlockHeight: 200,
				},
			},
			initialCurrentSession:   100,
			initialPreviousSession:  0,
			initialBlockHeight:      150, // outside rollover window [99, 110] of session 100
			expectedCurrentSession:  200,
			expectedPreviousSession: 100,
			expectedInRollover:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lfn := newMockLazyFullNode()
			lfn.rolloverState.currentSessionEndHeight = tt.initialCurrentSession
			lfn.rolloverState.previousSessionEndHeight = tt.initialPreviousSession
			lfn.rolloverState.currentBlockHeight = tt.initialBlockHeight

			lfn.updateSessionEndHeight(tt.session)

			if tt.shouldReturn && lfn.rolloverState.currentSessionEndHeight != tt.expectedCurrentSession {
				t.Errorf("updateSessionEndHeight() should have returned early, but state changed")
				return
			}

			if lfn.rolloverState.currentSessionEndHeight != tt.expectedCurrentSession {
				t.Errorf("currentSessionEndHeight = %v, want %v", lfn.rolloverState.currentSessionEndHeight, tt.expectedCurrentSession)
			}

			if lfn.rolloverState.previousSessionEndHeight != tt.expectedPreviousSession {
				t.Errorf("previousSessionEndHeight = %v, want %v", lfn.rolloverState.previousSessionEndHeight, tt.expectedPreviousSession)
			}

			if !tt.shouldReturn && lfn.rolloverState.isInSessionRollover != tt.expectedInRollover {
				t.Errorf("isInSessionRollover = %v, want %v", lfn.rolloverState.isInSessionRollover, tt.expectedInRollover)
			}
		})
	}
}

func Test_isInRolloverWindow(t *testing.T) {
	tests := []struct {
		name             string
		blockHeight      int64
		sessionEndHeight int64
		expected         bool
	}{
		{
			name:             "zero session end height returns false",
			blockHeight:      100,
			sessionEndHeight: 0,
			expected:         false,
		},
		{
			name:             "zero block height returns false",
			blockHeight:      0,
			sessionEndHeight: 100,
			expected:         false,
		},
		{
			name:             "both zero returns false",
			blockHeight:      0,
			sessionEndHeight: 0,
			expected:         false,
		},
		{
			name:             "block height before rollover window",
			blockHeight:      98,
			sessionEndHeight: 100,
			expected:         false, // 98 < 99 (rollover start)
		},
		{
			name:             "block height at rollover window start",
			blockHeight:      99,
			sessionEndHeight: 100,
			expected:         true, // 99 == 99 (rollover start)
		},
		{
			name:             "block height at session end",
			blockHeight:      100,
			sessionEndHeight: 100,
			expected:         true, // 100 is within [99, 110]
		},
		{
			name:             "block height in grace period",
			blockHeight:      105,
			sessionEndHeight: 100,
			expected:         true, // 105 is within [99, 110]
		},
		{
			name:             "block height at rollover window end",
			blockHeight:      110,
			sessionEndHeight: 100,
			expected:         true, // 110 == 110 (rollover end)
		},
		{
			name:             "block height after rollover window",
			blockHeight:      111,
			sessionEndHeight: 100,
			expected:         false, // 111 > 110 (rollover end)
		},
		{
			name:             "large session end height",
			blockHeight:      1000005,
			sessionEndHeight: 1000000,
			expected:         true, // 1000005 is within [999999, 1000010]
		},
		{
			name:             "edge case - session end height 1",
			blockHeight:      1,
			sessionEndHeight: 1,
			expected:         true, // 1 is within [0, 11]
		},
		{
			name:             "edge case - block height 0 with session end height 1",
			blockHeight:      0,
			sessionEndHeight: 1,
			expected:         false, // block height 0 returns false regardless
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lfn := newMockLazyFullNode()

			result := lfn.isInRolloverWindow(tt.blockHeight, tt.sessionEndHeight)

			if result != tt.expected {
				t.Errorf("isInRolloverWindow(%v, %v) = %v, want %v",
					tt.blockHeight, tt.sessionEndHeight, result, tt.expected)
			}
		})
	}
}

// Test thread safety of getSessionRolloverState
func Test_getSessionRolloverState_ThreadSafety(t *testing.T) {
	lfn := newMockLazyFullNode()
	lfn.rolloverState.isInSessionRollover = true

	// Launch multiple goroutines to access the state concurrently
	const numGoroutines = 100
	results := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			result := lfn.getSessionRolloverState()
			results <- result
		}()
	}

	// Collect all results
	for i := 0; i < numGoroutines; i++ {
		result := <-results
		if result != true {
			t.Errorf("getSessionRolloverState() = %v, want true", result)
		}
	}
}

// Test thread safety of updateSessionEndHeight
func Test_updateSessionEndHeight_ThreadSafety(t *testing.T) {
	lfn := newMockLazyFullNode()

	// Launch multiple goroutines to update state concurrently
	const numGoroutines = 50
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(sessionEndHeight int64) {
			defer wg.Done()
			testSession := sessiontypes.Session{
				Header: &sessiontypes.SessionHeader{
					SessionEndBlockHeight: sessionEndHeight,
				},
			}
			lfn.updateSessionEndHeight(testSession)
		}(int64(100 + i))
	}

	wg.Wait()

	// Verify final state is consistent (no race conditions caused panics)
	finalHeight := lfn.rolloverState.currentSessionEndHeight
	if finalHeight < 100 || finalHeight >= 150 {
		t.Errorf("final currentSessionEndHeight = %v, expected between 100 and 149", finalHeight)
	}
}

// Benchmark tests
func BenchmarkGetSessionRolloverState(b *testing.B) {
	lfn := newMockLazyFullNode()
	lfn.rolloverState.isInSessionRollover = true

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lfn.getSessionRolloverState()
	}
}

func BenchmarkIsInRolloverWindow(b *testing.B) {
	lfn := newMockLazyFullNode()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lfn.isInRolloverWindow(100, 95)
	}
}
