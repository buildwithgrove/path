package concurrency

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewConcurrencyLimiter(t *testing.T) {
	t.Run("with positive limit", func(t *testing.T) {
		cl := NewConcurrencyLimiter(10)
		require.NotNil(t, cl)
		require.Equal(t, 10, cl.maxConcurrent)
		require.Equal(t, 10, cap(cl.semaphore))
	})

	t.Run("with zero limit uses default", func(t *testing.T) {
		cl := NewConcurrencyLimiter(0)
		require.NotNil(t, cl)
		require.Equal(t, defaultMaxConcurrentRequests, cl.maxConcurrent)
		require.Equal(t, defaultMaxConcurrentRequests, cap(cl.semaphore))
	})

	t.Run("with negative limit uses default", func(t *testing.T) {
		cl := NewConcurrencyLimiter(-5)
		require.NotNil(t, cl)
		require.Equal(t, defaultMaxConcurrentRequests, cl.maxConcurrent)
		require.Equal(t, defaultMaxConcurrentRequests, cap(cl.semaphore))
	})
}

func TestConcurrencyLimiterAcquire(t *testing.T) {
	t.Run("acquire within limit", func(t *testing.T) {
		cl := NewConcurrencyLimiter(3)
		ctx := context.Background()

		require.True(t, cl.Acquire(ctx))
		require.Equal(t, int64(1), cl.getActiveRequests())

		require.True(t, cl.Acquire(ctx))
		require.Equal(t, int64(2), cl.getActiveRequests())

		require.True(t, cl.Acquire(ctx))
		require.Equal(t, int64(3), cl.getActiveRequests())

		cl.Release()
		cl.Release()
		cl.Release()
		require.Equal(t, int64(0), cl.getActiveRequests())
	})

	t.Run("acquire blocks when at limit", func(t *testing.T) {
		cl := NewConcurrencyLimiter(2)
		ctx := context.Background()

		require.True(t, cl.Acquire(ctx))
		require.True(t, cl.Acquire(ctx))

		acquired := make(chan bool)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			acquired <- cl.Acquire(ctx)
		}()

		select {
		case result := <-acquired:
			require.False(t, result, "should have timed out")
		case <-time.After(100 * time.Millisecond):
			t.Fatal("acquire should have returned with timeout")
		}

		cl.Release()
		cl.Release()
	})

	t.Run("acquire respects context cancellation", func(t *testing.T) {
		cl := NewConcurrencyLimiter(1)
		ctx := context.Background()

		require.True(t, cl.Acquire(ctx))

		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()

		require.False(t, cl.Acquire(cancelCtx))
		require.Equal(t, int64(1), cl.getActiveRequests())

		cl.Release()
	})
}

func TestConcurrencyLimiterTryAcquireWithTimeout(t *testing.T) {
	t.Run("acquires when available", func(t *testing.T) {
		cl := NewConcurrencyLimiter(2)

		require.True(t, cl.tryAcquireWithTimeout(100*time.Millisecond))
		require.Equal(t, int64(1), cl.getActiveRequests())

		require.True(t, cl.tryAcquireWithTimeout(100*time.Millisecond))
		require.Equal(t, int64(2), cl.getActiveRequests())

		cl.Release()
		cl.Release()
	})

	t.Run("times out when at limit", func(t *testing.T) {
		cl := NewConcurrencyLimiter(1)

		require.True(t, cl.tryAcquireWithTimeout(100*time.Millisecond))

		start := time.Now()
		require.False(t, cl.tryAcquireWithTimeout(50*time.Millisecond))
		elapsed := time.Since(start)

		require.Greater(t, elapsed, 40*time.Millisecond)
		require.Less(t, elapsed, 60*time.Millisecond)

		cl.Release()
	})
}

