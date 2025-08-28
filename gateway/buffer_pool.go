package gateway

import (
	"bytes"
	"io"
	"sync"
)

const (
	// DefaultInitialBufferSize is the initial size of the buffer pool.
	// Start with 256KB buffers - can grow as needed
	DefaultInitialBufferSize = 256 * 1024

	// TODO_CONSIDERATION: Consider making this configurable
	// DefaultMaxBufferSize is the maximum size of the buffer pool.
	// Set the max buffer size to 4MB to avoid memory bloat.
	DefaultMaxBufferSize = 4 * 1024 * 1024
)

// bufferPool manages a pool of reusable byte buffers to reduce GC pressure.
// Buffers are sized appropriately for typical HTTP response bodies.
type bufferPool struct {
	pool          sync.Pool
	maxReaderSize int64
}

func NewBufferPool(maxReaderSize int64) *bufferPool {
	return &bufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, DefaultInitialBufferSize))
			},
		},
		maxReaderSize: maxReaderSize,
	}
}

// getBuffer retrieves a buffer from the pool.
func (bp *bufferPool) getBuffer() *bytes.Buffer {
	buf := bp.pool.Get().(*bytes.Buffer)
	buf.Reset() // Ensure buffer is clean
	return buf
}

// putBuffer returns a buffer to the pool.
// Buffers larger than maxBufferSize are not returned to avoid memory bloat.
func (bp *bufferPool) putBuffer(buf *bytes.Buffer) {
	// Don't pool huge buffers to avoid memory bloat
	if buf.Cap() > DefaultMaxBufferSize {
		return
	}
	bp.pool.Put(buf)
}

// readWithBuffer reads from an io.Reader using a pooled buffer.
func (bp *bufferPool) readWithBuffer(r io.Reader) ([]byte, error) {
	buf := bp.getBuffer()
	defer bp.putBuffer(buf)

	limitedReader := io.LimitReader(r, bp.maxReaderSize)
	_, err := buf.ReadFrom(limitedReader)
	if err != nil {
		return nil, err
	}

	// Make a copy of the bytes to return
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}
