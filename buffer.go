package invoker

import (
	"bytes"
	"sync"
)

// SafeBuffer introduces a mutex lock, making all operations safe
// and non-conflicting: read operation cannot modify buffer until write is finished.
//
// Used in buffer polling contexts, see NewWatchedSafeBuffer of this package.
type SafeBuffer struct {
	buf *bytes.Buffer
	mux *sync.Mutex
}

// NewSafeBuffer constructs a new safe buffer.
func NewSafeBuffer() *SafeBuffer {
	return &SafeBuffer{
		buf: new(bytes.Buffer),
		mux: new(sync.Mutex),
	}
}

func (b *SafeBuffer) Read(p []byte) (n int, err error) {
	b.mux.Lock()
	n, err = b.buf.Read(p)
	b.mux.Unlock()

	return n, err
}

func (b *SafeBuffer) Write(p []byte) (n int, err error) {
	b.mux.Lock()
	n, err = b.buf.Write(p)
	b.mux.Unlock()

	return n, err
}

func (b *SafeBuffer) String() string {
	b.mux.Lock()
	s := b.buf.String()
	b.mux.Unlock()

	return s
}

func (b *SafeBuffer) Bytes() []byte {
	b.mux.Lock()
	s := b.buf.Bytes()
	b.mux.Unlock()

	return s
}
