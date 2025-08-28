package gateway

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConcurrencyLimiter(t *testing.T) {
	t.Run("with positive limit", func(t *testing.T) {
		cl := NewConcurrencyLimiter(10)
		assert.NotNil(t, cl)
		assert.Equal(t, 10, cl.maxConcurrent)
		assert.Equal(t, 10, cap(cl.semaphore))
	})

	t.Run("with zero limit uses default", func(t *testing.T) {
		cl := NewConcurrencyLimiter(0)
		assert.NotNil(t, cl)
		assert.Equal(t, maxConcurrentStuff, cl.maxConcurrent)
		assert.Equal(t, maxConcurrentStuff, cap(cl.semaphore))
	})

	t.Run("with negative limit uses default", func(t *testing.T) {
		cl := NewConcurrencyLimiter(-5)
		assert.NotNil(t, cl)
		assert.Equal(t, maxConcurrentStuff, cl.maxConcurrent)
		assert.Equal(t, maxConcurrentStuff, cap(cl.semaphore))
	})
}

func TestConcurrencyLimiterAcquire(t *testing.T) {
	t.Run("acquire within limit", func(t *testing.T) {
		cl := NewConcurrencyLimiter(3)
		ctx := context.Background()

		assert.True(t, cl.acquire(ctx))
		assert.Equal(t, int64(1), cl.getActiveRequests())

		assert.True(t, cl.acquire(ctx))
		assert.Equal(t, int64(2), cl.getActiveRequests())

		assert.True(t, cl.acquire(ctx))
		assert.Equal(t, int64(3), cl.getActiveRequests())

		cl.release()
		cl.release()
		cl.release()
		assert.Equal(t, int64(0), cl.getActiveRequests())
	})

	t.Run("acquire blocks when at limit", func(t *testing.T) {
		cl := NewConcurrencyLimiter(2)
		ctx := context.Background()

		assert.True(t, cl.acquire(ctx))
		assert.True(t, cl.acquire(ctx))

		acquired := make(chan bool)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			acquired <- cl.acquire(ctx)
		}()

		select {
		case result := <-acquired:
			assert.False(t, result, "should have timed out")
		case <-time.After(100 * time.Millisecond):
			t.Fatal("acquire should have returned with timeout")
		}

		cl.release()
		cl.release()
	})

	t.Run("acquire respects context cancellation", func(t *testing.T) {
		cl := NewConcurrencyLimiter(1)
		ctx := context.Background()

		assert.True(t, cl.acquire(ctx))

		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()

		assert.False(t, cl.acquire(cancelCtx))
		assert.Equal(t, int64(1), cl.getActiveRequests())

		cl.release()
	})
}

func TestConcurrencyLimiterTryAcquireWithTimeout(t *testing.T) {
	t.Run("acquires when available", func(t *testing.T) {
		cl := NewConcurrencyLimiter(2)

		assert.True(t, cl.tryAcquireWithTimeout(100*time.Millisecond))
		assert.Equal(t, int64(1), cl.getActiveRequests())

		assert.True(t, cl.tryAcquireWithTimeout(100*time.Millisecond))
		assert.Equal(t, int64(2), cl.getActiveRequests())

		cl.release()
		cl.release()
	})

	t.Run("times out when at limit", func(t *testing.T) {
		cl := NewConcurrencyLimiter(1)

		assert.True(t, cl.tryAcquireWithTimeout(100*time.Millisecond))

		start := time.Now()
		assert.False(t, cl.tryAcquireWithTimeout(50*time.Millisecond))
		elapsed := time.Since(start)

		assert.Greater(t, elapsed, 40*time.Millisecond)
		assert.Less(t, elapsed, 60*time.Millisecond)

		cl.release()
	})
}

func TestConcurrencyLimiterRelease(t *testing.T) {
	t.Run("release decrements active count", func(t *testing.T) {
		cl := NewConcurrencyLimiter(3)
		ctx := context.Background()

		cl.acquire(ctx)
		cl.acquire(ctx)
		cl.acquire(ctx)
		assert.Equal(t, int64(3), cl.getActiveRequests())

		cl.release()
		assert.Equal(t, int64(2), cl.getActiveRequests())

		cl.release()
		assert.Equal(t, int64(1), cl.getActiveRequests())

		cl.release()
		assert.Equal(t, int64(0), cl.getActiveRequests())
	})

	t.Run("release without acquire is safe", func(t *testing.T) {
		cl := NewConcurrencyLimiter(2)
		assert.Equal(t, int64(0), cl.getActiveRequests())

		cl.release()
		assert.Equal(t, int64(0), cl.getActiveRequests())
	})

	t.Run("release allows blocked acquire", func(t *testing.T) {
		cl := NewConcurrencyLimiter(1)
		ctx := context.Background()

		require.True(t, cl.acquire(ctx))

		acquired := make(chan bool)
		go func() {
			ctx := context.Background()
			acquired <- cl.acquire(ctx)
		}()

		time.Sleep(10 * time.Millisecond)

		cl.release()

		select {
		case result := <-acquired:
			assert.True(t, result)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("acquire should have succeeded after release")
		}

		cl.release()
	})
}

