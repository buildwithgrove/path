package user

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

// LimiterManager manages rate limiters for different userAppIDs.
type limiterManager struct {
	limiters map[UserAppID]*limiterEntry
	mu       sync.RWMutex
	ttl      time.Duration
}

// limiterEntry holds the rate limiter and the last time it was accessed.
type limiterEntry struct {
	limiter    *rate.Limiter
	lastAccess atomic.Int64
}

// NewLimiterManager creates a new LimiterManager with the specified TTL for idle entries.
func newLimiterManager() *limiterManager {
	manager := &limiterManager{
		limiters: make(map[UserAppID]*limiterEntry),
		ttl:      30 * time.Second,
	}

	go manager.cleanup() // Start the cleanup process

	return manager
}

// getLimiter returns the rate limiter for a given userAppID, creating one with the specified limit if it doesn't exist.
func (m *limiterManager) getLimiter(userAppID UserAppID, limit int) *rate.Limiter {
	m.mu.RLock()
	entry, exists := m.limiters[userAppID]
	m.mu.RUnlock()

	if exists {
		entry.lastAccess.Store(time.Now().UnixNano())
		return entry.limiter
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// If the limiter does not exist, create a new one with the desired rate.
	limiter := rate.NewLimiter(rate.Limit(limit), limit)
	newEntry := &limiterEntry{limiter: limiter}
	newEntry.lastAccess.Store(time.Now().UnixNano())
	m.limiters[userAppID] = newEntry

	return limiter
}

// cleanup periodically removes limiters that haven't been accessed in a while.
func (m *limiterManager) cleanup() {
	for {
		time.Sleep(m.ttl) // Wait for the TTL period

		var toDelete []UserAppID

		m.mu.RLock()
		for userAppID, entry := range m.limiters {
			if time.Since(time.Unix(0, entry.lastAccess.Load())) > m.ttl {
				toDelete = append(toDelete, userAppID)
			}
		}
		m.mu.RUnlock()

		if len(toDelete) > 0 {
			m.mu.Lock()
			defer m.mu.Unlock()
			for _, userAppID := range toDelete {
				delete(m.limiters, userAppID)
				fmt.Printf("Removed limiter for %s due to inactivity\n", userAppID)
			}
		}
	}
}
