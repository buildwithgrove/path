package shannon

import (
	"bytes"
	"io"
	"sync"
)

// bufferPool manages a pool of reusable byte buffers to reduce GC pressure.
// Buffers are sized appropriately for typical HTTP response bodies.
type bufferPool struct {
	pool sync.Pool
}

// Global buffer pool instance
var globalBufferPool = &bufferPool{
	pool: sync.Pool{
		New: func() interface{} {
			// Start with 64KB buffers - can grow as needed
			return bytes.NewBuffer(make([]byte, 0, 64*1024))
		},
	},
}

// getBuffer retrieves a buffer from the pool.
func (bp *bufferPool) getBuffer() *bytes.Buffer {
	buf := bp.pool.Get().(*bytes.Buffer)
	buf.Reset() // Ensure buffer is clean
	return buf
}

// putBuffer returns a buffer to the pool.
// Buffers larger than 1MB are not returned to avoid memory bloat.
func (bp *bufferPool) putBuffer(buf *bytes.Buffer) {
	// Don't pool huge buffers to avoid memory bloat
	if buf.Cap() > 1024*1024 {
		return
	}
	bp.pool.Put(buf)
}

// readWithBuffer reads from an io.Reader using a pooled buffer.
func readWithBuffer(r io.Reader, maxSize int64) ([]byte, error) {
	buf := globalBufferPool.getBuffer()
	defer globalBufferPool.putBuffer(buf)
	
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