package shannon

import (
	"bytes"
	"io"
	"sync"
)

const (
	// defaultInitialBufferSize is the initial size of the buffer pool.
	// Start with 256KB buffers - can grow as needed
	defaultInitialBufferSize = 256 * 1024

	// TODO_CONSIDERATION: Consider making this configurable
	// maxBufferSize is the maximum size of the buffer pool.
	// Set the max buffer size to 4MB to avoid memory bloat.
	maxBufferSize = 4 * 1024 * 1024
)

// bufferPool manages a pool of reusable byte buffers to reduce GC pressure.
// Buffers are sized appropriately for typical HTTP response bodies.
type bufferPool struct {
	pool sync.Pool
}

func NewBufferPool() *bufferPool {
	return &bufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, defaultInitialBufferSize))
			},
		},
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
	if buf.Cap() > maxBufferSize {
		return
	}
	bp.pool.Put(buf)
}

// readWithBuffer reads from an io.Reader using a pooled buffer.
func (bp *bufferPool) readWithBuffer(r io.Reader, maxSize int64) ([]byte, error) {
	buf := bp.getBuffer()
	defer bp.putBuffer(buf)

	limitedReader := io.LimitReader(r, maxSize)
	_, err := buf.ReadFrom(limitedReader)
	if err != nil {
		return nil, err
	}

	// Make a copy of the bytes to return
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}