func TestConcurrencyLimiterRelease(t *testing.T) {
	t.Run("release decrements active count", func(t *testing.T) {
		cl := NewConcurrencyLimiter(3)
		ctx := context.Background()

		cl.Acquire(ctx)
		cl.Acquire(ctx)
		cl.Acquire(ctx)
		require.Equal(t, int64(3), cl.getActiveRequests())

		cl.Release()
		require.Equal(t, int64(2), cl.getActiveRequests())

		cl.Release()
		require.Equal(t, int64(1), cl.getActiveRequests())

		cl.Release()
		require.Equal(t, int64(0), cl.getActiveRequests())
	})

	t.Run("release without acquire is safe", func(t *testing.T) {
		cl := NewConcurrencyLimiter(2)
		require.Equal(t, int64(0), cl.getActiveRequests())

		cl.Release()
		require.Equal(t, int64(0), cl.getActiveRequests())
	})

	t.Run("release allows blocked acquire", func(t *testing.T) {
		cl := NewConcurrencyLimiter(1)
		ctx := context.Background()

		require.True(t, cl.Acquire(ctx))

		acquired := make(chan bool)
		go func() {
			ctx := context.Background()
			acquired <- cl.Acquire(ctx)
		}()

		time.Sleep(10 * time.Millisecond)

		cl.Release()

		select {
		case result := <-acquired:
			require.True(t, result)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("acquire should have succeeded after release")
		}

		cl.Release()
	})
}

func TestConcurrencyLimiterGetActiveRequests(t *testing.T) {
	cl := NewConcurrencyLimiter(10)
	ctx := context.Background()

	require.Equal(t, int64(0), cl.getActiveRequests())

	cl.Acquire(ctx)
	require.Equal(t, int64(1), cl.getActiveRequests())

	cl.Acquire(ctx)
	require.Equal(t, int64(2), cl.getActiveRequests())

	cl.Release()
	require.Equal(t, int64(1), cl.getActiveRequests())

	cl.Release()
	require.Equal(t, int64(0), cl.getActiveRequests())
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
					if cl.Acquire(ctx) {
						time.Sleep(time.Microsecond)
						cl.Release()
					}
				}
			}()
		}

		wg.Wait()
		require.Equal(t, int64(0), cl.getActiveRequests())
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
				if cl.Acquire(ctx) {
					current := atomic.AddInt32(&activeCount, 1)
					for {
						max := atomic.LoadInt32(&maxActiveCount)
						if current <= max || atomic.CompareAndSwapInt32(&maxActiveCount, max, current) {
							break
						}
					}
					time.Sleep(10 * time.Millisecond)
					atomic.AddInt32(&activeCount, -1)
					cl.Release()
				}
			}()
		}

		wg.Wait()
		require.LessOrEqual(t, maxActiveCount, int32(maxConcurrent))
		require.Equal(t, int64(0), cl.getActiveRequests())
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
					require.GreaterOrEqual(t, count, int64(0))
					require.LessOrEqual(t, count, int64(100))
				}
			}
		}()

		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					if cl.Acquire(ctx) {
						time.Sleep(time.Millisecond)
						cl.Release()
					}
				}
			}()
		}

		time.Sleep(100 * time.Millisecond)
		close(stopReading)
		wg.Wait()

		require.Equal(t, int64(0), cl.getActiveRequests())
	})
}

func TestConcurrencyLimiterRaceConditions(t *testing.T) {
	cl := NewConcurrencyLimiter(10)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Use context with timeout to avoid blocking forever
			if cl.Acquire(ctx) {
				// If we acquired, release it after a short time
				time.Sleep(time.Microsecond)
				cl.Release()
			}
		}()
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cl.Release()
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
			if cl.Acquire(ctx) {
				cl.Release()
			}
		}
	})
}

func BenchmarkConcurrencyLimiterGetActiveRequests(b *testing.B) {
	cl := NewConcurrencyLimiter(1000)
	ctx := context.Background()

	for i := 0; i < 500; i++ {
		cl.Acquire(ctx)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cl.getActiveRequests()
		}
	})

	for i := 0; i < 500; i++ {
		cl.Release()
	}
}
