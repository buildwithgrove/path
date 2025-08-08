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

func Test_updateSessionStartHeight(t *testing.T) {
	tests := []struct {
		name                       string
		session                    sessiontypes.Session
		initialSessionStartHeight  int64
		initialBlockHeight         int64
		expectedSessionStartHeight int64
		expectedInRollover         bool
		shouldReturn               bool // whether function should return early
	}{
		{
			name: "nil session header returns early",
			session: sessiontypes.Session{
				Header: nil,
			},
			initialSessionStartHeight:  100,
			initialBlockHeight:         105,
			expectedSessionStartHeight: 100, // unchanged
			expectedInRollover:         false,
			shouldReturn:               true,
		},
		{
			name: "new session start height - block height in rollover window",
			session: sessiontypes.Session{
				Header: &sessiontypes.SessionHeader{
					SessionStartBlockHeight: 100,
				},
			},
			initialSessionStartHeight:  0,
			initialBlockHeight:         105, // 105 is in rollover window [99, 110] of session start 100
			expectedSessionStartHeight: 100,
			expectedInRollover:         true,
		},
		{
			name: "new session start height - block height outside rollover window",
			session: sessiontypes.Session{
				Header: &sessiontypes.SessionHeader{
					SessionStartBlockHeight: 100,
				},
			},
			initialSessionStartHeight:  0,
			initialBlockHeight:         120, // 120 is outside rollover window [99, 110] of session start 100
			expectedSessionStartHeight: 100,
			expectedInRollover:         false,
		},
		{
			name: "same session start height - no change",
			session: sessiontypes.Session{
				Header: &sessiontypes.SessionHeader{
					SessionStartBlockHeight: 100,
				},
			},
			initialSessionStartHeight:  100,
			initialBlockHeight:         105,
			expectedSessionStartHeight: 100,
			expectedInRollover:         true, // 105 is in rollover window [99, 110] of session start 100
		},
		{
			name: "zero session start height - no rollover",
			session: sessiontypes.Session{
				Header: &sessiontypes.SessionHeader{
					SessionStartBlockHeight: 0,
				},
			},
			initialSessionStartHeight:  100,
			initialBlockHeight:         105,
			expectedSessionStartHeight: 0,
			expectedInRollover:         false, // zero session start height means no rollover
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lfn := newMockLazyFullNode()
			lfn.rolloverState.currentSessionStartHeight = tt.initialSessionStartHeight
			lfn.rolloverState.currentBlockHeight = tt.initialBlockHeight

			lfn.updateSessionStartHeight(tt.session)

			if tt.shouldReturn && lfn.rolloverState.currentSessionStartHeight != tt.expectedSessionStartHeight {
				t.Errorf("updateSessionStartHeight() should have returned early, but state changed")
				return
			}

			if lfn.rolloverState.currentSessionStartHeight != tt.expectedSessionStartHeight {
				t.Errorf("currentSessionStartHeight = %v, want %v", lfn.rolloverState.currentSessionStartHeight, tt.expectedSessionStartHeight)
			}

			if !tt.shouldReturn && lfn.rolloverState.isInSessionRollover != tt.expectedInRollover {
				t.Errorf("isInSessionRollover = %v, want %v", lfn.rolloverState.isInSessionRollover, tt.expectedInRollover)
			}
		})
	}
}

func Test_isInRolloverWindow(t *testing.T) {
	tests := []struct {
		name               string
		blockHeight        int64
		sessionStartHeight int64
		expected           bool
	}{
		{
			name:               "zero session start height returns false",
			blockHeight:        100,
			sessionStartHeight: 0,
			expected:           false,
		},
		{
			name:               "zero block height returns false",
			blockHeight:        0,
			sessionStartHeight: 100,
			expected:           false,
		},
		{
			name:               "both zero returns false",
			blockHeight:        0,
			sessionStartHeight: 0,
			expected:           false,
		},
		{
			name:               "block height before rollover window",
			blockHeight:        98,
			sessionStartHeight: 100,
			expected:           false, // 98 < 99 (rollover start)
		},
		{
			name:               "block height at rollover window start",
			blockHeight:        99,
			sessionStartHeight: 100,
			expected:           true, // 99 == 99 (rollover start)
		},
		{
			name:               "block height at session start",
			blockHeight:        100,
			sessionStartHeight: 100,
			expected:           true, // 100 is within [99, 110]
		},
		{
			name:               "block height in grace period",
			blockHeight:        105,
			sessionStartHeight: 100,
			expected:           true, // 105 is within [99, 110]
		},
		{
			name:               "block height at rollover window end",
			blockHeight:        110,
			sessionStartHeight: 100,
			expected:           true, // 110 == 110 (rollover end)
		},
		{
			name:               "block height after rollover window",
			blockHeight:        111,
			sessionStartHeight: 100,
			expected:           false, // 111 > 110 (rollover end)
		},
		{
			name:               "large session start height",
			blockHeight:        1000005,
			sessionStartHeight: 1000000,
			expected:           true, // 1000005 is within [999999, 1000010]
		},
		{
			name:               "edge case - session start height 1",
			blockHeight:        1,
			sessionStartHeight: 1,
			expected:           true, // 1 is within [0, 11]
		},
		{
			name:               "edge case - block height 0 with session start height 1",
			blockHeight:        0,
			sessionStartHeight: 1,
			expected:           false, // block height 0 returns false regardless
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lfn := newMockLazyFullNode()

			result := lfn.isInRolloverWindow(tt.blockHeight, tt.sessionStartHeight)

			if result != tt.expected {
				t.Errorf("isInRolloverWindow(%v, %v) = %v, want %v",
					tt.blockHeight, tt.sessionStartHeight, result, tt.expected)
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

// Test thread safety of updateSessionStartHeight
func Test_updateSessionStartHeight_ThreadSafety(t *testing.T) {
	lfn := newMockLazyFullNode()

	// Launch multiple goroutines to update state concurrently
	const numGoroutines = 50
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(sessionStartHeight int64) {
			defer wg.Done()
			testSession := sessiontypes.Session{
				Header: &sessiontypes.SessionHeader{
					SessionStartBlockHeight: sessionStartHeight,
				},
			}
			lfn.updateSessionStartHeight(testSession)
		}(int64(100 + i))
	}

	wg.Wait()

	// Verify final state is consistent (no race conditions caused panics)
	finalHeight := lfn.rolloverState.currentSessionStartHeight
	if finalHeight < 100 || finalHeight >= 150 {
		t.Errorf("final currentSessionStartHeight = %v, expected between 100 and 149", finalHeight)
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
