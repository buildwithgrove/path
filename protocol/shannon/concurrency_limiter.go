package shannon

import (
	"context"
	"sync"
	"time"
)

// concurrencyLimiter provides bounded concurrency control for HTTP requests.
// It uses a weighted semaphore pattern to limit the number of concurrent operations.
type concurrencyLimiter struct {
	semaphore chan struct{}
	maxConcurrent int
	activeRequests int64
	mu sync.RWMutex
}

// newConcurrencyLimiter creates a limiter that bounds concurrent operations.
func newConcurrencyLimiter(maxConcurrent int) *concurrencyLimiter {
	if maxConcurrent <= 0 {
		maxConcurrent = 1000 // Default reasonable limit
	}
	
	return &concurrencyLimiter{
		semaphore: make(chan struct{}, maxConcurrent),
		maxConcurrent: maxConcurrent,
	}
}

// acquire blocks until a slot is available or context is cancelled.
// Returns true if acquired, false if context was cancelled.
func (cl *concurrencyLimiter) acquire(ctx context.Context) bool {
	select {
	case cl.semaphore <- struct{}{}:
		cl.mu.Lock()
		cl.activeRequests++
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
		cl.mu.Unlock()
	default:
		// Should never happen if acquire/release are properly paired
	}
}

// getActiveRequests returns the current number of active requests.
func (cl *concurrencyLimiter) getActiveRequests() int64 {
	cl.mu.RLock()
	defer cl.mu.RUnlock()
	return cl.activeRequests
}