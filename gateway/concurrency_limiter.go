package gateway

import (
	"context"
	"sync"
	"time"

	shannonmetrics "github.com/buildwithgrove/path/metrics/protocol/shannon"
)

// TODO_IMPROVE: Make this configurable via settings
const (
	defaultMaxConcurrentRequests = 1000000
)

// concurrencyLimiter bounds concurrent operations via semaphore pattern.
// Prevents resource exhaustion and tracks active request counts.
type concurrencyLimiter struct {
	semaphore      chan struct{}
	maxConcurrent  int
	activeRequests int64
	mu             sync.RWMutex
}

// NewConcurrencyLimiter creates a limiter that bounds concurrent operations.
func NewConcurrencyLimiter(maxConcurrent int) *concurrencyLimiter {
	if maxConcurrent <= 0 {
		maxConcurrent = defaultMaxConcurrentRequests // Default reasonable limit
	}

	return &concurrencyLimiter{
		semaphore:     make(chan struct{}, maxConcurrent),
		maxConcurrent: maxConcurrent,
	}
}

// acquire blocks until a slot is available or context is canceled.
// Returns true if acquired, false if context was canceled.
func (cl *concurrencyLimiter) acquire(ctx context.Context) bool {
	select {
	case cl.semaphore <- struct{}{}:
		cl.mu.Lock()
		cl.activeRequests++
		// Track active relays for observability
		shannonmetrics.SetActiveRelays(cl.activeRequests)
		cl.mu.Unlock()
		return true
	case <-ctx.Done():
		return false
	}
}

// tryAcquireWithTimeout attempts to acquire a slot with timeout.
func (cl *concurrencyLimiter) tryAcquireWithTimeout(timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return cl.acquire(ctx)
}

// release returns a slot to the pool.
func (cl *concurrencyLimiter) release() {
	select {
	case <-cl.semaphore:
		cl.mu.Lock()
		cl.activeRequests--
		// Track active relays for observability
		shannonmetrics.SetActiveRelays(cl.activeRequests)
		cl.mu.Unlock()
	default:
		// TODO_TECHDEBT: Log acquire/release mismatch for debugging
	}
}

// getActiveRequests returns the current number of active requests.
func (cl *concurrencyLimiter) getActiveRequests() int64 {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	// Refresh metric with current count
	shannonmetrics.SetActiveRelays(cl.activeRequests)

	return cl.activeRequests
}