func TestConcurrencyLimiterGetActiveRequests(t *testing.T) {
	cl := NewConcurrencyLimiter(10)
	ctx := context.Background()

	assert.Equal(t, int64(0), cl.getActiveRequests())

	cl.acquire(ctx)
	assert.Equal(t, int64(1), cl.getActiveRequests())

	cl.acquire(ctx)
	assert.Equal(t, int64(2), cl.getActiveRequests())

	cl.release()
	assert.Equal(t, int64(1), cl.getActiveRequests())

	cl.release()
	assert.Equal(t, int64(0), cl.getActiveRequests())
}

func TestConcurrencyLimiterConcurrentAccess(t *testing.T) {
	t.Run("concurrent acquire and release", func(t *testing.T) {
		cl := NewConcurrencyLimiter(10)
		ctx := context.Background()

		var wg sync.WaitGroup
		numGoroutines := 50
		numOperations := 100

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < numOperations; j++ {
					if cl.acquire(ctx) {
						time.Sleep(time.Microsecond)
						cl.release()
					}
				}
			}()
		}

		wg.Wait()
		assert.Equal(t, int64(0), cl.getActiveRequests())
	})

	t.Run("enforces max concurrent limit", func(t *testing.T) {
		maxConcurrent := 5
		cl := NewConcurrencyLimiter(maxConcurrent)
		ctx := context.Background()

		var activeCount int32
		var maxActiveCount int32
		var wg sync.WaitGroup

		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if cl.acquire(ctx) {
					current := atomic.AddInt32(&activeCount, 1)
					for {
						max := atomic.LoadInt32(&maxActiveCount)
						if current <= max || atomic.CompareAndSwapInt32(&maxActiveCount, max, current) {
							break
						}
					}
					time.Sleep(10 * time.Millisecond)
					atomic.AddInt32(&activeCount, -1)
					cl.release()
				}
			}()
		}

		wg.Wait()
		assert.LessOrEqual(t, maxActiveCount, int32(maxConcurrent))
		assert.Equal(t, int64(0), cl.getActiveRequests())
	})

	t.Run("concurrent getActiveRequests", func(t *testing.T) {
		cl := NewConcurrencyLimiter(100)
		ctx := context.Background()

		var wg sync.WaitGroup
		stopReading := make(chan struct{})

		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopReading:
					return
				default:
					count := cl.getActiveRequests()
					assert.GreaterOrEqual(t, count, int64(0))
					assert.LessOrEqual(t, count, int64(100))
				}
			}
		}()

		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					if cl.acquire(ctx) {
						time.Sleep(time.Millisecond)
						cl.release()
					}
				}
			}()
		}

		time.Sleep(100 * time.Millisecond)
		close(stopReading)
		wg.Wait()

		assert.Equal(t, int64(0), cl.getActiveRequests())
	})
}

func TestConcurrencyLimiterRaceConditions(t *testing.T) {
	cl := NewConcurrencyLimiter(10)
	ctx := context.Background()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cl.acquire(ctx)
		}()
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cl.release()
		}()
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cl.getActiveRequests()
		}()
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cl.tryAcquireWithTimeout(1 * time.Millisecond)
		}()
	}

	wg.Wait()
}

func BenchmarkConcurrencyLimiterAcquireRelease(b *testing.B) {
	cl := NewConcurrencyLimiter(1000)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if cl.acquire(ctx) {
				cl.release()
			}
		}
	})
}

func BenchmarkConcurrencyLimiterGetActiveRequests(b *testing.B) {
	cl := NewConcurrencyLimiter(1000)
	ctx := context.Background()

	for i := 0; i < 500; i++ {
		cl.acquire(ctx)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cl.getActiveRequests()
		}
	})

	for i := 0; i < 500; i++ {
		cl.release()
	}
}