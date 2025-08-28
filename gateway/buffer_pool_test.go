package gateway

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBufferPool(t *testing.T) {
	bp := NewBufferPool(DefaultMaxBufferSize)
	assert.NotNil(t, bp)
	assert.NotNil(t, bp.pool)
}

func TestBufferPoolGetBuffer(t *testing.T) {
	bp := NewBufferPool(DefaultMaxBufferSize)

	t.Run("get buffer returns clean buffer", func(t *testing.T) {
		buf := bp.getBuffer()
		assert.NotNil(t, buf)
		assert.Equal(t, 0, buf.Len())
		assert.GreaterOrEqual(t, buf.Cap(), DefaultInitialBufferSize)
	})

	t.Run("multiple gets return different buffers", func(t *testing.T) {
		buf1 := bp.getBuffer()
		buf2 := bp.getBuffer()
		assert.NotSame(t, buf1, buf2)
	})

	t.Run("reused buffer is reset", func(t *testing.T) {
		buf := bp.getBuffer()
		buf.WriteString("test data")
		assert.Greater(t, buf.Len(), 0)

		bp.putBuffer(buf)

		buf2 := bp.getBuffer()
		assert.Equal(t, 0, buf2.Len())
		assert.GreaterOrEqual(t, buf2.Cap(), DefaultInitialBufferSize)
	})
}

func TestBufferPoolPutBuffer(t *testing.T) {
	bp := NewBufferPool(DefaultMaxBufferSize)

	t.Run("puts buffer back to pool", func(t *testing.T) {
		buf := bp.getBuffer()
		originalCap := buf.Cap()
		buf.WriteString("some test data")

		bp.putBuffer(buf)

		buf2 := bp.getBuffer()
		assert.Equal(t, 0, buf2.Len())
		assert.Equal(t, originalCap, buf2.Cap())
	})

	t.Run("does not pool oversized buffers", func(t *testing.T) {
		buf := bytes.NewBuffer(make([]byte, 0, DefaultMaxBufferSize+1))
		bp.putBuffer(buf)

		newBuf := bp.getBuffer()
		assert.NotEqual(t, DefaultMaxBufferSize+1, newBuf.Cap())
		assert.GreaterOrEqual(t, newBuf.Cap(), DefaultInitialBufferSize)
	})

	t.Run("pools buffer at max size", func(t *testing.T) {
		buf := bytes.NewBuffer(make([]byte, 0, DefaultMaxBufferSize))
		bp.putBuffer(buf)

		newBuf := bp.getBuffer()
		assert.LessOrEqual(t, newBuf.Cap(), DefaultMaxBufferSize)
	})
}

func TestBufferPoolReadWithBuffer(t *testing.T) {
	bp := NewBufferPool(DefaultMaxBufferSize)

	t.Run("reads small data", func(t *testing.T) {
		testData := "Hello, World!"
		reader := strings.NewReader(testData)

		data, err := bp.readWithBuffer(reader)
		require.NoError(t, err)
		assert.Equal(t, testData, string(data))
	})

	t.Run("reads large data", func(t *testing.T) {
		testData := strings.Repeat("x", 100000)
		reader := strings.NewReader(testData)

		data, err := bp.readWithBuffer(reader)
		require.NoError(t, err)
		assert.Equal(t, testData, string(data))
	})

	t.Run("respects max size limit", func(t *testing.T) {
		testData := "This is a longer string that will be truncated"
		reader := strings.NewReader(testData)
		maxSize := int64(10)

		data, err := bp.readWithBuffer(reader)
		require.NoError(t, err)
		assert.Equal(t, testData[:maxSize], string(data))
		assert.Equal(t, int(maxSize), len(data))
	})

	t.Run("handles empty reader", func(t *testing.T) {
		reader := strings.NewReader("")

		data, err := bp.readWithBuffer(reader)
		require.NoError(t, err)
		assert.Empty(t, data)
	})

	t.Run("handles reader error", func(t *testing.T) {
		reader := &errorReader{err: io.ErrUnexpectedEOF}

		data, err := bp.readWithBuffer(reader)
		assert.Equal(t, io.ErrUnexpectedEOF, err)
		assert.Nil(t, data)
	})

	t.Run("returns independent copy of data", func(t *testing.T) {
		testData := "Original data"
		reader := strings.NewReader(testData)

		data1, err := bp.readWithBuffer(reader)
		require.NoError(t, err)

		reader2 := strings.NewReader("Modified data")
		data2, err := bp.readWithBuffer(reader2)
		require.NoError(t, err)

		assert.Equal(t, "Original data", string(data1))
		assert.Equal(t, "Modified data", string(data2))
	})
}

func TestBufferPoolConcurrency(t *testing.T) {
	bp := NewBufferPool(DefaultMaxBufferSize)

	t.Run("concurrent get and put", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 100
		numOperations := 100

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numOperations; j++ {
					buf := bp.getBuffer()
					buf.WriteString("test data")
					assert.Greater(t, buf.Len(), 0)
					bp.putBuffer(buf)
				}
			}(i)
		}

		wg.Wait()
	})

	t.Run("concurrent readWithBuffer", func(t *testing.T) {
		var wg sync.WaitGroup
		numGoroutines := 50

		testData := "Concurrent test data"

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				reader := strings.NewReader(testData)
				data, err := bp.readWithBuffer(reader)
				assert.NoError(t, err)
				assert.Equal(t, testData, string(data))
			}()
		}

		wg.Wait()
	})

	t.Run("pool reuse under concurrency", func(t *testing.T) {
		var wg sync.WaitGroup
		bufferChan := make(chan *bytes.Buffer, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				buf := bp.getBuffer()
				bufferChan <- buf
			}()
		}

		go func() {
			wg.Wait()
			close(bufferChan)
		}()

		buffers := make([]*bytes.Buffer, 0, 10)
		for buf := range bufferChan {
			buffers = append(buffers, buf)
		}

		for _, buf := range buffers {
			bp.putBuffer(buf)
		}

		reusedCount := 0
		for i := 0; i < 10; i++ {
			newBuf := bp.getBuffer()
			for _, oldBuf := range buffers {
				if newBuf == oldBuf {
					reusedCount++
					break
				}
			}
		}

		assert.Greater(t, reusedCount, 0, "Some buffers should be reused")
	})
}

func TestBufferPoolMemoryEfficiency(t *testing.T) {
	bp := NewBufferPool(DefaultMaxBufferSize)

	t.Run("buffer grows as needed", func(t *testing.T) {
		buf := bp.getBuffer()
		initialCap := buf.Cap()

		largeData := strings.Repeat("x", DefaultInitialBufferSize*2)
		buf.WriteString(largeData)

		assert.Greater(t, buf.Cap(), initialCap)
		assert.Equal(t, len(largeData), buf.Len())
	})

	t.Run("large buffer not returned to pool", func(t *testing.T) {
		buf := bytes.NewBuffer(make([]byte, 0, DefaultMaxBufferSize+1000))
		buf.WriteString("large buffer data")
		bp.putBuffer(buf)

		newBuf := bp.getBuffer()
		assert.LessOrEqual(t, newBuf.Cap(), DefaultInitialBufferSize)
	})

	t.Run("readWithBuffer doesn't leak memory", func(t *testing.T) {
		testData := strings.Repeat("x", 1000)

		for i := 0; i < 100; i++ {
			reader := strings.NewReader(testData)
			data, err := bp.readWithBuffer(reader)
			require.NoError(t, err)
			assert.Equal(t, len(testData), len(data))
		}
	})
}

func BenchmarkBufferPoolGetPut(b *testing.B) {
	bp := NewBufferPool(DefaultMaxBufferSize)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := bp.getBuffer()
			buf.WriteString("benchmark data")
			bp.putBuffer(buf)
		}
	})
}

func BenchmarkBufferPoolReadWithBuffer(b *testing.B) {
	bp := NewBufferPool(DefaultMaxBufferSize)
	testData := strings.Repeat("x", 10000)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			reader := strings.NewReader(testData)
			data, err := bp.readWithBuffer(reader)
			if err != nil {
				b.Fatal(err)
			}
			if len(data) != len(testData) {
				b.Fatal("data length mismatch")
			}
		}
	})
}

func BenchmarkBufferPoolVsNewBuffer(b *testing.B) {
	bp := NewBufferPool(DefaultMaxBufferSize)

	b.Run("with pool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := bp.getBuffer()
			buf.WriteString("test data")
			bp.putBuffer(buf)
		}
	})

	b.Run("without pool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(make([]byte, 0, DefaultInitialBufferSize))
			buf.WriteString("test data")
		}
	})
}

type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}
